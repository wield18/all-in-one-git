package consumer

import "time"

// ticker发送时不能阻塞,所以说只要发送是提前检查下就行别搞缓冲吗?也行
var minimumDataChanSize = 16
var minimumBetchSize = 4
var tickGap = time.Second
var tickerChanSize = 1 // 1阻塞ticker
var consumeDelay = time.Second
var buffLen = 8192
