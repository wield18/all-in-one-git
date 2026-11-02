package consumer

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/sync/semaphore"
)

type Signal[k comparable] struct {
	Key k // 记个标识
	// maxParallel          int        // 最大协程并行 // 这样的话添加了个下限,一个topic,其实哈没太大用,因为我是算力池,已经有算力限制了,只要保证大算力的也能申请到算力就行
	computingPower       int64      // 池子的总算力
	computingPowerExtend int64      // 可以向下层池子申请的算力
	currentPowerExpand   int64      // 当前已经向下申请的算力
	mu                   sync.Mutex // 向下申请时必须加锁
	Weighted             *semaphore.Weighted
	next                 *Signal[k] // 链式制度,这个申请完了就申请下个,但你得知道你到底申请的是哪个Signal,得放回去
}

// TryAcquire 这个如果满足条件会去底层阻塞获得
func (s *Signal[k]) TryAcquire(ctx context.Context, weight int64) bool {
	ok := s.Weighted.TryAcquire(weight)
	if !ok && weight <= s.computingPowerExtend-s.currentPowerExpand { // 可以向下申请时, 一般有数值必然s.next不为nil,但还是保险下
		s.mu.Lock()
		defer s.mu.Unlock()
		if weight <= s.computingPowerExtend-s.currentPowerExpand {
			if s.next != nil { // 可以向下申请时, 一般有数值必然s.next不为nil,但还是保险下
				s.currentPowerExpand += weight
				s.next.Acquire(ctx, weight) // 强制获得底层的,我得记录一下它获得的是底层的还是当前层的
				return true
			}
			fmt.Printf("当前池子(key: %v)有下层申请余额但下层池子是nil \n", s.Key)
			return ok
		}
	}
	return ok
}

// Acquire 这里只能在当前层获得,不能去底层
// 阻塞获得,官方实现是带队列的申请,也就是先到先得,所以不用担心分段获得啥的
func (s *Signal[k]) Acquire(ctx context.Context, weight int64) {
	s.Weighted.Acquire(ctx, weight)
}

// 我咋知道到底啥时候才把从底层获得的还回去?
// 直接还吧来了个release的直接看看能不能把它的currentPowerExpand给减掉点,不能小于零
// 可能得疯狂上锁解锁,但最好这样
func (s *Signal[k]) Release(weight int64) {
	if weight <= s.currentPowerExpand { // 如果小于当前的扩展直接还回去
		s.mu.Lock()
		defer s.mu.Unlock()
		if weight <= s.currentPowerExpand {
			s.currentPowerExpand -= weight
			// 这里就别管next还会为nil的事了
			// 都获得完了,还可能为nil?
			s.next.Release(weight)
		}
		return
	}
	s.Weighted.Release(weight)
}
