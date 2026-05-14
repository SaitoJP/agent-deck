package tmux

import "testing"

// Issue #556: tmux-layer detection + pattern defaults for GitHub Copilot CLI.

func TestDetectToolFromCommand_Copilot(t *testing.T) {
	tests := []struct {
		name    string
		command string
		want    string
	}{
		{"bare copilot", "copilot", "copilot"},
		{"copilot with resume flag", "copilot --resume", "copilot"},
		{"copilot via npx", "npx @github/copilot", "copilot"},
		{"uppercase binary", "COPILOT", "copilot"},
		{
			"copilot wrapped with exec and claude model",
			`export AGENTDECK_INSTANCE_ID=abc; exec copilot --model claude-sonnet-4.6`,
			"copilot",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := detectToolFromCommand(tt.command); got != tt.want {
				t.Fatalf("detectToolFromCommand(%q) = %q, want %q", tt.command, got, tt.want)
			}
		})
	}
}

func TestDefaultRawPatterns_Copilot(t *testing.T) {
	raw := DefaultRawPatterns("copilot")
	if raw == nil {
		t.Fatal("expected non-nil RawPatterns for copilot")
	}
	if len(raw.BusyPatterns) == 0 {
		t.Error("copilot should have busy patterns")
	}
	if len(raw.PromptPatterns) == 0 {
		t.Error("copilot should have prompt patterns")
	}
}

func TestCopilotBusyIndicator_DoesNotMatchRunningWordInPromptTranscript(t *testing.T) {
	s := NewSession("copilot-running-word", "/tmp")
	s.Command = "copilot"

	content := `named conductor online
Running child sessions: 3
copilot>
`
	if s.hasBusyIndicator(content) {
		t.Fatal("hasBusyIndicator() = true, want false for plain transcript text containing Running")
	}
}

func TestCopilotBusyIndicator_MatchesEscToCancelFooter(t *testing.T) {
	s := NewSession("copilot-esc-cancel", "/tmp")
	s.Command = "copilot"

	content := "Thinking\nEsc to cancel\n"
	if !s.hasBusyIndicator(content) {
		t.Fatal("hasBusyIndicator() = false, want true when Copilot footer shows Esc to cancel")
	}
}

func TestCopilotBusyIndicator_IgnoresEscToCancelInTranscript(t *testing.T) {
	s := NewSession("copilot-esc-cancel-prose", "/tmp")
	s.Command = "copilot"

	content := `The UI says Esc to cancel while processing.
Here is the summary.
copilot>
`
	if s.hasBusyIndicator(content) {
		t.Fatal("hasBusyIndicator() = true, want false for transcript prose mentioning Esc to cancel")
	}
}

func TestCopilotBusyIndicator_MatchesThinkingLine(t *testing.T) {
	s := NewSession("copilot-thinking", "/tmp")
	s.Command = "copilot"

	content := "Thinking\ncopilot>\n"
	if !s.hasBusyIndicator(content) {
		t.Fatal("hasBusyIndicator() = false, want true for Copilot Thinking line")
	}
}

func TestCopilotBusyIndicator_MatchesThinkingLineWithAnimatedPrefix(t *testing.T) {
	for _, symbol := range []string{"●", "◉", "◎", "○"} {
		s := NewSession("copilot-thinking-prefix-"+symbol, "/tmp")
		s.Command = "copilot"

		content := symbol + " Thinking (Esc to cancel · 887 B)\n"
		if !s.hasBusyIndicator(content) {
			t.Fatalf("hasBusyIndicator() = false, want true for Copilot Thinking line with animated prefix %q", symbol)
		}
	}
}

func TestCopilotBusyIndicator_MatchesRandomStatusTextWithAnimatedPrefix(t *testing.T) {
	for _, symbol := range []string{"●", "◉", "◎", "○"} {
		s := NewSession("copilot-random-status-"+symbol, "/tmp")
		s.Command = "copilot"

		content := symbol + " Testing symbol set (Esc to cancel · 2.7 KiB)\n"
		if !s.hasBusyIndicator(content) {
			t.Fatalf("hasBusyIndicator() = false, want true for random Copilot status text with prefix %q", symbol)
		}
	}
}
