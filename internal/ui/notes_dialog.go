package ui

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// NotesDialog handles multi-line notes editing for a group.
type NotesDialog struct {
	visible   bool
	groupPath string
	groupName string
	editor    textarea.Model
	width     int
	height    int
}

// NewNotesDialog creates a new notes dialog.
func NewNotesDialog() *NotesDialog {
	ta := textarea.New()
	ta.Placeholder = "Type your notes here..."
	ta.CharLimit = 4096
	ta.SetWidth(40)
	ta.SetHeight(10)
	ta.ShowLineNumbers = false

	return &NotesDialog{
		editor: ta,
	}
}

// Show displays the dialog pre-filled with existing notes.
func (n *NotesDialog) Show(groupPath, groupName, existingNotes string) {
	n.visible = true
	n.groupPath = groupPath
	n.groupName = groupName
	n.editor.SetValue(existingNotes)
	n.editor.Focus()
	n.editor.CursorEnd()
}

// Hide closes the dialog.
func (n *NotesDialog) Hide() {
	n.visible = false
	n.editor.Blur()
}

// IsVisible returns whether the dialog is visible.
func (n *NotesDialog) IsVisible() bool {
	return n.visible
}

// GroupPath returns the group this dialog is editing notes for.
func (n *NotesDialog) GroupPath() string {
	return n.groupPath
}

// Value returns the current editor content.
func (n *NotesDialog) Value() string {
	return n.editor.Value()
}

// SetSize sets the dialog size.
func (n *NotesDialog) SetSize(width, height int) {
	n.width = width
	n.height = height

	// Responsive editor width
	editorW := 50
	if width > 0 && width < editorW+14 {
		editorW = width - 14
		if editorW < 30 {
			editorW = 30
		}
	}
	n.editor.SetWidth(editorW)
}

// Update forwards a key message to the textarea.
func (n *NotesDialog) Update(msg tea.KeyMsg) (*NotesDialog, tea.Cmd) {
	var cmd tea.Cmd
	n.editor, cmd = n.editor.Update(msg)
	return n, cmd
}

// View renders the dialog.
func (n *NotesDialog) View() string {
	if !n.visible {
		return ""
	}

	// Responsive dialog width
	dialogWidth := 56
	if n.width > 0 && n.width < dialogWidth+10 {
		dialogWidth = n.width - 10
		if dialogWidth < 34 {
			dialogWidth = 34
		}
	}
	titleWidth := dialogWidth - 4

	title := "Notes: " + n.groupName
	titleStyle := DialogTitleStyle.Width(titleWidth)
	hintStyle := lipgloss.NewStyle().Foreground(ColorComment)
	hint := hintStyle.Render("Ctrl+S save │ Esc cancel")

	dialogContent := lipgloss.JoinVertical(
		lipgloss.Center,
		titleStyle.Render(title),
		"",
		n.editor.View(),
		"",
		hint,
	)

	dialog := DialogBoxStyle.
		Width(dialogWidth).
		Render(dialogContent)

	return lipgloss.Place(
		n.width,
		n.height,
		lipgloss.Center,
		lipgloss.Center,
		dialog,
	)
}
