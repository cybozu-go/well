// +build windows

package cmd

import "net"

func isMaster() bool {
	return true
}

// SystemdListeners returns (nil, nil) on Windows.
func SystemdListeners() ([]net.Listener, error) {
	return nil, nil
}

// Run, on Windows, simply calls g.Listen then g.Serve.
func (g *Graceful) Run() {
	env := g.Env
	if env == nil {
		env = defaultEnv
	}

	// prepare listener files
	listeners, err := g.Listen()
	if err != nil {
		env.Cancel(err)
		return
	}
	g.Serve(listeners)
}
