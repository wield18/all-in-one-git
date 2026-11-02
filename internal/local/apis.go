package local

import (
	"all-in-one/internal/base"
	counttikcer "all-in-one/internal/local/count_ticker"
	tokenstorage "all-in-one/internal/local/token_storage"
	videobuf "all-in-one/internal/local/video_buf"
	"all-in-one/internal/repository"
	"time"
)

// 所有local服务的初始化
// 这里的服务不能在base里,因为它是基于base的
var (
	// TokenStorage = tokenstorage.New()
	TokenStorage *tokenstorage.UserCodeStorage
	// ViewCount    = counttikcer.NewCountTicker(time.Second*1, repository.IncrVideoViewsCount)
	ViewCount *counttikcer.ViewCountTicker
	// ActiveCount  = counttikcer.NewActiveCountTicker(time.Second*1, repository.IncrActiveCount)
	ActiveCount *counttikcer.ActiveCountTicker
	// VideoBuf     = videobuf.New() // 这里不能直接new
	VideoBuf *videobuf.VideoBufServer
)

func InitLocal() {
	TokenStorage = tokenstorage.New(base.P)
	ViewCount = counttikcer.NewCountTicker(time.Hour, base.Mysql, repository.IncrVideoViewsCount)
	ActiveCount = counttikcer.NewActiveCountTicker(time.Hour, base.Mysql, repository.IncrActiveCount)
	VideoBuf = videobuf.New(base.Redis, base.ES)
}
