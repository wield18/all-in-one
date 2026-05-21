// Package types 底层类型
package types

import "fmt"

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
