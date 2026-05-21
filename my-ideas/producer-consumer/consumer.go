// Package producerconsumer 就一个本地异步消费的东西
package producerconsumer

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wield18/all-in-one/config"
	consumerFunc "github.com/wield18/all-in-one/config/consumer"
	"github.com/wield18/all-in-one/my-ideas/producer-consumer/consumer"
	"github.com/wield18/all-in-one/my-ideas/producer-consumer/types"
)

var poolServer = &PoolServer{}

type PoolServer struct {
	Pool *ConsumerPool
}

// 暴露出生化方法
func InitPool(conf *config.Pool) *PoolServer {

	pool, err := NewConsumerPool(conf.QueueSize, conf.CorGo, conf.MaxGo, conf.MaxIdleGoTime)
	if err != nil {
		panic(err.Error())
	}
	// topic 跟对应的 consumeFunc 得注册
	aMap := consumerFunc.GetConsumerMap()
	for topic, fun := range aMap {
		err := pool.RegisterConsumer(topic, fun)
		fmt.Printf("topic: %s init success\n", topic)
		if err != nil {
			panic(err.Error())
		}
	}
	pool.Start()
	poolServer.Pool = pool
	return poolServer
}

// 引用获得
func GetPoolServer() *PoolServer {
	return poolServer
}

// 真正运行时的各种状态
var (
	created  int32 = 0
	running  int32 = 1
	stopping int32 = 2
	dead     int32 = 3
	locked   int32 = 4
)

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
			var fun consumer.Consumer
			oneConsuemr, ok := p.Group.Load(data.Topic)
			if !ok {
				fmt.Printf("%v: %s", errTopicIsNotFound, data.Topic)
				continue
			}
			fun = oneConsuemr.(consumer.Consumer)
			ctx := context.WithValue(p.ctx, consumer.DATA, data.Data)
			err := fun.Consume(ctx)
			if err != nil {
				if errors.Is(err, &types.PanicError{}) {
					data.State = consumer.ConsumePanic
				} else {
					data.State = consumer.ConsumeError
				}
			} else {
				data.State = consumer.ConsumeSuccess
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

// 阻塞发送,所有data的传递必须是指针传递,当然值传递也无妨,只要对应的解析consumer也是值解构,我自己规定指针传递
func (p *ConsumerPool) BlockSubmit(topic string, data interface{},
	callBack func(consumeState int32, err error)) error {
	for {
		err := p.Submit(topic, data, callBack)
		if err == nil {
			return nil
		} else if err == errQueueFull || err == errStateLocked { // 可重试的
			continue
		} else {
			return err
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
		// p.Group.Range(func(key, value interface{}) bool {
		// 	fmt.Println("看这里-------------------")
		// 	fmt.Printf("key: %v\n", key)
		// 	// 返回 true 继续遍历，返回 false 停止遍历
		// 	return true
		// })
		return fmt.Errorf("%w: %s", errTopicIsNotFound, topic)
	}

	state := atomic.LoadInt32(&p.state)
	if state == stopping || state == dead {
		return fmt.Errorf("%w: %d", errSubmitFailed, state)
	}
	// 能到这的有 locked created running
	da := &SubmitData{
		Topic:    topic,
		Data:     data,
		State:    consumer.ConsumeCreated,
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

func (p *ConsumerPool) RegisterConsumer(topic string, oneConsumer consumer.Consumer) error {
	// 先查一遍
	if _, ok := p.Group.Load(topic); ok {
		return fmt.Errorf("%w: %s", errTopicAlreadyRegisted, topic)
	}
	oneConsumer = &consumer.ConsumerFuncWrapper{Consumer: oneConsumer}
	p.Group.Store(topic, oneConsumer)
	p.Group.Range(func(key, value interface{}) bool {
		fmt.Println("看这里-------------------")
		fmt.Printf("key: %v\n", key)
		// 返回 true 继续遍历，返回 false 停止遍历
		return true
	})
	return nil
}

func (p *ConsumerPool) UpdateConsumer(topic string, oneConsumer consumer.Consumer) {
	oneConsumer = &consumer.ConsumerFuncWrapper{Consumer: oneConsumer}
	p.Group.Store(topic, oneConsumer)
}
