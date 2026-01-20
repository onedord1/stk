package batch

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/systask/systask/internal/config"
	"github.com/systask/systask/internal/ssh"
)

// ExecutionState represents command execution state
type ExecutionState int

const (
	StatePending ExecutionState = iota
	StateRunning
	StateSuccess
	StateError
)

// HostResult represents result for one host
type HostResult struct {
	Host   ssh.HostEntry
	State  ExecutionState
	Output string
	Error  error
	Time   time.Duration
}

// Executor manages batch command execution
type Executor struct {
	view        *tview.Flex
	hostList    *tview.List
	commandView *tview.InputField
	outputView  *tview.TextView
	statusTable *tview.Table
	theme       *config.Theme
	sshClient   *ssh.Client
	sshExecutor *ssh.Executor

	hosts         []ssh.HostEntry
	selectedHosts map[int]bool
	results       []HostResult
	mu            sync.Mutex

	app *tview.Application
}

// NewExecutor creates a new batch executor
func NewExecutor(theme *config.Theme, client *ssh.Client, executor *ssh.Executor) *Executor {
	e := &Executor{
		theme:         theme,
		sshClient:     client,
		sshExecutor:   executor,
		selectedHosts: make(map[int]bool),
	}
	e.build()
	return e
}

// build constructs the batch executor view
func (e *Executor) build() {
	// Host selection list
	e.hostList = tview.NewList()
	e.hostList.ShowSecondaryText(false)
	e.hostList.SetBackgroundColor(e.theme.Background)
	e.hostList.SetBorder(true)
	e.hostList.SetBorderColor(e.theme.Border)
	e.hostList.SetTitle(" Select Hosts (Space to toggle) ")
	e.hostList.SetTitleColor(e.theme.Primary)
	e.hostList.SetMainTextColor(e.theme.Foreground)
	e.hostList.SetSelectedBackgroundColor(e.theme.Muted)
	e.hostList.SetSelectedTextColor(e.theme.Primary)

	// Command input
	e.commandView = tview.NewInputField()
	e.commandView.SetLabel(" Command: ")
	e.commandView.SetLabelColor(e.theme.Primary)
	e.commandView.SetFieldBackgroundColor(e.theme.Muted)
	e.commandView.SetFieldTextColor(e.theme.Foreground)
	e.commandView.SetBackgroundColor(e.theme.Background)
	e.commandView.SetPlaceholder("Enter command to execute on selected hosts...")
	e.commandView.SetPlaceholderTextColor(e.theme.Border)
	e.commandView.SetBorder(true)
	e.commandView.SetBorderColor(e.theme.Border)

	// Status table
	e.statusTable = tview.NewTable()
	e.statusTable.SetBorders(false)
	e.statusTable.SetSelectable(true, false)
	e.statusTable.SetBackgroundColor(e.theme.Background)
	e.statusTable.SetSelectedStyle(tcell.StyleDefault.
		Background(e.theme.Muted).
		Foreground(e.theme.Primary))
	e.statusTable.SetBorder(true)
	e.statusTable.SetBorderColor(e.theme.Border)
	e.statusTable.SetTitle(" Execution Status ")
	e.statusTable.SetTitleColor(e.theme.Primary)

	// Headers
	headers := []string{"STATUS", "HOST", "TIME", "MESSAGE"}
	for i, h := range headers {
		cell := tview.NewTableCell(h).
			SetTextColor(e.theme.Secondary).
			SetSelectable(false).
			SetExpansion(1)
		if i == 0 || i == 2 {
			cell.SetExpansion(0)
		}
		e.statusTable.SetCell(0, i, cell)
	}
	e.statusTable.SetFixed(1, 0)

	// Output view
	e.outputView = tview.NewTextView()
	e.outputView.SetDynamicColors(true)
	e.outputView.SetScrollable(true)
	e.outputView.SetBackgroundColor(e.theme.Background)
	e.outputView.SetBorder(true)
	e.outputView.SetBorderColor(e.theme.Border)
	e.outputView.SetTitle(" Output ")
	e.outputView.SetTitleColor(e.theme.Primary)

	// Layout
	leftPanel := tview.NewFlex().SetDirection(tview.FlexRow)
	leftPanel.AddItem(e.hostList, 0, 1, true)
	leftPanel.AddItem(e.commandView, 3, 0, false)

	rightPanel := tview.NewFlex().SetDirection(tview.FlexRow)
	rightPanel.AddItem(e.statusTable, 0, 1, false)
	rightPanel.AddItem(e.outputView, 0, 1, false)

	e.view = tview.NewFlex()
	e.view.AddItem(leftPanel, 0, 1, true)
	e.view.AddItem(rightPanel, 0, 2, false)
	e.view.SetBackgroundColor(e.theme.Background)

	// Key handlers
	e.hostList.SetInputCapture(e.handleHostListInput)
	e.commandView.SetInputCapture(e.handleCommandInput)
	e.commandView.SetDoneFunc(e.handleCommandDone)
	e.statusTable.SetSelectionChangedFunc(e.handleStatusSelect)
}

// handleHostListInput handles host list key events
func (e *Executor) handleHostListInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyRune:
		switch event.Rune() {
		case ' ':
			// Toggle selection
			idx := e.hostList.GetCurrentItem()
			e.toggleHost(idx)
			return nil
		case 'a':
			// Select all
			e.selectAll()
			return nil
		case 'n':
			// Select none
			e.selectNone()
			return nil
		case 'j':
			return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
		case 'k':
			return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
		}
	case tcell.KeyTab:
		// Focus command input
		if e.app != nil {
			e.app.SetFocus(e.commandView)
		}
		return nil
	}

	return event
}

// handleCommandInput handles command field input
func (e *Executor) handleCommandInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyTab:
		// Focus back to host list
		if e.app != nil {
			e.app.SetFocus(e.hostList)
		}
		return nil
	case tcell.KeyBacktab:
		// Focus back to host list
		if e.app != nil {
			e.app.SetFocus(e.hostList)
		}
		return nil
	}
	return event
}

// handleCommandDone handles command input completion
func (e *Executor) handleCommandDone(key tcell.Key) {
	if key == tcell.KeyEnter {
		command := e.commandView.GetText()
		if command != "" {
			e.executeCommand(command)
		}
	}
}

// handleStatusSelect handles status table selection
func (e *Executor) handleStatusSelect(row, col int) {
	if row > 0 && row <= len(e.results) {
		result := e.results[row-1]
		e.showOutput(result)
	}
}

// toggleHost toggles host selection
func (e *Executor) toggleHost(idx int) {
	if idx < 0 || idx >= len(e.hosts) {
		return
	}

	if e.selectedHosts[idx] {
		delete(e.selectedHosts, idx)
	} else {
		e.selectedHosts[idx] = true
	}

	e.updateHostList()
}

// selectAll selects all hosts
func (e *Executor) selectAll() {
	for i := range e.hosts {
		e.selectedHosts[i] = true
	}
	e.updateHostList()
}

// selectNone deselects all hosts
func (e *Executor) selectNone() {
	e.selectedHosts = make(map[int]bool)
	e.updateHostList()
}

// updateHostList updates the host list display
func (e *Executor) updateHostList() {
	current := e.hostList.GetCurrentItem()
	e.hostList.Clear()

	for i, host := range e.hosts {
		icon := "○"
		if e.selectedHosts[i] {
			icon = "●"
		}

		name := host.Name
		if name == "" {
			name = host.Hostname
		}

		text := fmt.Sprintf("%s %s", icon, name)
		shortcut := rune(0)
		if i < 9 {
			shortcut = rune('1' + i)
		}

		e.hostList.AddItem(text, "", shortcut, nil)
	}

	// Restore selection
	if current >= 0 && current < e.hostList.GetItemCount() {
		e.hostList.SetCurrentItem(current)
	}
}

// SetHosts sets the available hosts
func (e *Executor) SetHosts(hosts []ssh.HostEntry) {
	e.hosts = hosts
	e.updateHostList()
}

// SetApp sets the tview application reference
func (e *Executor) SetApp(app *tview.Application) {
	e.app = app
}

// Focus sets focus to the host list for Tab navigation
func (e *Executor) Focus() {
	if e.app != nil {
		e.app.SetFocus(e.hostList)
	}
}

// executeCommand runs command on selected hosts
func (e *Executor) executeCommand(command string) {
	// Get selected hosts
	var selectedHosts []ssh.HostEntry
	for idx := range e.selectedHosts {
		if idx < len(e.hosts) {
			selectedHosts = append(selectedHosts, e.hosts[idx])
		}
	}

	if len(selectedHosts) == 0 {
		e.outputView.SetText(fmt.Sprintf("[%s]No hosts selected![white]", colorToTag(e.theme.Error)))
		return
	}

	// Initialize results
	e.mu.Lock()
	e.results = make([]HostResult, len(selectedHosts))
	for i, host := range selectedHosts {
		e.results[i] = HostResult{
			Host:  host,
			State: StatePending,
		}
	}
	e.mu.Unlock()

	e.updateStatusTable()
	e.outputView.SetText(fmt.Sprintf("Executing: %s\n\nRunning on %d hosts...", command, len(selectedHosts)))

	// Execute in parallel
	go func() {
		ctx := context.Background()
		var wg sync.WaitGroup

		for i, host := range selectedHosts {
			wg.Add(1)
			go func(idx int, h ssh.HostEntry) {
				defer wg.Done()

				// Update to running
				e.mu.Lock()
				e.results[idx].State = StateRunning
				e.mu.Unlock()

				if e.app != nil {
					e.app.QueueUpdateDraw(func() {
						e.updateStatusTable()
					})
				}

				// Execute command
				start := time.Now()
				output, err := e.sshClient.RunCommand(h, command)
				elapsed := time.Since(start)

				// Update result
				e.mu.Lock()
				e.results[idx].Time = elapsed
				e.results[idx].Output = output
				if err != nil {
					e.results[idx].State = StateError
					e.results[idx].Error = err
				} else {
					e.results[idx].State = StateSuccess
				}
				e.mu.Unlock()

				if e.app != nil {
					e.app.QueueUpdateDraw(func() {
						e.updateStatusTable()
					})
				}
			}(i, host)
		}

		wg.Wait()

		// Wait for context cancellation
		select {
		case <-ctx.Done():
		default:
		}

		if e.app != nil {
			e.app.QueueUpdateDraw(func() {
				e.showSummary()
			})
		}
	}()
}

// updateStatusTable updates the status table
func (e *Executor) updateStatusTable() {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Clear existing rows (except header)
	for row := e.statusTable.GetRowCount() - 1; row > 0; row-- {
		e.statusTable.RemoveRow(row)
	}

	for i, result := range e.results {
		row := i + 1

		// Status icon
		var icon string
		switch result.State {
		case StatePending:
			icon = fmt.Sprintf("[%s]○[white]", colorToTag(e.theme.Muted))
		case StateRunning:
			icon = fmt.Sprintf("[%s]◐[white]", colorToTag(e.theme.Warning))
		case StateSuccess:
			icon = fmt.Sprintf("[%s]●[white]", colorToTag(e.theme.Success))
		case StateError:
			icon = fmt.Sprintf("[%s]✗[white]", colorToTag(e.theme.Error))
		}

		e.statusTable.SetCell(row, 0, tview.NewTableCell(icon).SetExpansion(0))

		// Host name
		name := result.Host.Name
		if name == "" {
			name = result.Host.Hostname
		}
		e.statusTable.SetCell(row, 1, tview.NewTableCell(name).
			SetTextColor(e.theme.Foreground).
			SetExpansion(1))

		// Time
		timeStr := "-"
		if result.Time > 0 {
			timeStr = result.Time.Round(time.Millisecond).String()
		}
		e.statusTable.SetCell(row, 2, tview.NewTableCell(timeStr).
			SetTextColor(e.theme.Muted).
			SetExpansion(0))

		// Message
		msg := ""
		if result.Error != nil {
			msg = result.Error.Error()
		} else if result.State == StateSuccess {
			lines := len(strings.Split(result.Output, "\n"))
			msg = fmt.Sprintf("%d lines", lines)
		}
		e.statusTable.SetCell(row, 3, tview.NewTableCell(truncate(msg, 40)).
			SetTextColor(e.theme.Muted).
			SetExpansion(1))
	}
}

// showOutput shows output for a result
func (e *Executor) showOutput(result HostResult) {
	name := result.Host.Name
	if name == "" {
		name = result.Host.Hostname
	}

	e.outputView.SetTitle(fmt.Sprintf(" Output: %s ", name))

	var text string
	if result.Error != nil {
		text = fmt.Sprintf("[%s]Error: %v[white]\n\n%s",
			colorToTag(e.theme.Error), result.Error, result.Output)
	} else {
		text = result.Output
	}

	e.outputView.SetText(text)
}

// showSummary shows execution summary
func (e *Executor) showSummary() {
	e.mu.Lock()
	defer e.mu.Unlock()

	success := 0
	failed := 0
	var totalTime time.Duration

	for _, r := range e.results {
		if r.State == StateSuccess {
			success++
		} else if r.State == StateError {
			failed++
		}
		totalTime += r.Time
	}

	summary := fmt.Sprintf(
		"[%s]Execution Complete[white]\n\n"+
			"[%s]Success:[white] %d\n"+
			"[%s]Failed:[white] %d\n"+
			"[%s]Total Time:[white] %s\n\n"+
			"Select a host in the status table to view output.",
		colorToTag(e.theme.Primary),
		colorToTag(e.theme.Success), success,
		colorToTag(e.theme.Error), failed,
		colorToTag(e.theme.Secondary), totalTime.Round(time.Millisecond),
	)

	e.outputView.SetTitle(" Summary ")
	e.outputView.SetText(summary)
}

// View returns the executor view
func (e *Executor) View() *tview.Flex {
	return e.view
}

// Helpers
func colorToTag(c tcell.Color) string {
	r, g, b := c.RGB()
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
