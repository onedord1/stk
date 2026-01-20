package disk

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/systask/systask/internal/config"
	"github.com/systask/systask/internal/ssh"
)

// Filesystem represents a mounted filesystem
type Filesystem struct {
	Device     string
	MountPoint string
	Type       string
	Size       uint64
	Used       uint64
	Available  uint64
	UsePercent float64
}

// Manager manages disk filesystems
type Manager struct {
	view       *tview.Flex
	table      *tview.Table
	detailView *tview.TextView
	barView    *tview.TextView
	theme      *config.Theme
	sshClient  *ssh.Client
	host       *ssh.HostEntry

	filesystems []Filesystem
	selectedIdx int
}

// NewManager creates a new disk manager
func NewManager(theme *config.Theme, client *ssh.Client) *Manager {
	m := &Manager{
		theme:     theme,
		sshClient: client,
	}
	m.build()
	return m
}

// build constructs the disk manager view
func (m *Manager) build() {
	// Filesystem table
	m.table = tview.NewTable()
	m.table.SetBorders(false)
	m.table.SetSelectable(true, false)
	m.table.SetBackgroundColor(m.theme.Background)
	m.table.SetSelectedStyle(tcell.StyleDefault.
		Background(m.theme.Muted).
		Foreground(m.theme.Primary))
	m.table.SetBorder(true)
	m.table.SetBorderColor(m.theme.Border)
	m.table.SetTitle(" Filesystems ")
	m.table.SetTitleColor(m.theme.Primary)

	// Set headers
	headers := []string{"DEVICE", "MOUNT", "TYPE", "SIZE", "USED", "AVAIL", "USE%"}
	for i, h := range headers {
		cell := tview.NewTableCell(h).
			SetTextColor(m.theme.Secondary).
			SetSelectable(false).
			SetExpansion(1)
		if i >= 3 && i <= 6 {
			cell.SetAlign(tview.AlignRight).SetExpansion(0)
		}
		m.table.SetCell(0, i, cell)
	}
	m.table.SetFixed(1, 0)

	// Bar visualization
	m.barView = tview.NewTextView()
	m.barView.SetDynamicColors(true)
	m.barView.SetBorder(true)
	m.barView.SetBorderColor(m.theme.Border)
	m.barView.SetTitle(" Usage Visualization ")
	m.barView.SetTitleColor(m.theme.Primary)
	m.barView.SetBackgroundColor(m.theme.Background)

	// Detail view
	m.detailView = tview.NewTextView()
	m.detailView.SetDynamicColors(true)
	m.detailView.SetBorder(true)
	m.detailView.SetBorderColor(m.theme.Border)
	m.detailView.SetTitle(" Details ")
	m.detailView.SetTitleColor(m.theme.Primary)
	m.detailView.SetBackgroundColor(m.theme.Background)

	// Bottom row
	bottomRow := tview.NewFlex()
	bottomRow.AddItem(m.barView, 0, 1, false)
	bottomRow.AddItem(m.detailView, 0, 1, false)

	// Layout
	m.view = tview.NewFlex().SetDirection(tview.FlexRow)
	m.view.AddItem(m.table, 0, 2, true)
	m.view.AddItem(bottomRow, 12, 0, false)
	m.view.SetBackgroundColor(m.theme.Background)

	// Selection handler
	m.table.SetSelectionChangedFunc(func(row, col int) {
		if row > 0 && row <= len(m.filesystems) {
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
	case 'j':
		return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
	case 'k':
		return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
	case 'u':
		// Show directory usage (du)
		if len(m.filesystems) > 0 {
			m.showDirUsage(m.filesystems[m.selectedIdx].MountPoint)
		}
		return nil
	}

	return event
}

// SetHost sets the current host
func (m *Manager) SetHost(host *ssh.HostEntry) {
	m.host = host
}

// Refresh updates the filesystem list
func (m *Manager) Refresh() error {
	if m.host == nil {
		return fmt.Errorf("no host set")
	}

	// Get filesystem list
	output, err := m.sshClient.RunCommand(*m.host,
		"df -T -B1 --exclude-type=tmpfs --exclude-type=devtmpfs --exclude-type=squashfs 2>/dev/null | tail -n +2")
	if err != nil {
		return err
	}

	m.filesystems = m.parseFilesystems(output)
	m.updateTable()
	m.updateBarView()

	return nil
}

// parseFilesystems parses df output
func (m *Manager) parseFilesystems(output string) []Filesystem {
	var filesystems []Filesystem
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 7 {
			continue
		}

		size, _ := strconv.ParseUint(fields[2], 10, 64)
		used, _ := strconv.ParseUint(fields[3], 10, 64)
		avail, _ := strconv.ParseUint(fields[4], 10, 64)

		percentStr := strings.TrimSuffix(fields[5], "%")
		percent, _ := strconv.ParseFloat(percentStr, 64)

		fs := Filesystem{
			Device:     fields[0],
			Type:       fields[1],
			Size:       size,
			Used:       used,
			Available:  avail,
			UsePercent: percent,
			MountPoint: fields[6],
		}

		filesystems = append(filesystems, fs)
	}

	return filesystems
}

// updateTable updates the filesystem table
func (m *Manager) updateTable() {
	// Clear existing rows (except header)
	for row := m.table.GetRowCount() - 1; row > 0; row-- {
		m.table.RemoveRow(row)
	}

	for i, fs := range m.filesystems {
		row := i + 1

		// Device
		m.table.SetCell(row, 0, tview.NewTableCell(truncate(fs.Device, 30)).
			SetTextColor(m.theme.Foreground).
			SetExpansion(1))

		// Mount point
		m.table.SetCell(row, 1, tview.NewTableCell(truncate(fs.MountPoint, 25)).
			SetTextColor(m.theme.Foreground).
			SetExpansion(1))

		// Type
		m.table.SetCell(row, 2, tview.NewTableCell(fs.Type).
			SetTextColor(m.theme.Muted).
			SetExpansion(0))

		// Size
		m.table.SetCell(row, 3, tview.NewTableCell(formatBytes(fs.Size)).
			SetTextColor(m.theme.Foreground).
			SetAlign(tview.AlignRight).
			SetExpansion(0))

		// Used
		m.table.SetCell(row, 4, tview.NewTableCell(formatBytes(fs.Used)).
			SetTextColor(m.theme.Foreground).
			SetAlign(tview.AlignRight).
			SetExpansion(0))

		// Available
		m.table.SetCell(row, 5, tview.NewTableCell(formatBytes(fs.Available)).
			SetTextColor(m.theme.Success).
			SetAlign(tview.AlignRight).
			SetExpansion(0))

		// Use percent
		percentColor := m.theme.Success
		if fs.UsePercent > 80 {
			percentColor = m.theme.Error
		} else if fs.UsePercent > 60 {
			percentColor = m.theme.Warning
		}
		m.table.SetCell(row, 6, tview.NewTableCell(fmt.Sprintf("%.0f%%", fs.UsePercent)).
			SetTextColor(percentColor).
			SetAlign(tview.AlignRight).
			SetExpansion(0))
	}

	// Select first row if exists
	if m.table.GetRowCount() > 1 {
		m.table.Select(1, 0)
	}
}

// updateBarView updates the usage bar visualization
func (m *Manager) updateBarView() {
	var bars []string
	barWidth := 30

	for _, fs := range m.filesystems {
		color := colorToTag(m.theme.Success)
		if fs.UsePercent > 80 {
			color = colorToTag(m.theme.Error)
		} else if fs.UsePercent > 60 {
			color = colorToTag(m.theme.Warning)
		}

		filled := int(fs.UsePercent / 100 * float64(barWidth))
		if filled > barWidth {
			filled = barWidth
		}

		bar := fmt.Sprintf("[%s]%s[%s]%s[white]",
			color, strings.Repeat("█", filled),
			colorToTag(m.theme.Muted), strings.Repeat("░", barWidth-filled))

		line := fmt.Sprintf("%-20s %s %5.1f%%",
			truncate(fs.MountPoint, 20), bar, fs.UsePercent)
		bars = append(bars, line)
	}

	m.barView.SetText(strings.Join(bars, "\n"))
}

// updateDetails shows selected filesystem details
func (m *Manager) updateDetails() {
	if m.selectedIdx < 0 || m.selectedIdx >= len(m.filesystems) {
		return
	}

	fs := m.filesystems[m.selectedIdx]
	primary := colorToTag(m.theme.Primary)

	// Get additional info
	inodeOutput, _ := m.sshClient.RunCommand(*m.host,
		fmt.Sprintf("df -i %s 2>/dev/null | tail -1", fs.MountPoint))

	var inodeInfo string
	if fields := strings.Fields(inodeOutput); len(fields) >= 5 {
		inodeInfo = fmt.Sprintf("Inodes: %s used, %s free", fields[2], fields[3])
	}

	details := fmt.Sprintf(
		"[%s]Device:[white] %s\n"+
			"[%s]Mount Point:[white] %s\n"+
			"[%s]Type:[white] %s\n"+
			"[%s]Total:[white] %s\n"+
			"[%s]Used:[white] %s (%.1f%%)\n"+
			"[%s]Available:[white] %s\n"+
			"[%s]%s[white]",
		primary, fs.Device,
		primary, fs.MountPoint,
		primary, fs.Type,
		primary, formatBytes(fs.Size),
		primary, formatBytes(fs.Used), fs.UsePercent,
		primary, formatBytes(fs.Available),
		primary, inodeInfo,
	)

	m.detailView.SetText(details)
}

// showDirUsage shows directory usage for a mount point
func (m *Manager) showDirUsage(mountPoint string) {
	if m.host == nil {
		return
	}

	output, err := m.sshClient.RunCommand(*m.host,
		fmt.Sprintf("du -h --max-depth=1 %s 2>/dev/null | sort -rh | head -20", mountPoint))
	if err != nil {
		return
	}

	m.detailView.SetTitle(fmt.Sprintf(" Directory Usage: %s ", mountPoint))
	m.detailView.SetText(output)
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
