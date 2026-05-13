package session

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

const copilotAgentDeckHookFileName = "agent-deck.json"

type copilotHookEntry struct {
	Type       string `json:"type"`
	Command    string `json:"command"`
	Matcher    string `json:"matcher,omitempty"`
	TimeoutSec int    `json:"timeoutSec,omitempty"`
}

type copilotHooksFile struct {
	Version         int                           `json:"version"`
	DisableAllHooks bool                          `json:"disableAllHooks,omitempty"`
	Hooks           map[string][]copilotHookEntry `json:"hooks"`
}

var copilotHookEventConfigs = []struct {
	Event   string
	Matcher string
}{
	{Event: "SessionStart"},
	{Event: "UserPromptSubmit"},
	{Event: "Stop"},
	{Event: "PermissionRequest"},
	{Event: "Notification", Matcher: "permission_prompt|elicitation_dialog"},
	{Event: "SessionEnd"},
}

func copilotHooksDir() string {
	return filepath.Join(getCopilotHomeDir(), "hooks")
}

func copilotHooksFilePath() string {
	return filepath.Join(copilotHooksDir(), copilotAgentDeckHookFileName)
}

func desiredCopilotHooksFile() copilotHooksFile {
	hooks := make(map[string][]copilotHookEntry, len(copilotHookEventConfigs))
	for _, cfg := range copilotHookEventConfigs {
		hooks[cfg.Event] = []copilotHookEntry{{
			Type:       "command",
			Command:    agentDeckHookCommand,
			Matcher:    cfg.Matcher,
			TimeoutSec: 30,
		}}
	}
	return copilotHooksFile{
		Version: 1,
		Hooks:   hooks,
	}
}

func InjectCopilotHooks() (bool, error) {
	path := copilotHooksFilePath()
	if CheckCopilotHooksInstalled() {
		return false, nil
	}

	cfg := desiredCopilotHooksFile()
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return false, fmt.Errorf("marshal copilot hooks: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return false, fmt.Errorf("create copilot hooks dir: %w", err)
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return false, fmt.Errorf("write copilot hooks tmp: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return false, fmt.Errorf("rename copilot hooks file: %w", err)
	}

	sessionLog.Info("copilot_hooks_installed", slog.String("path", path))
	return true, nil
}

func RemoveCopilotHooks() (bool, error) {
	path := copilotHooksFilePath()
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("remove copilot hooks: %w", err)
	}
	sessionLog.Info("copilot_hooks_removed", slog.String("path", path))
	return true, nil
}

func CheckCopilotHooksInstalled() bool {
	data, err := os.ReadFile(copilotHooksFilePath())
	if err != nil {
		return false
	}

	var cfg copilotHooksFile
	if err := json.Unmarshal(data, &cfg); err != nil {
		return false
	}
	if cfg.Version != 1 {
		return false
	}
	for _, event := range copilotHookEventConfigs {
		entries, ok := cfg.Hooks[event.Event]
		if !ok || len(entries) != 1 {
			return false
		}
		entry := entries[0]
		if entry.Type != "command" || entry.Command != agentDeckHookCommand || entry.Matcher != event.Matcher {
			return false
		}
	}
	return true
}

func ensureCopilotHooksInstalled() {
	if _, err := InjectCopilotHooks(); err != nil {
		sessionLog.Warn("copilot_hooks_install_failed", slog.String("error", err.Error()))
	}
}
