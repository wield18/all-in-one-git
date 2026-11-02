package consumer

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func TestChan(t *testing.T) {
	abc := make(chan int, 1)
	abc <- 2
	fmt.Println(len(abc))
	fmt.Println(<-abc)
	fmt.Println(len(abc))
}

func TestShutDown(t *testing.T) {
	abc := make(chan struct{})
	close(abc)
	for {
		select {
		case v := <-abc:
			fmt.Println(v)
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func TestInterface(t *testing.T) {
	abc := 123
	def := Abc(abc).(int)
	fmt.Println(def)
}

func Abc(abc interface{}) interface{} {
	return abc
}

// 不能并发着写
func TestMap(t *testing.T) {
	aMap := make(map[int64]struct{})
	aInt := int64(0)
	ctx, cancel := context.WithCancel(context.Background())
	worker := func() {
		for {
			id := atomic.AddInt64(&aInt, 1)
			aMap[id] = struct{}{}
			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}
	for i := 0; i < 4; i++ {
		go worker()
	}

	time.Sleep(time.Second)
	cancel()
	fmt.Println(len(aMap) == int(aInt))
}

func TestTick(t *testing.T) {
	ticker := time.NewTicker(time.Second)
	ticker.Stop()
	ch := make(chan struct{})
	close(ch)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			time.Sleep(100 * time.Millisecond)
			fmt.Println("oh_no")
		case <-ch: // 正常的chan会疯狂塞nil
			time.Sleep(100 * time.Millisecond)
			fmt.Println("oh_no chan")
		}
	}

}

func TestCtx(t *testing.T) {
	ticker := time.NewTicker(time.Second)
	ticker.Stop()
	ch := make(chan struct{})
	close(ch)
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	for {
		select {
		case <-ctx.Done():
			time.Sleep(100 * time.Millisecond)
			fmt.Println("oh_no ctx")
		case <-ticker.C:
			time.Sleep(100 * time.Millisecond)
			fmt.Println("oh_no")
		case <-ch: // 正常的chan会疯狂塞nil
			time.Sleep(100 * time.Millisecond)
			fmt.Println("oh_no chan")
		}
	}

}

func TestCtx2(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	cancel()
	for {
		select {
		case <-ctx.Done():
			fmt.Println("oh_no ctx")
			time.Sleep(100 * time.Millisecond)
		default:
			fmt.Println("default")
			time.Sleep(100 * time.Millisecond)
		}
	}

}
