package consumer

import (
	"context"
	"fmt"
	"testing"
	"time"

	"golang.org/x/sync/semaphore"
)

// func TestNewPool(t *testing.T) {
// 	// NewSignalPool()
// 	def := NewSignal[string]().WithK("def").WithComputingPower(100)
// 	abc := NewSignal[string]().WithK("abc").WithComputingPower(100).WithNext(def, 50)
// 	p := NewSignalPool(abc)
// 	for k, v := range p {
// 		fmt.Println(k, v.computingPowerExtend)
// 	}

// }

// // 我得一直获得
// func TestTryToExtend(t *testing.T) {
// 	def := NewSignal[string]().WithK("def").WithComputingPower(100)
// 	abc := NewSignal[string]().WithK("abc").WithComputingPower(100).WithNext(def, 50)
// 	p := NewSignalPool(abc)
// 	ctx, cancel := context.WithCancel(context.Background())

// 	var powerGain int32
// 	worker := func() {
// 		for {
// 			ok := p["abc"].TryAcquire(ctx, 1)
// 			if ok {
// 				atomic.AddInt32(&powerGain, 1)
// 			}
// 			select {
// 			case <-ctx.Done():
// 				return
// 			default:
// 			}
// 		}
// 	}
// 	for i := 0; i < 4; i++ {
// 		go worker()
// 	}
// 	time.Sleep(time.Second)
// 	cancel()
// 	fmt.Println(powerGain)

// 	// 尝试放回去看看下层是不是优先被放回,下层被放回了没
// 	ok := p["def"].TryAcquire(ctx, 100)
// 	if !ok {
// 		fmt.Println("还没释放")
// 	}
// 	p["abc"].Release(50)
// 	ok = p["def"].TryAcquire(ctx, 100)
// 	if ok {
// 		fmt.Println("底层的被释放掉了")
// 	}
// 	p["abc"].Release(100)

// 	// 直接尝试获得底层signal看能不能获得全部
// 	fmt.Println(p["abc"].currentPowerExpand)

// 	ctx, cancel = context.WithCancel(context.Background())

// 	smallWorker := func() {
// 		// 这里在疯狂的申请并放回去,放回去的少,
// 		for {
// 			p["abc"].TryAcquire(ctx, 2) // 这会尝试获得,并不会阻塞,会保证池子一直有一个或俩个
// 			p["abc"].Release(1)
// 			select {
// 			case <-ctx.Done():
// 				return
// 			default:
// 			}
// 		}
// 	}

// 	bigWorker := func() {
// 		time.Sleep(time.Millisecond * 100)
// 		// 这里只是尝试获得但并不放回
// 		for {
// 			p["abc"].Acquire(ctx, 20)
// 			fmt.Println("获得了")
// 			select {
// 			case <-ctx.Done():
// 				return
// 			default:
// 			}
// 		}
// 	}
// 	go smallWorker()
// 	go bigWorker()
// 	time.Sleep(time.Second * 1)
// 	cancel()
// }

// 虽说ai说Acquire有队列我得试试,可以
func TestQueue(t *testing.T) {
	weighted := semaphore.NewWeighted(100)
	ctx, cancel := context.WithCancel(context.Background())

	smallWorker := func() {
		// 这里在疯狂的申请并放回去,放回去的少,
		for {
			weighted.TryAcquire(2) // 这会尝试获得,并不会阻塞,会保证池子一直有一个或俩个
			weighted.Release(1)
			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}

	bigWorker := func() {
		time.Sleep(time.Millisecond * 100)
		// 这里只是尝试获得但并不放回
		for {
			weighted.Acquire(ctx, 20)
			fmt.Println("获得了")
			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}
	go smallWorker()
	go bigWorker()
	time.Sleep(time.Second * 1)
	cancel()
}
