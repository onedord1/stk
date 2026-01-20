package processes

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/systask/systask/internal/config"
	"github.com/systask/systask/internal/ssh"
)

// Process represents a running process
type Process struct {
	PID     int
	User    string
	CPU     float64
	Memory  float64
	VSZ     uint64
	RSS     uint64
	TTY     string
	State   string
	Started string
	Command string
}

// SortBy represents sort criteria
type SortBy int

const (
	SortByCPU SortBy = iota
	SortByMemory
	SortByPID
	SortByName
)

// Manager manages processes
type Manager struct {
	view       *tview.Flex
	table      *tview.Table
	detailView *tview.TextView
	theme      *config.Theme
	sshClient  *ssh.Client
	host       *ssh.HostEntry
	app        *tview.Application

	processes   []Process
	selectedIdx int
	sortBy      SortBy
	filter      string

	onKill func(pid int, signal string)
}

// NewManager creates a new process manager
func NewManager(theme *config.Theme, client *ssh.Client) *Manager {
	m := &Manager{
		theme:     theme,
		sshClient: client,
		sortBy:    SortByCPU,
	}
	m.build()
	return m
}

// build constructs the process manager view
func (m *Manager) build() {
	// Process table
	m.table = tview.NewTable()
	m.table.SetBorders(false)
	m.table.SetSelectable(true, false)
	m.table.SetBackgroundColor(m.theme.Background)
	m.table.SetSelectedStyle(tcell.StyleDefault.
		Background(m.theme.Muted).
		Foreground(m.theme.Primary))
	m.table.SetBorder(true)
	m.table.SetBorderColor(m.theme.Border)
	m.table.SetTitle(" Processes (top 50) ")
	m.table.SetTitleColor(m.theme.Primary)

	// Set headers
	headers := []string{"PID", "USER", "CPU%", "MEM%", "STATE", "COMMAND"}
	for i, h := range headers {
		cell := tview.NewTableCell(h).
			SetTextColor(m.theme.Secondary).
			SetSelectable(false).
			SetExpansion(1)
		if i < 4 {
			cell.SetExpansion(0).SetAlign(tview.AlignRight)
		}
		m.table.SetCell(0, i, cell)
	}
	m.table.SetFixed(1, 0)

	// Detail view
	m.detailView = tview.NewTextView()
	m.detailView.SetDynamicColors(true)
	m.detailView.SetBorder(true)
	m.detailView.SetBorderColor(m.theme.Border)
	m.detailView.SetTitle(" Process Details ")
	m.detailView.SetTitleColor(m.theme.Primary)
	m.detailView.SetBackgroundColor(m.theme.Background)

	// Layout
	m.view = tview.NewFlex().SetDirection(tview.FlexRow)
	m.view.AddItem(m.table, 0, 3, true)
	m.view.AddItem(m.detailView, 8, 0, false)
	m.view.SetBackgroundColor(m.theme.Background)

	// Selection handler
	m.table.SetSelectionChangedFunc(func(row, col int) {
		if row > 0 && row <= len(m.processes) {
			m.selectedIdx = row - 1
			m.updateDetails()
		}
	})

	// Key handler
	m.table.SetInputCapture(m.handleInput)
}

// handleInput processes key events
func (m *Manager) handleInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case 'k':
		if event.Modifiers() == tcell.ModNone {
			return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
		}
		// Kill with SIGTERM
		if len(m.processes) > 0 {
			m.killProcess(m.processes[m.selectedIdx].PID, "SIGTERM")
		}
		return nil
	case 'K':
		// Kill with SIGKILL
		if len(m.processes) > 0 {
			m.killProcess(m.processes[m.selectedIdx].PID, "SIGKILL")
		}
		return nil
	case 'j':
		return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
	case 't':
		m.sortBy = SortByCPU
		m.sortProcesses()
		m.updateTable()
		return nil
	case 'm':
		m.sortBy = SortByMemory
		m.sortProcesses()
		m.updateTable()
		return nil
	case 'p':
		m.sortBy = SortByPID
		m.sortProcesses()
		m.updateTable()
		return nil
	case 'n':
		m.sortBy = SortByName
		m.sortProcesses()
		m.updateTable()
		return nil
	}

	return event
}

// killProcess sends a signal to a process
func (m *Manager) killProcess(pid int, signal string) {
	if m.host == nil {
		return
	}

	if m.onKill != nil {
		m.onKill(pid, signal)
	}

	cmd := fmt.Sprintf("sudo kill -%s %d", signal, pid)
	m.sshClient.RunCommand(*m.host, cmd)
	m.Refresh()
}

// SetHost sets the current host
func (m *Manager) SetHost(host *ssh.HostEntry) {
	m.host = host
}

// SetApp sets the tview application reference
func (m *Manager) SetApp(app *tview.Application) {
	m.app = app
}

// Focus sets focus to the process table
func (m *Manager) Focus() {
	if m.app != nil {
		m.app.SetFocus(m.table)
	}
}

// Refresh updates the process list
func (m *Manager) Refresh() error {
	if m.host == nil {
		return fmt.Errorf("no host set")
	}

	// Get process list using ps
	output, err := m.sshClient.RunCommand(*m.host,
		"ps aux --no-headers | head -50")
	if err != nil {
		return err
	}

	m.processes = m.parseProcesses(output)
	m.sortProcesses()
	m.updateTable()

	return nil
}

// parseProcesses parses ps aux output
func (m *Manager) parseProcesses(output string) []Process {
	var processes []Process
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 11 {
			continue
		}

		pid, _ := strconv.Atoi(fields[1])
		cpu, _ := strconv.ParseFloat(fields[2], 64)
		mem, _ := strconv.ParseFloat(fields[3], 64)
		vsz, _ := strconv.ParseUint(fields[4], 10, 64)
		rss, _ := strconv.ParseUint(fields[5], 10, 64)

		// Command is the rest of the line
		command := strings.Join(fields[10:], " ")

		proc := Process{
			PID:     pid,
			User:    fields[0],
			CPU:     cpu,
			Memory:  mem,
			VSZ:     vsz,
			RSS:     rss,
			TTY:     fields[6],
			State:   fields[7],
			Started: fields[8] + " " + fields[9],
			Command: command,
		}

		processes = append(processes, proc)
	}

	return processes
}

// sortProcesses sorts by current criteria
func (m *Manager) sortProcesses() {
	switch m.sortBy {
	case SortByCPU:
		sort.Slice(m.processes, func(i, j int) bool {
			return m.processes[i].CPU > m.processes[j].CPU
		})
	case SortByMemory:
		sort.Slice(m.processes, func(i, j int) bool {
			return m.processes[i].Memory > m.processes[j].Memory
		})
	case SortByPID:
		sort.Slice(m.processes, func(i, j int) bool {
			return m.processes[i].PID < m.processes[j].PID
		})
	case SortByName:
		sort.Slice(m.processes, func(i, j int) bool {
			return m.processes[i].Command < m.processes[j].Command
		})
	}
}

// updateTable updates the process table
func (m *Manager) updateTable() {
	// Clear existing rows (except header)
	for row := m.table.GetRowCount() - 1; row > 0; row-- {
		m.table.RemoveRow(row)
	}

	for i, proc := range m.processes {
		// Apply filter
		if m.filter != "" && !strings.Contains(strings.ToLower(proc.Command), strings.ToLower(m.filter)) {
			continue
		}

		row := i + 1

		// PID
		m.table.SetCell(row, 0, tview.NewTableCell(fmt.Sprintf("%6d", proc.PID)).
			SetTextColor(m.theme.Foreground).
			SetAlign(tview.AlignRight).
			SetExpansion(0))

		// User
		m.table.SetCell(row, 1, tview.NewTableCell(truncate(proc.User, 10)).
			SetTextColor(m.theme.Muted).
			SetExpansion(0))

		// CPU%
		cpuColor := m.theme.Foreground
		if proc.CPU > 50 {
			cpuColor = m.theme.Warning
		}
		if proc.CPU > 80 {
			cpuColor = m.theme.Error
		}
		m.table.SetCell(row, 2, tview.NewTableCell(fmt.Sprintf("%5.1f", proc.CPU)).
			SetTextColor(cpuColor).
			SetAlign(tview.AlignRight).
			SetExpansion(0))

		// MEM%
		memColor := m.theme.Foreground
		if proc.Memory > 20 {
			memColor = m.theme.Warning
		}
		if proc.Memory > 50 {
			memColor = m.theme.Error
		}
		m.table.SetCell(row, 3, tview.NewTableCell(fmt.Sprintf("%5.1f", proc.Memory)).
			SetTextColor(memColor).
			SetAlign(tview.AlignRight).
			SetExpansion(0))

		// State
		m.table.SetCell(row, 4, tview.NewTableCell(proc.State).
			SetTextColor(m.theme.Muted).
			SetExpansion(0))

		// Command
		m.table.SetCell(row, 5, tview.NewTableCell(truncate(proc.Command, 60)).
			SetTextColor(m.theme.Foreground).
			SetExpansion(1))
	}

	// Select first row if exists
	if m.table.GetRowCount() > 1 {
		m.table.Select(1, 0)
	}
}

// updateDetails shows selected process details
func (m *Manager) updateDetails() {
	if m.selectedIdx < 0 || m.selectedIdx >= len(m.processes) {
		return
	}

	proc := m.processes[m.selectedIdx]
	primary := colorToTag(m.theme.Primary)

	details := fmt.Sprintf(
		"[%s]PID:[white] %d    [%s]User:[white] %s    [%s]State:[white] %s\n"+
			"[%s]CPU:[white] %.1f%%    [%s]Memory:[white] %.1f%%    [%s]RSS:[white] %s\n\n"+
			"[%s]Command:[white] %s",
		primary, proc.PID, primary, proc.User, primary, proc.State,
		primary, proc.CPU, primary, proc.Memory, primary, formatBytes(proc.RSS*1024),
		primary, proc.Command,
	)

	m.detailView.SetText(details)
}

// SetFilter sets the process filter
func (m *Manager) SetFilter(filter string) {
	m.filter = filter
	m.updateTable()
}

// OnKill sets the kill callback
func (m *Manager) OnKill(fn func(pid int, signal string)) {
	m.onKill = fn
}

// View returns the manager view
func (m *Manager) View() *tview.Flex {
	return m.view
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

func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
