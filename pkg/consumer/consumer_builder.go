package consumer

func NewSignalConsumer[v any](signal *Signal[string]) *SignalConsumer[v] {
	return &SignalConsumer[v]{
		tickerChan: make(chan struct{}, tickerChanSize),
		signal:     signal,
	}
}

func (c *SignalConsumer[v]) WithSize(dataChanSize, betchSize, weight int) *SignalConsumer[v] {
	if dataChanSize < minimumDataChanSize {
		dataChanSize = minimumDataChanSize
	}
	if betchSize < minimumBetchSize {
		betchSize = minimumBetchSize
	}
	if weight < 1 {
		weight = 1
	}
	c.betchSize = betchSize
	c.dataChan = make(chan *ConsumeData[v], dataChanSize)
	c.weight = int64(weight)
	return c
}

func (c *SignalConsumer[v]) WithSignal(signal *Signal[string]) *SignalConsumer[v] {
	c.signal = signal
	return c
}

func (c *SignalConsumer[v]) WithConsumerFunc(fun ConsumeFunc[v]) *SignalConsumer[v] {
	// c.consumeFunc = NewPanicWrapperedConsumerFunc(fun)
	c.consumeFunc = fun
	return c
}

// func NewPanicWrapperedConsumerFunc[v any](fun consumeFunc[v]) *PanicWrapperForConsumerFunc[v] {
// 	return &PanicWrapperForConsumerFunc[v]{coreFunc: fun}
// }
