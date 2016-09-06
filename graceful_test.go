// +build !windows

package cmd

import "testing"

func TestGraceful(t *testing.T) {
	t.Skip(`Graceful cannot be tested by go test as it executes
itself in another process of go test.
Instead, we test Graceful in a test program under "test/graceful".`)
}
