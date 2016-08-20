package cmd

import (
	"context"
	"sync"

	"github.com/cybozu-go/log"
)

// Environment implements context-based goroutine management.
type Environment struct {
	ctx    context.Context
	cancel context.CancelFunc
	stopCh chan struct{}
	wg     sync.WaitGroup

	mu      sync.RWMutex
	stopped bool
	err     error
}

// NewEnvironment creates a new Environment.
func NewEnvironment() *Environment {
	ctx, cancel := context.WithCancel(context.Background())
	e := &Environment{
		ctx:    ctx,
		cancel: cancel,
		stopCh: make(chan struct{}),
	}
	handleSignal(ctx, e)
	return e
}

// Context returns the base context of the environment.
func (e *Environment) Context() context.Context {
	return e.ctx
}

func (e *Environment) stop(err error) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.stopped {
		return false
	}

	e.stopped = true
	e.err = err
	close(e.stopCh) // unleash Wait()
	return true
}

// StopNoCancel just declares no further Go is called.
//
// This returns true if the caller is the first that calls Stop
// or StopNoCancel.
func (e *Environment) StopNoCancel() bool {
	return e.stop(nil)
}

// Stop cancels the base context.
//
// Passed err will be returned by Wait().
// Once stopped, Go() will not start new goroutines.
//
// This returns true if the caller is the first that calls Stop
// or StopNoCancel.  For second and later calls, Stop does nothing
// and returns false.
func (e *Environment) Stop(err error) bool {
	e.cancel()
	return e.stop(err)
}

// Wait waits for Stop or StopNoCancel being called.
//
// The returned err is the one passed to Stop.
// err can be tested by IsSignaled to determine whether the
// program got SIGINT or SIGTERM.
func (e *Environment) Wait() error {
	<-e.stopCh
	log.Info("cmd: waiting for all goroutines to complete", nil)
	e.wg.Wait()
	e.cancel() // in case for StopNoCancel

	e.mu.Lock()
	defer e.mu.Unlock()

	return e.err
}

// Go starts a goroutine that executes f.
//
// f takes a drived context from the base context.  The context
// will be canceled when f returns.
//
// Goroutines started by this function will be waited for by
// Wait until all such goroutines return.
//
// If f returns non-nil error, Stop is called immediately
// with that error.
//
// f should watch ctx.Done() channel and return quickly when the
// channel is closed.
func (e *Environment) Go(f func(ctx context.Context) error) {
	e.mu.RLock()
	if e.stopped {
		e.mu.RUnlock()
		return
	}
	e.wg.Add(1)
	e.mu.RUnlock()

	go func() {
		ctx, cancel := context.WithCancel(e.ctx)
		defer cancel()
		err := f(ctx)
		if err != nil {
			e.Stop(err)
		}
		e.wg.Done()
	}()
}
