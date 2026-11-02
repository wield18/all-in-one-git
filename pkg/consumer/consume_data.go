package consumer

import httperr "all-in-one/pkg/ekit/http_err"

// 一个正常
// 如果callbacker是nil的话
// 虽说返回的是ConsumeDataInterface,
// 但底层还是 &ConsumeData[v]
func NewConsumeData[v any](data v, callbacker ConsumeCallBackerInter) ConsumeDataInterface {
	return &ConsumeData[v]{
		Data:       data,
		callbacker: callbacker,
	}
}

func NewConsumeDataNilCallBack[v any](data v) ConsumeDataInterface {
	return &ConsumeData[v]{
		Data: data,
	}
}

// 获得一堆的ConsumeData
// 使用 nil callback
// 虽说返回的是ConsumeDataInterface,
// 但底层还是 &ConsumeData[v]
func NewConsumeDatasByNilCallbacker[v any](datas []v) []ConsumeDataInterface {
	l := len(datas)
	res := make([]ConsumeDataInterface, 0, l)
	for i := 0; i < l; i++ {
		res = append(res, NewConsumeData(datas[i], nil))
	}
	return res
}

type ConsumeDataInterface interface {
	SetCallbacker(callbacker ConsumeCallBackerInter) // 设置callback
	GetCallbacker() ConsumeCallBackerInter           // 获得callback
}

// 数据
type ConsumeData[v any] struct {
	// topic string
	// level      int // 给SignalPool判断等级用 // 最外层使用其实不用写在这
	// retryID    int64 // 这个东西一般零值,是去map找
	Data       v
	callbacker ConsumeCallBackerInter
	Err        httperr.IError
}

// 这样拿的话会有额外的struct复制消耗，如果不是指针
// 包里有个法可以直接遍历获得
func Datas[v any](datas []*ConsumeData[v]) []v {
	res := make([]v, 0, len(datas))
	for _, data := range datas {
		res = append(res, data.Data)
	}
	return res
}

func Errs[v any](datas []*ConsumeData[v]) []error {
	res := make([]error, 0, len(datas))
	for _, data := range datas {
		res = append(res, data.Err)
	}
	return res
}

func (c *ConsumeData[v]) GetCallbacker() ConsumeCallBackerInter {
	return c.callbacker
}

func (c *ConsumeData[v]) SetCallbacker(callbacker ConsumeCallBackerInter) {
	c.callbacker = callbacker
}

// 取消引用
func (c *ConsumeData[v]) flap() {
	var zero v
	// c.topic = ""
	c.Data = zero
	var zeroErr httperr.IError
	c.Err = zeroErr
	c.callbacker = nil
}

// WithErr 覆盖重写
func (c *ConsumeData[v]) WithErr(e httperr.IError) {
	c.Err = e
}
