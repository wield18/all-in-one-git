package consumerouter

import (
	"all-in-one/internal/base"
	"all-in-one/internal/entity"
	"all-in-one/internal/repository"
	"all-in-one/pkg/consumer"
	httperr "all-in-one/pkg/ekit/http_err"
	"context"
)

func NewConsumePool() *consumer.ConsumerPool {
	p := consumer.NewConsumerPool()
	defer p.StartTicker()
	signal := consumer.NewSignal[string](100)
	register(p, signal)
	return p
}

const (
	TopicTest        = "test"
	RedisKeyValueSet = "RedisKeyValueSet"
	RedisKeyDel      = "RedisKeyDel"
	// 404 不存在或不一致 500
	RedisStringValueCheck        = "RedisStringValueCheck"
	RedisKeyValueSetNX           = "RedisKeyValueSetNX"
	MysqlCreateLoginLocation     = "MysqlCreateLocation"
	MysqlUserViewVideos          = "MysqlUserViewVideos"
	MysqlUserActive              = "MysqlUserActive" // 这只是创建当前天的 active, 重复更新lastActiveTime
	RedisRestoreKey              = "RedisRestoreKey"
	MysqlUpdateVideoCommentCount = "MysqlUpdateVideoCommentCount"
)

// 这里为何不成功也是因为拿旧索引来初始化了
func register(p *consumer.ConsumerPool, signal *consumer.Signal[string]) {
	{
		p.Register(TopicTest,
			consumer.NewSignalConsumer[entity.KeyValue](signal).
				WithConsumerFunc(
					ConsumeKeyValue,
				).WithSize(128, 16, 1))
	}
	{
		p.Register(RedisKeyValueSet,
			consumer.NewSignalConsumer[entity.KeyValueRedis](signal).
				WithConsumerFunc(
					repository.RedisKeyValueSet(base.Redis),
				).WithSize(128, 16, 1))
		p.Register(RedisKeyDel,
			consumer.NewSignalConsumer[string](signal).
				WithConsumerFunc(
					repository.RedisKeyDel(base.Redis),
				).WithSize(128, 16, 1))
		p.Register(RedisStringValueCheck,
			consumer.NewSignalConsumer[entity.KeyValue](signal).
				WithConsumerFunc(
					repository.RedisStringValueCheck(base.Redis),
				).WithSize(128, 16, 1))
		p.Register(RedisKeyValueSetNX,
			consumer.NewSignalConsumer[entity.KeyValueRedis](signal).
				WithConsumerFunc(
					repository.RedisKeyValueSetNX(base.Redis),
				).WithSize(128, 16, 1))
	}
	{
		p.Register(RedisRestoreKey,
			consumer.NewSignalConsumer[*entity.KeyTimeoutRedis](signal).
				WithConsumerFunc(
					repository.RedisRestoreKey(base.Redis),
				).WithSize(128, 16, 1))

	}
	{
		p.Register(MysqlCreateLoginLocation,
			consumer.NewSignalConsumer[entity.LoginLocation](signal).
				WithConsumerFunc(
					validateStructWrap(repository.UserLocaionCreateBatch(base.Mysql)),
				).WithSize(128, 16, 1))

	}
	{
		p.Register(MysqlUserViewVideos,
			consumer.NewSignalConsumer[entity.View](signal).
				WithConsumerFunc(
					validateStructWrap(repository.UserViewVideos(base.Mysql)),
				).WithSize(128, 16, 1))

		p.Register(MysqlUserActive,
			consumer.NewSignalConsumer[*entity.Active](signal).
				WithConsumerFunc(
					validateStructWrap(repository.UserActive(base.Mysql)),
				).WithSize(128, 16, 1))
	}
	{

		p.Register(MysqlUpdateVideoCommentCount,
			consumer.NewSignalConsumer[entity.VideoDetle](signal).
				WithConsumerFunc(repository.VideoCommentCount(base.Mysql)).WithSize(128, 16, 1))
	}
}

// const (
// 	ValidateWrong = "validate wrong"
// )

// 对于struct进行validator的tag验证,如果没通过data会带err
func validateStructWrap[v any](fun consumer.ConsumeFunc[v]) consumer.ConsumeFunc[v] {
	return func(ctx context.Context, datas []*consumer.ConsumeData[v]) {
		l := len(datas)
		count := 0
		// 验证
		for i := 0; i < l; i++ {
			err := base.V.Struct(datas[i].Data)
			if err != nil { // 这是validate没过去,可能是编程人没干好也可能是真就没传好还过来了
				datas[i].WithErr(httperr.ErrBadRequest(err.Error()))
				count++
			}
		}
		if count == l {
			return
		}
		// 继续
		fun(ctx, datas)
	}
}
