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

// BlockDevice represents a block device from lsblk
type BlockDevice struct {
	Name       string
	Size       string
	Type       string // disk, part, lvm, crypt
	Mountpoint string
	FSType     string
	Model      string
	RO         bool   // read-only
	RM         bool   // removable
	Parent     string // parent device for partitions
}

// Manager manages disk filesystems
type Manager struct {
	view        *tview.Flex
	table       *tview.Table
	devTable    *tview.Table
	detailView  *tview.TextView
	barView     *tview.TextView
	actionView  *tview.TextView
	pages       *tview.Pages
	confirmForm *tview.Form
	theme       *config.Theme
	sshClient   *ssh.Client
	host        *ssh.HostEntry
	app         *tview.Application

	filesystems  []Filesystem
	blockDevices []BlockDevice
	selectedIdx  int
	devIdx       int
	currentTab   string // "filesystems" or "devices"
	showConfirm  bool
}

// NewManager creates a new disk manager
func NewManager(theme *config.Theme, client *ssh.Client) *Manager {
	m := &Manager{
		theme:      theme,
		sshClient:  client,
		currentTab: "filesystems",
	}
	m.build()
	return m
}

// build constructs the disk manager view
func (m *Manager) build() {
	// Tab bar
	tabs := tview.NewTextView()
	tabs.SetDynamicColors(true)
	tabs.SetBackgroundColor(m.theme.Background)
	tabs.SetText(m.getTabsText())

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
	m.table.SetTitle(" 💿 Filesystems ")
	m.table.SetTitleColor(m.theme.Primary)

	// Set filesystem headers
	fsHeaders := []string{"DEVICE", "MOUNT", "TYPE", "SIZE", "USED", "AVAIL", "USE%"}
	for i, h := range fsHeaders {
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

	// Block devices table
	m.devTable = tview.NewTable()
	m.devTable.SetBorders(false)
	m.devTable.SetSelectable(true, false)
	m.devTable.SetBackgroundColor(m.theme.Background)
	m.devTable.SetSelectedStyle(tcell.StyleDefault.
		Background(m.theme.Muted).
		Foreground(m.theme.Primary))
	m.devTable.SetBorder(true)
	m.devTable.SetBorderColor(m.theme.Border)
	m.devTable.SetTitle(" 🔧 Block Devices & Partitions ")
	m.devTable.SetTitleColor(m.theme.Primary)

	// Set device headers
	devHeaders := []string{"NAME", "SIZE", "TYPE", "FSTYPE", "MOUNTPOINT", "MODEL"}
	for i, h := range devHeaders {
		cell := tview.NewTableCell(h).
			SetTextColor(m.theme.Secondary).
			SetSelectable(false).
			SetExpansion(1)
		m.devTable.SetCell(0, i, cell)
	}
	m.devTable.SetFixed(1, 0)

	// Pages for switching between tables
	m.pages = tview.NewPages()
	m.pages.AddPage("filesystems", m.table, true, true)
	m.pages.AddPage("devices", m.devTable, true, false)

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

	// Action bar
	m.actionView = tview.NewTextView()
	m.actionView.SetDynamicColors(true)
	m.actionView.SetBackgroundColor(m.theme.Muted)
	m.actionView.SetText(m.getActionText())

	// Bottom row
	bottomRow := tview.NewFlex()
	bottomRow.AddItem(m.barView, 0, 1, false)
	bottomRow.AddItem(m.detailView, 0, 1, false)

	// Top area with tabs
	topArea := tview.NewFlex().SetDirection(tview.FlexRow)
	topArea.AddItem(tabs, 1, 0, false)
	topArea.AddItem(m.pages, 0, 1, true)

	// Layout
	m.view = tview.NewFlex().SetDirection(tview.FlexRow)
	m.view.AddItem(topArea, 0, 2, true)
	m.view.AddItem(bottomRow, 10, 0, false)
	m.view.AddItem(m.actionView, 1, 0, false)
	m.view.SetBackgroundColor(m.theme.Background)

	// Selection handlers
	m.table.SetSelectionChangedFunc(func(row, col int) {
		if row > 0 && row <= len(m.filesystems) {
			m.selectedIdx = row - 1
			m.updateDetails()
		}
	})

	m.devTable.SetSelectionChangedFunc(func(row, col int) {
		if row > 0 && row <= len(m.blockDevices) {
			m.devIdx = row - 1
			m.updateDeviceDetails()
		}
	})

	// Key handlers
	m.table.SetInputCapture(m.handleInput)
	m.devTable.SetInputCapture(m.handleDeviceInput)
}

// getTabsText returns the tab bar text
func (m *Manager) getTabsText() string {
	highlight := colorToTag(m.theme.Highlight)
	muted := colorToTag(m.theme.Muted)

	fs := "Filesystems"
	dev := "Block Devices"

	if m.currentTab == "filesystems" {
		fs = fmt.Sprintf("[%s][ Filesystems ][white]", highlight)
		dev = fmt.Sprintf("[%s]  Block Devices  [white]", muted)
	} else {
		fs = fmt.Sprintf("[%s]  Filesystems  [white]", muted)
		dev = fmt.Sprintf("[%s][ Block Devices ][white]", highlight)
	}

	return fmt.Sprintf("  %s    %s    [%s](Tab to switch)[white]", fs, dev, muted)
}

// getActionText returns the action bar text
func (m *Manager) getActionText() string {
	highlight := colorToTag(m.theme.Highlight)
	if m.currentTab == "filesystems" {
		return fmt.Sprintf("  [%s]u[white]=DirUsage  [%s]Tab[white]=Devices  [%s]r[white]=Refresh  [%s]ESC[white]=Home",
			highlight, highlight, highlight, highlight)
	}
	return fmt.Sprintf("  [%s]c[white]=Create  [%s]d[white]=Delete  [%s]f[white]=Format  [%s]m[white]=Mount/Unmount  [%s]Tab[white]=Filesystems  [%s]ESC[white]=Home",
		highlight, highlight, highlight, highlight, highlight, highlight)
}

// handleInput processes key events for filesystem table
func (m *Manager) handleInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyTab:
		m.switchTab()
		return nil
	}

	switch event.Rune() {
	case 'j':
		return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
	case 'k':
		return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
	case 'u':
		if len(m.filesystems) > 0 {
			m.showDirUsage(m.filesystems[m.selectedIdx].MountPoint)
		}
		return nil
	}

	return event
}

// handleDeviceInput processes key events for device table
func (m *Manager) handleDeviceInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyTab:
		m.switchTab()
		return nil
	}

	switch event.Rune() {
	case 'j':
		return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
	case 'k':
		return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
	case 'c':
		m.showCreatePartition()
		return nil
	case 'd':
		m.showDeletePartition()
		return nil
	case 'f':
		m.showFormatPartition()
		return nil
	case 'm':
		m.toggleMount()
		return nil
	case 'i':
		m.showDeviceInfo()
		return nil
	}

	return event
}

// switchTab switches between filesystem and device views
func (m *Manager) switchTab() {
	if m.currentTab == "filesystems" {
		m.currentTab = "devices"
		m.pages.SwitchToPage("devices")
		if m.app != nil {
			m.app.SetFocus(m.devTable)
		}
	} else {
		m.currentTab = "filesystems"
		m.pages.SwitchToPage("filesystems")
		if m.app != nil {
			m.app.SetFocus(m.table)
		}
	}
	m.actionView.SetText(m.getActionText())
}

// showCreatePartition shows the create partition dialog
func (m *Manager) showCreatePartition() {
	if len(m.blockDevices) == 0 || m.devIdx >= len(m.blockDevices) {
		m.detailView.SetText("[red]No device selected[white]")
		return
	}

	dev := m.blockDevices[m.devIdx]
	primary := colorToTag(m.theme.Primary)
	warning := colorToTag(m.theme.Warning)

	// Only allow creating on disk devices
	if dev.Type != "disk" {
		m.detailView.SetTitle(" ⚠️ Error ")
		m.detailView.SetText(fmt.Sprintf("[%s]Cannot create partition here[white]\n\n"+
			"You can only create partitions on disk devices.\n"+
			"Selected device type: %s\n\n"+
			"Select a disk (type=disk) to create a partition.",
			warning, dev.Type))
		return
	}

	// Show create partition instructions
	m.detailView.SetTitle(" 📝 Create Partition ")
	m.detailView.SetText(fmt.Sprintf(`[%s]Create Partition on %s[white]

[%s]⚠️ WARNING: This operation will modify disk %s[white]

To create a partition, run these commands in Terminal (t):

[%s]1. Open parted:[white]
   :run sudo parted /dev/%s

[%s]2. Create partition (example):[white]
   :run sudo parted /dev/%s mkpart primary ext4 0%% 100%%

[%s]3. Format the new partition:[white]
   :run sudo mkfs.ext4 /dev/%s1

[%s]Press 'r' to refresh after creating partition.[white]`,
		primary, dev.Name,
		warning, dev.Name,
		primary, dev.Name,
		primary, dev.Name,
		primary, dev.Name,
		primary))
}

// showDeletePartition shows delete partition confirmation
func (m *Manager) showDeletePartition() {
	if len(m.blockDevices) == 0 || m.devIdx >= len(m.blockDevices) {
		return
	}

	dev := m.blockDevices[m.devIdx]
	errColor := colorToTag(m.theme.Error)
	warning := colorToTag(m.theme.Warning)
	primary := colorToTag(m.theme.Primary)

	if dev.Type != "part" {
		m.detailView.SetTitle(" ⚠️ Error ")
		m.detailView.SetText(fmt.Sprintf("[%s]Cannot delete this device[white]\n\n"+
			"You can only delete partition devices.\n"+
			"Selected device type: %s",
			warning, dev.Type))
		return
	}

	if dev.Mountpoint != "" {
		m.detailView.SetTitle(" ⚠️ Warning ")
		m.detailView.SetText(fmt.Sprintf("[%s]Partition is mounted![white]\n\n"+
			"Mount point: %s\n\n"+
			"Unmount the partition first with 'm' key,\n"+
			"or run:\n\n"+
			"  :run sudo umount %s",
			warning, dev.Mountpoint, dev.Mountpoint))
		return
	}

	// Show delete confirmation
	m.detailView.SetTitle(" 🗑️ Delete Partition ")
	m.detailView.SetText(fmt.Sprintf(`[%s]╔══════════════════════════════════════╗
║      ⚠️  DANGER! DATA LOSS!  ⚠️       ║
╚══════════════════════════════════════╝[white]

[%s]Partition:[white] /dev/%s
[%s]Size:[white] %s
[%s]FSType:[white] %s

[%s]This will PERMANENTLY DELETE all data![white]

To delete, run in Terminal (t):

[%s]:run sudo parted /dev/%s rm <partition_number>[white]

Example for partition 1:
  :run sudo parted /dev/%s rm 1

[%s]Press 'r' to refresh after deletion.[white]`,
		errColor,
		primary, dev.Name,
		primary, dev.Size,
		primary, dev.FSType,
		errColor,
		primary, dev.Parent,
		dev.Parent,
		primary))
}

// showFormatPartition shows format partition dialog
func (m *Manager) showFormatPartition() {
	if len(m.blockDevices) == 0 || m.devIdx >= len(m.blockDevices) {
		return
	}

	dev := m.blockDevices[m.devIdx]
	errColor := colorToTag(m.theme.Error)
	warning := colorToTag(m.theme.Warning)
	primary := colorToTag(m.theme.Primary)
	highlight := colorToTag(m.theme.Highlight)

	if dev.Type != "part" {
		m.detailView.SetTitle(" ⚠️ Error ")
		m.detailView.SetText(fmt.Sprintf("[%s]Cannot format this device[white]\n\n"+
			"You can only format partition devices.\n"+
			"Selected device type: %s",
			warning, dev.Type))
		return
	}

	if dev.Mountpoint != "" {
		m.detailView.SetTitle(" ⚠️ Warning ")
		m.detailView.SetText(fmt.Sprintf("[%s]Partition is mounted![white]\n\n"+
			"Unmount first with 'm' key before formatting.",
			warning))
		return
	}

	m.detailView.SetTitle(" 📀 Format Partition ")
	m.detailView.SetText(fmt.Sprintf(`[%s]╔══════════════════════════════════════╗
║      ⚠️  DANGER! DATA LOSS!  ⚠️       ║
╚══════════════════════════════════════╝[white]

[%s]Partition:[white] /dev/%s
[%s]Current FS:[white] %s
[%s]Size:[white] %s

[%s]Available filesystem types:[white]
  [%s]ext4[white]   - Linux standard (recommended)
  [%s]xfs[white]    - High performance
  [%s]btrfs[white]  - Modern with snapshots
  [%s]ntfs[white]   - Windows compatible
  [%s]vfat[white]   - FAT32 (USB drives)

[%s]To format, run in Terminal (t):[white]

  :run sudo mkfs.ext4 /dev/%s
  :run sudo mkfs.xfs /dev/%s
  :run sudo mkfs.btrfs /dev/%s

[%s]Press 'r' to refresh after formatting.[white]`,
		errColor,
		primary, dev.Name,
		primary, dev.FSType,
		primary, dev.Size,
		primary,
		highlight, highlight, highlight, highlight, highlight,
		primary,
		dev.Name, dev.Name, dev.Name,
		primary))
}

// toggleMount mounts or unmounts the selected partition
func (m *Manager) toggleMount() {
	if len(m.blockDevices) == 0 || m.devIdx >= len(m.blockDevices) {
		return
	}

	dev := m.blockDevices[m.devIdx]
	primary := colorToTag(m.theme.Primary)
	success := colorToTag(m.theme.Success)
	warning := colorToTag(m.theme.Warning)

	if dev.Type != "part" {
		m.detailView.SetText(fmt.Sprintf("[%s]Select a partition to mount/unmount[white]", warning))
		return
	}

	if dev.Mountpoint != "" {
		// Unmount
		m.detailView.SetTitle(" 📤 Unmounting... ")
		m.detailView.SetText(fmt.Sprintf("[%s]Unmounting /dev/%s from %s...[white]",
			primary, dev.Name, dev.Mountpoint))

		go func() {
			cmd := fmt.Sprintf("sudo umount /dev/%s 2>&1", dev.Name)
			output, err := m.sshClient.RunCommand(*m.host, cmd)
			if m.app != nil {
				m.app.QueueUpdateDraw(func() {
					if err != nil {
						m.detailView.SetTitle(" ❌ Unmount Failed ")
						m.detailView.SetText(fmt.Sprintf("[red]Failed to unmount:[white]\n\n%s", output))
					} else {
						m.detailView.SetTitle(" ✅ Unmounted ")
						m.detailView.SetText(fmt.Sprintf("[%s]Successfully unmounted /dev/%s[white]", success, dev.Name))
						m.refreshDevices()
					}
				})
			}
		}()
	} else {
		// Show mount instructions
		m.detailView.SetTitle(" 📥 Mount Partition ")
		m.detailView.SetText(fmt.Sprintf(`[%s]Mount /dev/%s[white]

To mount this partition:

[%s]1. Create mount point:[white]
   :run sudo mkdir -p /mnt/%s

[%s]2. Mount the partition:[white]
   :run sudo mount /dev/%s /mnt/%s

[%s]For permanent mount, add to /etc/fstab[white]`,
			primary, dev.Name,
			primary, dev.Name,
			primary, dev.Name, dev.Name,
			primary))
	}
}

// showDeviceInfo shows detailed device info
func (m *Manager) showDeviceInfo() {
	if len(m.blockDevices) == 0 || m.devIdx >= len(m.blockDevices) {
		return
	}

	dev := m.blockDevices[m.devIdx]

	// Get detailed info
	output, _ := m.sshClient.RunCommand(*m.host,
		fmt.Sprintf("sudo fdisk -l /dev/%s 2>/dev/null | head -20", strings.Split(dev.Name, " ")[0]))

	m.detailView.SetTitle(fmt.Sprintf(" 📋 Device Info: %s ", dev.Name))
	m.detailView.SetText(output)
}

// SetHost sets the current host
func (m *Manager) SetHost(host *ssh.HostEntry) {
	m.host = host
}

// SetApp sets the tview application reference
func (m *Manager) SetApp(app *tview.Application) {
	m.app = app
}

// Focus sets focus to the current table
func (m *Manager) Focus() {
	if m.app != nil {
		if m.currentTab == "devices" {
			m.app.SetFocus(m.devTable)
		} else {
			m.app.SetFocus(m.table)
		}
	}
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

	// Also refresh block devices
	m.refreshDevices()

	return nil
}

// refreshDevices refreshes the block device list
func (m *Manager) refreshDevices() {
	output, _ := m.sshClient.RunCommand(*m.host,
		`lsblk -o NAME,SIZE,TYPE,FSTYPE,MOUNTPOINT,MODEL -n 2>/dev/null`)

	m.blockDevices = m.parseBlockDevices(output)
	m.updateDeviceTable()
}

// parseBlockDevices parses lsblk output
func (m *Manager) parseBlockDevices(output string) []BlockDevice {
	var devices []BlockDevice
	lines := strings.Split(output, "\n")
	currentDisk := ""

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Parse indented lines for partitions
		indent := 0
		for _, c := range line {
			if c == ' ' || c == '├' || c == '└' || c == '│' || c == '─' {
				indent++
			} else {
				break
			}
		}

		// Clean the line
		line = strings.TrimLeft(line, " ├└│─")
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		dev := BlockDevice{
			Name: fields[0],
			Size: fields[1],
			Type: fields[2],
		}

		if len(fields) > 3 {
			dev.FSType = fields[3]
		}
		if len(fields) > 4 {
			dev.Mountpoint = fields[4]
		}
		if len(fields) > 5 {
			dev.Model = strings.Join(fields[5:], " ")
		}

		// Track parent for partitions
		if dev.Type == "disk" {
			currentDisk = dev.Name
		} else if dev.Type == "part" {
			dev.Parent = currentDisk
		}

		devices = append(devices, dev)
	}

	return devices
}

// updateDeviceTable updates the device table
func (m *Manager) updateDeviceTable() {
	// Clear existing rows
	for row := m.devTable.GetRowCount() - 1; row > 0; row-- {
		m.devTable.RemoveRow(row)
	}

	for i, dev := range m.blockDevices {
		row := i + 1

		// Name with indent for partitions
		namePrefix := ""
		if dev.Type == "part" {
			namePrefix = "  └─ "
		}

		typeColor := m.theme.Foreground
		if dev.Type == "disk" {
			typeColor = m.theme.Primary
		} else if dev.Type == "part" {
			typeColor = m.theme.Success
		}

		m.devTable.SetCell(row, 0, tview.NewTableCell(namePrefix+dev.Name).
			SetTextColor(typeColor).
			SetExpansion(1))

		m.devTable.SetCell(row, 1, tview.NewTableCell(dev.Size).
			SetTextColor(m.theme.Foreground).
			SetExpansion(0))

		m.devTable.SetCell(row, 2, tview.NewTableCell(dev.Type).
			SetTextColor(m.theme.Muted).
			SetExpansion(0))

		m.devTable.SetCell(row, 3, tview.NewTableCell(dev.FSType).
			SetTextColor(m.theme.Warning).
			SetExpansion(0))

		mountColor := m.theme.Muted
		if dev.Mountpoint != "" {
			mountColor = m.theme.Success
		}
		m.devTable.SetCell(row, 4, tview.NewTableCell(dev.Mountpoint).
			SetTextColor(mountColor).
			SetExpansion(1))

		m.devTable.SetCell(row, 5, tview.NewTableCell(truncate(dev.Model, 20)).
			SetTextColor(m.theme.Muted).
			SetExpansion(1))
	}

	if m.devTable.GetRowCount() > 1 {
		m.devTable.Select(1, 0)
	}
}

// updateDeviceDetails shows details for selected device
func (m *Manager) updateDeviceDetails() {
	if m.devIdx < 0 || m.devIdx >= len(m.blockDevices) {
		return
	}

	dev := m.blockDevices[m.devIdx]
	primary := colorToTag(m.theme.Primary)
	highlight := colorToTag(m.theme.Highlight)

	details := fmt.Sprintf(`[%s]Device:[white] /dev/%s
[%s]Size:[white] %s
[%s]Type:[white] %s
[%s]Filesystem:[white] %s
[%s]Mount:[white] %s
[%s]Model:[white] %s

[%s]Actions:[white]
  c=Create  d=Delete
  f=Format  m=Mount/Unmount
  i=Info`,
		primary, dev.Name,
		primary, dev.Size,
		primary, dev.Type,
		primary, dev.FSType,
		primary, dev.Mountpoint,
		primary, dev.Model,
		highlight)

	m.detailView.SetTitle(" Device Details ")
	m.detailView.SetText(details)
}

// parseFilesystems parses df output
func (m *Manager) parseFilesystems(output string) []Filesystem {
	var filesystems []Filesystem
	lines := strings.Split(output, "\n")

	for _, line := range lines {
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
	for row := m.table.GetRowCount() - 1; row > 0; row-- {
		m.table.RemoveRow(row)
	}

	for i, fs := range m.filesystems {
		row := i + 1

		m.table.SetCell(row, 0, tview.NewTableCell(truncate(fs.Device, 30)).
			SetTextColor(m.theme.Foreground).
			SetExpansion(1))

		m.table.SetCell(row, 1, tview.NewTableCell(truncate(fs.MountPoint, 25)).
			SetTextColor(m.theme.Foreground).
			SetExpansion(1))

		m.table.SetCell(row, 2, tview.NewTableCell(fs.Type).
			SetTextColor(m.theme.Muted).
			SetExpansion(0))

		m.table.SetCell(row, 3, tview.NewTableCell(formatBytes(fs.Size)).
			SetTextColor(m.theme.Foreground).
			SetAlign(tview.AlignRight).
			SetExpansion(0))

		m.table.SetCell(row, 4, tview.NewTableCell(formatBytes(fs.Used)).
			SetTextColor(m.theme.Foreground).
			SetAlign(tview.AlignRight).
			SetExpansion(0))

		m.table.SetCell(row, 5, tview.NewTableCell(formatBytes(fs.Available)).
			SetTextColor(m.theme.Success).
			SetAlign(tview.AlignRight).
			SetExpansion(0))

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

	m.detailView.SetTitle(" Details ")
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
