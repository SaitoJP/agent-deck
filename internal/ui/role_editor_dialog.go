package ui

import (
	"fmt"
	"strings"

	"github.com/asheshgoplani/agent-deck/internal/session"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// RoleEditorDialog is a dedicated large-format editor for persistent per-session
// role instructions.
type RoleEditorDialog struct {
	visible      bool
	sessionID    string
	sessionTitle string
	groupName    string
	width        int
	height       int
	editor       textarea.Model
}

func NewRoleEditorDialog() *RoleEditorDialog {
	ta := textarea.New()
	ta.ShowLineNumbers = false
	ta.Prompt = "  "
	ta.SetPromptFunc(2, func(lineIdx int) string {
		cursorScreenLine := min(ta.Line(), max(0, ta.Height()-1))
		if lineIdx == cursorScreenLine {
			return "▌ "
		}
		return "  "
	})
	ta.CharLimit = 0
	ta.Placeholder = "# Role\n\nDescribe how this session should behave.\n\n- responsibilities\n- boundaries\n- output style"
	ta.FocusedStyle.LineNumber = lipgloss.NewStyle().Foreground(ColorComment)
	ta.FocusedStyle.CursorLineNumber = lipgloss.NewStyle().Foreground(ColorAccent).Bold(true)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle().Background(ColorBg).Foreground(ColorText)
	ta.FocusedStyle.Text = lipgloss.NewStyle().Foreground(ColorText)
	ta.FocusedStyle.Placeholder = lipgloss.NewStyle().Foreground(ColorComment)
	ta.BlurredStyle.LineNumber = lipgloss.NewStyle().Foreground(ColorComment)
	ta.BlurredStyle.CursorLineNumber = lipgloss.NewStyle().Foreground(ColorAccent).Bold(true)
	ta.BlurredStyle.CursorLine = lipgloss.NewStyle().Background(ColorSurface).Foreground(ColorText)
	ta.BlurredStyle.Text = lipgloss.NewStyle().Foreground(ColorText)
	ta.Blur()
	return &RoleEditorDialog{editor: ta}
}

func (d *RoleEditorDialog) Show(inst *session.Instance) {
	d.visible = true
	d.sessionID = inst.ID
	d.sessionTitle = inst.Title
	d.groupName = displayGroupName(inst.GroupPath)
	d.editor.SetValue(inst.RoleInstructions)
	d.editor.Focus()
}

func (d *RoleEditorDialog) Hide() {
	d.visible = false
	d.editor.Blur()
}

func (d *RoleEditorDialog) IsVisible() bool {
	if d == nil {
		return false
	}
	return d.visible
}

func (d *RoleEditorDialog) SetSize(w, h int) {
	if d == nil {
		return
	}
	d.width, d.height = w, h
}
func (d *RoleEditorDialog) SessionID() string { return d.sessionID }
func (d *RoleEditorDialog) Value() string {
	if d == nil {
		return ""
	}
	return strings.TrimRight(d.editor.Value(), "\n")
}

func (d *RoleEditorDialog) Update(msg tea.KeyMsg) (*RoleEditorDialog, tea.Cmd) {
	if !d.visible {
		return d, nil
	}
	switch msg.String() {
	case "esc", "ctrl+s":
		return d, nil
	default:
		var cmd tea.Cmd
		d.editor, cmd = d.editor.Update(msg)
		return d, cmd
	}
}

func (d *RoleEditorDialog) View() string {
	if !d.visible {
		return ""
	}

	screenWidth := max(d.width, 40)
	screenHeight := max(d.height, 16)
	maxDialogWidth := max(24, screenWidth-4)
	maxDialogHeight := max(12, screenHeight-2)

	dialogWidth := min(120, maxDialogWidth)
	if dialogWidth < 60 {
		dialogWidth = maxDialogWidth
	}
	dialogHeight := min(34, maxDialogHeight)
	if dialogHeight < 14 {
		dialogHeight = maxDialogHeight
	}

	headerLines := 5
	footerLines := 1
	gapLines := 2

	outerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorCyan).
		Background(ColorSurface).
		Padding(1, 2)
	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Padding(0, 1)
	innerDialogHeight := max(8, dialogHeight-outerStyle.GetVerticalFrameSize())
	innerDialogWidth := max(16, dialogWidth-outerStyle.GetHorizontalFrameSize())
	contentHeight := innerDialogHeight - headerLines - footerLines - gapLines
	contentHeight = max(panelStyle.GetVerticalFrameSize()+1, contentHeight)
	panelInnerWidth := func(total int) int {
		inner := total - panelStyle.GetHorizontalFrameSize()
		return max(8, inner)
	}
	panelInnerHeight := func(total int) int {
		inner := total - panelStyle.GetVerticalFrameSize()
		return max(1, inner)
	}

	editorHeight := max(panelStyle.GetVerticalFrameSize()+1, contentHeight)
	editorWidth := innerDialogWidth

	editorInnerWidth := max(12, panelInnerWidth(editorWidth))
	d.editor.SetWidth(editorInnerWidth)
	d.editor.SetHeight(max(3, panelInnerHeight(editorHeight)-4))
	d.editor.Focus()

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorCyan)
	metaStyle := lipgloss.NewStyle().Foreground(ColorComment)
	panelTitleStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorAccent)
	helpStyle := lipgloss.NewStyle().Foreground(ColorComment)
	positionStyle := lipgloss.NewStyle().Foreground(ColorAccent)
	truncate := func(s string, width int) string {
		if width <= 0 {
			return ""
		}
		return runewidth.Truncate(stripControlCharsPreserveANSI(s), width, "...")
	}

	editorPanel := panelStyle.
		Render(panelTitleStyle.Render("Editor") + "\n" + d.editor.View())

	var content strings.Builder
	content.WriteString(titleStyle.Render("Role Instructions"))
	content.WriteString("\n")
	content.WriteString(metaStyle.Render(truncate(fmt.Sprintf("session: %s  •  group: %s", d.sessionTitle, d.groupName), innerDialogWidth)))
	content.WriteString("\n")
	content.WriteString(metaStyle.Render(truncate("Persistent text for this session. Applies automatically on fresh starts; resume keeps the existing conversation context.", innerDialogWidth)))
	content.WriteString("\n\n")
	content.WriteString(editorPanel)
	content.WriteString("\n\n")
	content.WriteString(renderRoleEditorFooter(
		innerDialogWidth,
		helpStyle,
		positionStyle,
		fmt.Sprintf("Ln %d/%d", d.editor.Line()+1, max(1, d.editor.LineCount())),
	))

	dialog := outerStyle.Render(content.String())
	return lipgloss.Place(screenWidth, screenHeight, lipgloss.Center, lipgloss.Center, dialog)
}

func renderRoleEditorFooter(width int, helpStyle, positionStyle lipgloss.Style, position string) string {
	footerText := "Ctrl+S Save • Enter NL • Esc Close • " + position
	return helpStyle.Render(runewidth.Truncate(footerText, width, "..."))
}
