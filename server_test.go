package cmd

import (
	"bytes"
	"context"
	"io"
	"net"
	"strconv"
	"testing"
	"time"
)

func listen(port int, t *testing.T) net.Listener {
	l, err := net.Listen("tcp", "localhost:"+strconv.Itoa(port))
	if err != nil {
		t.Skip(err)
	}
	return l
}

func connect(port int, t *testing.T) net.Conn {
	conn, err := net.Dial("tcp", "localhost:"+strconv.Itoa(port))
	if err != nil {
		t.Fatal(err)
	}
	return conn
}

func TestServer(t *testing.T) {
	t.Parallel()

	l := listen(15555, t)
	handler := func(ctx context.Context, conn net.Conn) {
		conn.Write([]byte{'h', 'e', 'l', 'l', 'o'})
		<-ctx.Done()

		// uncomment the next line delays the test
		//time.Sleep(2 * time.Second)
	}

	env := NewEnvironment()
	s := &Server{
		Handler: handler,
		Env:     env,
	}
	s.Serve(l)

	conn := connect(15555, t)
	buf := make([]byte, 5)
	_, err := io.ReadFull(conn, buf)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buf, []byte{'h', 'e', 'l', 'l', 'o'}) {
		t.Error(`!bytes.Equal(buf, []byte{'h', 'e', 'l', 'l', 'o'})`)
	}

	env.Stop(nil)
	err = env.Wait()
	if err != nil {
		t.Error(err)
	}

	if s.TimedOut() {
		t.Error(`s.TimedOut()`)
	}
}

func TestServerTimeout(t *testing.T) {
	t.Parallel()

	l := listen(15556, t)
	handler := func(ctx context.Context, conn net.Conn) {
		conn.Write([]byte{'h', 'e', 'l', 'l', 'o'})
		<-ctx.Done()
		time.Sleep(1 * time.Second)
	}

	env := NewEnvironment()
	s := &Server{
		Handler:         handler,
		ShutdownTimeout: 100 * time.Millisecond,
		Env:             env,
	}
	s.Serve(l)

	conn := connect(15556, t)
	buf := make([]byte, 5)
	_, err := io.ReadFull(conn, buf)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buf, []byte{'h', 'e', 'l', 'l', 'o'}) {
		t.Error(`!bytes.Equal(buf, []byte{'h', 'e', 'l', 'l', 'o'})`)
	}

	env.Stop(nil)
	err = env.Wait()
	if err != nil {
		t.Error(err)
	}

	if !s.TimedOut() {
		t.Error(`!s.TimedOut()`)
	}
}
