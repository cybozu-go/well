//go:build !windows
// +build !windows

package well

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/cybozu-go/log"
)

func TestLogConfig(t *testing.T) {
	t.Parallel()

	var c1, c2 struct {
		Log LogConfig `toml:"log" json:"log"`
	}

	_, err := toml.DecodeFile("testdata/log.toml", &c1)
	if err != nil {
		t.Fatal(err)
	}

	if c1.Log.Filename != "/abc/def" {
		t.Error(`c1.Log.Filename != "/abc/def"`)
	}
	if c1.Log.Level != "debug" {
		t.Error(`c1.Log.Level != "debug"`)
	}
	if c1.Log.Format != "json" {
		t.Error(`c1.Log.Format != "json"`)
	}

	f, err := os.Open("testdata/log.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(&c2)
	if err != nil {
		t.Fatal(err)
	}

	if c2.Log.Filename != "/abc/def" {
		t.Error(`c2.Log.Filename != "/abc/def"`)
	}
	if c2.Log.Level != "debug" {
		t.Error(`c2.Log.Level != "debug"`)
	}
	if c2.Log.Format != "json" {
		t.Error(`c2.Log.Format != "json"`)
	}
}

func TestLogConfigApply(t *testing.T) {
	t.Parallel()

	c := &LogConfig{
		Filename: "",
		Level:    "info",
		Format:   "json",
	}

	err := c.Apply()
	if err != nil {
		t.Fatal(err)
	}

	logger := log.DefaultLogger()
	if logger.Threshold() != log.LvInfo {
		t.Error(`logger.Threshold() != log.LvInfo`)
	}
	if logger.Formatter().String() != "json" {
		t.Error(`logger.Formatter().String() != "json"`)
	}

	c.Format = "bad_format"
	err = c.Apply()
	if err == nil {
		t.Error(c.Format + " should cause an error")
	}

	c.Level = "bad_level"
	c.Format = "json"
	err = c.Apply()
	if err == nil {
		t.Error(c.Level + " should cause an error")
	}
}

func TestLogFlags(t *testing.T) {
	t.Parallel()
	t.Skip("this test redirects log outputs to a temp file.")

	c := &LogConfig{
		Filename: "",
		Level:    "info",
		Format:   "json",
	}

	f, err := os.CreateTemp("", "gotest")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	defer os.Remove(f.Name())

	flag.Set("logfile", f.Name())
	flag.Set("loglevel", "debug")
	flag.Set("logformat", "plain")

	err = c.Apply()
	if err != nil {
		t.Fatal(err)
	}

	logger := log.DefaultLogger()

	if logger.Threshold() != log.LvDebug {
		t.Error(`logger.Threshold() != log.LvDebug`)
	}
	if logger.Formatter().String() != "plain" {
		t.Error(`logger.Formatter().String() != "plain"`)
	}

	err = log.Critical("hoge fuga", nil)
	if err != nil {
		t.Fatal(err)
	}
	syscall.Kill(os.Getpid(), syscall.SIGUSR1)
	time.Sleep(10 * time.Millisecond)

	g, err := os.Open(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(g)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Contains(data, []byte("hoge fuga")) {
		t.Error(`!bytes.Contains(data, []byte("hoge fuga"))`)
	}
}
