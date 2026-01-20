package services

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/systask/systask/internal/config"
	"github.com/systask/systask/internal/ssh"
)

// Service represents a systemd service
type Service struct {
	Name        string
	State       string
	SubState    string
	Description string
	Enabled     bool
	Running     bool
	ActiveSince string
	MainPID     string
	Memory      string
	CPU         string
}

// Manager manages systemd services
type Manager struct {
	view       *tview.Flex
	tabs       *tview.TextView
	table      *tview.Table
	runningTbl *tview.Table
	detailView *tview.TextView
	logsView   *tview.TextView
	actionView *tview.TextView
	pages      *tview.Pages
	logPages   *tview.Pages
	theme      *config.Theme
	sshClient  *ssh.Client
	host       *ssh.HostEntry
	app        *tview.Application

	services    []Service
	running     []Service
	currentTab  string
	selectedIdx int
	filter      string
	showLogs    bool
}

// NewManager creates a new service manager
func NewManager(theme *config.Theme, client *ssh.Client) *Manager {
	m := &Manager{
		theme:      theme,
		sshClient:  client,
		currentTab: "running",
	}
	m.build()
	return m
}

// build constructs the service manager view
func (m *Manager) build() {
	// Tab bar
	m.tabs = tview.NewTextView()
	m.tabs.SetDynamicColors(true)
	m.tabs.SetBackgroundColor(m.theme.Muted)
	m.tabs.SetTextAlign(tview.AlignCenter)
	m.updateTabs()

	// Running services table
	m.runningTbl = m.createTable()
	m.runningTbl.SetInputCapture(m.handleInput)

	// All services table
	m.table = m.createTable()
	m.table.SetInputCapture(m.handleInput)

	// Pages for tabs
	m.pages = tview.NewPages()
	m.pages.AddPage("running", m.runningTbl, true, true)
	m.pages.AddPage("all", m.table, true, false)

	// Content box
	contentBox := tview.NewFlex().SetDirection(tview.FlexRow)
	contentBox.AddItem(m.pages, 0, 1, true)
	contentBox.SetBorder(true)
	contentBox.SetBorderColor(m.theme.Border)
	contentBox.SetTitle(" ⚙️ Systemd Services ")
	contentBox.SetTitleColor(m.theme.Primary)
	contentBox.SetBackgroundColor(m.theme.Background)

	// Detail view (status + info)
	m.detailView = tview.NewTextView()
	m.detailView.SetDynamicColors(true)
	m.detailView.SetScrollable(true)
	m.detailView.SetBorder(true)
	m.detailView.SetBorderColor(m.theme.Border)
	m.detailView.SetTitle(" Service Details ")
	m.detailView.SetTitleColor(m.theme.Primary)
	m.detailView.SetBackgroundColor(m.theme.Background)

	// Logs view
	m.logsView = tview.NewTextView()
	m.logsView.SetDynamicColors(true)
	m.logsView.SetScrollable(true)
	m.logsView.SetBorder(true)
	m.logsView.SetBorderColor(m.theme.Border)
	m.logsView.SetTitle(" Live Logs ")
	m.logsView.SetTitleColor(m.theme.Primary)
	m.logsView.SetBackgroundColor(m.theme.Background)

	// Log pages (switch between detail and logs)
	m.logPages = tview.NewPages()
	m.logPages.AddPage("detail", m.detailView, true, true)
	m.logPages.AddPage("logs", m.logsView, true, false)

	// Action bar
	m.actionView = tview.NewTextView()
	m.actionView.SetDynamicColors(true)
	m.actionView.SetBackgroundColor(m.theme.Muted)
	m.actionView.SetText(m.getActionsText())

	// Right panel
	rightPanel := tview.NewFlex().SetDirection(tview.FlexRow)
	rightPanel.AddItem(m.logPages, 0, 1, false)
	rightPanel.AddItem(m.actionView, 3, 0, false)

	// Main content
	mainContent := tview.NewFlex()
	mainContent.AddItem(contentBox, 0, 2, true)
	mainContent.AddItem(rightPanel, 0, 1, false)

	// Layout
	m.view = tview.NewFlex().SetDirection(tview.FlexRow)
	m.view.AddItem(m.tabs, 1, 0, false)
	m.view.AddItem(mainContent, 0, 1, true)
	m.view.SetBackgroundColor(m.theme.Background)
}

// createTable creates a service table
func (m *Manager) createTable() *tview.Table {
	table := tview.NewTable()
	table.SetBorders(false)
	table.SetSelectable(true, false)
	table.SetBackgroundColor(m.theme.Background)
	table.SetSelectedStyle(tcell.StyleDefault.
		Background(m.theme.Muted).
		Foreground(m.theme.Primary))

	headers := []string{"STATUS", "SERVICE", "ENABLED", "DESCRIPTION"}
	for i, h := range headers {
		cell := tview.NewTableCell(h).
			SetTextColor(m.theme.Secondary).
			SetSelectable(false).
			SetExpansion(1)
		if i == 0 || i == 2 {
			cell.SetExpansion(0)
		}
		table.SetCell(0, i, cell)
	}
	table.SetFixed(1, 0)

	return table
}

// updateTabs updates the tab display
func (m *Manager) updateTabs() {
	runningStyle := "[white]"
	allStyle := "[white]"

	runningCount := len(m.running)
	allCount := len(m.services)

	if m.currentTab == "running" {
		runningStyle = fmt.Sprintf("[%s::b]", colorToTag(m.theme.Primary))
	} else {
		allStyle = fmt.Sprintf("[%s::b]", colorToTag(m.theme.Primary))
	}

	m.tabs.SetText(fmt.Sprintf("  %s● Running (%d) (1)[white]  │  %s◉ All Services (%d) (2)[white]  ",
		runningStyle, runningCount, allStyle, allCount))
}

// getActionsText returns the actions help text
func (m *Manager) getActionsText() string {
	highlight := colorToTag(m.theme.Highlight)
	logsKey := "L=Logs"
	if m.showLogs {
		logsKey = "L=Details"
	}
	return fmt.Sprintf(
		"  [%s]s[white]=Start  [%s]S[white]=Stop  [%s]r[white]=Restart  [%s]e[white]=Enable  [%s]d[white]=Disable  [%s]%s[white]  [%s]Tab[white]=Switch",
		highlight, highlight, highlight, highlight, highlight, highlight, logsKey, highlight)
}

// handleInput handles table input
func (m *Manager) handleInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyTab:
		if m.currentTab == "running" {
			m.currentTab = "all"
			m.pages.SwitchToPage("all")
		} else {
			m.currentTab = "running"
			m.pages.SwitchToPage("running")
		}
		m.updateTabs()
		return nil
	}

	services := m.services
	table := m.table
	if m.currentTab == "running" {
		services = m.running
		table = m.runningTbl
	}

	if len(services) == 0 {
		return event
	}

	row, _ := table.GetSelection()
	if row <= 0 || row > len(services) {
		return event
	}
	service := services[row-1]

	switch event.Rune() {
	case 's': // Start
		m.serviceAction("start", service.Name)
		return nil
	case 'S': // Stop
		m.serviceAction("stop", service.Name)
		return nil
	case 'r': // Restart
		m.serviceAction("restart", service.Name)
		return nil
	case 'e': // Enable
		m.serviceAction("enable", service.Name)
		return nil
	case 'd': // Disable
		m.serviceAction("disable", service.Name)
		return nil
	case 'l', 'L': // Toggle logs
		m.toggleLogs(service.Name)
		return nil
	case 'f': // Follow logs
		m.followLogs(service.Name)
		return nil
	case '/': // Filter
		m.showFilter()
		return nil
	case 'j':
		return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
	case 'k':
		return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
	case '1':
		m.currentTab = "running"
		m.updateTabs()
		m.pages.SwitchToPage("running")
		return nil
	case '2':
		m.currentTab = "all"
		m.updateTabs()
		m.pages.SwitchToPage("all")
		return nil
	}

	return event
}

// SetHost sets the current host
func (m *Manager) SetHost(host *ssh.HostEntry) {
	m.host = host
}

// SetApp sets the tview application reference
func (m *Manager) SetApp(app *tview.Application) {
	m.app = app
}

// Refresh updates the service list
func (m *Manager) Refresh() error {
	if m.host == nil {
		return fmt.Errorf("no host set")
	}

	// Get all services
	output, err := m.sshClient.RunCommand(*m.host,
		`systemctl list-units --type=service --all --no-pager --no-legend | awk '{print $1"|"$2"|"$3"|"$4"|"$5" "$6" "$7" "$8" "$9}'`)
	if err != nil {
		return err
	}

	m.services = m.parseServices(output)

	// Filter running services
	m.running = make([]Service, 0)
	for _, s := range m.services {
		if s.Running {
			m.running = append(m.running, s)
		}
	}

	m.updateTable(m.table, m.services)
	m.updateTable(m.runningTbl, m.running)
	m.updateTabs()

	// Show summary
	enabled := 0
	for _, s := range m.services {
		if s.Enabled {
			enabled++
		}
	}
	m.detailView.SetText(fmt.Sprintf(
		"[%s]Systemd Services Summary[white]\n\n"+
			"Total Services: %d\n"+
			"[%s]Running: %d[white]\n"+
			"Enabled: %d\n\n"+
			"[%s]Select a service for details[white]",
		colorToTag(m.theme.Primary), len(m.services),
		colorToTag(m.theme.Success), len(m.running),
		enabled,
		colorToTag(m.theme.Muted)))

	// Set up selection handler for both tables
	m.runningTbl.SetSelectionChangedFunc(func(row, col int) {
		if row > 0 && row <= len(m.running) {
			m.showServiceDetails(m.running[row-1])
		}
	})

	m.table.SetSelectionChangedFunc(func(row, col int) {
		if row > 0 && row <= len(m.services) {
			m.showServiceDetails(m.services[row-1])
		}
	})

	return nil
}

// parseServices parses systemctl output
func (m *Manager) parseServices(output string) []Service {
	var services []Service
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Split(line, "|")
		if len(fields) < 5 {
			continue
		}

		name := strings.TrimSuffix(fields[0], ".service")

		service := Service{
			Name:        name,
			State:       fields[2],
			SubState:    fields[3],
			Description: fields[4],
			Running:     fields[3] == "running",
		}

		// Check if enabled
		if strings.Contains(fields[1], "enabled") || strings.Contains(strings.ToLower(fields[2]), "active") {
			service.Enabled = true
		}

		services = append(services, service)
	}

	// Sort by name
	sort.Slice(services, func(i, j int) bool {
		return services[i].Name < services[j].Name
	})

	return services
}

// updateTable updates the service table
func (m *Manager) updateTable(table *tview.Table, services []Service) {
	// Clear existing rows (except header)
	for row := table.GetRowCount() - 1; row > 0; row-- {
		table.RemoveRow(row)
	}

	for i, svc := range services {
		row := i + 1

		// Status icon
		icon := "○"
		iconColor := m.theme.Muted
		if svc.Running {
			icon = "●"
			iconColor = m.theme.Success
		} else if svc.State == "failed" {
			icon = "✗"
			iconColor = m.theme.Error
		}
		table.SetCell(row, 0, tview.NewTableCell(icon).
			SetTextColor(iconColor).
			SetExpansion(0))

		// Name
		table.SetCell(row, 1, tview.NewTableCell(svc.Name).
			SetTextColor(m.theme.Foreground).
			SetExpansion(1))

		// Enabled
		enabled := ""
		enabledColor := m.theme.Muted
		if svc.Enabled {
			enabled = "✓"
			enabledColor = m.theme.Secondary
		}
		table.SetCell(row, 2, tview.NewTableCell(enabled).
			SetTextColor(enabledColor).
			SetExpansion(0))

		// Description
		table.SetCell(row, 3, tview.NewTableCell(truncate(svc.Description, 40)).
			SetTextColor(m.theme.Muted).
			SetExpansion(1))
	}

	if table.GetRowCount() > 1 {
		table.Select(1, 0)
	}
}

// showServiceDetails shows service status details
func (m *Manager) showServiceDetails(service Service) {
	m.showLogs = false
	m.logPages.SwitchToPage("detail")
	m.actionView.SetText(m.getActionsText())

	// Get detailed status
	output, _ := m.sshClient.RunCommand(*m.host,
		fmt.Sprintf("systemctl status %s.service 2>&1 | head -20", service.Name))

	primary := colorToTag(m.theme.Primary)
	muted := colorToTag(m.theme.Muted)

	status := "Stopped"
	statusColor := m.theme.Error
	if service.Running {
		status = "Running"
		statusColor = m.theme.Success
	}

	enabled := "Disabled"
	if service.Enabled {
		enabled = "Enabled"
	}

	details := fmt.Sprintf(
		"[%s]Service: %s[white]\n\n"+
			"[%s]Status:[white] [%s]%s[white]\n"+
			"[%s]Enabled:[white] %s\n"+
			"[%s]State:[white] %s (%s)\n\n"+
			"[%s]Status Output:[white]\n%s",
		primary, service.Name,
		primary, colorToTag(statusColor), status,
		primary, enabled,
		primary, service.State, service.SubState,
		muted, output,
	)

	m.detailView.SetTitle(fmt.Sprintf(" Service: %s ", service.Name))
	m.detailView.SetText(details)
}

// toggleLogs toggles between details and logs view
func (m *Manager) toggleLogs(serviceName string) {
	m.showLogs = !m.showLogs
	m.actionView.SetText(m.getActionsText())

	if m.showLogs {
		m.showServiceLogs(serviceName)
		m.logPages.SwitchToPage("logs")
	} else {
		m.logPages.SwitchToPage("detail")
	}
}

// showServiceLogs shows service logs
func (m *Manager) showServiceLogs(serviceName string) {
	output, _ := m.sshClient.RunCommand(*m.host,
		fmt.Sprintf("journalctl -u %s.service --no-pager -n 50 2>&1", serviceName))

	m.logsView.SetTitle(fmt.Sprintf(" Logs: %s ", serviceName))
	m.logsView.SetText(output)
}

// followLogs shows continuously updated logs
func (m *Manager) followLogs(serviceName string) {
	m.showLogs = true
	m.logPages.SwitchToPage("logs")

	m.logsView.SetTitle(fmt.Sprintf(" Live Logs: %s (updating...) ", serviceName))

	// Get more logs
	output, _ := m.sshClient.RunCommand(*m.host,
		fmt.Sprintf("journalctl -u %s.service --no-pager -n 100 2>&1", serviceName))
	m.logsView.SetText(output)
}

// serviceAction performs a service action
func (m *Manager) serviceAction(action, name string) {
	cmd := fmt.Sprintf("sudo systemctl %s %s.service 2>&1", action, name)
	output, err := m.sshClient.RunCommand(*m.host, cmd)

	if err != nil {
		m.detailView.SetText(fmt.Sprintf("[%s]Error: %v[white]\n\n%s",
			colorToTag(m.theme.Error), err, output))
	} else {
		m.detailView.SetText(fmt.Sprintf("[%s]✓ %s %s: success[white]\n\n%s",
			colorToTag(m.theme.Success), action, name, output))
		m.Refresh()
	}
}

// showFilter shows filter input
func (m *Manager) showFilter() {
	m.detailView.SetText(fmt.Sprintf(
		"[%s]Filter Services[white]\n\n"+
			"Use the search command:\n  /<search term>",
		colorToTag(m.theme.Primary)))
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
