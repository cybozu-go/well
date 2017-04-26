// +build !windows

package cmd

import (
	"os"
	"os/signal"
	"syscall"
)

const (
	sigPipeExit = 2
)

// handleSigPipe catches SIGPIPE and exit abnormally with status code 2.
//
// Background:
//
// systemd interprets programs exited with SIGPIPE as
// "successfully exited" because its default SuccessExitStatus=
// includes SIGPIPE.
// https://www.freedesktop.org/software/systemd/man/systemd.service.html
//
// Normal Go programs ignore SIGPIPE for file descriptors other than
// stdout(1) and stderr(2).  If SIGPIPE is raised from stdout or stderr,
// Go programs exit with a SIGPIPE signal.
// https://golang.org/pkg/os/signal/#hdr-SIGPIPE
//
// journald is a service tightly integrated in systemd.  Go programs
// running as a systemd service will normally connect its stdout and
// stderr pipes to journald.  Unfortunately though, journald often
// dies and restarts due to bugs, and once it restarts, programs
// whose stdout and stderr was connected to journald will receive
// SIGPIPE at their next writes to stdout or stderr.
//
// Combined these specifications and problems all together, Go programs
// running as systemd services often die with SIGPIPE, but systemd will
// not restarts them as they "successfully exited" except when the service
// is configured with SuccessExitStatus= without SIGPIPE or Restart=always.
//
// Handling SIGPIPE manually and exiting with abnormal status code 2
// can work around the problem.
func handleSigPipe() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGPIPE)
	go func() {
		<-c
		os.Exit(sigPipeExit)
	}()
}
