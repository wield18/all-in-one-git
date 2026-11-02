package consumer

import (
	httperr "all-in-one/pkg/ekit/http_err"
	"context"
	"fmt"
	"sync"
	"time"
)

// 不太像池子,像个消息分发器
// CunsumerInterface的实现现在只有一个直接go routine的
type ConsumerPool struct {
	aMap   map[string]CunsumerInterface // topic -> consumer, 这里的topic还是进行下
	ctx    context.Context
	cancel context.CancelFunc
	rwmu   sync.RWMutex // 保证线程安全
}

// BlockSubmitNoCallback
// 只需要传consume层对应的data就行
func (p *ConsumerPool) BlockSubmit(topic string, data ConsumeDataInterface) httperr.IError {
	var err httperr.IError
	ch := make(chan struct{}, 1)
	defer close(ch)
	abc := data.GetCallbacker()
	if abc != nil {
		data.SetCallbacker(func(data interface{}, e httperr.IError) {
			abc(data, e)
			err = e
			ch <- struct{}{}
		})
	} else {
		data.SetCallbacker(func(data interface{}, e httperr.IError) {
			err = e
			ch <- struct{}{}
		})
	}
	// 这里边锁了
	p.Submit(topic, data)
	// 我得等它执行完说白了
	<-ch
	return err
}

// BlockSubmitMany 基本一个topic发多个
// 没有东西传过来或者说传过来的东西验证不通过直接panic
// 返回的err只有可能是 errCtxDone
func (p *ConsumerPool) BlockSubmitMany(topic string, datas ...ConsumeDataInterface) []httperr.IError {
	l := len(datas)
	if l == 0 {
		panic(errNoData)
	}
	// topic检查
	consumer, err := p.check(topic)
	if err != nil {
		panic(err)
	}
	ch := make(chan struct{}, 1)
	errs := make([]httperr.IError, l)
	defer close(ch)
	// 先检查并包装函数
	for i, data := range datas {
		// 必须调用池子的check
		err = consumer.check(data)
		if err != nil {
			panic(err)
		}
		abc := data.GetCallbacker()
		if abc != nil {
			data.SetCallbacker(func(data interface{}, e httperr.IError) {
				abc(data, e)
				errs[i] = e
				ch <- struct{}{}
			})
			continue
		}
		data.SetCallbacker(func(data interface{}, e httperr.IError) {
			errs[i] = e
			ch <- struct{}{}
		})

	}
	// 启动所有,因为预先进行了所有的检查
	// 这里边不可能panic
	// 这里需要锁是因为 直接调用内部 没有锁保护
	p.rwmu.RLock()
	select {
	case <-p.ctx.Done():
		for i := 0; i < l; i++ {
			errs[i] = errCtxDone
		}
		return errs
	default:
	}
	for _, data := range datas {
		consumer.submit(data) // 直接调用内部
	}
	p.rwmu.RUnlock()

	// 等待所有
	for i := 0; i < l; i++ {
		<-ch
	}
	return errs
}

type TopicAndData struct {
	Topic string
	Data  ConsumeDataInterface
}

// 没有类型安全
// type TopicAndPureData struct {
// 	Topic string
// 	Data  any
// }

func NewTopicAndData(topic string, data ConsumeDataInterface) TopicAndData {
	return TopicAndData{Topic: topic, Data: data}
}

// BlockSubmitManyByPair 一次发送多个
// 因为是pair写法,所以不好初始化,但效果是能发任意的topic任意的data
// 这个err只有检查错误
// 如果没有任何pair直接返回
func (p *ConsumerPool) BlockSubmitManyByPair(datas ...TopicAndData) []httperr.IError {
	l := len(datas)
	if l == 0 {
		panic(errNoData)
	}
	var err error
	ch := make(chan struct{}, 1)
	errs := make([]httperr.IError, l)
	defer close(ch)
	// 先检查并包装函数
	for i, data := range datas {
		// 必须调用池子的check
		err = p.Check(data.Topic, data.Data)
		if err != nil {
			panic(err)
		}
		abc := data.Data.GetCallbacker()
		if abc != nil {
			data.Data.SetCallbacker(func(data interface{}, e httperr.IError) {
				abc(data, e)
				errs[i] = e
				ch <- struct{}{}
			})
			continue
		}
		data.Data.SetCallbacker(func(data interface{}, e httperr.IError) {
			errs[i] = e
			ch <- struct{}{}
		})

	}
	// 启动所有,因为预先进行了所有的检查,这里不可能有发送错误
	// 内部锁了
	for _, data := range datas {
		p.Submit(data.Topic, data.Data)
	}
	// 等待所有
	for i := 0; i < l; i++ {
		<-ch
	}
	return errs
}

// Check 检查俩个地方
// 1. topic找不到的错误
// 2. 是否有基本的断言(内部检查)
func (p *ConsumerPool) Check(topic string, data interface{}) error {
	consumer, err := p.check(topic)
	if err != nil {
		return err
	}
	return consumer.check(data)
}

// check 检查
// topic是否找到
func (p *ConsumerPool) check(topic string) (CunsumerInterface, error) {
	consumer, ok := p.aMap[topic]
	if !ok {
		return consumer, NewTopicNotFound(topic)
	}
	return consumer, nil
}

// 一旦检查不通过直接panic
func (p *ConsumerPool) Submit(topic string, data ConsumeDataInterface) {
	consumer, err := p.check(topic)
	if err != nil {
		panic(err)
	}

	p.rwmu.RLock()
	defer p.rwmu.RUnlock()
	select {
	case <-p.ctx.Done():
		return
	default:
	}
	consumer.submit(data) // 这里有断言失败的可能断言失败直接就panic了 嗯 其他没有
}

// 要求相同topic
func (p *ConsumerPool) SubmitMany(topic string, datas ...ConsumeDataInterface) {
	l := len(datas)
	if l == 0 {
		panic(errNoData)
	}
	// 基本的topic检查
	consumer, err := p.check(topic)
	if err != nil {
		panic(err)
	}
	// 启动所有,这里不可能有发送错误
	// consumer.submit 的时候进行内部的断言检查不行直接panic
	p.rwmu.RLock()
	defer p.rwmu.RUnlock()
	select {
	case <-p.ctx.Done():
		return
	default:
	}
	for _, data := range datas {
		consumer.submit(data)
	}
}

// SubmitManyByPair 任意的topic跟consumer组合
func (p *ConsumerPool) SubmitManyByPair(datas ...TopicAndData) {
	l := len(datas)
	if l == 0 {
		panic(errNoData)
	}
	// p.Submit 已经锁了
	for _, data := range datas {
		p.Submit(data.Topic, data.Data)
	}
}

// 这里要调用shutdown的话必须保证没有人会继续发东西
// 就算发消息,我得预防
// 这里调用cancel时可能StartTicker还在tick呢,让后我把所有chan的都关了,直接panic
func (p *ConsumerPool) ShutDown() {
	p.rwmu.Lock()
	defer p.rwmu.Unlock()
	p.cancel()
}

// 一个新的pool没有任何东西
func NewConsumerPool() *ConsumerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &ConsumerPool{
		ctx:    ctx,
		cancel: cancel,
		aMap:   make(map[string]CunsumerInterface),
	}
}

func (p *ConsumerPool) Register(topic string, handle CunsumerInterface) {
	p.rwmu.Lock()
	defer p.rwmu.Unlock()
	p.aMap[topic] = handle
}

// StartTicker 使用 NewConsumerPool 时需嗲用 StartTicker 起动ticker通知
func (p *ConsumerPool) StartTicker() error {
	for _, oneConsumer := range p.aMap {
		oneConsumer.startListening(p.ctx)
	}
	ticker := time.NewTicker(tickGap)
	ticking := func() {
		for {
			select {
			case <-p.ctx.Done():
				fmt.Println("ConsumerPool StartTicker Stop")
				ticker.Stop()
				return
			case <-ticker.C:
				// 调用stop直接过不来,加写锁,把底层的tick调用完,然后再cancel,会退出这个协程,并直接退出下层的所有的consumer
				// 我这里就算锁了,它可能是在这等待,我必须得拿个标志位才行
				p.rwmu.Lock()
				select {
				case <-p.ctx.Done():
					return
				default:
				}
				for _, oneConsumer := range p.aMap {
					oneConsumer.tick()
				}
				p.rwmu.Unlock()
			}
		}
	}
	go ticking()
	return nil
}
