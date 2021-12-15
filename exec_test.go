//go:build !windows
// +build !windows

package well

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/cybozu-go/log"
)

func TestUTF8StringFromBytes(t *testing.T) {
	t.Parallel()

	str := "世\xc5\x33\x34"
	if utf8.ValidString(str) {
		t.Error(str + " should be invalid")
	}

	vstr := UTF8StringFromBytes([]byte(str))
	if !utf8.ValidString(vstr) {
		t.Error(vstr + ` should be valid`)
	}
}

func TestLogCmd(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, testUUID)

	logger := log.NewLogger()
	logger.SetFormatter(log.JSONFormat{})
	buf := new(bytes.Buffer)
	logger.SetOutput(buf)

	cmd := CommandContext(ctx, "true", "1", "2")
	cmd.Severity = log.LvDebug
	cmd.Logger = logger
	err := cmd.Run()
	if err != nil {
		t.Fatal(err)
	}
	if len(buf.Bytes()) > 0 {
		t.Error(`log should not be recorded`)
	}

	cmd = CommandContext(ctx, "true", "1", "2")
	cmd.Severity = log.LvInfo
	cmd.Logger = logger
	err = cmd.Run()
	if err != nil {
		t.Fatal(err)
	}
	var execlog ExecLog
	err = json.Unmarshal(buf.Bytes(), &execlog)
	if err != nil {
		t.Fatal(err)
	}

	if time.Since(execlog.LoggedAt) > time.Minute {
		t.Error(time.Since(execlog.LoggedAt) > time.Minute)
	}
	if execlog.Severity != "info" {
		t.Error(`execlog.Severity != "info"`)
	}
	if execlog.Type != "exec" {
		t.Error(`execlog.Type != "exec"`)
	}
	if execlog.Args[1] != "1" {
		t.Error(`execlog.Args[1] != "1"`)
	}
	if execlog.RequestID != testUUID {
		t.Error(`execlog.RequestID != testUUID`)
	}
}

func TestLogCmdError(t *testing.T) {
	ctx := context.Background()
	logger := log.NewLogger()
	logger.SetFormatter(log.JSONFormat{})
	buf := new(bytes.Buffer)
	logger.SetOutput(buf)

	cmd := CommandContext(ctx, "/bin/sh", "-c", "echo hoge fuga 1>&2; exit 3")
	cmd.Severity = log.LvDebug
	cmd.Logger = logger
	err := cmd.Run()
	if err == nil {
		t.Fatal("the command should fail")
	}

	var execlog ExecLog
	err = json.Unmarshal(buf.Bytes(), &execlog)
	if err != nil {
		t.Fatal(err)
	}

	if execlog.Severity != "error" {
		t.Error(`execlog.Severity != "error"`)
	}
	if execlog.Type != "exec" {
		t.Error(`execlog.Type != "exec"`)
	}
	t.Log(execlog.Error)
	if execlog.Stderr != "hoge fuga\n" {
		t.Error(`execlog.Stderr != "hoge fuga\n"`)
	}
}
