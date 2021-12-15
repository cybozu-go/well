package well

import (
	"os"
	"runtime"
	"testing"
)

func TestIsSystemdService(t *testing.T) {
	t.Parallel()

	if runtime.GOOS != "linux" {
		if IsSystemdService() {
			t.Error(`IsSystemdService()`)
		}
		return
	}

	if os.Getenv("GITHUB_ACTIONS") == "true" {
		if !IsSystemdService() {
			t.Error(`GitHub Actions is run as a systemd service`)
		}
		return
	}

	if IsSystemdService() {
		t.Error(`IsSystemdService()`)
	}

	err := os.Setenv("JOURNAL_STREAM", "10:20")
	if err != nil {
		t.Fatal(err)
	}

	if !IsSystemdService() {
		t.Error(`!IsSystemdService()`)
	}
}
