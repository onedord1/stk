package terminal

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/systask/systask/internal/config"
	"github.com/systask/systask/internal/ssh"
	"golang.org/x/term"
)

// Terminal provides fullscreen SSH terminal with quick actions
type Terminal struct {
	view     *tview.Flex
	textView *tview.TextView
	quickBar *tview.Flex
	theme    *config.Theme
	host     *ssh.HostEntry
	app      *tview.Application
	session  *exec.Cmd
	zenMode  bool

	// Callbacks
	onExit        func()
	onQuickAction func(string)
}

// NewTerminal creates a new terminal view
func NewTerminal(theme *config.Theme) *Terminal {
	t := &Terminal{
		theme:   theme,
		zenMode: true,
	}
	t.build()
	return t
}

// build constructs the terminal view
func (t *Terminal) build() {
	// Main terminal output area
	t.textView = tview.NewTextView()
	t.textView.SetDynamicColors(true)
	t.textView.SetScrollable(true)
	t.textView.SetBackgroundColor(tcell.ColorBlack)
	t.textView.SetTextColor(tcell.ColorWhite)
	t.textView.SetWrap(true)
	t.textView.SetWordWrap(true)

	// Quick action bar (hidden in zen mode)
	t.quickBar = tview.NewFlex()
	t.quickBar.SetBackgroundColor(t.theme.Muted)

	actions := []struct {
		key  string
		name string
		code string
	}{
		{"F1", "Help", "help"},
		{"F2", "Users", "users"},
		{"F3", "Docker", "docker"},
		{"F4", "Services", "services"},
		{"F5", "Logs", "logs"},
		{"F6", "Files", "sftp"},
		{"F10", "Exit", "exit"},
	}

	for _, a := range actions {
		label := tview.NewTextView()
		label.SetDynamicColors(true)
		label.SetTextAlign(tview.AlignCenter)
		label.SetText(fmt.Sprintf("[%s]%s[white]%s ",
			colorToTag(t.theme.Highlight), a.key, a.name))
		label.SetBackgroundColor(t.theme.Muted)
		t.quickBar.AddItem(label, 12, 0, false)
	}

	// Main layout
	t.view = tview.NewFlex().SetDirection(tview.FlexRow)
	t.view.AddItem(t.textView, 0, 1, true)
	t.view.AddItem(t.quickBar, 1, 0, false)
	t.view.SetBackgroundColor(tcell.ColorBlack)

	// Key handling
	t.view.SetInputCapture(t.handleInput)
}

// handleInput handles terminal input
func (t *Terminal) handleInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyF1:
		if t.onQuickAction != nil {
			t.onQuickAction("help")
		}
		return nil
	case tcell.KeyF2:
		if t.onQuickAction != nil {
			t.onQuickAction("users")
		}
		return nil
	case tcell.KeyF3:
		if t.onQuickAction != nil {
			t.onQuickAction("docker")
		}
		return nil
	case tcell.KeyF4:
		if t.onQuickAction != nil {
			t.onQuickAction("services")
		}
		return nil
	case tcell.KeyF5:
		if t.onQuickAction != nil {
			t.onQuickAction("logs")
		}
		return nil
	case tcell.KeyF6:
		if t.onQuickAction != nil {
			t.onQuickAction("sftp")
		}
		return nil
	case tcell.KeyF10, tcell.KeyEsc:
		if t.onExit != nil {
			t.onExit()
		}
		return nil
	case tcell.KeyF11:
		t.toggleZenMode()
		return nil
	}
	return event
}

// toggleZenMode toggles fullscreen mode
func (t *Terminal) toggleZenMode() {
	t.zenMode = !t.zenMode
	if t.zenMode {
		t.view.RemoveItem(t.quickBar)
	} else {
		t.view.AddItem(t.quickBar, 1, 0, false)
	}
}

// SetHost sets the host to connect to
func (t *Terminal) SetHost(host *ssh.HostEntry) {
	t.host = host
}

// SetApp sets the tview application
func (t *Terminal) SetApp(app *tview.Application) {
	t.app = app
}

// OnExit sets the exit callback
func (t *Terminal) OnExit(fn func()) {
	t.onExit = fn
}

// OnQuickAction sets the quick action callback
func (t *Terminal) OnQuickAction(fn func(string)) {
	t.onQuickAction = fn
}

// Connect establishes SSH connection and starts terminal
func (t *Terminal) Connect() error {
	if t.host == nil {
		return fmt.Errorf("no host set")
	}

	// Build SSH command
	args := []string{}

	if t.host.User != "" {
		args = append(args, "-l", t.host.User)
	}

	if t.host.Port != 0 && t.host.Port != 22 {
		args = append(args, "-p", fmt.Sprintf("%d", t.host.Port))
	}

	if t.host.KeyFile != "" {
		args = append(args, "-i", t.host.KeyFile)
	}

	args = append(args, t.host.Hostname)

	t.textView.SetText(fmt.Sprintf("Connecting to %s@%s...\n", t.host.User, t.host.Hostname))

	return nil
}

// RunInteractiveSSH runs SSH in interactive mode (outside tview)
func RunInteractiveSSH(host *ssh.HostEntry) error {
	args := []string{}

	if host.User != "" {
		args = append(args, "-l", host.User)
	}

	if host.Port != 0 && host.Port != 22 {
		args = append(args, "-p", fmt.Sprintf("%d", host.Port))
	}

	if host.KeyFile != "" {
		args = append(args, "-i", host.KeyFile)
	}

	args = append(args, host.Hostname)

	cmd := exec.Command("ssh", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Put terminal in raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	return cmd.Run()
}

// View returns the terminal view
func (t *Terminal) View() *tview.Flex {
	return t.view
}

// Write implements io.Writer for terminal output
func (t *Terminal) Write(p []byte) (n int, err error) {
	t.app.QueueUpdateDraw(func() {
		fmt.Fprint(t.textView, string(p))
		t.textView.ScrollToEnd()
	})
	return len(p), nil
}

var _ io.Writer = (*Terminal)(nil)

// Helper
func colorToTag(c tcell.Color) string {
	r, g, b := c.RGB()
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}
