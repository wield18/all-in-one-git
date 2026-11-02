package redisx

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
)

// Redis配置
type Redis struct {
	Address  string `yaml:"Address"`
	Password string `yaml:"Password"`
	DB       int    `yaml:"DB"`
	PoolSize int    `yaml:"PoolSize"`
}

func NewRedis(config Redis) *redis.Client {
	// 创建 Redis 客户端
	redisDb := redis.NewClient(&redis.Options{
		Addr:     config.Address,  // Redis 地址
		// 你用的是开源版 Redis，它不支持这个命令。但你用的 go-redis 客户端（版本 9.23.0）在默认的 ModeAuto 模式下，
		// 每次建立连接时都会尝试发送 CLIENT MAINT_NOTIFICATIONS 去探测服务器是否支持。
		MaintNotificationsConfig: &maintnotifications.Config{ // 
			Mode: maintnotifications.ModeDisabled,
		},
		Password: config.Password, // 密码（与 redis.conf 中的 requirepass 一致）
		DB:       config.DB,       // 使用哪个数据库（0-15）

		// 可选：连接池配置
		PoolSize:     config.PoolSize, // 连接池大小
		MinIdleConns: 10,              // 最小空闲连接数
		MaxRetries:   3,               // 最大重试次数

		// 超时配置
		DialTimeout:  5 * time.Second,  // 连接超时
		ReadTimeout:  10 * time.Second, // 读超时
		WriteTimeout: 10 * time.Second, // 写超时
	})
	redisDb.AddHook(loggingHook{}) // 关键：挂载 Hook
	// 测试连接
	ctx := context.Background()
	err := redisDb.Ping(ctx).Err()
	if err != nil {
		panic(fmt.Sprintf("Redis 连接失败: %v", err))
	}
	return redisDb

}
