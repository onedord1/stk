package ui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/systask/systask/internal/config"
)

// StatusBar manages the bottom status bar
type StatusBar struct {
	view  *tview.Flex
	theme *config.Theme

	modeText   *tview.TextView
	statusText *tview.TextView
	alertText  *tview.TextView
	helpText   *tview.TextView
	timeText   *tview.TextView

	currentMode string
}

// NewStatusBar creates a new status bar
func NewStatusBar(theme *config.Theme) *StatusBar {
	sb := &StatusBar{
		theme:       theme,
		currentMode: "NORMAL",
	}
	sb.build()
	return sb
}

// build constructs the status bar
func (sb *StatusBar) build() {
	// Mode indicator (left)
	sb.modeText = tview.NewTextView()
	sb.modeText.SetBackgroundColor(sb.theme.Primary)
	sb.modeText.SetTextColor(sb.theme.Background)
	sb.modeText.SetTextAlign(tview.AlignCenter)
	sb.modeText.SetText(" NORMAL ")

	// Status text (center-left)
	sb.statusText = tview.NewTextView()
	sb.statusText.SetBackgroundColor(sb.theme.Muted)
	sb.statusText.SetTextColor(sb.theme.Foreground)
	sb.statusText.SetDynamicColors(true)
	sb.statusText.SetText(" Ready ")

	// Alert area (center)
	sb.alertText = tview.NewTextView()
	sb.alertText.SetBackgroundColor(sb.theme.Muted)
	sb.alertText.SetTextColor(sb.theme.Warning)
	sb.alertText.SetDynamicColors(true)
	sb.alertText.SetTextAlign(tview.AlignCenter)

	// Help text (center-right)
	sb.helpText = tview.NewTextView()
	sb.helpText.SetBackgroundColor(sb.theme.Muted)
	sb.helpText.SetTextColor(sb.theme.Border)
	sb.helpText.SetTextAlign(tview.AlignRight)
	sb.helpText.SetText(" ESC=Home ?=Help ")

	// Time (right)
	sb.timeText = tview.NewTextView()
	sb.timeText.SetBackgroundColor(sb.theme.Primary)
	sb.timeText.SetTextColor(sb.theme.Background)
	sb.timeText.SetTextAlign(tview.AlignCenter)
	sb.updateTime()

	// Build layout
	sb.view = tview.NewFlex()
	sb.view.SetBackgroundColor(sb.theme.Muted)
	sb.view.AddItem(sb.modeText, 10, 0, false)
	sb.view.AddItem(sb.statusText, 0, 2, false)
	sb.view.AddItem(sb.alertText, 0, 1, false)
	sb.view.AddItem(sb.helpText, 20, 0, false)
	sb.view.AddItem(sb.timeText, 10, 0, false)
}

// SetMode updates the mode indicator
func (sb *StatusBar) SetMode(mode string) {
	sb.currentMode = mode

	modeColors := map[string]tcell.Color{
		"NORMAL":  sb.theme.Primary,
		"COMMAND": sb.theme.Warning,
		"SEARCH":  sb.theme.Secondary,
		"INSERT":  sb.theme.Success,
		"VISUAL":  sb.theme.Highlight,
	}

	if color, ok := modeColors[mode]; ok {
		sb.modeText.SetBackgroundColor(color)
	}
	sb.modeText.SetText(fmt.Sprintf(" %s ", mode))
}

// SetStatus updates the status text
func (sb *StatusBar) SetStatus(text string) {
	sb.statusText.SetText(fmt.Sprintf(" %s ", text))
}

// SetAlert displays an alert message
func (sb *StatusBar) SetAlert(text string, isError bool) {
	if isError {
		sb.alertText.SetTextColor(sb.theme.Error)
	} else {
		sb.alertText.SetTextColor(sb.theme.Warning)
	}
	sb.alertText.SetText(text)
}

// ClearAlert clears the alert message
func (sb *StatusBar) ClearAlert() {
	sb.alertText.SetText("")
}

// SetSuccess shows a success message
func (sb *StatusBar) SetSuccess(text string) {
	sb.alertText.SetTextColor(sb.theme.Success)
	sb.alertText.SetText(text)
}

// updateTime updates the time display
func (sb *StatusBar) updateTime() {
	sb.timeText.SetText(time.Now().Format(" 15:04:05 "))
}

// StartClock starts the clock updater
func (sb *StatusBar) StartClock(app *tview.Application) {
	go func() {
		ticker := time.NewTicker(time.Second)
		for range ticker.C {
			app.QueueUpdateDraw(func() {
				sb.updateTime()
			})
		}
	}()
}

// View returns the underlying view
func (sb *StatusBar) View() *tview.Flex {
	return sb.view
}

// GetMode returns the current mode
func (sb *StatusBar) GetMode() string {
	return sb.currentMode
}
