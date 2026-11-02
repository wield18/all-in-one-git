package consumer

import (
	httperr "all-in-one/pkg/ekit/http_err"
	"context"
	"fmt"
	"runtime"
	"time"
)

// CunsumerInterface 写成接口模式主要是方便注入,调用对应方法
// 对于我具体咋样进行的消费,不要管
type CunsumerInterface interface {
	startListening(context.Context) // 启动内部监听
	// consume(int)                     // 我为何一定要传参数一样传过去呢?
	submit(interface{})      // 数据不合法,断言不成功直接panic 阻塞或非阻塞提交
	check(interface{}) error // 检查断言是否成功, 留给上层panic
	tick()                   // 获得外层的ticker通知
	shutDown()               //内部维护关闭
}

// signal管理的Consummer
type CunsumerWithSignalInterface interface {
	CunsumerInterface
	SetSignal(signal *Signal[string])
}

// SignalConsumer 初代实现
// 使用信号量分配角度申请goroutine开放
type SignalConsumer[v any] struct {
	tickerChan  chan struct{}        // ticker chan
	dataChan    chan *ConsumeData[v] // 数据
	betchSize   int                  // 每批数据
	signal      *Signal[string]      // 用户分级申请算力,这里不维护,上层会把对应的Signal传过来直接用
	ctx         context.Context      // 主要是上层的ctx直接赋值,这是接口
	weight      int64                // 单次申请需要多少算力
	consumeFunc ConsumeFunc[v]       // 这里必须绑定对应的函数
}

// 直接设置
func (c *SignalConsumer[v]) SetSignal(signal *Signal[string]) {
	c.WithSignal(signal)
}

var _ CunsumerWithSignalInterface = &SignalConsumer[any]{}

// pool层关闭了tick不再接收
func (s *SignalConsumer[v]) tick() {
	if len(s.tickerChan) == 1 {
		return
	}
	s.tickerChan <- struct{}{}
}

// shutDown 一是不再监听
// 关闭所有chan
// 等待剩下全部完成
// 不做存储
func (s *SignalConsumer[v]) shutDown() {
	close(s.tickerChan)
	close(s.dataChan)
	// 阻塞启动所有的consumer
	// 直到dataChan没东西了
	datas := make([]*ConsumeData[v], 0)
	for data := range s.dataChan {
		datas = append(datas, data)
	}
	// l == 0
	// l == 10 s.betchSize == 5
	// count = 2
	// for 2
	// if / % betchSize != 0 {

	// }

	// 这里可能一个数据都没有,也可能正好被除尽
	l := len(datas)
	if l == 0 {
		return
	}
	count := l / s.betchSize
	// 也可能正好被除尽
	for i := 0; i < count; i++ {
		go s.consume(datas[s.betchSize*i : s.betchSize*(i+1)])
	}
	// 这里必须得有数据时才行
	if l%s.betchSize != 0 {
		go s.consume(datas[s.betchSize*(count):])
	}
}

func (s *SignalConsumer[v]) startListening(ctx context.Context) {
	s.ctx = ctx // 直接写入ctx
	listening := func() {
		datas := make([]*ConsumeData[v], 0)
		for {
			select {
			case <-s.ctx.Done():
				if len(datas) != 0 {
					go s.consume(datas)
				}
				s.shutDown()
				return
			case <-s.tickerChan:
				if len(datas) != 0 {
					go s.consume(datas)
					datas = make([]*ConsumeData[v], 0)
				}
			case data := <-s.dataChan:
				datas = append(datas, data)
				// 如果想搁这运行获得的这个datas不能直接传
				// 或者说我定义个其他方法来
				if len(datas) == s.betchSize {
					go s.consume(datas)
					datas = make([]*ConsumeData[v], 0) // 重置
				}
			}
		}
	}
	go listening()
}

// 这里是主要消费跟回调的部分
// 消费使用的还是我那个
// []*ConsumeData[v] 包装了一个WithErr保存错误的方法
// 而回调方法传入的 data, err都是[]*ConsumeData[v]的Data跟Err字段
// 如果Data太大了建议指针存储
func (s *SignalConsumer[v]) consume(datas []*ConsumeData[v]) {
	// 在这进行panic recover
	defer func() {
		if r := recover(); r != nil {
			s.signal.Release(s.weight) // 释放
			buf := make([]byte, buffLen)
			buf = buf[:runtime.Stack(buf, false)]
			fmt.Println(string(buf)) // 如果panic直接底层输出
		}
	}()

	ok := s.signal.TryAcquire(s.ctx, s.weight)
	if !ok {
		time.Sleep(consumeDelay)
		s.signal.Acquire(s.ctx, s.weight) // 必须获得
	}
	// 运行
	s.consumeFunc(s.ctx, datas) // 这里的err哈其实有点panic的感觉,如果内部一个返回了err

	// 所有callback使用完成所有data断联系
	// 满足其中一个就出不来
	for _, data := range datas {
		if data.callbacker == nil {
			data.flap()
			continue
		}
		data.callbacker(data.Data, data.Err)
		data.flap()
	}
	// 释放信号量, 如果前边崩了哈
	s.signal.Release(s.weight)
}

// check 就简单断言下
func (c *SignalConsumer[v]) check(data interface{}) error {
	dat, ok := data.(*ConsumeData[v])
	if !ok {
		return httperr.NewAssertWrong(dat, data)
	}
	return nil
}

// checkReturnData 返回数据不是吗
func (c *SignalConsumer[v]) checkReturnData(data interface{}) (*ConsumeData[v], error) {
	dat, ok := data.(*ConsumeData[v])
	if !ok {
		return dat, httperr.NewAssertWrong(dat, data)
	}
	return dat, nil
}

// submit 返回false只有可能是断言失败了
func (s *SignalConsumer[v]) submit(data interface{}) {
	dat, err := s.checkReturnData(data)
	if err != nil {
		panic(err)
	}
	s.dataChan <- dat

}
