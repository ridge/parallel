package parallel

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunNoSubtasksSuccess(t *testing.T) {
	ctx := context.Background()
	err := Run(ctx, func(ctx context.Context, spawn SpawnFn) error {
		return nil
	})
	require.NoError(t, err)
}

func TestRunNoSubtasksError(t *testing.T) {
	ctx := context.Background()
	err := Run(ctx, func(ctx context.Context, spawn SpawnFn) error {
		return errors.New("oops")
	})
	require.EqualError(t, err, "oops")
}

func TestRunSubtaskExit(t *testing.T) {
	ctx := context.Background()
	seq := make(chan int)
	step1 := make(chan struct{})
	step2 := make(chan struct{})
	var err error
	go func() {
		err = Run(ctx, func(ctx context.Context, spawn SpawnFn) error {
			spawn("exit1", Exit, func(ctx context.Context) error {
				<-step1
				seq <- 2
				return nil
			})
			spawn("exit2", Exit, func(ctx context.Context) error {
				seq <- 1
				<-step2
				seq <- 3
				return nil
			})
			return nil
		})
		seq <- 4
	}()
	require.Equal(t, 1, <-seq)
	close(step1)
	require.Equal(t, 2, <-seq)
	close(step2)
	require.Equal(t, 3, <-seq)
	require.Equal(t, 4, <-seq)
	require.NoError(t, err)
}

func TestRunSubtaskContinue(t *testing.T) {
	ctx := context.Background()
	seq := make(chan int)
	step1 := make(chan struct{})
	step2 := make(chan struct{})
	var err error
	go func() {
		err = Run(ctx, func(ctx context.Context, spawn SpawnFn) error {
			spawn("continue1", Continue, func(ctx context.Context) error {
				<-step1
				seq <- 2
				return nil
			})
			spawn("continue2", Continue, func(ctx context.Context) error {
				seq <- 1
				<-step2
				seq <- 3
				return nil
			})
			return nil
		})
		seq <- 4
	}()
	require.Equal(t, 1, <-seq)
	close(step1)
	require.Equal(t, 2, <-seq)
	close(step2)
	require.Equal(t, 3, <-seq)
	require.Equal(t, 4, <-seq)
	require.NoError(t, err)
}

// Fail is the actual enum for handling mode, so it should be present
func TestRunSubtaskFail(t *testing.T) {
	ctx := context.Background()
	seq := make(chan int)
	step1 := make(chan struct{})
	step2 := make(chan struct{})
	var err error
	go func() {
		err = Run(ctx, func(ctx context.Context, spawn SpawnFn) error {
			spawn("fail1", Fail, func(ctx context.Context) error {
				<-step1
				seq <- 2
				return nil
			})
			spawn("fail2", Fail, func(ctx context.Context) error {
				seq <- 1
				<-step2
				seq <- 3
				<-ctx.Done()
				seq <- 4
				return nil
			})
			return nil
		})
		seq <- 5
	}()
	require.Equal(t, 1, <-seq)
	close(step1)
	require.Equal(t, 2, <-seq)
	close(step2)
	require.Equal(t, 3, <-seq)
	require.Equal(t, 4, <-seq)
	require.Equal(t, 5, <-seq)
	require.EqualError(t, err, "task fail1 terminated unexpectedly")
}

func TestRunSubtaskError(t *testing.T) {
	ctx := context.Background()
	seq := make(chan int)
	step1 := make(chan struct{})
	step2 := make(chan struct{})
	var err error
	go func() {
		err = Run(ctx, func(ctx context.Context, spawn SpawnFn) error {
			spawn("error1", Exit, func(ctx context.Context) error {
				<-step1
				seq <- 2
				return errors.New("oops1")
			})
			spawn("error2", Exit, func(ctx context.Context) error {
				seq <- 1
				<-step2
				seq <- 3
				<-ctx.Done()
				seq <- 4
				return errors.New("oops2")
			})
			return nil
		})
		seq <- 5
	}()
	require.Equal(t, 1, <-seq)
	close(step1)
	require.Equal(t, 2, <-seq)
	close(step2)
	require.Equal(t, 3, <-seq)
	require.Equal(t, 4, <-seq)
	require.Equal(t, 5, <-seq)
	require.EqualError(t, err, "oops1")
}

func TestRunSubtaskInitError(t *testing.T) {
	ctx := context.Background()
	seq := make(chan int)
	step1 := make(chan struct{})
	step2 := make(chan struct{})
	var err error
	go func() {
		err = Run(ctx, func(ctx context.Context, spawn SpawnFn) error {
			spawn("error", Exit, func(ctx context.Context) error {
				<-step1
				seq <- 1
				return errors.New("oops1")
			})
			<-step2
			seq <- 2
			<-ctx.Done()
			seq <- 3
			return errors.New("oops2")
		})
		seq <- 4
	}()
	close(step1)
	require.Equal(t, 1, <-seq)
	close(step2)
	require.Equal(t, 2, <-seq)
	require.Equal(t, 3, <-seq)
	require.Equal(t, 4, <-seq)
	require.EqualError(t, err, "oops1")
}

func TestRunShutdownNotOK(t *testing.T) {
	ctx := context.Background()
	seq := make(chan int)
	step1 := make(chan struct{})
	step2 := make(chan struct{})
	var err error
	go func() {
		err = Run(ctx, func(ctx context.Context, spawn SpawnFn) error {
			spawn("exitSucc", Exit, func(ctx context.Context) error {
				<-step1
				seq <- 2
				<-step2
				return nil
			})
			spawn("shutdownFail", Exit, func(ctx context.Context) error {
				seq <- 1
				<-ctx.Done()
				return errors.New("failed shutdown")
			})
			return nil
		})
		seq <- 3
	}()
	require.Equal(t, 1, <-seq)
	close(step1)
	require.Equal(t, 2, <-seq)
	close(step2)
	require.Equal(t, 3, <-seq)
	require.Equal(t, err, errors.New("failed shutdown"))
}

func TestRunShutdownCancel(t *testing.T) {
	ctx := context.Background()
	seq := make(chan int)
	step1 := make(chan struct{})
	step2 := make(chan struct{})
	var err error
	go func() {
		err = Run(ctx, func(ctx context.Context, spawn SpawnFn) error {
			spawn("exitSucc", Exit, func(ctx context.Context) error {
				<-step1
				seq <- 2
				<-step2
				return nil
			})
			spawn("shutdownFail", Exit, func(ctx context.Context) error {
				seq <- 1
				<-ctx.Done()
				return ctx.Err()
			})
			return nil
		})
		seq <- 3
	}()
	require.Equal(t, 1, <-seq)
	close(step1)
	require.Equal(t, 2, <-seq)
	close(step2)
	require.Equal(t, 3, <-seq)
	require.NoError(t, err)
}

func TestRunCancel(t *testing.T) {
	ctx := context.Background()
	seq := make(chan int)
	var err error
	go func() {
		err = Run(ctx, func(ctx context.Context, spawn SpawnFn) error {
			spawn("exitWithCancel", Exit, func(ctx context.Context) error {
				ctx, cancel := context.WithCancel(ctx)
				cancel()
				return ctx.Err()
			})
			return nil
		})
		seq <- 1
	}()
	require.Equal(t, 1, <-seq)
	require.Equal(t, err, context.Canceled)
}

// Fail is the actual way for handling the tasks, so it should be present
func TestExitFailTaskOnCancel(t *testing.T) {
	ctx := context.Background()
	seq := make(chan int)
	var err error
	go func() {
		err = Run(ctx, func(ctx context.Context, spawn SpawnFn) error {
			spawn("daemon", Fail, func(ctx context.Context) error {
				<-ctx.Done()
				seq <- 2
				return nil
			})
			spawn("shutdown", Exit, func(ctx context.Context) error {
				seq <- 1
				return nil
			})
			return nil
		})
		seq <- 3
	}()
	require.Equal(t, 1, <-seq)
	require.Equal(t, 2, <-seq)
	require.Equal(t, 3, <-seq)
	require.NoError(t, err)
}
