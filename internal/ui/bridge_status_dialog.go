package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/asheshgoplani/agent-deck/internal/session"
)

// bridgeLogTailLines is the number of log lines shown in the dialog.
const bridgeLogTailLines = 20

// BridgeStatusDialog shows the conductor bridge daemon status and connected conductors.
type BridgeStatusDialog struct {
	visible       bool
	width, height int
	status        session.BridgeStatus
	runningNames  map[string]bool
	showLogs      bool
	logLines      []string
	message       string
	messageTime   time.Time
}

// NewBridgeStatusDialog creates a new bridge status dialog.
func NewBridgeStatusDialog() *BridgeStatusDialog {
	return &BridgeStatusDialog{}
}

// Show opens the dialog with current bridge status.
func (d *BridgeStatusDialog) Show(status session.BridgeStatus, runningNames map[string]bool) {
	d.visible = true
	d.status = status
	d.runningNames = runningNames
	d.showLogs = false
	d.logLines = nil
	d.message = ""
}

// Hide closes the dialog.
func (d *BridgeStatusDialog) Hide() {
	d.visible = false
	d.showLogs = false
	d.logLines = nil
	d.message = ""
}

// IsVisible returns whether the dialog is currently shown.
func (d *BridgeStatusDialog) IsVisible() bool {
	return d.visible
}

// SetSize updates the dialog dimensions.
func (d *BridgeStatusDialog) SetSize(w, h int) {
	d.width = w
	d.height = h
}

// SetMessage sets a transient feedback message.
func (d *BridgeStatusDialog) SetMessage(msg string) {
	d.message = msg
	d.messageTime = time.Now()
}

// Refresh re-reads bridge status and running names.
func (d *BridgeStatusDialog) Refresh(status session.BridgeStatus, runningNames map[string]bool) {
	d.status = status
	d.runningNames = runningNames
	if d.showLogs {
		d.loadLogs()
	}
}

// ToggleLogs toggles the log tail view.
func (d *BridgeStatusDialog) ToggleLogs() {
	d.showLogs = !d.showLogs
	if d.showLogs {
		d.loadLogs()
	} else {
		d.logLines = nil
	}
}

// loadLogs reads the tail of the bridge log file.
func (d *BridgeStatusDialog) loadLogs() {
	logPath, err := session.BridgeLogPath()
	if err != nil {
		d.logLines = []string{"(error: " + err.Error() + ")"}
		return
	}
	data, err := os.ReadFile(logPath)
	if err != nil {
		d.logLines = []string{"(no log file found)"}
		return
	}
	allLines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	start := len(allLines) - bridgeLogTailLines
	if start < 0 {
		start = 0
	}
	d.logLines = allLines[start:]
}

// View renders the bridge status dialog.
func (d *BridgeStatusDialog) View() string {
	if !d.visible {
		return ""
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorAccent)

	sectionStyle := lipgloss.NewStyle().
		Foreground(ColorCyan).
		Bold(true)

	labelStyle := lipgloss.NewStyle().
		Foreground(ColorComment)

	valueStyle := lipgloss.NewStyle().
		Foreground(ColorText)

	footerStyle := lipgloss.NewStyle().
		Foreground(ColorComment).
		Italic(true)

	keyStyle := lipgloss.NewStyle().
		Foreground(ColorPurple)

	runningDot := lipgloss.NewStyle().Foreground(ColorGreen).Render("●")
	stoppedDot := lipgloss.NewStyle().Foreground(ColorTextDim).Render("○")

	var lines []string
	lines = append(lines, titleStyle.Render("Bridge Status"))
	lines = append(lines, "")

	// Feedback message (auto-dismiss after 5s)
	if d.message != "" && time.Since(d.messageTime) < 5*time.Second {
		lines = append(lines, "  "+lipgloss.NewStyle().Foreground(ColorYellow).Render(d.message))
		lines = append(lines, "")
	}

	// Daemon status
	if d.status.Running {
		lines = append(lines, fmt.Sprintf("  %s %s", runningDot, valueStyle.Render("Daemon running")))
	} else {
		lines = append(lines, fmt.Sprintf("  %s %s", stoppedDot, lipgloss.NewStyle().Foreground(ColorRed).Render("Daemon stopped")))
	}
	lines = append(lines, "")

	if !d.showLogs {
		// Platforms
		lines = append(lines, sectionStyle.Render("PLATFORMS"))
		if d.status.HasSlack() {
			mode := d.status.Slack.ListenMode
			if mode == "" {
				mode = "mentions"
			}
			lines = append(lines, fmt.Sprintf("  %s Slack  %s %s",
				runningDot,
				labelStyle.Render("channel:"),
				valueStyle.Render(d.status.Slack.ChannelID),
			))
			lines = append(lines, fmt.Sprintf("           %s %s",
				labelStyle.Render("mode:"),
				valueStyle.Render(mode),
			))
		} else {
			lines = append(lines, fmt.Sprintf("  %s %s", stoppedDot, labelStyle.Render("Slack not configured")))
		}

		if d.status.HasTelegram() {
			lines = append(lines, fmt.Sprintf("  %s Telegram  %s %s",
				runningDot,
				labelStyle.Render("user:"),
				valueStyle.Render(fmt.Sprintf("%d", d.status.Telegram.UserID)),
			))
		} else {
			lines = append(lines, fmt.Sprintf("  %s %s", stoppedDot, labelStyle.Render("Telegram not configured")))
		}

		if d.status.HasDiscord() {
			lines = append(lines, fmt.Sprintf("  %s Discord  %s %s",
				runningDot,
				labelStyle.Render("channel:"),
				valueStyle.Render(fmt.Sprintf("%d", d.status.Discord.ChannelID)),
			))
		} else {
			lines = append(lines, fmt.Sprintf("  %s %s", stoppedDot, labelStyle.Render("Discord not configured")))
		}
		lines = append(lines, "")

		// Conductors
		lines = append(lines, sectionStyle.Render("CONDUCTORS"))
		if len(d.status.Conductors) == 0 {
			lines = append(lines, "  "+labelStyle.Render("No conductors configured"))
		} else {
			for _, c := range d.status.Conductors {
				dot := stoppedDot
				if d.runningNames[session.ConductorSessionTitle(c.Name)] {
					dot = runningDot
				}
				label := fmt.Sprintf("%s %s (%s)", dot, c.Name, c.Profile)
				if c.Description != "" {
					label += fmt.Sprintf(" — %s", c.Description)
				}
				hb := "off"
				if c.HeartbeatEnabled {
					interval := c.HeartbeatInterval
					if interval <= 0 {
						interval = 15
					}
					hb = fmt.Sprintf("%dm", interval)
				}
				label += fmt.Sprintf("  %s %s", labelStyle.Render("hb:"), valueStyle.Render(hb))
				lines = append(lines, "  "+valueStyle.Render(label))
			}
		}
	} else {
		// Log view
		lines = append(lines, sectionStyle.Render("LOGS")+" "+labelStyle.Render(fmt.Sprintf("(last %d lines)", bridgeLogTailLines)))
		logStyle := lipgloss.NewStyle().Foreground(ColorText)
		if len(d.logLines) == 0 {
			lines = append(lines, "  "+labelStyle.Render("(empty)"))
		} else {
			dialogWidth := d.dialogWidth()
			maxLineWidth := dialogWidth - 6
			if maxLineWidth < 20 {
				maxLineWidth = 20
			}
			for _, line := range d.logLines {
				if len(line) > maxLineWidth {
					line = line[:maxLineWidth-1] + "…"
				}
				lines = append(lines, "  "+logStyle.Render(line))
			}
		}
	}

	lines = append(lines, "")
	var footerParts []string
	footerParts = append(footerParts, keyStyle.Render("R")+" restart")
	if d.showLogs {
		footerParts = append(footerParts, keyStyle.Render("L")+" hide logs")
	} else {
		footerParts = append(footerParts, keyStyle.Render("L")+" logs")
	}
	footerParts = append(footerParts, keyStyle.Render("Esc")+" close")
	lines = append(lines, footerStyle.Render(strings.Join(footerParts, "  |  ")))

	content := strings.Join(lines, "\n")

	dialogWidth := d.dialogWidth()
	box := DialogBoxStyle.
		Width(dialogWidth).
		Render(content)

	return centerInScreen(box, d.width, d.height)
}

// dialogWidth returns the computed dialog width.
func (d *BridgeStatusDialog) dialogWidth() int {
	dialogWidth := 70
	if d.showLogs {
		dialogWidth = 90
	}
	if d.width > 0 && d.width < dialogWidth+10 {
		dialogWidth = d.width - 10
		if dialogWidth < 40 {
			dialogWidth = 40
		}
	}
	return dialogWidth
}
