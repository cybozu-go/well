package cmd

import "context"

var (
	defaultEnv *Environment
)

func init() {
	defaultEnv = NewEnvironment()
}

// Context returns the base context of the global environment.
func Context() context.Context {
	return defaultEnv.Context()
}

// Stop cancels the base context of the global environment.
//
// Passed err will be returned by Wait().
// Once stopped, Go() will not start new goroutines.
//
// This returns true if the caller is the first that calls Stop.
// For second and later calls, Stop does nothing and returns false.
func Stop(err error) bool {
	return defaultEnv.Stop(err)
}

// Wait waits for Stop being called.
//
// The returned err is the one passed to Stop.
// err can be tested by IsSignaled to determine whether the
// program got SIGINT or SIGTERM.
func Wait() error {
	return defaultEnv.Wait()
}

// Go starts a goroutine that executes f in the global environment.
//
// Goroutines started by this function will be waited for by
// Wait until all such goroutines return.
//
// If f returns non-nil error, Stop is called with that error.
//
// f should watch ctx.Done() channel and return quickly when the
// channel is canceled.
func Go(f func(ctx context.Context) error) {
	defaultEnv.Go(f)
}
