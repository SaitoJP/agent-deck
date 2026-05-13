package session

import "strings"

// NormalizeHookEventName collapses CLI/runtime-specific spellings onto the
// canonical names used throughout agent-deck's hook/status logic.
func NormalizeHookEventName(event string) string {
	switch strings.TrimSpace(event) {
	case "sessionStart":
		return "SessionStart"
	case "userPromptSubmitted":
		return "UserPromptSubmit"
	case "agentStop":
		return "Stop"
	case "permissionRequest":
		return "PermissionRequest"
	case "notification":
		return "Notification"
	case "sessionEnd":
		return "SessionEnd"
	default:
		return strings.TrimSpace(event)
	}
}

func IsPermissionLikeHookEvent(event string) bool {
	switch NormalizeHookEventName(event) {
	case "PermissionRequest", "Notification":
		return true
	default:
		return false
	}
}

func IsTerminalPromptHookEvent(event string) bool {
	switch NormalizeHookEventName(event) {
	case "Stop", "PermissionRequest", "Notification":
		return true
	default:
		return false
	}
}
