// Package producerconsumer 就一个本地异步消费的东西
package producerconsumer

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

var (
	buffLen  = 2048
	errPanic = errors.New("Panic")
	DATA     = "data"
)

type PanicError struct {
	Value interface{}
	Stack string
}

func (e *PanicError) Error() string {
	return fmt.Sprintf("panic: %v\n%s", e.Value, e.Stack)
}

func (e *PanicError) Is(target error) bool {
	_, ok := target.(*PanicError)
	return ok
}

// 真正运行时的各种状态
var (
	created  int32 = 0
	running  int32 = 1
	stopping int32 = 2
	dead     int32 = 3
	locked   int32 = 4
)

// 具体所有consume的状态
var (
	consumeCreated int32 = 0
	consumeSuccess int32 = 1
	consumeError   int32 = 2
	consumePanic   int32 = 3
)

type Consumer interface {
	Consume(ctx context.Context) error
}

type ConsumerFunc func(ctx context.Context) error

func (c ConsumerFunc) Consume(ctx context.Context) error { return c(ctx) }

type ConsumerFuncWrapper struct {
	Consumer Consumer
}

func (c *ConsumerFuncWrapper) Consume(ctx context.Context) (err error) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, buffLen)
			buf = buf[:runtime.Stack(buf, false)]
			err = &PanicError{Value: errPanic, Stack: string(buf)}
		}
	}()
	return c.Consumer.Consume(ctx)
}

type ConsumerPool struct {
	Group           sync.Map // 只暴露一个Group 其他内部管理
	queue           chan SubmitData
	corGO           int32
	maxGO           int32
	currentPoolSize int32
	ctx             context.Context
	cancel          context.CancelCauseFunc
	state           int32
	maxIdleGoTime   time.Duration // 一个Goroutine最长等待时间
}

type SubmitData struct {
	Topic    string
	Data     interface{}
	State    int32
	CallBack func(consumeState int32, err error)
}

var (
	goCountWrong             = "corgo > maxgo, So maxGo = corGo"
	errGoRoutine             = errors.New("go routine can not <= 0")
	errQueueSizeZero         = errors.New("queue size can not be zero")
	errMaxTimeZero           = errors.New("max can not be zero")
	errPoolStateIsntCreatred = errors.New("pool state isn't Created")
	errTopicIsNotFound       = errors.New("topic is not found")
	errSubmitFailed          = errors.New("submit failed")
	errQueueFull             = errors.New("queue full")
	errUserInterrupt         = errors.New("user interrupt")
	errShutDownState         = errors.New("state is not running")
	errStateLocked           = errors.New("pool locked")
	errTopicAlreadyRegisted  = errors.New("this topic's consumer already in group")
	errTopicNil              = errors.New("topic is nil")
	errCallBackIsNil         = errors.New("callback is nil")
	errWrongState            = errors.New("wrong state")
)

// NewConsumerPool:
func NewConsumerPool(queueSize, corGo, maxGo int, maxIdleGoTime time.Duration) (*ConsumerPool, error) {
	if queueSize <= 0 {
		return nil, errQueueSizeZero
	}
	if corGo <= 0 || maxGo <= 0 {
		return nil, errGoRoutine
	}
	if maxIdleGoTime <= 0 {
		return nil, errMaxTimeZero
	}
	if maxGo < corGo {
		maxGo = corGo
		fmt.Println(goCountWrong)
	}

	ctx, cancel := context.WithCancelCause(context.Background())
	p := &ConsumerPool{
		queue:         make(chan SubmitData, queueSize),
		corGO:         int32(corGo),
		maxGO:         int32(maxGo),
		state:         created,
		maxIdleGoTime: maxIdleGoTime,
		ctx:           ctx,
		cancel:        cancel,
	}
	return p, nil
}

func (p *ConsumerPool) Start() error {

	// 先创建几个routine
	if atomic.CompareAndSwapInt32(&p.state, created, locked) {
		defer atomic.CompareAndSwapInt32(&p.state, locked, running)
		atomic.AddInt32(&p.currentPoolSize, p.corGO)
		for i := 0; i < int(p.corGO); i++ {
			go p.goroutine(i)
		}
		return nil
	}

	return errPoolStateIsntCreatred

}

func (p *ConsumerPool) goroutine(id int) {
	idleTimer := time.NewTimer(0)
	if !idleTimer.Stop() {
		<-idleTimer.C
	}
	for {
		select { // 这里任务放第一个 因为有义务把所有任务全完成 再结束
		// 任务
		case data, ok := <-p.queue:
			if !ok {
				atomic.AddInt32(&p.currentPoolSize, -1)
				return
			}
			var fun Consumer
			consuemr, ok := p.Group.Load(data.Topic)
			if !ok {
				fmt.Printf("%v: %s", errTopicIsNotFound, data.Topic)
				continue
			}
			fun = consuemr.(Consumer)
			err := fun.Consume(p.ctx)
			if err != nil {
				if errors.Is(err, &PanicError{}) {
					data.State = consumePanic
				} else {
					data.State = consumeError
				}
			} else {
				data.State = consumeSuccess
			}
			data.CallBack(data.State, err)

			// 查看一下这个goroutine是否需要留存
			// 这里判断不是原子的，虽说可能会导致当前的协程数量小于一下corGo
			// 但不太影响吧
			if p.currentPoolSize >= p.corGO { // 加idleTime
				idleTimer.Reset(p.maxIdleGoTime)
			}
		// 结束
		case <-p.ctx.Done():
			atomic.AddInt32(&p.currentPoolSize, -1)
			return
		// 超时
		case <-idleTimer.C:
			atomic.AddInt32(&p.currentPoolSize, -1)
			return
		}
	}
}

func (p *ConsumerPool) Submit(topic string, data interface{},
	callBack func(consumeState int32, err error)) error {
	if topic == "" {
		return errTopicNil
	}
	if callBack == nil {
		return errCallBackIsNil
	}
	if _, ok := p.Group.Load(topic); !ok {
		return errTopicIsNotFound
	}

	state := atomic.LoadInt32(&p.state)
	if state == stopping || state == dead {
		return fmt.Errorf("%w: %d", errSubmitFailed, state)
	}
	// 能到这的有 locked created running
	da := &SubmitData{
		Topic:    topic,
		Data:     data,
		State:    consumeCreated,
		CallBack: callBack,
	}
	if atomic.CompareAndSwapInt32(&p.state, created, locked) {
		defer atomic.CompareAndSwapInt32(&p.state, locked, created)
		// lock的时候再判断这东西保证安全不能阻塞
		if len(p.queue) == cap(p.queue) {
			return errQueueFull
		}
		p.queue <- *da
		if p.currentPoolSize <= p.maxGO {
			id := atomic.AddInt32(&p.currentPoolSize, 1)
			go p.goroutine(int(id))
		}
		return nil
	}
	if atomic.CompareAndSwapInt32(&p.state, running, locked) {
		defer atomic.CompareAndSwapInt32(&p.state, locked, running)
		if len(p.queue) == cap(p.queue) {
			return errQueueFull
		}
		p.queue <- *da
		if p.currentPoolSize <= p.maxGO {
			id := atomic.AddInt32(&p.currentPoolSize, 1)
			go p.goroutine(int(id))
		}
		return nil
	}

	// 可能别人一直没解锁 也可能在cas时他上锁了我没上成
	return errStateLocked
}

// 阻塞，等待所有完成
func (p *ConsumerPool) ShutDown() error {

	state := atomic.LoadInt32(&p.state)
	if state == stopping || state == dead || state == created {
		return fmt.Errorf("%w: %d", errWrongState, state)
	}
	p.cancel(errUserInterrupt)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		// 如果locked直接stopping locked时我只是创建在最后直接改成stopping
		// 随所会把创建部分给跑一便，但之后就不能再创建了 这俩判断一定不能拼成一个
		if atomic.CompareAndSwapInt32(&p.state, locked, stopping) {
			// 进入stopping
			for range ticker.C {
				if curGo := atomic.LoadInt32(&p.currentPoolSize); curGo == 0 {
					close(p.queue)
					atomic.CompareAndSwapInt32(&p.state, stopping, dead)
					return nil
				}
			}
		}
		if atomic.CompareAndSwapInt32(&p.state, running, stopping) {
			// 进入stopping
			for range ticker.C {
				if curGo := atomic.LoadInt32(&p.currentPoolSize); curGo == 0 {
					close(p.queue)
					atomic.CompareAndSwapInt32(&p.state, stopping, dead)
					return nil
				}
			}
		}
		// 再次保险一次 减少失误的可能性 但也可能失误
		state := atomic.LoadInt32(&p.state)
		if state == running || state == locked {
			continue
		}
		// 这个基本不可能
		return fmt.Errorf("%w: %d", errShutDownState, state)
	}
}

func (p *ConsumerPool) RegisterConsumer(topic string, consumer Consumer) error {
	if _, ok := p.Group.Load(topic); ok {
		return fmt.Errorf("%w: %s", errTopicAlreadyRegisted, topic)
	}
	consumer = &ConsumerFuncWrapper{Consumer: consumer}
	p.Group.Store(topic, consumer)
	return nil
}

func (p *ConsumerPool) UpdateConsumer(topic string, consumer Consumer) {
	p.Group.Store(topic, consumer)
}
