package well

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"runtime"
	"testing"
	"time"

	"github.com/cybozu-go/log"
	"golang.org/x/net/http2"
)

const (
	testUUID = "cad48be9-285c-4b70-8177-33e41550a3c8"
)

func newMux(env *Environment, sleepCh chan struct{}) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		v := r.Context().Value(RequestIDContextKey)
		if v == nil {
			http.Error(w, "No request ID in context", http.StatusInternalServerError)
			return
		}

		fmt.Printf("req=%+v\n", r)
		if r.ProtoMajor == 2 {
			if _, ok := w.(StdResponseWriter2); !ok {
				http.Error(w, "not implement StdResponseWriter2", http.StatusInternalServerError)
				return
			}
			w.Write([]byte("hello2"))
			return
		}

		if _, ok := w.(StdResponseWriter); !ok {
			http.Error(w, "not implement StdResponseWriter", http.StatusInternalServerError)
			return
		}
		w.Write([]byte("hello"))
	})
	mux.HandleFunc("/sleep", func(w http.ResponseWriter, r *http.Request) {
		close(sleepCh)
		time.Sleep(1 * time.Second)
	})
	mux.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		env.Cancel(nil)
	})
	return mux
}

func newHTTPClient() *http.Client {
	tr := &http.Transport{
		DisableCompression:  true,
		MaxIdleConnsPerHost: 10,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
	}
	err := http2.ConfigureTransport(tr)
	if err != nil {
		panic(err)
	}
	return &http.Client{
		Transport: tr,
	}
}

func TestHTTPServer(t *testing.T) {
	t.Parallel()

	env := NewEnvironment(context.Background())
	logger := log.NewLogger()
	out := new(bytes.Buffer)
	logger.SetOutput(out)
	logger.SetFormatter(log.JSONFormat{})

	s := &HTTPServer{
		Server: &http.Server{
			Addr:        "localhost:16555",
			Handler:     newMux(env, nil),
			ReadTimeout: 3 * time.Second,
		},
		AccessLog: logger,
		Env:       env,
	}
	err := s.ListenAndServe()
	if err != nil {
		t.Fatal(err)
	}

	cl := newHTTPClient()

	req, err := http.NewRequest("GET", "http://localhost:16555/hello", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set(requestIDHeader, testUUID)
	resp, err := cl.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("%d %s", resp.StatusCode, string(data))
	}
	if !bytes.Equal(data, []byte("hello")) {
		t.Error(`!bytes.Equal(data, []byte("hello"))`)
	}

	resp, err = cl.Get("http://localhost:16555/notfound")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf(`resp.StatusCode != http.StatusNotFound`)
	}

	resp, err = cl.Get("http://localhost:16555/shutdown")
	if err != nil {
		t.Fatal(err)
	}
	data, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("%d %s", resp.StatusCode, string(data))
	}

	cl.CloseIdleConnections()

	waitStart := time.Now()
	err = env.Wait()
	if err != nil {
		t.Error(err)
	}
	if time.Since(waitStart) > time.Second {
		t.Error("too long to shutdown")
	}

	testAccessLog(bytes.NewReader(out.Bytes()), t, 5)
}

func TestHTTPServer2(t *testing.T) {
	t.Parallel()

	env := NewEnvironment(context.Background())
	logger := log.NewLogger()
	out := new(bytes.Buffer)
	logger.SetOutput(out)
	logger.SetFormatter(log.JSONFormat{})

	s := &HTTPServer{
		Server: &http.Server{
			Addr:        "localhost:16556",
			Handler:     newMux(env, nil),
			ReadTimeout: 3 * time.Second,
		},
		AccessLog: logger,
		Env:       env,
	}
	err := s.ListenAndServeTLS("testdata/cert.pem", "testdata/key.pem")
	if err != nil {
		t.Fatal(err)
	}

	cl := newHTTPClient()

	req, err := http.NewRequest("GET", "https://localhost:16556/hello", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set(requestIDHeader, testUUID)
	resp, err := cl.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("%d %s", resp.StatusCode, string(data))
	}
	if !bytes.Equal(data, []byte("hello2")) {
		t.Error(`!bytes.Equal(data, []byte("hello2"))`)
	}

	resp, err = cl.Get("https://localhost:16556/notfound")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf(`resp.StatusCode != http.StatusNotFound`)
	}

	resp, err = cl.Get("https://localhost:16556/shutdown")
	if err != nil {
		t.Fatal(err)
	}
	data, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("%d %s", resp.StatusCode, string(data))
	}

	waitStart := time.Now()
	err = env.Wait()
	if err != nil {
		t.Error(err)
	}
	if time.Since(waitStart) > 3*time.Second {
		t.Error("too long to shutdown")
	}

	testAccessLog(bytes.NewReader(out.Bytes()), t, 6)
}

func testAccessLog(r io.Reader, t *testing.T, helloLength int64) {
	decoder := json.NewDecoder(r)

	accessLogs := make([]*AccessLog, 0, 3)
	for decoder.More() {
		al := new(AccessLog)
		err := decoder.Decode(al)
		if err != nil {
			t.Fatal(err)
		}
		accessLogs = append(accessLogs, al)
	}

	if len(accessLogs) != 3 {
		t.Fatal(`len(accessLogs) != 3`)
	}

	helloLog := accessLogs[0]
	notfoundLog := accessLogs[1]

	if time.Since(helloLog.LoggedAt) > time.Minute {
		t.Error(`time.Since(helloLog.LoggedAt) > time.Minute`)
	}
	if time.Since(notfoundLog.LoggedAt) > time.Minute {
		t.Error(`time.Since(notfoundLog.LoggedAt) > time.Minute`)
	}
	if helloLog.Severity != "info" {
		t.Error(`helloLog.Severity != "info"`)
	}
	if notfoundLog.Severity != "warning" {
		t.Error(`notfoundLog.Severity != "warning"`)
	}
	if helloLog.Type != "access" {
		t.Error(`helloLog.Type != "access"`)
	}
	if helloLog.StatusCode != http.StatusOK {
		t.Error(`helloLog.StatusCode != http.StatusOK`)
	}
	if notfoundLog.StatusCode != http.StatusNotFound {
		t.Error(`notfoundLog.StatusCode != http.StatusNotFound`)
	}
	if helloLog.Method != "GET" {
		t.Error(`helloLog.Method != "GET"`)
	}
	if helloLog.RequestURI != "/hello" {
		t.Error(`helloLog.RequestURI != "/hello"`)
	}
	if notfoundLog.RequestURI != "/notfound" {
		t.Error(`notfoundLog.RequestURI != "/notfound"`)
	}
	if helloLog.ResponseLength != helloLength {
		t.Error(`helloLog.ResponseLength != helloLength`)
	}
	if helloLog.RequestID != testUUID {
		t.Error(`helloLog.RequestID != testUUID`)
	}
}

func TestHTTPServerTimeout(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows doesn't support FileListener")
	}
	t.Parallel()

	env := NewEnvironment(context.Background())
	sleepCh := make(chan struct{})
	s := &HTTPServer{
		Server: &http.Server{
			Addr:    "localhost:16557",
			Handler: newMux(env, sleepCh),
		},
		ShutdownTimeout: 50 * time.Millisecond,
		Env:             env,
	}
	err := s.ListenAndServe()
	if err != nil {
		t.Fatal(err)
	}

	cl := newHTTPClient()
	go func() {
		resp, err := cl.Get("http://localhost:16557/sleep")
		if err != nil {
			return
		}
		resp.Body.Close()
	}()

	<-sleepCh
	resp, err := cl.Get("http://localhost:16557/shutdown")
	if err != nil {
		t.Fatal(err)
	}
	data, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("%d %s", resp.StatusCode, string(data))
	}

	err = env.Wait()
	if err != nil {
		t.Error(err)
	}
	if !s.TimedOut() {
		t.Error(`!s.TimedOut()`)
	}
}

// Client tests

type testClientHandler struct{}

func (h testClientHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	uuid := r.Header.Get(requestIDHeader)
	if uuid == testUUID {
		return
	}
	http.Error(w, "invalid UUID", http.StatusInternalServerError)
}

func TestHTTPClient(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	env := NewEnvironment(ctx)
	ctx = WithRequestID(ctx, testUUID)

	s := &HTTPServer{
		Server: &http.Server{
			Addr:    "localhost:16558",
			Handler: testClientHandler{},
		},
		Env: env,
	}
	err := s.ListenAndServe()
	if err != nil {
		t.Fatal(err)
	}

	logger := log.NewLogger()
	logger.SetFormatter(log.JSONFormat{})
	buf := new(bytes.Buffer)
	logger.SetOutput(buf)

	cl := HTTPClient{
		Client:   &http.Client{},
		Severity: log.LvDebug,
		Logger:   logger,
	}
	req, err := http.NewRequest("GET", "http://localhost:16558", nil)
	if err != nil {
		t.Fatal(err)
	}

	req = req.WithContext(ctx)
	resp, err := cl.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Error("bad response:", resp.StatusCode)
	}

	if len(buf.Bytes()) != 0 {
		t.Error("should not be logged")
	}

	req, err = http.NewRequest("GET", "http://localhost:16558", nil)
	if err != nil {
		t.Fatal(err)
	}

	// raise threshold
	logger.SetThreshold(log.LvDebug)

	req = req.WithContext(ctx)
	resp, err = cl.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	var reqlog RequestLog
	err = json.Unmarshal(buf.Bytes(), &reqlog)
	if err != nil {
		t.Fatal(err)
	}

	if time.Since(reqlog.LoggedAt) > time.Minute {
		t.Error(`time.Since(reqlog.LoggedAt) > time.Minute`)
	}
	if reqlog.Severity != "debug" {
		t.Error(`reqlog.Severity != "debug"`)
	}
	if reqlog.Type != "http" {
		t.Error(`reqlog.Type != "http"`)
	}
	if reqlog.ResponseTime > 60.0 {
		t.Error(`reqlog.ResponseTime > 60.0`)
	}
	if reqlog.StatusCode != 200 {
		t.Error(`reqlog.StatusCode != 200`)
	}
	if reqlog.URLString != "http://localhost:16558" {
		t.Error(`reqlog.URLString != "http://localhost:16558"`)
	}
	if time.Since(reqlog.StartAt) > time.Minute {
		t.Error(`time.Since(reqlog.StartAt) > time.Minute`)
	}
	if reqlog.RequestID != testUUID {
		t.Error(`reqlog.RequestID != testUUID`)
	}

	env.Cancel(nil)
	err = env.Wait()
	if err != nil {
		t.Error(err)
	}
}
