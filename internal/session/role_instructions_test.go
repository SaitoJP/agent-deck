package session

import "testing"

func TestComposeStartupMessage_FreshSessionPrependsRole(t *testing.T) {
	inst := NewInstanceWithTool("fresh", "/tmp/fresh", "copilot")
	inst.RoleInstructions = "# Role\nBe concise."

	got := inst.ComposeStartupMessage("Summarize the repo.")
	want := "# Role\nBe concise.\n\n---\n\nAfter adopting the role instructions above, handle this initial request:\nSummarize the repo."
	if got != want {
		t.Fatalf("ComposeStartupMessage() = %q, want %q", got, want)
	}
}

func TestComposeStartupMessage_ResumeSessionSkipsRoleBootstrap(t *testing.T) {
	inst := NewInstanceWithTool("resume", "/tmp/resume", "copilot")
	inst.RoleInstructions = "Persistent role"
	inst.CopilotSessionID = "existing-session"

	got := inst.ComposeStartupMessage("Continue.")
	if got != "Continue." {
		t.Fatalf("ComposeStartupMessage() = %q, want %q", got, "Continue.")
	}
}

func TestComposeStartupMessage_FreshWithoutInitialMessageUsesRoleOnly(t *testing.T) {
	inst := NewInstanceWithTool("fresh-role-only", "/tmp/fresh-role-only", "codex")
	inst.RoleInstructions = "You are the release manager."

	got := inst.ComposeStartupMessage("")
	if got != "You are the release manager." {
		t.Fatalf("ComposeStartupMessage() = %q, want %q", got, "You are the release manager.")
	}
}
