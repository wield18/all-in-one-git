package repository

import (
	"all-in-one/internal/entity"
	"all-in-one/pkg/consumer"
	httperr "all-in-one/pkg/ekit/http_err"
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
)

const (
	redisKeyNotExist = "RedisKeyNotExist"
	strNotEqual      = "StrNotEqual"
)

var (
	errExpired = errors.New("expired")
)

// withPipeLined 底层方法
func withPipeLined(ctx context.Context, redisDB *redis.Client, datas []*consumer.ConsumeData[entity.KeyValueRedis],
	operateFunc func(pipe redis.Pipeliner) error) {
	cmds, err := redisDB.Pipelined(ctx, operateFunc)
	// 可能网络原因都没执行，也可能执行一部分
	if err != nil {
		// 这得获得所有的data就行
		for _, data := range datas {
			if data.Err.IsZero() {
				data.WithErr(httperr.ErrInternalServerError(err.Error()))
			}
		}
		return
	}
	// 跳过已经错的
	for i := 0; i < len(datas); i++ {
		if er := cmds[i].Err(); er != nil && datas[i].Err.IsZero() {
			datas[i].WithErr(httperr.ErrInternalServerError(er.Error()))
		}

	}
}

// 这里可以重新刷新key根据如果小于指定的时间
func RedisRestoreKey(redisDB *redis.Client) func(ctx context.Context, datas []*consumer.ConsumeData[*entity.KeyTimeoutRedis]) {
	return func(ctx context.Context, datas []*consumer.ConsumeData[*entity.KeyTimeoutRedis]) {
		cmds, err := redisDB.Pipelined(ctx, func(p redis.Pipeliner) error {
			for _, data := range datas {
				if data.Data.Rate == 0 {
					data.Data.Rate = 1
				}
				p.Eval(ctx, `
				local ttl = redis.call('TTL', KEYS[1])
				if ttl == -2 then return -2 end
				if ttl == -1 then
					redis.call('EXPIRE', KEYS[1], ARGV[1])
					return tonumber(ARGV[1])
				end
				if ttl < tonumber(ARGV[1]) * tonumber(ARGV[2]) then
					redis.call('EXPIRE', KEYS[1], ARGV[1])
					return tonumber(ARGV[1])
				end
				return ttl
			`, []string{data.Data.Key}, int(data.Data.Timeout.Seconds()), data.Data.Rate)
			}
			return nil
		})
		if err != nil {
			// 这得获得所有的data就行
			for _, data := range datas {
				if data.Err.IsZero() {
					data.WithErr(httperr.ErrInternalServerError(err.Error()))
				}
			}
			return
		}
		for i, cmd := range cmds {
			res, err := cmd.(*redis.Cmd).Int64()
			if err != nil && datas[i].Err.IsZero() {
				datas[i].WithErr(httperr.ErrInternalServerError(err.Error()))
				return
			}
			if res == -2 && datas[i].Err.IsZero() {
				datas[i].WithErr(httperr.ErrForbidden(errExpired.Error()))
				return
			}
			datas[i].Data.TTl = res
		}

	}
}

func RedisKeyValueSet(redisDB *redis.Client) func(ctx context.Context, datas []*consumer.ConsumeData[entity.KeyValueRedis]) {
	return func(ctx context.Context, datas []*consumer.ConsumeData[entity.KeyValueRedis]) {
		withPipeLined(ctx, redisDB, datas, func(pipe redis.Pipeliner) error {
			for _, data := range datas {
				dat := data.Data
				pipe.Set(ctx, dat.Key, dat.Value, dat.Timeout)
			}
			return nil
		})
	}
}

// 删除键
func RedisKeyDel(redisDB *redis.Client) func(ctx context.Context, datas []*consumer.ConsumeData[string]) {
	return func(ctx context.Context, datas []*consumer.ConsumeData[string]) {
		dats := consumer.Datas(datas)
		err := redisDB.Del(ctx, dats...).Err()
		if err != nil {
			for _, data := range datas {
				data.WithErr(httperr.ErrInternalServerError(err.Error()))
			}
		}
	}
}

// RedisStringValueCheck
// 字符串存储键值检查
// 这里主要通过明文对照一下
// 因为mget可能返回nil所以[]interface{}存储
func RedisStringValueCheck(redisDB *redis.Client) func(ctx context.Context, datas []*consumer.ConsumeData[entity.KeyValue]) {
	return func(ctx context.Context, datas []*consumer.ConsumeData[entity.KeyValue]) {
		keys := make([]string, 0)
		for _, data := range datas {
			keys = append(keys, data.Data.Key)
		}

		result, err := redisDB.MGet(ctx, keys...).Result()
		// 发送失败
		if err != nil {
			for _, data := range datas {
				data.WithErr(httperr.ErrInternalServerError(err.Error()))
			}
			return
		}
		// 分别比对
		for i, data := range datas {
			str, ok := result[i].(string)
			if !ok {
				data.WithErr(httperr.ErrNotFound(redisKeyNotExist))
				continue
			}
			if data.Data.Value != str {
				data.WithErr(httperr.ErrNotFound(strNotEqual))
			}
		}
	}

}

// NX NX
func RedisKeyValueSetNX(redisDB *redis.Client) func(ctx context.Context, datas []*consumer.ConsumeData[entity.KeyValueRedis]) {
	return func(ctx context.Context, datas []*consumer.ConsumeData[entity.KeyValueRedis]) {
		withPipeLined(ctx, redisDB, datas, func(pipe redis.Pipeliner) error {
			for _, data := range datas {
				dat := data.Data
				pipe.SetNX(ctx, dat.Key, dat.Value, dat.Timeout).Err()
			}
			return nil

		})
	}
}

// // RedisBindKeyValue 绑定value值
// func RedisBindKeyValue(redisDB *redis.Client) func(ctx context.Context, datas []*consumer.ConsumeData[*entity.KeyValueBind]) {
// 	return func(ctx context.Context, datas []*consumer.ConsumeData[*entity.KeyValueBind]) {
// 		// 这里的keys的顺序是从 datas 来的,这里的键如果相同呢?
// 		// 没有问题
// 		keys := make([]string, 0)
// 		for _, data := range datas {
// 			keys = append(keys, data.Data.Key)
// 		}
// 		// 这里只能说跟keys顺序对应
// 		values, err := redisDB.MGet(ctx, keys...).Result()
// 		if err != nil {
// 			for _, data := range datas {
// 				if data.Err.IsZero() {
// 					data.WithErr(httperr.ErrInternalServerError(err.Error()))
// 				}
// 			}
// 			return
// 		}

// 		var i, j int
// 		l1, l2 := len(datas), len(keys)
// 		// 这里得使用双指针对应下
// 		for i < l1 && j < l2 {
// 			if !datas[i].Err.IsZero() {
// 				i++
// 				continue
// 			}
// 			// 这里必然一一对应,因为上边err部分已经把空给过滤了
// 			if datas[i].Data.Key == keys[j] { // 这是第一个key不为空的时候
// 				if str, ok := values[j].(string); ok {
// 					datas[i].Data.Value = str
// 				}
// 				i++
// 				j++
// 				continue
// 			}
// 			// 一般不可能break
// 			break
// 		}
// 	}

// }

// 简单玩玩就行
// 它只能在一个键上加上过期,可能键不存在
// func RedisExpire(redisDB *redis.Client) func() (ctx context.Context, datas []*consumer.ConsumeData[entity.KeyTimeoutRedis]) {
// 	return func() (ctx context.Context, datas []*consumer.ConsumeData[entity.KeyTimeoutRedis]) {
// 		for _, data := range datas {
// 			dat := data.Data
// 			if !entity.IsValidRedisExpireOptionalParameter(dat.T) {
// 				data.WithErr(internal.NewInvalidPatameter(entity.T))
// 				continue
// 			}

// 			// 检查
// 			var err error
// 			if err = internal.CheckColumns(
// 				pairv.NewPair(entity.Key, internal.CheckEmptyStrFunc(dat.Key)),
// 				pairv.NewPair(entity.Timeout, internal.CheckZeroNumberFunc(dat.Timeout)),
// 			); err != nil {
// 				data.WithErr(base.NewIError500FromErr(err))
// 				continue
// 			}

// 			// 选择
// 			switch dat.T {
// 			case entity.NX:
// 				err = redisDB.ExpireNX(ctx, dat.Key, dat.Timeout).Err()
// 			case entity.XX:
// 				err = redisDB.ExpireXX(ctx, dat.Key, dat.Timeout).Err()
// 			case entity.GT:
// 				err = redisDB.ExpireGT(ctx, dat.Key, dat.Timeout).Err()
// 			case entity.LT:
// 				err = redisDB.ExpireLT(ctx, dat.Key, dat.Timeout).Err()
// 			}
// 			if err != nil {
// 				data.WithErr(base.NewIError500FromErr(err))
// 			}
// 		}
// 	}

// }
// func RedisIncr(redisDB *redis.Client) func() (ctx context.Context, datas []*consumer.ConsumeData[entity.KeyNumberRedis]) {
// 	return func() (ctx context.Context, datas []*consumer.ConsumeData[entity.KeyNumberRedis]) {

// 		for _, data := range datas {
// 			dat := data.Data
// 			var err error
// 			if internal.CheckEmptyStr(dat.Key) {
// 				data.WithErr(internal.NewInvalidPatameter(entity.Key))
// 				continue
// 			}
// 			switch dat.Number {
// 			case 0:
// 				err = redisDB.Incr(ctx, dat.Key).Err()
// 			default:
// 				err = redisDB.IncrBy(ctx, dat.Key, dat.Number).Err()
// 			}
// 			if err != nil {
// 				data.WithErr(base.NewIError500FromErr(err))
// 			}
// 		}
// 	}

// }
// func RedisPush(redisDB *redis.Client) func() (ctx context.Context, datas []*consumer.ConsumeData[entity.PushRedis]) {
// 	return func() (ctx context.Context, datas []*consumer.ConsumeData[entity.PushRedis]) {
// 		for _, data := range datas {
// 			dat := data.Data
// 			var err error
// 			if err = internal.CheckColumns(
// 				pairv.NewPair(entity.Key, internal.CheckEmptyStrFunc(dat.Key)),
// 				pairv.NewPair(entity.Val, internal.CheckEmptyStrsFunc(dat.Vals)),
// 			); err != nil {
// 				data.WithErr(internal.NewInvalidPatameter(entity.Key))
// 				continue
// 			}
// 			if !dat.IsRight {
// 				if !dat.UseXX {
// 					err = redisDB.LPush(ctx, dat.Key, dat.Vals).Err()
// 				} else {
// 					err = redisDB.LPushX(ctx, dat.Key, dat.Vals).Err()
// 				}
// 			} else {
// 				if !dat.UseXX {
// 					err = redisDB.RPush(ctx, dat.Key, dat.Vals).Err()
// 				} else {
// 					err = redisDB.RPushX(ctx, dat.Key, dat.Vals).Err()
// 				}
// 			}
// 			if err != nil {
// 				data.WithErr(base.NewIError500FromErr(err))
// 			}
// 		}
// 	}

// }

// // RedisPOP pop 一定要绑定值吗?绑定
// func RedisPOP(redisDB *redis.Client) func() (ctx context.Context, datas []*consumer.ConsumeData[*entity.PopRedis]) {
// 	return func() (ctx context.Context, datas []*consumer.ConsumeData[*entity.PopRedis]) {
// 		for _, data := range datas {
// 			var err error
// 			dat := data.Data
// 			if err = internal.CheckColumns(
// 				pairv.NewPair(entity.Key, internal.CheckEmptyStrFunc(dat.Key)),
// 				pairv.NewPair(entity.Val, internal.CheckNumberLessThenZeroFunc(dat.Count)),
// 			); err != nil {
// 				data.WithErr(base.NewIError500FromErr(err))
// 				continue
// 			}
// 			var results interface{}
// 			if dat.Count == 0 {
// 				if !dat.IsRight {
// 					results, err = redisDB.LPop(ctx, dat.Key).Result()
// 				} else {
// 					results, err = redisDB.RPop(ctx, dat.Key).Result()
// 				}
// 			} else {
// 				if !dat.IsRight {
// 					results, err = redisDB.LPopCount(ctx, dat.Key, dat.Count).Result()
// 				} else {
// 					results, err = redisDB.RPopCount(ctx, dat.Key, dat.Count).Result()
// 				}
// 			}
// 			if err != nil {
// 				data.WithErr(base.NewIError500FromErr(err))
// 				continue
// 			}
// 			if str, ok := results.(string); ok {
// 				dat.Vals = []string{str}
// 				continue
// 			}
// 			dat.Vals = results.([]string)
// 		}
// 	}

// }
// func RedisRange(redisDB *redis.Client) func() (ctx context.Context, datas []*consumer.ConsumeData[*entity.LRangeRedis]) {
// 	return func() (ctx context.Context, datas []*consumer.ConsumeData[*entity.LRangeRedis]) {
// 		for _, data := range datas {
// 			var err error
// 			dat := data.Data
// 			if err = internal.CheckColumns(
// 				pairv.NewPair(entity.Key, internal.CheckEmptyStrFunc(dat.Key)),
// 				pairv.NewPair(entity.Start, internal.CheckNumberLessThenZeroFunc(dat.Start)),
// 				pairv.NewPair(entity.End, internal.CheckNumberLessThenZeroFunc(dat.End)),
// 			); err != nil {
// 				data.WithErr(base.NewIError500FromErr(err))
// 				continue
// 			}
// 			if dat.End == 0 {
// 				dat.Vals, err = redisDB.LRange(ctx, dat.Key, dat.Start, -1).Result()
// 			} else {
// 				dat.Vals, err = redisDB.LRange(ctx, dat.Key, dat.Start, dat.End).Result()
// 			}

// 			if err != nil {
// 				data.WithErr(base.NewIError500FromErr(err))
// 				continue
// 			}

// 		}
// 	}

// }
// func RedisLen(redisDB *redis.Client) func() (ctx context.Context, datas []*consumer.ConsumeData[*entity.KeyNumberRedis]) {
// 	return func() (ctx context.Context, datas []*consumer.ConsumeData[*entity.KeyNumberRedis]) {
// 		for _, data := range datas {
// 			var err error
// 			dat := data.Data
// 			if internal.CheckEmptyStr(dat.Key) {
// 				data.WithErr(internal.NewInvalidPatameter(entity.Key))
// 				continue
// 			}
// 			dat.Number, err = redisDB.LLen(ctx, dat.Key).Result()
// 			if err != nil {
// 				data.WithErr(base.NewIError500FromErr(err))
// 				continue
// 			}

// 		}
// 	}

// }

// // RedisSADD
// // redis 返回的成功的count我这里不记录,如果成功了,必然都在里边
// func RedisSADD(redisDB *redis.Client) func() (ctx context.Context, datas []*consumer.ConsumeData[entity.KeyValues]) {
// 	return func() (ctx context.Context, datas []*consumer.ConsumeData[entity.KeyValues]) {
// 		for _, data := range datas {
// 			var err error
// 			dat := data.Data
// 			if err = internal.CheckColumns(
// 				pairv.NewPair(entity.Key, internal.CheckEmptyStrFunc(dat.Key)),
// 				pairv.NewPair(entity.Vals, internal.CheckEmptyStrsFunc(dat.Vals)),
// 			); err != nil {
// 				data.WithErr(base.NewIError500FromErr(err))
// 				continue
// 			}
// 			err = redisDB.SAdd(ctx, dat.Key, dat.Vals).Err()
// 			if err != nil {
// 				data.WithErr(base.NewIError500FromErr(err))
// 			}
// 		}
// 	}

// }

// // RedisSREM
// // redis 返回的成功的count我这里不记录,如果成功了,必然都删掉了
// func RedisSREM(redisDB *redis.Client) func() (ctx context.Context, datas []*consumer.ConsumeData[entity.KeyValues]) {
// 	return func() (ctx context.Context, datas []*consumer.ConsumeData[entity.KeyValues]) {
// 		for _, data := range datas {
// 			var err error
// 			dat := data.Data
// 			if err = internal.CheckColumns(
// 				pairv.NewPair(entity.Key, internal.CheckEmptyStrFunc(dat.Key)),
// 				pairv.NewPair(entity.Vals, internal.CheckEmptyStrsFunc(dat.Vals)),
// 			); err != nil {
// 				data.WithErr(base.NewIError500FromErr(err))
// 				continue
// 			}
// 			err = redisDB.SRem(ctx, dat.Key, dat.Vals).Err()
// 			if err != nil {
// 				data.WithErr(base.NewIError500FromErr(err))
// 			}
// 		}
// 	}

// }

// // RedisSPOP 随机删
// func RedisSPOP(redisDB *redis.Client) func() (ctx context.Context, datas []*consumer.ConsumeData[*entity.SPopRedis]) {
// 	return func() (ctx context.Context, datas []*consumer.ConsumeData[*entity.SPopRedis]) {
// 		for _, data := range datas {
// 			var err error
// 			dat := data.Data
// 			if err = internal.CheckColumns(
// 				pairv.NewPair(entity.Key, internal.CheckEmptyStrFunc(dat.Key)),
// 				pairv.NewPair(entity.Count, internal.CheckNumberLessThenZeroFunc(dat.Count)),
// 			); err != nil {
// 				data.WithErr(base.NewIError500FromErr(err))
// 			}
// 			var results interface{}
// 			switch dat.Count {
// 			case 0:
// 				results, err = redisDB.SPop(ctx, dat.Key).Result()
// 			default:
// 				results, err = redisDB.SPopN(ctx, dat.Key, dat.Count).Result()
// 			}

// 			if err != nil {
// 				data.WithErr(base.NewIError500FromErr(err))
// 				continue

// 			}
// 			if str, ok := results.(string); ok {
// 				dat.Vals = []string{str}
// 				continue
// 			}
// 			dat.Vals = results.([]string)
// 		}
// 	}

// }

// // RedisSRANDMEMBER 随机获得
// func RedisSRANDMEMBER(redisDB *redis.Client) func() (ctx context.Context, datas []*consumer.ConsumeData[*entity.SPopRedis]) {
// 	return func() (ctx context.Context, datas []*consumer.ConsumeData[*entity.SPopRedis]) {
// 		for _, data := range datas {
// 			var err error
// 			dat := data.Data
// 			if err = internal.CheckColumns(
// 				pairv.NewPair(entity.Key, internal.CheckEmptyStrFunc(dat.Key)),
// 				pairv.NewPair(entity.Count, internal.CheckNumberLessThenZeroFunc(dat.Count)),
// 			); err != nil {
// 				data.WithErr(base.NewIError500FromErr(err))
// 			}
// 			var results interface{}
// 			switch dat.Count {
// 			case 0:
// 				results, err = redisDB.SRandMember(ctx, dat.Key).Result()
// 			default:
// 				results, err = redisDB.SRandMemberN(ctx, dat.Key, dat.Count).Result()
// 			}

// 			if err != nil {
// 				data.WithErr(base.NewIError500FromErr(err))
// 				continue

// 			}
// 			if str, ok := results.(string); ok {
// 				dat.Vals = []string{str}
// 				continue
// 			}
// 			dat.Vals = results.([]string)
// 		}
// 	}

// }

// // RedisSMEMBERS 获得所有的 member
// func RedisSMEMBERS(redisDB *redis.Client) func() (ctx context.Context, datas []*consumer.ConsumeData[*entity.KeyValues]) {
// 	return func() (ctx context.Context, datas []*consumer.ConsumeData[*entity.KeyValues]) {
// 		for _, data := range datas {
// 			var err error
// 			dat := data.Data
// 			if internal.CheckEmptyStr(dat.Key) {
// 				data.WithErr(internal.NewInvalidPatameter(entity.Key))
// 			}
// 			dat.Vals, err = redisDB.SMembers(ctx, dat.Key).Result()
// 			if err != nil {
// 				data.WithErr(base.NewIError500FromErr(err))
// 				continue

// 			}

// 		}
// 	}

// }

// // RedisSCard 获得总数
// func RedisSCard(redisDB *redis.Client) func() (ctx context.Context, datas []*consumer.ConsumeData[*entity.KeyNumberRedis]) {
// 	return func() (ctx context.Context, datas []*consumer.ConsumeData[*entity.KeyNumberRedis]) {
// 		for _, data := range datas {
// 			var err error
// 			dat := data.Data
// 			if internal.CheckEmptyStr(dat.Key) {
// 				data.WithErr(internal.NewInvalidPatameter(entity.Key))
// 				continue
// 			}
// 			dat.Number, err = redisDB.SCard(ctx, dat.Key).Result()
// 			if err != nil {
// 				data.WithErr(base.NewIError500FromErr(err))
// 				continue
// 			}
// 		}
// 	}

// }

// // RedisIsMember 没有err就是存在
// func RedisIsMember(redisDB *redis.Client) func() (ctx context.Context, datas []*consumer.ConsumeData[*entity.KeyValue]) {
// 	return func() (ctx context.Context, datas []*consumer.ConsumeData[*entity.KeyValue]) {
// 		for _, data := range datas {
// 			var err error
// 			dat := data.Data
// 			if err = internal.CheckColumns(
// 				pairv.NewPair(entity.Key, internal.CheckEmptyStrFunc(dat.Key)),
// 				pairv.NewPair(entity.Val, internal.CheckEmptyStrFunc(dat.Value)),
// 			); err != nil {
// 				data.WithErr(base.NewIError500FromErr(err))
// 				continue
// 			}
// 			data.WithErr(redisDB.SIsMember(ctx, dat.Key, dat.Value).Err())
// 		}
// 	}

// }
// func RedisSetOperate(redisDB *redis.Client) func() (ctx context.Context, datas []*consumer.ConsumeData[*entity.SetOperationRedis]) {
// 	return func() (ctx context.Context, datas []*consumer.ConsumeData[*entity.SetOperationRedis]) {
// 		for _, data := range datas {
// 			dat := data.Data
// 			if !entity.IsValidRedisSetOperateParameter(dat.T) {
// 				data.WithErr(internal.NewInvalidPatameter(entity.T))
// 				continue
// 			}
// 			if internal.CheckEmptyStrsAndLen(dat.Keys, 2) {
// 				data.WithErr(internal.NewInvalidPatameter(entity.Keys))
// 				continue
// 			}
// 			var err error
// 			switch dat.T {
// 			case entity.SINTER:
// 				dat.Result, err = redisDB.SInter(ctx, dat.Keys...).Result()
// 			case entity.SUNION:
// 				dat.Result, err = redisDB.SUnion(ctx, dat.Keys...).Result()
// 			case entity.SDIFF:
// 				dat.Result, err = redisDB.SDiff(ctx, dat.Keys...).Result()
// 			}
// 			if err != nil {
// 				data.WithErr(base.NewIError500FromErr(err))
// 				continue
// 			}
// 			if dat.NewSetKey != "" {
// 				err = redisDB.SAdd(ctx, dat.NewSetKey, dat.Result).Err()
// 			}
// 			if err != nil {
// 				data.WithErr(base.NewIError500FromErr(err))
// 				continue
// 			}
// 		}
// 	}

// }

// // GT、LT 和 NX 这三个选项在 Redis 命令中本身就是互斥的
// // zadd key NX GT CH INCR score member
// // zadd key nx 可incr, ch没意义 members
// // zadd key 非nx gt/lt ch members

// // incr这个方法不暴露吧
// // zadd key xx gt/lt ch incr member 只能操作一个

// func RedisZADD(redisDB *redis.Client) func() (ctx context.Context, datas []*consumer.ConsumeData[*entity.ZAddRedis]) {
// 	return func() (ctx context.Context, datas []*consumer.ConsumeData[*entity.ZAddRedis]) {
// 		for _, data := range datas {
// 			var err error
// 			dat := data.Data
// 			if err = internal.CheckColumns(
// 				pairv.NewPair(
// 					fmt.Sprint([]string{entity.UseNX, entity.UseGT, entity.UseLT}),
// 					internal.CheckMutualBitwiseFunc(dat.Options, entity.NX, entity.GT, entity.LT)),
// 				pairv.NewPair(
// 					fmt.Sprint([]string{entity.UseXX, entity.UseNX}),
// 					internal.CheckMutualBitwiseFunc(dat.Options, entity.XX, entity.NX),
// 				),
// 				pairv.NewPair(
// 					fmt.Sprint([]string{entity.UseGT, entity.UseLT}),
// 					internal.CheckMutualBitwiseFunc(dat.Options, entity.GT, entity.LT),
// 				),
// 				pairv.NewPair(entity.Key, internal.CheckEmptyStrFunc(dat.Key)),
// 				pairv.NewPair(entity.Memners, internal.CheckRedisZFunc(dat.Memners)),
// 			); err != nil {
// 				data.WithErr(base.NewIError500FromErr(err))
// 				continue
// 			}

// 			args := redis.ZAddArgs{Members: dat.Memners}
// 			if dat.Options&entity.NX == dat.Options { // 说明data里有nx,有nx其他都不管了
// 				args.NX = true
// 			} else {
// 				if dat.Options&entity.XX == dat.Options {
// 					args.XX = true
// 				}
// 				if dat.Options&entity.CH == dat.Options {
// 					args.Ch = true
// 				}
// 				if dat.Options&entity.GT == dat.Options {
// 					args.GT = true
// 				}
// 				if dat.Options&entity.LT == dat.Options {
// 					args.LT = true
// 				}
// 			}

// 			err = redisDB.ZAddArgs(ctx, dat.Key, args).Err()
// 			if err != nil {
// 				data.WithErr(base.NewIError500FromErr(err))
// 			}
// 		}
// 	}

// }
// func RedisZREM(redisDB *redis.Client) func() (ctx context.Context, datas []*consumer.ConsumeData[entity.ZRemRedis]) {
// 	return func() (ctx context.Context, datas []*consumer.ConsumeData[entity.ZRemRedis]) {
// 		for _, data := range datas {
// 			var err error
// 			dat := data.Data
// 			if err = internal.CheckColumns(
// 				pairv.NewPair(entity.Key, internal.CheckEmptyStrFunc(dat.Key)),
// 				pairv.NewPair(entity.Memners, internal.CheckEmptyStrsAndLenFunc(dat.Memners, 1)),
// 			); err != nil {
// 				data.WithErr(base.NewIError500FromErr(err))
// 				continue
// 			}
// 			data.WithErr(redisDB.ZRem(ctx, dat.Key, dat.Memners).Err())

// 		}
// 	}

// }
// func RedisZScore(redisDB *redis.Client) func() (ctx context.Context, datas []*consumer.ConsumeData[*entity.ZScoreRedis]) {
// 	return func() (ctx context.Context, datas []*consumer.ConsumeData[*entity.ZScoreRedis]) {
// 		for _, data := range datas {
// 			var err error
// 			dat := data.Data
// 			if err = internal.CheckColumns(
// 				pairv.NewPair(entity.Key, internal.CheckEmptyStrFunc(dat.Key)),
// 				pairv.NewPair(entity.Member, internal.CheckEmptyStrFunc(dat.Memner)),
// 			); err != nil {
// 				data.WithErr(base.NewIError500FromErr(err))
// 				continue
// 			}
// 			var score float64
// 			score, err = redisDB.ZScore(ctx, dat.Key, dat.Memner).Result()
// 			if err != nil {
// 				data.WithErr(base.NewIError500FromErr(err))
// 				continue
// 			}
// 			dat.Score = score
// 		}
// 	}

// }
// func RedisZIncr(redisDB *redis.Client) func() (ctx context.Context, datas []*consumer.ConsumeData[*entity.ZScoreRedis]) {
// 	return func() (ctx context.Context, datas []*consumer.ConsumeData[*entity.ZScoreRedis]) {
// 		for _, data := range datas {
// 			var err error
// 			dat := data.Data
// 			if err = internal.CheckColumns(
// 				pairv.NewPair(entity.Key, internal.CheckEmptyStrFunc(dat.Key)),
// 				pairv.NewPair(entity.Member, internal.CheckEmptyStrFunc(dat.Memner)),
// 				pairv.NewPair(entity.Score, internal.CheckZeroNumberFunc(dat.Score)),
// 			); err != nil {
// 				data.WithErr(base.NewIError500FromErr(err))
// 				continue
// 			}
// 			var score float64
// 			score, err = redisDB.ZIncrBy(ctx, dat.Key, dat.Score, dat.Memner).Result()
// 			if err != nil {
// 				data.WithErr(base.NewIError500FromErr(err))
// 				continue
// 			}
// 			dat.Score = score
// 		}
// 	}

// }
// func RedisZCARD(redisDB *redis.Client) func() (ctx context.Context, datas []*consumer.ConsumeData[*entity.KeyNumberRedis]) {
// 	return func() (ctx context.Context, datas []*consumer.ConsumeData[*entity.KeyNumberRedis]) {
// 		for _, data := range datas {
// 			var err error
// 			dat := data.Data
// 			if internal.CheckEmptyStr(dat.Key) {
// 				data.WithErr(internal.NewInvalidPatameter(entity.Key))
// 				continue
// 			}
// 			var l int64
// 			l, err = redisDB.ZCard(ctx, dat.Key).Result()
// 			if err != nil {
// 				data.WithErr(base.NewIError500FromErr(err))
// 				continue
// 			}
// 			dat.Number = l
// 		}
// 	}

// }
// func RedisZCOUNT(redisDB *redis.Client) func() (ctx context.Context, datas []*consumer.ConsumeData[*entity.ZCountMinMaxRedis]) {
// 	return func() (ctx context.Context, datas []*consumer.ConsumeData[*entity.ZCountMinMaxRedis]) {
// 		for _, data := range datas {
// 			var err error
// 			dat := data.Data
// 			if err = internal.CheckColumns(
// 				pairv.NewPair(entity.Key, internal.CheckEmptyStrFunc(dat.Key)),
// 				pairv.NewPair(entity.Key, internal.CheckOrderFunc(dat.Min, dat.Max)),
// 			); err != nil {
// 				data.WithErr(base.NewIError500FromErr(err))
// 				continue
// 			}
// 			var count int64
// 			count, err = redisDB.ZCount(
// 				ctx, dat.Key, strconv.FormatInt(dat.Min, 10), strconv.FormatInt(dat.Max, 10)).Result()
// 			if err != nil {
// 				data.WithErr(base.NewIError500FromErr(err))
// 				continue
// 			}
// 			dat.Count = count
// 			redisDB.ZRangeArgs(ctx, redis.ZRangeArgs{})

// 		}
// 	}

// }

// // ZRANGE
// // 默认WithScore
// // 默认byScore
// // result使用interface所以可能是 []redis.Z或[]string
// func RedisZRANGE(redisDB *redis.Client) func() (ctx context.Context, datas []*consumer.ConsumeData[*entity.ZRangeBYSCORERedis]) {
// 	return func() (ctx context.Context, datas []*consumer.ConsumeData[*entity.ZRangeBYSCORERedis]) {
// 		for _, data := range datas {
// 			var err error
// 			dat := data.Data
// 			if err = internal.CheckColumns(
// 				pairv.NewPair(entity.Key, internal.CheckEmptyStrFunc(dat.Key)),
// 				pairv.NewPair(entity.Key, internal.CheckOrderFunc(dat.Start, dat.End)),
// 				pairv.NewPair(entity.Offset, internal.CheckNumberLessThenZeroFunc(dat.Page-1)),
// 			); err != nil {
// 				data.WithErr(base.NewIError500FromErr(err))
// 				continue
// 			}
// 			args := redis.ZRangeArgs{ByScore: true}
// 			args.Rev = dat.Rev

// 			if dat.Pagesize != 0 {
// 				args.Count = dat.Pagesize
// 				args.Offset = (dat.Page - 1) * dat.Pagesize
// 			}
// 			dat.Results, data.Err = redisDB.ZRangeArgsWithScores(ctx, args).Result()
// 		}
// 	}

// }

// // ZRANK
// // 默认使用 redis.RankScore
// func RedisZRANK(redisDB *redis.Client) func() (ctx context.Context, datas []*consumer.ConsumeData[*entity.ZRankRedis]) {
// 	return func() (ctx context.Context, datas []*consumer.ConsumeData[*entity.ZRankRedis]) {
// 		for _, data := range datas {
// 			var err error
// 			dat := data.Data
// 			if err = internal.CheckColumns(
// 				pairv.NewPair(entity.Key, internal.CheckEmptyStrFunc(dat.Key)),
// 				pairv.NewPair(entity.Member, internal.CheckEmptyStrFunc(dat.Member)),
// 			); err != nil {
// 				data.WithErr(base.NewIError500FromErr(err))
// 				continue
// 			}
// 			if !dat.Rev {

// 				dat.Results, data.Err = redisDB.ZRankWithScore(ctx, dat.Key, dat.Member).Result()
// 			} else {
// 				dat.Results, data.Err = redisDB.ZRevRankWithScore(ctx, dat.Key, dat.Member).Result()
// 			}

// 		}
// 	}

// }
