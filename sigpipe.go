// +build !windows

package cmd

import (
	"os"
	"os/signal"
	"syscall"
)

func ignoreSigPipe() {
	// signal.Ignore does NOT ignore signals; instead, it just stop
	// relaying signals to the channel.
	//signal.Ignore(syscall.SIGPIPE)

	// unbuffered channel will effectively ignore the signal
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGPIPE)
}
