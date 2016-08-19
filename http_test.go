package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/cybozu-go/log"
)

func newMux(env *Environment, sleepCh chan struct{}) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})
	mux.HandleFunc("/sleep", func(w http.ResponseWriter, r *http.Request) {
		close(sleepCh)
		time.Sleep(1 * time.Second)
	})
	mux.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		env.Stop(nil)
	})
	return mux
}

func newHTTPClient() *http.Client {
	tr := &http.Transport{
		DisableCompression:  true,
		MaxIdleConnsPerHost: 10,
	}
	return &http.Client{
		Transport: tr,
	}
}

func TestHTTPServer(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
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
	resp, err := cl.Get("http://localhost:16555/hello")
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

	waitStart := time.Now()
	err = env.Wait()
	if err != nil {
		t.Error(err)
	}
	if time.Since(waitStart) > time.Second {
		t.Error("too long to shutdown")
	}

	testAccessLog(bytes.NewReader(out.Bytes()), t)
}

func testAccessLog(r io.Reader, t *testing.T) {
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
	t.Logf("%#v", *helloLog)
	t.Logf("%#v", *notfoundLog)

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
	if helloLog.ResponseLength != 5 {
		t.Error(`helloLog.ResponseLength != 5`)
	}
}

func TestHTTPServerTimeout(t *testing.T) {
	t.Parallel()

	env := NewEnvironment()
	sleepCh := make(chan struct{})
	s := &HTTPServer{
		Server: &http.Server{
			Addr:    "localhost:16556",
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
		resp, err := cl.Get("http://localhost:16556/sleep")
		if err != nil {
			return
		}
		resp.Body.Close()
	}()

	<-sleepCh
	resp, err := cl.Get("http://localhost:16556/shutdown")
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
