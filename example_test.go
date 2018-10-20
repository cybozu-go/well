package well_test

import (
	"context"
	"flag"
	"net/http"

	"github.com/BurntSushi/toml"
	"github.com/cybozu-go/log"
	"github.com/cybozu-go/well"
)

func doSomething() error {
	return nil
}

// The most basic usage of the framework.
func Example_basic() {
	flag.Parse()
	well.LogConfig{}.Apply()

	well.Go(func(ctx context.Context) error {
		err := doSomething()

		if err != nil {
			// non-nil error will be passed to Cancel
			// by the framework.
			return err
		}

		// on success, nil should be returned.
		return nil
	})

	// some more Go
	//well.Go(func(ctx context.Context) error {})

	// Stop declares no Go calls will be made from this point.
	// Calling Stop is optional if Cancel is guaranteed to be called
	// at some point.
	well.Stop()

	// Wait waits for all goroutines started by Go to complete,
	// or one of such goroutine returns non-nil error.
	err := well.Wait()
	if err != nil {
		log.ErrorExit(err)
	}
}

// HTTP server that stops gracefully.
func Example_http() {
	flag.Parse() // must precedes LogConfig.Apply
	well.LogConfig{}.Apply()

	// log accesses in JSON Lines format.
	accessLog := log.NewLogger()
	accessLog.SetFormatter(log.JSONFormat{})

	// HTTP server.
	serv := &well.HTTPServer{
		Server: &http.Server{
			Handler: http.FileServer(http.Dir("/path/to/directory")),
		},
		AccessLog: accessLog,
	}

	// ListenAndServe is overridden to start a goroutine by well.Go.
	err := serv.ListenAndServe()
	if err != nil {
		log.ErrorExit(err)
	}

	// Wait waits for SIGINT or SIGTERM.
	// In this case, well.Stop can be omitted.
	err = well.Wait()

	// Use IsSignaled to determine err is the result of a signal.
	if err != nil && !well.IsSignaled(err) {
		log.ErrorExit(err)
	}
}

// Load logging configurations from TOML file.
func ExampleLogConfig() {
	// compile-time defaults
	config := &well.LogConfig{
		Level:  "error",
		Format: "json",
	}

	// load from TOML
	_, err := toml.DecodeFile("/path/to/config.toml", config)
	if err != nil {
		log.ErrorExit(err)
	}

	// Apply gives priority to command-line flags, if any.
	flag.Parse()
	err = config.Apply()
	if err != nil {
		log.ErrorExit(err)
	}
}

// Barrier wait for gorutines.
func ExampleNewEnvironment() {
	// An independent environment.
	env := well.NewEnvironment(context.Background())

	env.Go(func(ctx context.Context) error {
		// do something
		return nil
	})
	// some more env.Go()

	// wait all goroutines started by env.Go().
	// These goroutines may also be canceled when the global
	// environment is canceled.
	env.Stop()
	err := env.Wait()
	if err != nil {
		log.ErrorExit(err)
	}

	// another environment for another barrier.
	env = well.NewEnvironment(context.Background())

	// env.Go, env.Stop, and env.Wait again.
}
