package parallel

import "context"

// A Task is the main function of a service or component, or any other task.
//
// This is a core concept of this package. The simple signature of a task makes
// it possible to seamlessly combine code using this package with code that
// isn't aware of it.
//
// When ctx is closed, the function should finish as soon as possible and return
// ctx.Err().
//
// A task can also finish for any other reason, returning an error if there was
// a problem, or nil if it was a finite job that has completed successfully.
type Task func(ctx context.Context) error
