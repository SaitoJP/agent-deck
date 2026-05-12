package session

import "strings"

// ShouldBootstrapRoleInstructions reports whether the persistent role text
// should be delivered as part of the next start. We only bootstrap on fresh
// starts; resume paths keep their prior conversation context and must not get
// duplicate role messages injected.
func (i *Instance) ShouldBootstrapRoleInstructions() bool {
	if i == nil || strings.TrimSpace(i.RoleInstructions) == "" {
		return false
	}

	switch {
	case IsClaudeCompatible(i.Tool):
		i.ensureClaudeSessionIDFromDisk()
		return i.ClaudeSessionID == ""
	case i.Tool == "copilot":
		return i.CopilotSessionID == ""
	case i.Tool == "gemini":
		return i.GeminiSessionID == ""
	case i.Tool == "opencode":
		return i.OpenCodeSessionID == ""
	case IsCodexCompatible(i.Tool):
		return i.CodexSessionID == ""
	default:
		return false
	}
}

// ComposeStartupMessage combines the persistent role text with an optional
// user-provided initial prompt. Role instructions are only prepended when this
// start is fresh; resume starts keep the existing conversation context.
func (i *Instance) ComposeStartupMessage(initialMessage string) string {
	userPrompt := strings.TrimSpace(initialMessage)
	role := strings.TrimSpace(i.RoleInstructions)
	if !i.ShouldBootstrapRoleInstructions() {
		return userPrompt
	}
	if role == "" {
		return userPrompt
	}
	if userPrompt == "" {
		return role
	}
	return role + "\n\n---\n\nAfter adopting the role instructions above, handle this initial request:\n" + userPrompt
}

// StartWithStartupMessage starts the session with its persistent role text
// and/or a caller-supplied initial prompt, selecting Start vs StartWithMessage
// automatically.
func (i *Instance) StartWithStartupMessage(initialMessage string) error {
	message := i.ComposeStartupMessage(initialMessage)
	if strings.TrimSpace(message) == "" {
		return i.Start()
	}
	return i.StartWithMessage(message)
}
