package parallel

import (
	"context"
	"fmt"
)

// SpawnFn is a function that starts a subtask in a goroutine.
//
// The task name is only for error messages. It is recommended that name is
// chosen uniquely within the parent task, but it's not enforced.
//
// The onExit mode specifies what happens if the subtask exits, see
// documentation for OnExit.
type SpawnFn func(name string, onExit OnExit, task Task)

// OnExit is an enumeration of exit handling modes. It specifies what should
// happen to the parent task if the subtask returns nil.
//
// Regardless of the chosen mode, if the subtask returns an error, it causes the
// parent task to shut down gracefully and return that error.
type OnExit int

const (
	// Continue means other subtasks of the parent task should continue to run.
	// Note that the parent task will return nil if its last remaining subtask
	// returns nil, even if Continue is specified.
	//
	// Use this mode for finite jobs that need to run once.
	Continue OnExit = iota

	// Exit means shut down the parent task gracefully.
	//
	// Use this mode for tasks that should be able to initiate graceful
	// shutdown, such as an HTTP server with a /quit endpoint that needs to
	// cause the process to exit.
	//
	// If any of other subtasks return an error, and it is not a (possibly
	// wrapped) context.Canceled, then the parent task will return the error.
	// Only first error from subtasks will be returned, the rest will be
	// discarded.
	//
	// If all other subtasks return nil or context.Canceled, the parent task
	// returns nil.
	Exit

	// Fail means shut down the parent task gracefully and return an error.
	//
	// Use this mode for subtasks that should never return unless their context
	// is closed.
	Fail
)

func (onExit OnExit) String() string {
	switch onExit {
	case Continue:
		return "Continue"
	case Exit:
		return "Exit"
	case Fail:
		return "Fail"
	default:
		return fmt.Sprintf("invalid OnExit mode: %d", onExit)
	}
}

// Run runs a task with several subtasks.
//
// The start function is the start-up sequence of the task. It receives a spawn
// function that can be used to launch subtasks. If the outer context closes, or
// any of the subtasks returns an error or panics, the inner context passed to
// start and to every task will also close, signalling all remaining goroutines
// to terminate.
//
// The start function should perform all necessary initialization and return.
// When it returns, it doesn't by itself cause the task context to close, and
// subtasks to exit.
//
// Run waits until the start function and all subtasks exit.
//
// If start returns an error, it becomes the return value of Run. Otherwise, Run
// returns the error or panic value from the first failed subtask.
//
// The subtasks can in turn be implemented using parallel.Run and have subtasks
// of their own.
//
// Example:
//
//  err := parallel.Run(ctx, func(ctx context.Context, spawn parallel.SpawnFn) error {
//      s1, err := service1.New(...)
//      if err != nil {
//          return err
//      }
//
//      s2, err := service2.New(...)
//      if err != nil {
//          return err
//      }
//
//      if err := s1.HeavyInit(ctx); err != nil {
//          return err
//      }
//
//      spawn("service1", parallel.Fail, s1.Run)
//      spawn("service2", parallel.Fail, s2.Run)
//      return nil
//  })
//
func Run(ctx context.Context, start func(ctx context.Context, spawn SpawnFn) error) error {
	g := NewGroup(ctx)

	if err := start(g.Context(), g.Spawn); err != nil {
		g.Exit(err)
	}

	return g.Wait()
}
