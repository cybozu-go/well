package cmd

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/cybozu-go/log"
)

// Server is a generic network server that accepts connections
// and invokes Handler in a goroutine for each connection.
//
// In addition, Serve method gracefully waits all its goroutines to
// complete before returning.
type Server struct {

	// Handler handles a connection.  This must not be nil.
	//
	// ctx is a derived context from the base context that will be
	// canceled when Handler returns.
	Handler func(ctx context.Context, conn net.Conn)

	// ShutdownTimeout is the timeout duration before the managed
	// goroutine started by Serve returns leaving active connections.
	//
	// Zero duration disables timeout.
	ShutdownTimeout time.Duration

	// Env is the environment where this server runs.
	//
	// The global environment is used if Env is nil.
	Env *Environment

	wg sync.WaitGroup
}

// Serve starts a managed goroutine to accept connections.
//
// Serve itself returns immediately.  The goroutine continues
// to accept and handle connections until the base context is
// canceled.
func (s *Server) Serve(l net.Listener) {
	env := s.Env
	if env == nil {
		env = defaultEnv
	}

	go func() {
		<-env.ctx.Done()
		l.Close()
	}()

	env.Go(func(ctx context.Context) error {
		for {
			conn, err := l.Accept()
			if err != nil {
				log.Debug("Listener.Accept error", map[string]interface{}{
					"addr":  l.Addr().String(),
					"error": err.Error(),
				})
				goto OUT
			}
			s.wg.Add(1)
			go func() {
				ctx, cancel := context.WithCancel(ctx)
				defer cancel()
				s.Handler(ctx, conn)
				s.wg.Done()
			}()
		}
	OUT:
		s.wait()
		return nil
	})
}

func (s *Server) wait() {
	ch := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(ch)
	}()

	if s.ShutdownTimeout == 0 {
		<-ch
		return
	}

	select {
	case <-ch:
	case <-time.After(s.ShutdownTimeout):
	}
}
