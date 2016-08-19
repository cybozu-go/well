package cmd

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cybozu-go/log"
)

const (
	defaultHTTPReadTimeout = 30 * time.Second

	// our request tracking header.
	uuidHeaderName = "X-Cybozu-Request-ID"
)

// HTTPServer is a wrapper for http.Server.
//
// This struct overrides Serve and ListenAndServe* methods.
//
// http.Server members are replaced as following:
//    - Handler is replaced with a wrapper handler.
//    - ReadTimeout is set to 30 seconds if it is zero.
//    - ConnState is replaced with the one provided by the framework.
type HTTPServer struct {
	*http.Server

	// AccessLog is a logger for access logs.
	// If this is nil, the default logger is used.
	AccessLog *log.Logger

	// ShutdownTimeout is the maximum duration the server waits for
	// all connections to be closed before shutdown.
	//
	// Zero duration disables timeout.
	ShutdownTimeout time.Duration

	// Env is the environment where this server runs.
	//
	// The global environment is used if Env is nil.
	Env *Environment

	handler  http.Handler
	wg       sync.WaitGroup
	timedout int32

	mu        sync.Mutex
	idleConns map[net.Conn]struct{}

	initOnce sync.Once
}

type writerString interface {
	WriteString(data string) (int, error)
}

// StdResponseWriter is the interface implemented by
// the ResponseWriter from http.Server.
//
// HTTPServer's ResponseWriter implements this as well.
type StdResponseWriter interface {
	http.ResponseWriter
	io.ReaderFrom
	http.Flusher
	http.CloseNotifier
	http.Hijacker
	writerString
}

type logResponseWriter struct {
	StdResponseWriter
	status int
	size   int64
}

func (w *logResponseWriter) WriteHeader(status int) {
	w.status = status
	w.StdResponseWriter.WriteHeader(status)
}

func (w *logResponseWriter) Write(data []byte) (int, error) {
	n, err := w.StdResponseWriter.Write(data)
	w.size += int64(n)
	return n, err
}

// ServeHTTP implements http.Handler interface.
func (s *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	lw := &logResponseWriter{w.(StdResponseWriter), http.StatusOK, 0}
	ctx, cancel := context.WithCancel(s.Env.ctx)
	defer cancel()
	s.handler.ServeHTTP(lw, r.WithContext(ctx))

	fields := map[string]interface{}{
		log.FnType:           "access",
		log.FnResponseTime:   time.Since(startTime).Seconds(),
		log.FnProtocol:       r.Proto,
		log.FnHTTPStatusCode: lw.status,
		log.FnHTTPMethod:     r.Method,
		log.FnURL:            r.RequestURI,
		log.FnHTTPHost:       r.Host,
		log.FnRequestSize:    r.ContentLength,
		log.FnResponseSize:   lw.size,
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		fields[log.FnRemoteAddress] = ip
	}
	ua := r.Header.Get("User-Agent")
	if len(ua) > 0 {
		fields[log.FnHTTPUserAgent] = ua
	}
	reqid := r.Header.Get(uuidHeaderName)
	if len(reqid) > 0 {
		fields[log.FnRequestID] = reqid
	}

	lv := log.LvInfo
	switch {
	case 500 <= lw.status:
		lv = log.LvError
	case 400 <= lw.status:
		lv = log.LvWarn
	}
	s.AccessLog.Log(lv, "cmd: "+http.StatusText(lw.status), fields)
}

func (s *HTTPServer) init() {
	if s.handler != nil {
		return
	}

	s.idleConns = make(map[net.Conn]struct{}, 100000)

	if s.Server.Handler == nil {
		panic("Handler must not be nil")
	}
	s.handler = s.Server.Handler
	s.Server.Handler = s
	if s.Server.ReadTimeout == 0 {
		s.Server.ReadTimeout = defaultHTTPReadTimeout
	}
	s.Server.ConnState = func(c net.Conn, state http.ConnState) {
		s.mu.Lock()
		if state == http.StateIdle {
			s.idleConns[c] = struct{}{}
		} else {
			delete(s.idleConns, c)
		}
		s.mu.Unlock()

		if state == http.StateNew {
			s.wg.Add(1)
			return
		}
		if state == http.StateHijacked || state == http.StateClosed {
			s.wg.Done()
		}
	}

	if s.AccessLog == nil {
		s.AccessLog = log.DefaultLogger()
	}

	if s.Env == nil {
		s.Env = defaultEnv
	}
	s.Env.Go(s.wait)
}

func (s *HTTPServer) wait(ctx context.Context) error {
	<-ctx.Done()

	s.Server.SetKeepAlivesEnabled(false)

	ch := make(chan struct{})

	// Interrupt conn.Read for idle connections.
	//
	// This must be run inside for-loop to catch connections
	// going idle at critical timing to acquire s.mu
	go func() {
	AGAIN:
		s.mu.Lock()
		for conn := range s.idleConns {
			conn.SetReadDeadline(time.Now())
		}
		s.mu.Unlock()
		select {
		case <-ch:
			return
		default:
		}
		time.Sleep(10 * time.Millisecond)
		goto AGAIN
	}()

	go func() {
		s.wg.Wait()
		close(ch)
	}()

	if s.ShutdownTimeout == 0 {
		<-ch
		return nil
	}

	select {
	case <-ch:
	case <-time.After(s.ShutdownTimeout):
		log.Warn("cmd: timeout waiting for shutdown", nil)
		atomic.StoreInt32(&s.timedout, 1)
	}
	return nil
}

// TimedOut returns true if the server shut down before all connections
// got closed.
func (s *HTTPServer) TimedOut() bool {
	return atomic.LoadInt32(&s.timedout) != 0
}

// Serve overrides http.Server's Serve method.
//
// Unlike the original, this method returns immediately just after
// starting a goroutine to accept connections.
//
// The framework automatically closes l when the environment's Stop
// is called.
//
// Serve always returns nil.
func (s *HTTPServer) Serve(l net.Listener) error {
	s.initOnce.Do(s.init)

	go func() {
		<-s.Env.ctx.Done()
		l.Close()
	}()

	go func() {
		s.Server.Serve(l)
	}()

	return nil
}

type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

// ListenAndServe overrides http.Server's method.
//
// Unlike the original, this method returns immediately just after
// starting a goroutine to accept connections.
//
// ListenAndServe returns non-nil error if and only if net.Listen failed.
func (s *HTTPServer) ListenAndServe() error {
	addr := s.Server.Addr
	if addr == "" {
		addr = ":http"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return s.Serve(tcpKeepAliveListener{ln.(*net.TCPListener)})
}

// ListenAndServeTLS overrides http.Server's method.
//
// Unlike the original, this method returns immediately just after
// starting a goroutine to accept connections.
//
// Another difference from the original is that certFile and keyFile
// must be specified.  If not, configure http.Server.TLSConfig
// manually and use Serve().
//
// HTTP/2 is always enabled.
//
// ListenAndServeTLS returns non-nil error if net.Listen failed
// or failed to load certificate files.
func (s *HTTPServer) ListenAndServeTLS(certFile, keyFile string) error {
	addr := s.Server.Addr
	if addr == "" {
		addr = ":https"
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}

	config := &tls.Config{
		NextProtos:               []string{"h2", "http/1.1"},
		Certificates:             []tls.Certificate{cert},
		PreferServerCipherSuites: true,
		ClientSessionCache:       tls.NewLRUClientSessionCache(0),
	}
	s.Server.TLSConfig = config

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	tlsListener := tls.NewListener(tcpKeepAliveListener{ln.(*net.TCPListener)}, config)
	return s.Serve(tlsListener)
}
