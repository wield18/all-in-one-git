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

type ViewCountTicker struct {
	exec   func(mdb *gorm.DB, aMap map[int64]int64)
	aMap   map[int64]map[int64]struct{} // videoID:userID:struct
	ctx    context.Context
	cancel context.CancelFunc
	mutex  sync.RWMutex
	mdb    *gorm.DB
}

// 传入的exec得能消费多个根单个
func NewCountTicker(tickerGap time.Duration, mdb *gorm.DB, exec func(mdb *gorm.DB, aMap map[int64]int64)) *ViewCountTicker {
	ctx, cancel := context.WithCancel(context.Background())
	c := ViewCountTicker{
		aMap: make(map[int64]map[int64]struct{}), ctx: ctx, cancel: cancel,
		exec: exec, mdb: mdb,
	}
	go c.startTicking(tickerGap)
	return &c
}

func (c *ViewCountTicker) startTicking(tickerGap time.Duration) {
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

func (c *ViewCountTicker) Exec() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	aMap := make(map[int64]int64)
	for k, v := range c.aMap {
		aMap[k] = int64(len(v))
	}
	c.exec(c.mdb, aMap)
	c.aMap = make(map[int64]map[int64]struct{})
}

func (c *ViewCountTicker) ShutDown() {
	c.cancel()
}
func (c *ViewCountTicker) ExecOne(key int64) {
	if v, ok := c.aMap[key]; ok {
		c.mutex.Lock()
		defer c.mutex.Unlock()
		c.exec(c.mdb, map[int64]int64{key: int64(len(v))})
		delete(c.aMap, key)
	}
}
func (c *ViewCountTicker) Incr(key int64, userID int64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if _, ok := c.aMap[key]; !ok {
		c.aMap[key] = map[int64]struct{}{userID: {}}
		return
	}
	if _, ok := c.aMap[key][userID]; !ok {
		c.aMap[key][userID] = struct{}{}
	}

}

func (c *ViewCountTicker) GetIncr(key int64) int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.aMap[key])
}
