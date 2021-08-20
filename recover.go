package parallel

import (
	"context"
	"fmt"
	"runtime/debug"
)

// ErrPanic is the error type that occurs when a subtask panics
type ErrPanic struct {
	Value interface{}
	Stack []byte
}

func (err ErrPanic) Error() string {
	return fmt.Sprintf("panic: %s", err.Value)
}

// Unwrap returns the error passed to panic, or nil if panic was called with
// something other than an error
func (err ErrPanic) Unwrap() error {
	if e, ok := err.Value.(error); ok {
		return e
	}
	return nil
}

// RunTask executes the task in the current goroutine, recovering from panics.
// A panic is logged, reported to monitoring and returned as ErrPanic.
func RunTask(ctx context.Context, task Task) (err error) {
	defer func() {
		if p := recover(); p != nil {
			panicErr := ErrPanic{Value: p, Stack: debug.Stack()}
			err = panicErr
		}
	}()
	return task(ctx)
}
