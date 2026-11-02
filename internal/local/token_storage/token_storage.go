package tokenstorage

import (
	"all-in-one/internal/base"
	"all-in-one/internal/base/constant"
	consumerouter "all-in-one/internal/consume_router"
	"all-in-one/internal/entity"
	"all-in-one/pkg/consumer"
	httperr "all-in-one/pkg/ekit/http_err"
	"strconv"
	"sync"
	"time"
)

// 这里的token是 code, _ := base.HashCryptoCode(strconv.FormatInt(id, 10))
// token := base.Sprintf(constant.UserToken, code)

// 这里直接存code

// 这里使用p来阻塞分发
type UserCodeStorage struct {
	aMap sync.Map
	p    *consumer.ConsumerPool
}

// 需要redis跟Crypto哈希
// 这个实现非常的具体不是很变通
func New(p *consumer.ConsumerPool) *UserCodeStorage {
	return &UserCodeStorage{p: p}
}

// 这里传入的token长度啥样我不管
// 存入生成的code
// base.HashCryptoCode(strconv.FormatInt(id, 10))
func (t *UserCodeStorage) Store(code string, deadLine time.Time) {
	t.aMap.Store(code, deadLine)
}

func (t *UserCodeStorage) Get(code string) (interface{}, bool) {
	return t.aMap.Load(code)
}

func (t *UserCodeStorage) Del(code string) {
	t.aMap.Delete(code)
}

// base.CheckCryptoCode(code, strconv.FormatInt(id, 10))
// 1.首先code不空
// 1.本地没找到->验证合法->redis查->本地存储
//
//	|->不合法->403
//	          |-> redis没查到->403重新登录
//
// |->本地找到,可进行redis续期
//
// 过了检查得进行本地是否过期
// 这里没过可能是真非法或者过期了
func (t *UserCodeStorage) CheckAndRestore(id int64, code string) (zero httperr.IError) {
	if code == "" {
		return errTokenEmpty
	}
	deadLine, ok := t.aMap.Load(code)
	if !ok { // 本地没查到
		// 再进行验证不是Gaojian发的
		if !base.CheckCryptoCode(code, strconv.FormatInt(id, 10)) {
			// 这里如果不是
			return errTokenUnknow
		}
		// 如果是我发的,我得看看redis存在不
		key := base.Sprintf(constant.UserToken, code)
		data := &entity.KeyTimeoutRedis{
			Key:     key,
			Timeout: constant.UserTokenDuration,
			Rate:    constant.RestoreRate,
		}

		ierr := t.p.BlockSubmit(consumerouter.RedisRestoreKey, consumer.NewConsumeData(data, nil))
		if !ierr.IsZero() {
			return ierr
		}
		t.Store(code, time.Now().Add(time.Duration(data.TTl*int64(time.Second))))
		return
	}
	// 本地查到了
	deadLineTime := deadLine.(time.Time)
	// 续期
	left := time.Since(deadLineTime)
	if left < 0 && left > -constant.UserTokenDurationThreshold {
		key := base.Sprintf(constant.UserToken, code)
		data := &entity.KeyTimeoutRedis{
			Key:     key,
			Timeout: constant.UserTokenDuration,
			Rate:    constant.RestoreRate,
		}

		ierr := t.p.BlockSubmit(consumerouter.RedisRestoreKey, consumer.NewConsumeData(data, nil))
		if !ierr.IsZero() {
			return ierr
		}
		t.Store(code, time.Now().Add(time.Duration(data.TTl*int64(time.Second))))
		return
	}
	// 本地过期了
	if left > 0 {
		t.aMap.Delete(code)
		return errTokenExpired
	}
	return
}

// Refersh 的基础在于token存储过多
// 估计以后会定期删除过期的token
func (t *UserCodeStorage) Refersh() {
	t.aMap.Clear()
}
