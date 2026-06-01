package harness

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestShortTmuxSocketPath_UsesTmp(t *testing.T) {
	socketPath, cleanup := shortTmuxSocketPath()
	defer cleanup()

	if filepath.Base(socketPath) != "tmux.sock" {
		t.Fatalf("socket path = %q, want basename tmux.sock", socketPath)
	}
	if !strings.HasPrefix(socketPath, "/tmp/") {
		t.Fatalf("socket path = %q, want /tmp prefix for short tmux socket paths", socketPath)
	}
}
