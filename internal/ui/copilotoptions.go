package ui

import (
	"strings"

	"github.com/asheshgoplani/agent-deck/internal/session"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CopilotOptionsPanel is a UI panel for GitHub Copilot CLI launch options.
type CopilotOptionsPanel struct {
	modelInput textinput.Model
	allowAll   bool
	focusIndex int
}

// NewCopilotOptionsPanel creates a new panel for NewDialog.
func NewCopilotOptionsPanel() *CopilotOptionsPanel {
	modelInput := textinput.New()
	modelInput.Placeholder = "claude-sonnet-4.6"
	modelInput.CharLimit = 128
	modelInput.Width = 36

	return &CopilotOptionsPanel{
		modelInput: modelInput,
	}
}

// SetDefaults applies default values from config.
func (p *CopilotOptionsPanel) SetDefaults(config *session.UserConfig) {
	if config == nil {
		p.modelInput.SetValue("")
		p.allowAll = false
		return
	}
	p.modelInput.SetValue(strings.TrimSpace(config.Copilot.DefaultModel))
	p.allowAll = config.Copilot.AllowAll
}

// SetFromOptions applies persisted CopilotOptions to the panel fields.
func (p *CopilotOptionsPanel) SetFromOptions(opts *session.CopilotOptions) {
	if opts == nil {
		return
	}
	p.modelInput.SetValue(strings.TrimSpace(opts.Model))
	p.allowAll = opts.AllowAll
	p.updateInputFocus()
}

// Focus sets focus to this panel.
func (p *CopilotOptionsPanel) Focus() {
	p.focusIndex = 0
	p.updateInputFocus()
}

// Blur removes focus from this panel.
func (p *CopilotOptionsPanel) Blur() {
	p.focusIndex = -1
	p.modelInput.Blur()
}

// IsFocused returns true if the panel has focus.
func (p *CopilotOptionsPanel) IsFocused() bool {
	return p.focusIndex >= 0
}

// AtTop returns true if focus is on the first element.
func (p *CopilotOptionsPanel) AtTop() bool {
	return p.focusIndex <= 0
}

// GetOptions returns current options as CopilotOptions.
func (p *CopilotOptionsPanel) GetOptions() *session.CopilotOptions {
	model := strings.TrimSpace(p.modelInput.Value())
	return &session.CopilotOptions{
		SessionMode: "new",
		Model:       model,
		AllowAll:    p.allowAll,
	}
}

// Update handles key events.
func (p *CopilotOptionsPanel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			p.focusIndex--
			if p.focusIndex < 0 {
				p.focusIndex = 1
			}
			p.updateInputFocus()
			return nil

		case "down", "tab":
			p.focusIndex++
			if p.focusIndex > 1 {
				p.focusIndex = 0
			}
			p.updateInputFocus()
			return nil

		case "shift+tab":
			p.focusIndex--
			if p.focusIndex < 0 {
				p.focusIndex = 1
			}
			p.updateInputFocus()
			return nil

		case " ":
			if p.focusIndex == 1 {
				p.allowAll = !p.allowAll
				return nil
			}
		}
	}

	if p.focusIndex == 0 {
		var cmd tea.Cmd
		p.modelInput, cmd = p.modelInput.Update(msg)
		return cmd
	}
	return nil
}

// View renders the options panel.
func (p *CopilotOptionsPanel) View() string {
	headerStyle := lipgloss.NewStyle().Foreground(ColorComment)
	valueStyle := lipgloss.NewStyle()
	if p.focusIndex == 0 {
		valueStyle = valueStyle.Foreground(ColorAccent)
	}

	var content strings.Builder
	content.WriteString(headerStyle.Render("─ Copilot Options ─"))
	content.WriteString("\n")
	content.WriteString(valueStyle.Render("Model: " + p.modelInput.View()))
	content.WriteString("\n")
	content.WriteString(renderCheckboxLine("Allow all - pass --allow-all", p.allowAll, p.focusIndex == 1))
	return content.String()
}

func (p *CopilotOptionsPanel) updateInputFocus() {
	if p.focusIndex == 0 {
		p.modelInput.Focus()
	} else {
		p.modelInput.Blur()
	}
}
