//go:build windows
// +build windows

package main

import "time"

func restart() {
	time.Sleep(10 * time.Millisecond)
}
