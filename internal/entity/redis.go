package entity

import "time"

type KeyValueRedis struct {
	Key     string
	Value   string
	Timeout time.Duration
}

// 这里的rate是key刷新的最大比例
// 如果为零则默认为1,也就是没有比例只要小于Timeout就刷新到Timeout
type KeyTimeoutRedis struct {
	Key     string
	Timeout time.Duration
	Rate    float64
	TTl     int64
}
