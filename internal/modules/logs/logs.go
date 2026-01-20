package logs

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/systask/systask/internal/config"
	"github.com/systask/systask/internal/ssh"
)

// LogSource represents a log source
type LogSource struct {
	Name    string
	Command string
}

// DefaultSources returns common log sources
func DefaultSources() []LogSource {
	return []LogSource{
		{Name: "syslog", Command: "sudo tail -500 /var/log/syslog 2>/dev/null || sudo tail -500 /var/log/messages"},
		{Name: "auth", Command: "sudo tail -500 /var/log/auth.log 2>/dev/null || sudo tail -500 /var/log/secure"},
		{Name: "dmesg", Command: "dmesg | tail -500"},
		{Name: "journal", Command: "journalctl -n 500 --no-pager"},
		{Name: "nginx", Command: "sudo tail -500 /var/log/nginx/access.log 2>/dev/null"},
		{Name: "apache", Command: "sudo tail -500 /var/log/apache2/access.log 2>/dev/null || sudo tail -500 /var/log/httpd/access_log 2>/dev/null"},
	}
}

// Viewer displays log content
type Viewer struct {
	view        *tview.Flex
	sourceList  *tview.List
	logView     *tview.TextView
	filterInput *tview.InputField
	helpView    *tview.TextView
	theme       *config.Theme
	sshClient   *ssh.Client
	host        *ssh.HostEntry
	app         *tview.Application

	sources       []LogSource
	currentSource int
	filter        string
	following     bool
	wrap          bool
	logLines      []string

	stopFollow chan struct{}
}

// NewViewer creates a new log viewer
func NewViewer(theme *config.Theme, client *ssh.Client) *Viewer {
	v := &Viewer{
		theme:     theme,
		sshClient: client,
		sources:   DefaultSources(),
	}
	v.build()
	return v
}

// build constructs the log viewer
func (v *Viewer) build() {
	// Source list
	v.sourceList = tview.NewList()
	v.sourceList.ShowSecondaryText(false)
	v.sourceList.SetBackgroundColor(v.theme.Background)
	v.sourceList.SetBorder(true)
	v.sourceList.SetBorderColor(v.theme.Border)
	v.sourceList.SetTitle(" Sources ")
	v.sourceList.SetTitleColor(v.theme.Primary)
	v.sourceList.SetMainTextColor(v.theme.Foreground)
	v.sourceList.SetSelectedBackgroundColor(v.theme.Muted)
	v.sourceList.SetSelectedTextColor(v.theme.Primary)

	for i, src := range v.sources {
		shortcut := rune(0)
		if i < 9 {
			shortcut = rune('1' + i)
		}
		v.sourceList.AddItem(src.Name, "", shortcut, nil)
	}

	v.sourceList.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		v.currentSource = index
		v.loadLog()
	})

	// Log view
	v.logView = tview.NewTextView()
	v.logView.SetDynamicColors(true)
	v.logView.SetScrollable(true)
	v.logView.SetBackgroundColor(v.theme.Background)
	v.logView.SetBorder(true)
	v.logView.SetBorderColor(v.theme.Border)
	v.logView.SetTitle(" Log Output ")
	v.logView.SetTitleColor(v.theme.Primary)
	v.logView.SetTextColor(v.theme.Foreground)
	v.logView.SetWrap(false)

	// Filter input
	v.filterInput = tview.NewInputField()
	v.filterInput.SetLabel("Filter: ")
	v.filterInput.SetLabelColor(v.theme.Secondary)
	v.filterInput.SetFieldBackgroundColor(v.theme.Muted)
	v.filterInput.SetFieldTextColor(v.theme.Foreground)
	v.filterInput.SetBackgroundColor(v.theme.Background)
	v.filterInput.SetPlaceholder("Enter filter pattern...")
	v.filterInput.SetPlaceholderTextColor(v.theme.Border)

	v.filterInput.SetChangedFunc(func(text string) {
		v.filter = text
		v.applyFilter()
	})

	// Help bar
	v.helpView = tview.NewTextView()
	v.helpView.SetDynamicColors(true)
	v.helpView.SetBackgroundColor(v.theme.Muted)
	v.helpView.SetText(v.getHelpText())

	// Layout
	sidebar := tview.NewFlex().SetDirection(tview.FlexRow)
	sidebar.AddItem(v.sourceList, 0, 1, true)

	main := tview.NewFlex().SetDirection(tview.FlexRow)
	main.AddItem(v.logView, 0, 1, false)
	main.AddItem(v.filterInput, 1, 0, false)
	main.AddItem(v.helpView, 1, 0, false)

	v.view = tview.NewFlex()
	v.view.AddItem(sidebar, 20, 0, true)
	v.view.AddItem(main, 0, 1, false)
	v.view.SetBackgroundColor(v.theme.Background)

	// Key handler
	v.sourceList.SetInputCapture(v.handleInput)
}

// getHelpText returns the help bar text
func (v *Viewer) getHelpText() string {
	highlight := colorToTag(v.theme.Highlight)
	return fmt.Sprintf("  [%s]f[white]=Follow  [%s]w[white]=Wrap  [%s]g/G[white]=Top/Bottom  [%s]/[white]=Filter  [%s]ESC[white]=Home",
		highlight, highlight, highlight, highlight, highlight)
}

// handleInput processes key events
func (v *Viewer) handleInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case 'f':
		v.toggleFollow()
		return nil
	case 'w':
		v.toggleWrap()
		return nil
	case 'g':
		v.logView.ScrollToBeginning()
		return nil
	case 'G':
		v.logView.ScrollToEnd()
		return nil
	case '/':
		// Focus filter input
		return nil
	case 'j':
		return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
	case 'k':
		return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
	}

	return event
}

// toggleWrap toggles word wrap mode
func (v *Viewer) toggleWrap() {
	v.wrap = !v.wrap
	v.logView.SetWrap(v.wrap)
	if v.wrap {
		v.logView.SetTitle(fmt.Sprintf(" %s [WRAP] ", v.sources[v.currentSource].Name))
	} else {
		v.logView.SetTitle(fmt.Sprintf(" %s ", v.sources[v.currentSource].Name))
	}
}

// SetHost sets the current host
func (v *Viewer) SetHost(host *ssh.HostEntry) {
	v.host = host
}

// SetApp sets the tview application reference
func (v *Viewer) SetApp(app *tview.Application) {
	v.app = app
}

// Focus sets focus to the source list
func (v *Viewer) Focus() {
	if v.app != nil {
		v.app.SetFocus(v.sourceList)
	}
}

// loadLog loads the current log source
func (v *Viewer) loadLog() {
	if v.host == nil || v.currentSource >= len(v.sources) {
		return
	}

	src := v.sources[v.currentSource]
	v.logView.SetTitle(fmt.Sprintf(" %s ", src.Name))

	output, err := v.sshClient.RunCommand(*v.host, src.Command)
	if err != nil {
		v.logView.SetText(fmt.Sprintf("[%s]Error: %v[white]", colorToTag(v.theme.Error), err))
		return
	}

	v.logLines = strings.Split(output, "\n")
	v.applyFilter()
}

// applyFilter filters and displays log lines
func (v *Viewer) applyFilter() {
	var filtered []string

	for _, line := range v.logLines {
		if v.filter == "" || strings.Contains(strings.ToLower(line), strings.ToLower(v.filter)) {
			// Highlight log levels
			highlighted := v.highlightLine(line)
			filtered = append(filtered, highlighted)
		}
	}

	v.logView.SetText(strings.Join(filtered, "\n"))
	v.logView.ScrollToEnd()
}

// highlightLine adds color highlighting to log lines
func (v *Viewer) highlightLine(line string) string {
	lower := strings.ToLower(line)

	// Error highlighting
	if strings.Contains(lower, "error") || strings.Contains(lower, "fail") ||
		strings.Contains(lower, "crit") || strings.Contains(lower, "emerg") {
		return fmt.Sprintf("[%s]%s[white]", colorToTag(v.theme.Error), line)
	}

	// Warning highlighting
	if strings.Contains(lower, "warn") || strings.Contains(lower, "alert") {
		return fmt.Sprintf("[%s]%s[white]", colorToTag(v.theme.Warning), line)
	}

	// Success/info highlighting
	if strings.Contains(lower, "success") || strings.Contains(lower, "started") ||
		strings.Contains(lower, "connected") {
		return fmt.Sprintf("[%s]%s[white]", colorToTag(v.theme.Success), line)
	}

	return line
}

// toggleFollow toggles follow mode
func (v *Viewer) toggleFollow() {
	if v.following {
		v.stopFollowing()
	} else {
		v.startFollowing()
	}
}

// startFollowing starts follow mode (real-time polling)
func (v *Viewer) startFollowing() {
	if v.host == nil || v.currentSource >= len(v.sources) {
		return
	}

	v.following = true
	v.stopFollow = make(chan struct{})

	wrapIndicator := ""
	if v.wrap {
		wrapIndicator = " [WRAP]"
	}
	v.logView.SetTitle(fmt.Sprintf(" %s [FOLLOWING]%s ", v.sources[v.currentSource].Name, wrapIndicator))

	// Poll every 1 second for near real-time updates
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-v.stopFollow:
				return
			case <-ticker.C:
				v.loadLog()
				if v.app != nil {
					v.app.QueueUpdateDraw(func() {
						v.logView.ScrollToEnd()
					})
				}
			}
		}
	}()
}

// stopFollowing stops follow mode
func (v *Viewer) stopFollowing() {
	v.following = false
	if v.stopFollow != nil {
		close(v.stopFollow)
	}
	v.logView.SetTitle(fmt.Sprintf(" %s ", v.sources[v.currentSource].Name))
}

// Refresh reloads the current log
func (v *Viewer) Refresh() error {
	v.loadLog()
	return nil
}

// View returns the viewer view
func (v *Viewer) View() *tview.Flex {
	return v.view
}

// AddSource adds a custom log source
func (v *Viewer) AddSource(name, command string) {
	v.sources = append(v.sources, LogSource{Name: name, Command: command})
	v.sourceList.AddItem(name, "", 0, nil)
}

// Helper
func colorToTag(c tcell.Color) string {
	r, g, b := c.RGB()
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}
