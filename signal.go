package cmd

import (
	"context"
	"errors"
	"os"
	"os/signal"

	"github.com/cybozu-go/log"
)

var (
	errSignaled = errors.New("signaled")
)

// IsSignaled returns true if err returned by Wait indicates that
// the program has received SIGINT or SIGTERM.
func IsSignaled(err error) bool {
	return err == errSignaled
}

// handleSignal should be called by Go.
func handleSignal(ctx context.Context) error {
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, stopSignals...)

	select {
	case <-ctx.Done():
		return nil
	case s := <-ch:
		log.Warn("cmd: got signal", map[string]interface{}{
			"signal": s.String(),
		})
		return errSignaled
	}
}
