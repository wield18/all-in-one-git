package consumer

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// var p = newConsumeRouter()

// // func TestMain(*testing.T) {
// // 	err := p.BlockSubmit(TopicTest, NewConsumeData(entity.KeyValue{
// // 		Key:   "gaojian",
// // 		Value: "guyuchun",
// // 	}, nil))
// // 	fmt.Println(err)
// // }

// // func TestMain2(*testing.T) {
// // 	abc := ConsumeData[entity.KeyValue]{Data: entity.KeyValue{
// // 		Key:   "gaojian",
// // 		Value: "guyuchun",
// // 	}, Err: errors.New("gaojian")}
// // 	err := p.BlockSubmitMany(TopicTest, &abc, NewConsumeData(entity.KeyValue{
// // 		Key:   "gaojian2",
// // 		Value: "guyuchun2",
// // 	}, nil))
// // 	fmt.Println(err)
// // }

// // func TestMain3(*testing.T) {
// // 	abc := ConsumeData[entity.KeyValue]{Data: entity.KeyValue{
// // 		Key:   "gaojian",
// // 		Value: "guyuchun",
// // 	}, Err: errors.New("gaojian")}
// // 	errs := p.BlockSubmitManyByPair(TopicAndData{Topic: TopicTest, Data: &abc}, TopicAndData{Topic: TopicTest, Data: NewConsumeData(entity.KeyValue{
// // 		Key:   "gaojian2",
// // 		Value: "guyuchun2",
// // 	}, nil)})
// // 	fmt.Println(errs)
// // }

// // func TestMain4(*testing.T) {
// // 	abc := ConsumeData[entity.KeyValue]{Data: entity.KeyValue{
// // 		Key:   "gaojian",
// // 		Value: "guyuchun",
// // 	}, Err: errors.New("gaojian"), callbacker: func(data interface{}, e error) { fmt.Println(e) }}
// // 	p.SubmitMany(TopicTest, &abc, NewConsumeData(entity.KeyValue{
// // 		Key:   "gaojian2",
// // 		Value: "guyuchun2",
// // 	}, func(data interface{}, e error) { fmt.Println(1) }))
// // 	time.Sleep(2 * time.Second)
// // }

// func newConsumeRouter() *ConsumerPool {
// 	p := NewConsumerPool()
// 	defer p.StartTicker()
// 	signal := NewSignal[string](100)
// 	register(p, signal)
// 	return p
// }

const (
	topicTest = "test"
)

type keyValue struct {
	Key   string
	Value string
}

// 这里为何不成功也是因为拿旧索引来初始化了
func register(p *ConsumerPool, signal *Signal[string]) {
	{
		p.Register(topicTest,
			NewSignalConsumer[keyValue](signal).
				WithConsumerFunc(consumeKeyValue).WithSize(128, 16, 1))
	}

}

func consumeKeyValue(ctx context.Context, datas []*ConsumeData[keyValue]) {
	for _, data := range datas {
		fmt.Println(data.Data.Key)
		fmt.Println(data.Data.Value)
	}
}

func TestConsumerPool(t *testing.T) {
	p := NewConsumerPool()
	register(p, NewSignal[string](100))
	p.StartTicker()
	err := p.BlockSubmit(topicTest, NewConsumeDataNilCallBack(keyValue{
		Key:   "gaojian",
		Value: "val",
	}))
	if !err.IsZero() {
		fmt.Println(err.Msg)
	}

	// panic: need (*consumer.ConsumeData[all-in-one/pkg/consumer.keyValue])(nil),
	// but gain &consumer.ConsumeData[interface {}]{Data:consumer.keyValue{Key:"gaojian", Value:"val"}, callbacker:(consumer.ConsumeCallBackerInter)(0x7ff60e073980), Err:httperr.IError{HttpCode:0, HttpErr:"", InnerCode:0, Msg:""}} [recovered, repanicked]
	p.ShutDown()
	time.Sleep(time.Second)
}

// // func TestConsumerPool2(t *testing.T) {
// // 	aMap := make(map[string]CunsumerInterface)
// // 	ctx := context.Background()
// // 	ctxTimeOut, _ := context.WithTimeout(ctx, 1*time.Second)
// // 	pool := NewConsumerPool(ctxTimeOut, aMap)

// // 	signal := NewSignal[string](100)
// // 	aMap["test1"] = NewConsumer[myInt]().WithSize(100, 32, 1).
// // 		WithConsumerFunc(func(ctx context.Context, data []*ConsumeData[myInt]) {}).WithSignal(signal)
// // 	pool.StartTicker()

// // 	var count int64
// // 	go func() {
// // 		for {
// // 			pool.Submit("test1", NewConsumerData(myInt(113), func(err error) { atomic.AddInt64(&count, 1) }))
// // 			select {
// // 			case <-ctxTimeOut.Done():
// // 				return
// // 			default:
// // 			}
// // 		}
// // 	}()

// // 	time.Sleep(time.Second * 2)
// // 	// pool.ShutDown()
// // 	fmt.Println(count)
// // 	time.Sleep(time.Second)
// // 	ok := signal.TryAcquire(context.Background(), 100)
// // 	if ok {
// // 		fmt.Println("good")
// // 	}
// // }

// // func TestMath(t *testing.T) {
// // 	abc := 7 & 10
// // 	fmt.Println(abc)
// // }

// // func TestAssert(*testing.T) {
// // 	abc := Abcd(NewConsumerData(myInt(113), func(err error) {}))
// // 	fmt.Println(abc.Data)
// // }

// // func Abcd(abc interface{}) *ConsumeData[myInt] {
// // 	return abc.(*ConsumeData[myInt])
// // }

// // type aCallBack struct {
// // 	fun func(e error)
// // }

// // func newACall(fun func(e error)) *aCallBack {
// // 	return &aCallBack{fun: fun}
// // }

// // func TestCallBack(*testing.T) {
// // 	// 检查一下直接栈堆溢出
// // 	// aCall := newACall(func(e error) { fmt.Println("gaojian") })
// // 	// aCall.fun = func(e error) { aCall.fun(e); fmt.Println("guyuchun") }
// // 	// aCall.fun(nil)

// // 	// 这样可以
// // 	aCall := newACall(func(e error) { fmt.Println("gaojian") })
// // 	abc := aCall.fun
// // 	aCall.fun = func(e error) { abc(e); fmt.Println("guyuchun") }
// // 	aCall.fun(nil)
// // }

// func TestChan1(t *testing.T) {
// 	ch := make(chan struct{})
// 	close(ch)
// 	<-ch
// }

// func testConsumerFunc(ctx context.Context, data []*ConsumeData[struct{}]) {

// }

// // func TestConsumerPool(t *testing.T) {
// // 	// ctx, _ := context.WithCancel(context.Background())
// // 	topic := "gaojian"
// // 	aMap := map[string]CunsumerWithSignalInterface{
// // 		topic: NewConsumer[struct{}]().WithConsumerFunc(testConsumerFunc).WithSize(16, 4, 1),
// // 	}
// // 	signal := NewSignal[string](100)
// // 	PackSignalPoolLocal(signal, aMap)
// // 	consumerV := NewConsumerPoolWithSignal(aMap)

// // 	var abc int64
// // 	// 这里会阻塞
// // 	consumerV.BlockSubmit(topic, NewConsumeData(struct{}{}, func(data interface{}, err error) {
// // 		atomic.AddInt64(&abc, 1)
// // 	}))
// // 	// 一堆阻塞
// // 	consumerV.BlockSubmitMany(topic,
// // 		NewConsumeData(struct{}{}, func(data interface{}, err error) { atomic.AddInt64(&abc, 1) }),
// // 		NewConsumeData(struct{}{}, func(data interface{}, err error) { atomic.AddInt64(&abc, 1) }),
// // 		NewConsumeData(struct{}{}, func(data interface{}, err error) { atomic.AddInt64(&abc, 1) }),
// // 		NewConsumeData(struct{}{}, func(data interface{}, err error) { atomic.AddInt64(&abc, 1) }),
// // 	)
// // 	pairs, _ := pairv.NewPairsPtr([]string{
// // 		topic,
// // 		topic,
// // 		topic,
// // 		topic,
// // 	}, []ConsumeDataInterface{
// // 		NewConsumeData(struct{}{}, func(data interface{}, err error) { atomic.AddInt64(&abc, 1) }),
// // 		NewConsumeData(struct{}{}, func(data interface{}, err error) { atomic.AddInt64(&abc, 1) }),
// // 		NewConsumeData(struct{}{}, func(data interface{}, err error) { atomic.AddInt64(&abc, 1) }),
// // 		NewConsumeData(struct{}{}, func(data interface{}, err error) { atomic.AddInt64(&abc, 1) }),
// // 	})
// // 	// 9
// // 	consumerV.BlockSubmitManyByPair(pairs...)
// // 	ch := make(chan struct{}, 100)

// // 	consumerV.Submit(topic,
// // 		NewConsumeData(struct{}{}, func(data interface{}, err error) { atomic.AddInt64(&abc, 1); ch <- struct{}{} }))

// // 	consumerV.SubmitMany(topic,
// // 		NewConsumeData(struct{}{}, func(data interface{}, err error) { atomic.AddInt64(&abc, 1); ch <- struct{}{} }),
// // 		NewConsumeData(struct{}{}, func(data interface{}, err error) { atomic.AddInt64(&abc, 1); ch <- struct{}{} }),
// // 		NewConsumeData(struct{}{}, func(data interface{}, err error) { atomic.AddInt64(&abc, 1); ch <- struct{}{} }),
// // 		NewConsumeData(struct{}{}, func(data interface{}, err error) { atomic.AddInt64(&abc, 1); ch <- struct{}{} }),
// // 	)
// // 	pairs, _ = pairv.NewPairsPtr([]string{
// // 		topic,
// // 		topic,
// // 		topic,
// // 		topic,
// // 	}, []ConsumeDataInterface{
// // 		NewConsumeData(struct{}{}, func(data interface{}, err error) { atomic.AddInt64(&abc, 1); ch <- struct{}{} }),
// // 		NewConsumeData(struct{}{}, func(data interface{}, err error) { atomic.AddInt64(&abc, 1); ch <- struct{}{} }),
// // 		NewConsumeData(struct{}{}, func(data interface{}, err error) { atomic.AddInt64(&abc, 1); ch <- struct{}{} }),
// // 		NewConsumeData(struct{}{}, func(data interface{}, err error) { atomic.AddInt64(&abc, 1); ch <- struct{}{} }),
// // 	})
// // 	consumerV.SubmitManyByPair(pairs...)
// // 	consumerV.ShutDown()
// // 	for {
// // 		select {
// // 		case <-ch:
// // 		}
// // 		if abc == 18 {
// // 			close(ch) // 直接关
// // 			break
// // 		}
// // 	}
// // 	fmt.Println("ok")
// // 	time.Sleep(1 * time.Second)
// // 	fmt.Println(abc)

// // }
