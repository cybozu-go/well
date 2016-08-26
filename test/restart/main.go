package main

import (
	"context"
	"errors"
	"flag"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/cybozu-go/cmd"
	"github.com/cybozu-go/log"
)

var (
	tcpAddr  = "localhost:18556"
	unixAddr string
)

func getTemporaryFilename() string {
	f, err := ioutil.TempFile("", "gotest")
	if err != nil {
		log.ErrorExit(err)
	}
	f.Close()
	os.Remove(f.Name())
	return f.Name()
}

func main() {
	flag.Parse()
	cmd.LogConfig{}.Apply()

	unixAddr = getTemporaryFilename()
	defer os.Remove(unixAddr)

	listen := func() ([]net.Listener, error) {
		ln1, err := net.Listen("tcp", tcpAddr)
		if err != nil {
			log.ErrorExit(err)
		}
		ln2, err := net.Listen("unix", unixAddr)
		if err != nil {
			log.ErrorExit(err)
		}
		// start client after creating listening socket.
		cmd.Go(testClient)
		return []net.Listener{ln1, ln2}, nil
	}

	g := &cmd.Graceful{
		Listen: listen,
		Serve:  serve,
	}
	g.Run()

	// rest are executed only in the master process.
	err := cmd.Wait()
	if err != nil && !cmd.IsSignaled(err) {
		log.ErrorExit(err)
	}
}

// serve implements a network server that can be stopped gracefully
// using cmd.Server.
func serve(listeners []net.Listener) {
	var counter int64
	handler := func(ctx context.Context, conn net.Conn) {
		n := atomic.AddInt64(&counter, 1)
		if n > 1 {
			time.Sleep(time.Duration(n) * time.Second)
		}
		conn.Write([]byte("hello " + strconv.FormatInt(n, 10)))
		conn.Close()
	}

	s := &cmd.Server{
		Handler: handler,
	}
	for _, ln := range listeners {
		s.Serve(ln)
	}
	err := cmd.Wait()
	if err != nil && !cmd.IsSignaled(err) {
		log.ErrorExit(err)
	}
}

func testClient(ctx context.Context) error {
	for i := 0; i < 5; i++ {
		err := ping("tcp", tcpAddr)
		if err != nil {
			return err
		}
		restart()
	}

	err := ping("unix", unixAddr)
	if err != nil {
		return err
	}

	cmd.Cancel(nil)
	return nil
}

func ping(network, addr string) error {
	conn, err := net.Dial(network, addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	st := time.Now()
	data, err := ioutil.ReadAll(conn)
	if err != nil {
		return err
	}
	log.Info("got data", map[string]interface{}{
		"data": string(data),
	})

	if time.Since(st) > time.Second {
		return errors.New("too long")
	}
	return nil
}

func restart() {
	syscall.Kill(os.Getpid(), syscall.SIGHUP)
	time.Sleep(10 * time.Millisecond)
}
