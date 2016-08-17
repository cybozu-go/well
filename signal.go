package cmd

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/cybozu-go/log"
)

var (
	errSignaled = errors.New("signaled")
)

// IsSignaled returns true if err indicates that the program has
// received SIGINT or SIGTERM.
func IsSignaled(err error) bool {
	return err == errSignaled
}

func handleSignal(ctx context.Context, e *Environment) {
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case <-ctx.Done():
			return
		case s := <-ch:
			if !e.Stop(errSignaled) {
				return
			}
			log.Warn("cmd: got signal", map[string]interface{}{
				"signal": s.String(),
			})
		}
	}()
}
