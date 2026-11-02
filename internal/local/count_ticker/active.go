package counttikcer

import (
	"context"
	"sync"
	"time"

	"gorm.io/gorm"
)

// 如果使用泛型得自己定义个方法来确定是否能添加,
// 而且因为泛型的问题那个方法只能针对那个类型,
// 而那个方法估计是跟这一起传过来,还有具体执行也是

// 如果使用接口
// 可以调用接口的方法来进行检查
// 而因为接口的存在可以直接对于一个接口绑定对应的执行方法

type ActiveCountTicker struct {
	exec   func(mdb *gorm.DB, aMap map[int64]int64)
	aMap   map[int64]int64
	ctx    context.Context
	cancel context.CancelFunc
	mutex  sync.RWMutex
	mdb    *gorm.DB
}

// 传入的exec得能消费多个根单个
func NewActiveCountTicker(tickerGap time.Duration, mdb *gorm.DB, exec func(mdb *gorm.DB, aMap map[int64]int64)) *ActiveCountTicker {
	ctx, cancel := context.WithCancel(context.Background())
	c := ActiveCountTicker{aMap: make(map[int64]int64), ctx: ctx, cancel: cancel,
		exec: exec, mdb: mdb,
	}
	go c.startTicking(tickerGap)
	return &c
}

func (c *ActiveCountTicker) startTicking(tickerGap time.Duration) {
	ticker := time.NewTicker(tickerGap)
	for {
		select {
		case <-c.ctx.Done():
			if len(c.aMap) != 0 {
				c.Exec()
			}
			return
		case <-ticker.C:
			if len(c.aMap) != 0 {
				c.Exec()
			}

		}
	}
}

func (c *ActiveCountTicker) Exec() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.exec(c.mdb, c.aMap)
	c.aMap = make(map[int64]int64)
}

func (c *ActiveCountTicker) ShutDown() {
	c.cancel()
}
func (c *ActiveCountTicker) ExecOne(key int64) {
	if v, ok := c.aMap[key]; ok {
		c.mutex.Lock()
		defer c.mutex.Unlock()
		c.exec(c.mdb, map[int64]int64{key: v})
		delete(c.aMap, key)
	}
}
func (c *ActiveCountTicker) Incr(userID int64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.aMap[userID] = c.aMap[userID] + 1

}

func (c *ActiveCountTicker) GetIncr(userID int64) int64 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.aMap[userID]
}
