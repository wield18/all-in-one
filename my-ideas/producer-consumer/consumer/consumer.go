// Package consumer 存放ConsumerFunc的 structs and interfaces
package consumer

import (
	"context"
	"errors"
	"runtime"

	types "github.com/wield18/all-in-one/my-ideas/producer-consumer/types"
)

var (
	buffLen  = 2048
	errPanic = errors.New("Panic")
)

const (
	DATA = "data"
)

// 具体所有consume的状态
var (
	ConsumeCreated int32 = 0
	ConsumeSuccess int32 = 1
	ConsumeError   int32 = 2
	ConsumePanic   int32 = 3
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
			err = &types.PanicError{Value: errPanic, Stack: string(buf)}
		}
	}()
	return c.Consumer.Consume(ctx)
}
