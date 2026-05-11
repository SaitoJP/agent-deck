package main

import (
	"fmt"
	"strings"

	"github.com/asheshgoplani/agent-deck/internal/session"
)

func applyCLICopilotModelOverride(inst *session.Instance, model string) error {
	model = strings.TrimSpace(model)
	if model == "" || inst == nil {
		return nil
	}
	if inst.Tool != "copilot" {
		return fmt.Errorf("--model only works with Copilot sessions")
	}

	inst.CopilotModel = model

	opts := inst.GetCopilotOptions()
	if opts == nil {
		opts = &session.CopilotOptions{SessionMode: "new"}
	}
	opts.Model = model
	return inst.SetCopilotOptions(opts)
}
