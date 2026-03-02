package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/asheshgoplani/agent-deck/internal/session"
)

// ConductorPickerDialog presents a list of configured conductors for the user to start.
type ConductorPickerDialog struct {
	visible       bool
	width, height int
	conductors    []session.ConductorMeta
	runningNames  map[string]bool // Already-running conductor names
	cursor        int
}

// NewConductorPickerDialog creates a new conductor picker dialog.
func NewConductorPickerDialog() *ConductorPickerDialog {
	return &ConductorPickerDialog{}
}

// Show opens the picker with the available conductors and running state.
func (d *ConductorPickerDialog) Show(conductors []session.ConductorMeta, runningNames map[string]bool) {
	d.visible = true
	d.conductors = conductors
	d.runningNames = runningNames
	d.cursor = 0
}

// Hide closes the dialog and resets state.
func (d *ConductorPickerDialog) Hide() {
	d.visible = false
	d.cursor = 0
	d.conductors = nil
	d.runningNames = nil
}

// IsVisible returns whether the dialog is currently shown.
func (d *ConductorPickerDialog) IsVisible() bool {
	return d.visible
}

// SetSize updates the dialog dimensions for centering.
func (d *ConductorPickerDialog) SetSize(w, h int) {
	d.width = w
	d.height = h
}

// GetSelected returns the conductor at the current cursor position, or nil.
func (d *ConductorPickerDialog) GetSelected() *session.ConductorMeta {
	if len(d.conductors) == 0 || d.cursor >= len(d.conductors) {
		return nil
	}
	return &d.conductors[d.cursor]
}

// Update handles key events for the picker.
func (d *ConductorPickerDialog) Update(msg tea.KeyMsg) {
	if !d.visible {
		return
	}

	switch msg.String() {
	case "j", "down", "ctrl+n":
		if len(d.conductors) > 0 {
			d.cursor = (d.cursor + 1) % len(d.conductors)
		}
	case "k", "up", "ctrl+p":
		if len(d.conductors) > 0 {
			d.cursor = (d.cursor - 1 + len(d.conductors)) % len(d.conductors)
		}
	}
}

// View renders the conductor picker dialog.
func (d *ConductorPickerDialog) View() string {
	if !d.visible {
		return ""
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorAccent)

	selectedStyle := lipgloss.NewStyle().
		Foreground(ColorAccent).
		Bold(true)

	normalStyle := lipgloss.NewStyle().
		Foreground(ColorText)

	footerStyle := lipgloss.NewStyle().
		Foreground(ColorComment).
		Italic(true)

	runningDot := lipgloss.NewStyle().Foreground(ColorGreen).Render("●")
	stoppedDot := lipgloss.NewStyle().Foreground(ColorTextDim).Render("○")

	var lines []string
	lines = append(lines, titleStyle.Render("Start Conductor"))
	lines = append(lines, "")

	if len(d.conductors) == 0 {
		lines = append(lines, normalStyle.Render("No conductors configured"))
	} else {
		for i, c := range d.conductors {
			dot := stoppedDot
			if d.runningNames[session.ConductorSessionTitle(c.Name)] {
				dot = runningDot
			}

			label := fmt.Sprintf("%s %s (%s)", dot, c.Name, c.Profile)
			if c.Description != "" {
				label += fmt.Sprintf(" - %q", c.Description)
			}

			if i == d.cursor {
				lines = append(lines, "> "+selectedStyle.Render(label))
			} else {
				lines = append(lines, "  "+normalStyle.Render(label))
			}
		}
	}

	lines = append(lines, "")
	lines = append(lines, footerStyle.Render("Enter start | Esc cancel | j/k navigate"))

	content := strings.Join(lines, "\n")

	dialogWidth := 50
	if d.width > 0 && d.width < dialogWidth+10 {
		dialogWidth = d.width - 10
		if dialogWidth < 30 {
			dialogWidth = 30
		}
	}

	box := DialogBoxStyle.
		Width(dialogWidth).
		Render(content)

	return centerInScreen(box, d.width, d.height)
}
