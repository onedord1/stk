package ui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/systask/systask/internal/config"
)

// Command represents a parsed command
type Command struct {
	Name string
	Args []string
}

// CommandPalette manages command input
type CommandPalette struct {
	view  *tview.InputField
	theme *config.Theme

	// State
	isActive   bool
	history    []string
	historyIdx int

	// Callbacks
	onExecute func(Command)
	onCancel  func()

	// Known commands for autocomplete
	commands []string
}

// NewCommandPalette creates a new command palette
func NewCommandPalette(theme *config.Theme) *CommandPalette {
	cp := &CommandPalette{
		theme:   theme,
		history: []string{},
		commands: []string{
			"theme", "connect", "disconnect", "quit", "q",
			"health", "services", "processes", "logs", "disks", "batch",
			"run", "exec", "ssh", "help", "refresh", "clear",
		},
	}
	cp.build()
	return cp
}

// build constructs the command palette
func (cp *CommandPalette) build() {
	cp.view = tview.NewInputField()
	cp.view.SetBackgroundColor(cp.theme.Background)
	cp.view.SetFieldBackgroundColor(cp.theme.Muted)
	cp.view.SetFieldTextColor(cp.theme.Foreground)
	cp.view.SetLabelColor(cp.theme.Primary)
	cp.view.SetLabel(": ")
	cp.view.SetPlaceholder("Enter command...")
	cp.view.SetPlaceholderTextColor(cp.theme.Border)

	// Autocomplete
	cp.view.SetAutocompleteFunc(func(currentText string) []string {
		if currentText == "" {
			return nil
		}

		var matches []string
		for _, cmd := range cp.commands {
			if strings.HasPrefix(cmd, strings.ToLower(currentText)) {
				matches = append(matches, cmd)
			}
		}
		return matches
	})

	// Handle input
	cp.view.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			text := strings.TrimSpace(cp.view.GetText())
			if text != "" {
				cp.history = append(cp.history, text)
				cp.historyIdx = len(cp.history)

				if cp.onExecute != nil {
					cp.onExecute(cp.parseCommand(text))
				}
			}
			cp.isActive = false
		case tcell.KeyEsc:
			cp.isActive = false
			if cp.onCancel != nil {
				cp.onCancel()
			}
		}
	})

	// History navigation
	cp.view.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyUp:
			if cp.historyIdx > 0 {
				cp.historyIdx--
				cp.view.SetText(cp.history[cp.historyIdx])
			}
			return nil
		case tcell.KeyDown:
			if cp.historyIdx < len(cp.history)-1 {
				cp.historyIdx++
				cp.view.SetText(cp.history[cp.historyIdx])
			} else {
				cp.historyIdx = len(cp.history)
				cp.view.SetText("")
			}
			return nil
		case tcell.KeyTab:
			// Accept autocomplete suggestion
			return event
		}
		return event
	})
}

// parseCommand parses command text into Command struct
func (cp *CommandPalette) parseCommand(text string) Command {
	parts := strings.Fields(text)
	if len(parts) == 0 {
		return Command{}
	}

	return Command{
		Name: strings.ToLower(parts[0]),
		Args: parts[1:],
	}
}

// Activate enables command mode
func (cp *CommandPalette) Activate() {
	cp.isActive = true
	cp.view.SetText("")
	cp.historyIdx = len(cp.history)
}

// Deactivate disables command mode
func (cp *CommandPalette) Deactivate() {
	cp.isActive = false
	cp.view.SetText("")
}

// IsActive returns if command palette is active
func (cp *CommandPalette) IsActive() bool {
	return cp.isActive
}

// OnExecute sets the execution callback
func (cp *CommandPalette) OnExecute(fn func(Command)) {
	cp.onExecute = fn
}

// OnCancel sets the cancel callback
func (cp *CommandPalette) OnCancel(fn func()) {
	cp.onCancel = fn
}

// View returns the underlying view
func (cp *CommandPalette) View() *tview.InputField {
	return cp.view
}

// AddCommand adds a command to autocomplete
func (cp *CommandPalette) AddCommand(cmd string) {
	cp.commands = append(cp.commands, cmd)
}

// SearchPalette manages search/filter input
type SearchPalette struct {
	view     *tview.InputField
	theme    *config.Theme
	isActive bool

	onSearch func(string)
	onCancel func()
}

// NewSearchPalette creates a new search palette
func NewSearchPalette(theme *config.Theme) *SearchPalette {
	sp := &SearchPalette{
		theme: theme,
	}
	sp.build()
	return sp
}

// build constructs the search palette
func (sp *SearchPalette) build() {
	sp.view = tview.NewInputField()
	sp.view.SetBackgroundColor(sp.theme.Background)
	sp.view.SetFieldBackgroundColor(sp.theme.Muted)
	sp.view.SetFieldTextColor(sp.theme.Foreground)
	sp.view.SetLabelColor(sp.theme.Secondary)
	sp.view.SetLabel("/ ")
	sp.view.SetPlaceholder("Search...")
	sp.view.SetPlaceholderTextColor(sp.theme.Border)

	sp.view.SetChangedFunc(func(text string) {
		if sp.onSearch != nil {
			sp.onSearch(text)
		}
	})

	sp.view.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			sp.isActive = false
		case tcell.KeyEsc:
			sp.isActive = false
			if sp.onCancel != nil {
				sp.onCancel()
			}
		}
	})
}

// Activate enables search mode
func (sp *SearchPalette) Activate() {
	sp.isActive = true
	sp.view.SetText("")
}

// Deactivate disables search mode
func (sp *SearchPalette) Deactivate() {
	sp.isActive = false
}

// IsActive returns if search is active
func (sp *SearchPalette) IsActive() bool {
	return sp.isActive
}

// OnSearch sets the search callback
func (sp *SearchPalette) OnSearch(fn func(string)) {
	sp.onSearch = fn
}

// OnCancel sets the cancel callback
func (sp *SearchPalette) OnCancel(fn func()) {
	sp.onCancel = fn
}

// View returns the underlying view
func (sp *SearchPalette) View() *tview.InputField {
	return sp.view
}
