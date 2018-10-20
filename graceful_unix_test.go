// +build !windows

package well

import (
	"net"
	"runtime"
	"testing"
)

func TestListenerFiles(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows doesn't support FileListener")
	}
	t.Parallel()

	ln, err := net.Listen("tcp", "localhost:18555")
	if err != nil {
		t.Skip(err)
	}
	defer ln.Close()

	fl, err := listenerFiles([]net.Listener{ln})
	if err != nil {
		t.Error(err)
	}
	if len(fl) != 1 {
		t.Error(`len(fl) != 1`)
	}
}
