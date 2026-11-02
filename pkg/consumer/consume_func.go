package consumer

import (
	httperr "all-in-one/pkg/ekit/http_err"
	"context"
)

// callback接口
// data是真正你传构建consumeData时传入的Data
// 这里一般管理的是异步的监听或者err data 的记录
type ConsumeCallBackerInter func(data interface{}, e httperr.IError)

// 批量消费接口
// 如果是一些简单的问题里边可以直接retry,真的无法解决的问题才可能去callback,但callback尽量不要再次发送过来,
// 一般callback,只是个等待监听完事,err只是放个样子,但如果是panic的err可能我以后会写入个文件,以后看,但不再跑一版,
// 要记录具体的啥样数据 报的panic
// 真正执行consume时,有问题内部直接解决,剩下就是不能解决的问题,所以说要不把data返回过来完事吧,要记录一下也简单
type ConsumeFunc[v any] func(ctx context.Context, data []*ConsumeData[v])

// 这里必须得是泛型,不该写err返回,只需要panic保护一下
// 如果说panic掉了,打印栈堆到log里,但损失的数据消费没消费完,暂时先丢失,如果想保留啥的
// type PanicWrapperForConsumerFunc[v any] struct {
// 	coreFunc consumeFunc[v]
// }

// func (p *PanicWrapperForConsumerFunc[v]) proceeWithErr(ctx context.Context, data []*ConsumeData[v]) (err error) {
// 	defer func() {
// 		if r := recover(); r != nil {
// 			buf := make([]byte, buffLen)
// 			buf = buf[:runtime.Stack(buf, false)]
// 			err = &PanicError{Stack: string(buf)}
// 			// 把err返回上去
// 		}
// 	}()
// 	p.coreFunc(ctx, data)
// 	return nil
// }

// func (p *PanicWrapperForConsumerFunc[v]) process(data []*ConsumerData[v]) {
// 	p.coreFunc(data)
// }
