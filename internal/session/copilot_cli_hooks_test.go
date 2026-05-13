package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetCopilotHomeDir_PrefersCopilotHome(t *testing.T) {
	t.Setenv("COPILOT_HOME", "/tmp/copilot-home")
	t.Setenv("COPILOT_CONFIG_DIR", "/tmp/copilot-config-dir")

	if got := getCopilotHomeDir(); got != "/tmp/copilot-home" {
		t.Fatalf("getCopilotHomeDir() = %q, want /tmp/copilot-home", got)
	}
}

func TestInjectCopilotHooks_Fresh(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("COPILOT_HOME", tmpDir)
	t.Setenv("COPILOT_CONFIG_DIR", "")

	installed, err := InjectCopilotHooks()
	if err != nil {
		t.Fatalf("InjectCopilotHooks failed: %v", err)
	}
	if !installed {
		t.Fatal("expected hooks to be newly installed")
	}

	data, err := os.ReadFile(filepath.Join(tmpDir, "hooks", "agent-deck.json"))
	if err != nil {
		t.Fatalf("read hook file: %v", err)
	}

	var cfg copilotHooksFile
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("parse hook file: %v", err)
	}
	if cfg.Version != 1 {
		t.Fatalf("version = %d, want 1", cfg.Version)
	}
	for _, event := range copilotHookEventConfigs {
		entries, ok := cfg.Hooks[event.Event]
		if !ok {
			t.Fatalf("missing hook event %q", event.Event)
		}
		if len(entries) != 1 {
			t.Fatalf("event %q entries = %d, want 1", event.Event, len(entries))
		}
		if entries[0].Command != agentDeckHookCommand {
			t.Fatalf("event %q command = %q, want %q", event.Event, entries[0].Command, agentDeckHookCommand)
		}
		if entries[0].Matcher != event.Matcher {
			t.Fatalf("event %q matcher = %q, want %q", event.Event, entries[0].Matcher, event.Matcher)
		}
	}
}

func TestInjectCopilotHooks_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("COPILOT_HOME", tmpDir)
	t.Setenv("COPILOT_CONFIG_DIR", "")

	first, err := InjectCopilotHooks()
	if err != nil {
		t.Fatalf("first inject failed: %v", err)
	}
	if !first {
		t.Fatal("expected first install true")
	}

	second, err := InjectCopilotHooks()
	if err != nil {
		t.Fatalf("second inject failed: %v", err)
	}
	if second {
		t.Fatal("expected second install false")
	}
}

func TestRemoveCopilotHooks(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("COPILOT_HOME", tmpDir)
	t.Setenv("COPILOT_CONFIG_DIR", "")

	if _, err := InjectCopilotHooks(); err != nil {
		t.Fatalf("inject failed: %v", err)
	}

	removed, err := RemoveCopilotHooks()
	if err != nil {
		t.Fatalf("remove failed: %v", err)
	}
	if !removed {
		t.Fatal("expected hooks to be removed")
	}
	if CheckCopilotHooksInstalled() {
		t.Fatal("expected hooks not installed after removal")
	}
}

func TestCheckCopilotHooksInstalled_DetectsDrift(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("COPILOT_HOME", tmpDir)
	t.Setenv("COPILOT_CONFIG_DIR", "")

	if err := os.MkdirAll(filepath.Join(tmpDir, "hooks"), 0o755); err != nil {
		t.Fatalf("mkdir hooks: %v", err)
	}

	drifted := `{
  "version": 1,
  "hooks": {
    "Stop": [
      { "type": "command", "command": "agent-deck hook-handler" }
    ],
    "Notification": [
      { "type": "command", "command": "agent-deck hook-handler", "matcher": "permission_prompt" }
    ]
  }
}`
	if err := os.WriteFile(filepath.Join(tmpDir, "hooks", "agent-deck.json"), []byte(drifted), 0o644); err != nil {
		t.Fatalf("write drifted hook file: %v", err)
	}

	if CheckCopilotHooksInstalled() {
		t.Fatal("expected drifted hook file to be reported as not installed")
	}
}

func TestInjectCopilotHooks_RewritesDriftedFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("COPILOT_HOME", tmpDir)
	t.Setenv("COPILOT_CONFIG_DIR", "")

	if err := os.MkdirAll(filepath.Join(tmpDir, "hooks"), 0o755); err != nil {
		t.Fatalf("mkdir hooks: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(tmpDir, "hooks", "agent-deck.json"),
		[]byte(`{"version":1,"hooks":{"Notification":[{"type":"command","command":"agent-deck hook-handler","matcher":"permission_prompt"}]}}`),
		0o644,
	); err != nil {
		t.Fatalf("write drifted hook file: %v", err)
	}

	installed, err := InjectCopilotHooks()
	if err != nil {
		t.Fatalf("InjectCopilotHooks failed: %v", err)
	}
	if !installed {
		t.Fatal("expected InjectCopilotHooks to rewrite drifted file")
	}

	data, err := os.ReadFile(filepath.Join(tmpDir, "hooks", "agent-deck.json"))
	if err != nil {
		t.Fatalf("read hook file: %v", err)
	}
	text := string(data)
	if !strings.Contains(text, `"Notification"`) || !strings.Contains(text, `"permission_prompt|elicitation_dialog"`) {
		t.Fatalf("rewritten hook file missing expected notification matcher: %s", text)
	}
	if !strings.Contains(text, `"Stop"`) || !strings.Contains(text, `"SessionStart"`) {
		t.Fatalf("rewritten hook file missing required events: %s", text)
	}
}
