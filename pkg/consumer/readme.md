这里主要是一个任务分发器。
统一使用 `httperr.IError`，可以直接决定返回的状态。

主要需要了解的类型如下：

```go
// ConsumerPool，发布任务、注册用的结构体，线程安全实例
type ConsumerPool struct {
	aMap   map[string]CunsumerInterface // topic -> consumer，这里的 topic 还需要再斟酌一下
	ctx    context.Context
	cancel context.CancelFunc
	rwmu   sync.RWMutex // 保证线程安全
}

// CunsumerInterface 消费者接口
type CunsumerInterface interface {
	startListening(context.Context) // 启动内部监听
	// consume(int)                 // 我为什么一定要把参数原样传过去呢？
	submit(interface{})      // 数据不合法、断言不成功会直接 panic；阻塞或非阻塞提交
	check(interface{}) error // 检查断言是否成功，留给上层 panic
	tick()                   // 接收外层的 ticker 通知
	shutDown()               // 内部维护关闭
}

// 基本消费 data 使用的接口
type ConsumeDataInterface interface {
	SetCallbacker(callbacker ConsumeCallBackerInter) // 设置 callback
	GetCallbacker() ConsumeCallBackerInter           // 获得 callback
}

// SignalConsumer 初代实现
// 从信号量分配的角度申请 goroutine 开放
// 这里认为一类结构体只能被一类结构体消费，就类似一个消费者只能消费一种数据
type SignalConsumer[v any] struct {
	tickerChan  chan struct{}        // ticker chan
	dataChan    chan *ConsumeData[v] // 数据
	betchSize   int                  // 每批数据
	signal      *Signal[string]      // 用户分级申请算力，这里不维护，上层会把对应的 Signal 传过来直接用
	ctx         context.Context      // 主要是上层的 ctx 直接赋值，这是接口
	weight      int64                // 单次申请需要多少算力
	consumeFunc ConsumeFunc[v]       // 这里必须绑定对应的函数
}

// Signal 管理是否开 goroutine，还是阻塞
// 可链式申请部分算力
type Signal[k comparable] struct {
	Key k // 记个标识
	computingPower       int64      // 池子的总算力
	computingPowerExtend int64      // 可以向下层池子申请的算力
	currentPowerExpand   int64      // 当前已经向下申请的算力
	mu                   sync.Mutex // 向下申请时必须加锁
	Weighted             *semaphore.Weighted
	next                 *Signal[k] // 链式制度，这个申请完了就申请下一个，但你得知道你到底申请的是哪个 Signal，得放回去
}

// ConsumeData 泛型实现真实的被消费的结构体，我只实现了一个
type ConsumeData[v any] struct {
	Data       v
	callbacker ConsumeCallBackerInter
	Err        httperr.IError
}

// ./pkg/ekit/htttperr
// IError可能作为返回体使用
// json只暴露httpCode跟对应err
type IError struct {
	HttpCode  int    `json:"code"`  // http code
	HttpErr   string `json:"error"` // 导致的http错误
	InnerCode int    `json:"-"`     // 内部code
	Msg       string `json:"-"`     // 真正的原因, 得记录
}

```

使用的话，主要在于 pool 的 consumer 注册，以及用 pool 来进行分发消费。

示例代码：

```go
func TestIntance(*testing.T) {
	p := NewConsumerPool()
	signal := NewSignal[string](100)
	register(p, signal)
	p.StartTicker()
	for i := 0; i < 10; i++ {
		p.Submit("TopicTest", NewConsumeData(keyValue{Key: "key-" + strconv.Itoa(i), Value: "value-" + strconv.Itoa(i)},
			func(data interface{}, e httperr.IError) {
				if !e.IsZero() {
					dat := data.(keyValue)
					fmt.Println(dat.Key)
					fmt.Println(dat.Value)
				}
			}))
	}
	err := p.BlockSubmit("TopicTest", NewConsumeData(keyValue{Key: "key-end", Value: "value-end"}, nil))
	if !err.IsZero() {
		fmt.Println("end")
	}
	p.ShutDown()
}

type keyValue struct {
	Key   string
	Value string
}

func register(p *ConsumerPool, signal *Signal[string]) {
	TopicTest := "TopicTest"
	{
		p.Register(TopicTest,
			NewSignalConsumer[keyValue](signal).
				WithConsumerFunc(
					ConsumeKeyValue,
				).WithSize(12, 4, 1))
	}
}

func ConsumeKeyValue(ctx context.Context, datas []*ConsumeData[keyValue]) {
	dats := []keyValue{}
	fmt.Println(dats)
	for _, data := range datas {
		dats = append(dats, data.Data)
		fmt.Println(data.Data.Key)
		fmt.Println(data.Data.Value)
		data.WithErr(httperr.ErrBadGateway("nope"))
	}
}
```

如果类型不匹配，会直接 panic，这一点需要知道。

主要需要知道 `p` 有以下方法：

```go
func (p *ConsumerPool) BlockSubmit(topic string, data ConsumeDataInterface) httperr.IError
func (p *ConsumerPool) BlockSubmitMany(topic string, datas ...ConsumeDataInterface) []httperr.IError
func (p *ConsumerPool) BlockSubmitManyByPair(datas ...TopicAndData) []httperr.IError
func (p *ConsumerPool) Check(topic string, data interface{}) error            // 主要是类型检查，外部用不着
func (p *ConsumerPool) check(topic string) (CunsumerInterface, error)          // 内部 consumer 类型检查
func (p *ConsumerPool) Submit(topic string, data ConsumeDataInterface)         // 发送不阻塞
func (p *ConsumerPool) SubmitMany(topic string, datas ...ConsumeDataInterface) // 向同一个 consumer 发送多个，不阻塞
func (p *ConsumerPool) SubmitManyByPair(datas ...TopicAndData)                 // 向不同 consumer 发送不同数据
func (p *ConsumerPool) ShutDown()                                              // 关闭服务时，结束阻塞并运行完剩下的数据
func (p *ConsumerPool) Register(topic string, handle CunsumerInterface)        // 注册时使用
func (p *ConsumerPool) StartTicker() error                                     // NewConsumerPool 之后记得启动 ticker，因为有锁，启动后再 Register 也无妨
```
