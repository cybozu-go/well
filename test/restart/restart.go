// +build !windows

package main

import (
	"os"
	"syscall"
	"time"
)

func restart() {
	syscall.Kill(os.Getpid(), syscall.SIGHUP)
	time.Sleep(10 * time.Millisecond)
}
