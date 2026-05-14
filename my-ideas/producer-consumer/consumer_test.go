package producerconsumer

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// 其实就是在支持context的同时过程函数还在时不时检查一下是否已经超时
// 也就有可能消费成功但消费超时了，其实哈这里的 maxProcessTime 没啥用
// 因为必须死等它消费结果 不能直接因其消费太慢而直接返回超时
func TestErrFuncCancel(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), time.Millisecond*500)
	err := MastFunc(ctx)
	fmt.Println(err)
	time.Sleep(time.Second)
}

func MastFunc(ctx context.Context) error {
	chanErr := make(chan error, 1)
	go errFunc(ctx, chanErr)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-chanErr:
		return err
	}
}

// 我除非在这个可能超时的func里时不时看一下是否超时才能提前终止
func errFunc(ctx context.Context, chanErr chan error) {
	time.Sleep(time.Second)
	if ctx.Err() != nil {
		return
	}

	fmt.Println("still running")
	chanErr <- errors.New("time came")
}

// func TestNewPool(t *testing.T) {
// 	testCases := struct{
// 		name string
// 		wantErr error
// 	}{}
// 	pool, err := NewConsumerPool(128, 4, 8, 1)
// 	if err != nil {
// 		return
// 	}

// }

func TestChanBase(t *testing.T) {
	ch := make(chan struct{}, 1)
	close(ch)
	select {
	case ch <- struct{}{}:
		fmt.Println("what?")
	default:
		fmt.Println("chan closed")
	}
}

// 判断时
// if errors.Is(err, &PanicError{}) {
//     data.State = consumePanic
// }

func TestGOGOGO(t *testing.T) {
	topic := "GO"
	pool, _ := NewConsumerPool(128, 1, 2, 1)
	pool.RegisterConsumer(topic, ConsumerFunc(func(ctx context.Context) error { return nil }))
	err := pool.Start()
	if err != nil {
		fmt.Println(err)
		return
	}
	var count int32
	start := time.Now()
	for i := 0; i < 1; i++ {
		go func() {
			for {
				if err := pool.Submit(topic, nil, func(consumeState int32, err error) { atomic.AddInt32(&count, 1) }); err != nil {
					if err == errQueueFull || err == errStateLocked {
						continue
					}
				}
			}
		}()
	}
	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		fmt.Println("当前运行了: ", count, "--用了:1秒")
		fmt.Println(len(pool.queue))
		if count == 10_000 {
			break
		}
	}
	fmt.Printf("总共用了: %v", time.Since(start))

}

func TestNewConsumerPool(t *testing.T) {
	testCases := []struct {
		name    string
		wantErr error
		fun     func() error
	}{
		{
			name:    "queueSize <= 0",
			wantErr: errQueueSizeZero,
			fun: func() error {
				_, err := NewConsumerPool(0, 0, 0, 0)
				return err
			},
		},
		{
			name:    "corGo <= 0 || maxGo <= 0",
			wantErr: errGoRoutine,
			fun: func() error {
				_, err := NewConsumerPool(1, 0, 0, 0)
				return err
			},
		},
		{
			name:    "maxGo < corGo",
			wantErr: nil,
			fun: func() error {
				_, err := NewConsumerPool(1, 2, 1, 1)
				return err
			},
		},
		{
			name:    "normal",
			wantErr: nil,
			fun: func() error {
				_, err := NewConsumerPool(1, 1, 1, 1)
				return err
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.fun()
			// assert ...
			fmt.Println(err)
		})
	}
}

func TestStart(t *testing.T) {
	testCases := []struct {
		name    string
		wantErr error
		getPool func() *ConsumerPool
	}{
		{
			name:    "errPoolStateIsntCreatred",
			wantErr: errPoolStateIsntCreatred,
			getPool: func() *ConsumerPool {
				p, _ := NewConsumerPool(1, 1, 1, 1)
				p.Start()
				return p
			},
		},
		{
			name:    "errPoolStateIsntCreatred",
			wantErr: errPoolStateIsntCreatred,
			getPool: func() *ConsumerPool {
				p, _ := NewConsumerPool(1, 1, 1, 1)
				p.state = dead
				return p
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := tc.getPool()
			err := p.Start()
			assert.Equal(t, tc.wantErr, err)
			// fmt.Println(err)
		})
	}
}

func TestSubmit(t *testing.T) {
	topic := "zxc"
	consumer := ConsumerFunc(func(ctx context.Context) error { return nil })
	testCases := []struct {
		name       string
		wantErr    error
		updatePool func(p *ConsumerPool) *ConsumerPool
		getAttrs   func() (string, interface{}, func(consumeState int32, err error))
	}{
		{
			name:    "errTopicNil",
			wantErr: errTopicNil,
			getAttrs: func() (string, interface{}, func(consumeState int32, err error)) {
				return "", nil, nil
			},
		},
		{
			name:    "errCallBackIsNil",
			wantErr: errCallBackIsNil,
			getAttrs: func() (string, interface{}, func(consumeState int32, err error)) {
				return topic, nil, nil
			},
		},
		{
			name:    "errTopicIsNotFound",
			wantErr: errTopicIsNotFound,
			getAttrs: func() (string, interface{}, func(consumeState int32, err error)) {
				return topic, nil, func(consumeState int32, err error) {}
			},
		},
		{
			name:    "state == stopping",
			wantErr: fmt.Errorf("%w: %d", errSubmitFailed, stopping),
			updatePool: func(p *ConsumerPool) *ConsumerPool {
				p.RegisterConsumer(topic, consumer)
				p.state = stopping
				return p
			},
			getAttrs: func() (string, interface{}, func(consumeState int32, err error)) {
				return topic, nil, func(consumeState int32, err error) {}
			},
		},
		{
			name:    "state == dead",
			wantErr: fmt.Errorf("%w: %d", errSubmitFailed, dead),
			updatePool: func(p *ConsumerPool) *ConsumerPool {
				p.RegisterConsumer(topic, consumer)
				p.ShutDown()
				return p
			},
			getAttrs: func() (string, interface{}, func(consumeState int32, err error)) {
				return topic, nil, func(consumeState int32, err error) {}
			},
		},
		{
			name:    "errStateLocked",
			wantErr: errStateLocked,
			updatePool: func(p *ConsumerPool) *ConsumerPool {
				p.RegisterConsumer(topic, consumer)
				p.state = locked
				return p
			},
			getAttrs: func() (string, interface{}, func(consumeState int32, err error)) {
				return topic, nil, func(consumeState int32, err error) {}
			},
		},
		{
			name:    "errQueueFull",
			wantErr: errQueueFull,
			updatePool: func(p *ConsumerPool) *ConsumerPool {
				p.RegisterConsumer(topic, consumer)
				p.Submit(topic, nil, func(consumeState int32, err error) {
					time.Sleep(time.Second)
				})
				// 占一下
				p.Submit(topic, nil, func(consumeState int32, err error) {})
				return p
			},
			getAttrs: func() (string, interface{}, func(consumeState int32, err error)) {
				return topic, nil, func(consumeState int32, err error) {}
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var p *ConsumerPool
			p, _ = NewConsumerPool(1, 1, 1, 1)
			p.Start()
			if tc.updatePool != nil {
				p = tc.updatePool(p)
			}
			err := p.Submit(tc.getAttrs())
			assert.Equal(t, tc.wantErr, err)
			// fmt.Println(err)
		})
	}
}

func TestChan(t *testing.T) {
	ch := make(chan struct{}, 1)
	fmt.Println("len: ", len(ch), "; cap: ", cap(ch))

	select {
	case ch <- struct{}{}:
		fmt.Println("加了一个")
		fmt.Println("len: ", len(ch), "; cap: ", cap(ch))
	default:
		fmt.Println("len: ", len(ch), "; cap: ", cap(ch))
	}

	select {
	case ch <- struct{}{}:
		fmt.Println("加了一个")
		fmt.Println("len: ", len(ch), "; cap: ", cap(ch))
	default:
		fmt.Println("len: ", len(ch), "; cap: ", cap(ch))
	}
}

func TestShutDown(t *testing.T) {
	// ctx,cancel := context.WithCancel(context.Background())
	testCases := []struct {
		name    string
		wantErr error
		getPool func() *ConsumerPool
	}{
		{
			name:    "errWrongState-stopping",
			wantErr: fmt.Errorf("%w: %d", errWrongState, stopping),
			getPool: func() *ConsumerPool {
				p, _ := NewConsumerPool(1, 1, 1, 1)
				p.state = stopping
				return p
			},
		},
		{
			name:    "errWrongState-dead",
			wantErr: fmt.Errorf("%w: %d", errWrongState, dead),
			getPool: func() *ConsumerPool {
				p, _ := NewConsumerPool(1, 1, 1, 1)
				p.Start()
				p.ShutDown()
				return p
			},
		},
		{
			name:    "errWrongState-created",
			wantErr: fmt.Errorf("%w: %d", errWrongState, created),
			getPool: func() *ConsumerPool {
				p, _ := NewConsumerPool(1, 1, 1, 1)
				return p
			},
		},
		{
			name: "normal",
			getPool: func() *ConsumerPool {
				p, _ := NewConsumerPool(1, 1, 1, 1)
				p.Start()
				return p
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := tc.getPool()
			err := p.ShutDown()
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
