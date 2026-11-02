package main

import (
	"all-in-one/config"
	"all-in-one/pkg/consumer"
	"all-in-one/pkg/emailqq"

	"github.com/mojocn/base64Captcha"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// 需要啥service
type servers struct {
	mdb      bool
	rdb      bool
	qq       bool
	captcha  bool
	slogv    bool
	consumer bool
}

// 测试使用
func StartAllService(conf *config.Config, needs servers) (
	mdb *gorm.DB,
	redis *redis.Client,
	qq *emailqq.QQDialer,
	captch *base64Captcha.Captcha,
	saveState func(),
	consumerV *consumer.ConsumerPool) {
	// var err error
	// if needs.mdb {
	// 	mdb = mysql.NewDB(conf.Mysql)
	// }
	// if needs.rdb {
	// 	redis = redisv.NewRedis(conf.Redis)
	// }
	// // 这个没有测试过连接

	// if needs.qq {
	// 	qq = emailqq.NewDialer(conf.QQ)

	// }
	// if needs.captcha {
	// 	captch, err = captcha.NewCaptcha(conf.CaptCha)
	// 	if err != nil {
	// 		panic(fmt.Errorf("%w", err))
	// 	}
	// }
	// if needs.slogv {
	// 	saveState, err = slogv.InitSlog(conf.Log)
	// 	if err != nil {
	// 		panic(fmt.Errorf("%w", err))
	// 	}
	// }
	// if needs.consumer {
	// 	consume.NewConsumeRouter()
	// }
	return
}
