package main

import (
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/asheshgoplani/agent-deck/internal/session"
)

func TestMCPInfoForJSON_NilOrEmpty(t *testing.T) {
	if got := mcpInfoForJSON(nil); got != nil {
		t.Fatalf("mcpInfoForJSON(nil) = %#v, want nil", got)
	}

	if got := mcpInfoForJSON(&session.MCPInfo{}); got != nil {
		t.Fatalf("mcpInfoForJSON(empty) = %#v, want nil", got)
	}
}

func TestMCPInfoForJSON_UsesSlicesAndIsMarshalable(t *testing.T) {
	info := &session.MCPInfo{
		Global:  []string{"global-a"},
		Project: []string{"project-a"},
		LocalMCPs: []session.LocalMCP{
			{Name: "local-a", SourcePath: "/tmp"},
		},
	}

	got := mcpInfoForJSON(info)
	if got == nil {
		t.Fatal("mcpInfoForJSON returned nil for populated MCP info")
	}

	local, ok := got["local"].([]string)
	if !ok {
		t.Fatalf("mcps.local type = %T, want []string", got["local"])
	}
	if len(local) != 1 || local[0] != "local-a" {
		t.Fatalf("mcps.local = %#v, want []string{\"local-a\"}", local)
	}

	payload := map[string]interface{}{"mcps": got}
	if _, err := json.Marshal(payload); err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
}

func TestPrintSessionHelp_IncludesResumeCommand(t *testing.T) {
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	defer r.Close()

	os.Stdout = w
	printSessionHelp()
	_ = w.Close()
	os.Stdout = origStdout

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if !strings.Contains(string(out), "resume <id>") {
		t.Fatalf("session help missing resume command, got:\n%s", string(out))
	}
}
