package sftp

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/systask/systask/internal/config"
	"github.com/systask/systask/internal/ssh"
)

// FileEntry represents a file or directory
type FileEntry struct {
	Name    string
	Path    string
	IsDir   bool
	Size    int64
	ModTime string
	Perms   string
}

// TransferJob represents a file transfer job
type TransferJob struct {
	Source      string
	Destination string
	SourceHost  *ssh.HostEntry
	DestHost    *ssh.HostEntry
	Status      string
	Progress    int
	Error       error
}

// Manager manages SFTP file transfers
type Manager struct {
	view         *tview.Flex
	srcPanel     *tview.Flex
	dstPanel     *tview.Flex
	srcList      *tview.Table
	dstList      *tview.Table
	srcPath      *tview.InputField
	dstPath      *tview.InputField
	hostSelector *tview.List
	transferList *tview.Table
	statusView   *tview.TextView
	helpView     *tview.TextView
	pages        *tview.Pages
	theme        *config.Theme
	sshClient    *ssh.Client
	app          *tview.Application

	// State
	hosts         []ssh.HostEntry
	srcHost       *ssh.HostEntry
	dstHost       *ssh.HostEntry
	srcFiles      []FileEntry
	dstFiles      []FileEntry
	srcCurrentDir string
	dstCurrentDir string
	srcSelected   map[int]bool
	dstSelected   map[int]bool
	focusLeft     bool
	transfers     []TransferJob
	selectingFor  string // "source" or "dest"
	srcIsLocal    bool
	dstIsLocal    bool

	// Clipboard for copy/paste
	clipboard       []FileEntry
	clipboardHost   *ssh.HostEntry
	clipboardIsLocal bool
	clipboardDir    string
	clipboardMove   bool // true = cut, false = copy
}

// NewManager creates a new SFTP manager
func NewManager(theme *config.Theme, client *ssh.Client) *Manager {
	m := &Manager{
		theme:         theme,
		sshClient:     client,
		srcCurrentDir: "/",
		dstCurrentDir: "/",
		srcSelected:   make(map[int]bool),
		dstSelected:   make(map[int]bool),
		focusLeft:     true,
		srcIsLocal:    true, // Default to local machine for source
	}
	m.build()
	return m
}

// build constructs the SFTP manager view
func (m *Manager) build() {
	// Source panel
	m.srcPath = tview.NewInputField()
	m.srcPath.SetLabel(" Path: ")
	m.srcPath.SetLabelColor(m.theme.Primary)
	m.srcPath.SetFieldBackgroundColor(m.theme.Muted)
	m.srcPath.SetFieldTextColor(m.theme.Foreground)
	m.srcPath.SetText("/")
	m.srcPath.SetBackgroundColor(m.theme.Background)

	m.srcList = m.createFileTable("source")

	srcBox := tview.NewFlex().SetDirection(tview.FlexRow)
	srcBox.AddItem(m.srcPath, 1, 0, false)
	srcBox.AddItem(m.srcList, 0, 1, true)
	srcBox.SetBorder(true)
	srcBox.SetBorderColor(m.theme.Primary)
	srcBox.SetTitle(" 📁 Source: LOCAL (h=change) ")
	srcBox.SetTitleColor(m.theme.Primary)
	srcBox.SetBackgroundColor(m.theme.Background)
	m.srcPanel = srcBox

	// Destination panel
	m.dstPath = tview.NewInputField()
	m.dstPath.SetLabel(" Path: ")
	m.dstPath.SetLabelColor(m.theme.Primary)
	m.dstPath.SetFieldBackgroundColor(m.theme.Muted)
	m.dstPath.SetFieldTextColor(m.theme.Foreground)
	m.dstPath.SetText("/")
	m.dstPath.SetBackgroundColor(m.theme.Background)

	m.dstList = m.createFileTable("dest")

	dstBox := tview.NewFlex().SetDirection(tview.FlexRow)
	dstBox.AddItem(m.dstPath, 1, 0, false)
	dstBox.AddItem(m.dstList, 0, 1, false)
	dstBox.SetBorder(true)
	dstBox.SetBorderColor(m.theme.Border)
	dstBox.SetTitle(" 📁 Destination: (h=select host) ")
	dstBox.SetTitleColor(m.theme.Primary)
	dstBox.SetBackgroundColor(m.theme.Background)
	m.dstPanel = dstBox

	// Host selector (hidden by default)
	m.hostSelector = tview.NewList()
	m.hostSelector.ShowSecondaryText(true)
	m.hostSelector.SetBackgroundColor(m.theme.Background)
	m.hostSelector.SetBorder(true)
	m.hostSelector.SetBorderColor(m.theme.Primary)
	m.hostSelector.SetTitle(" Select Host (Backspace=back) ")
	m.hostSelector.SetTitleColor(m.theme.Primary)
	m.hostSelector.SetSelectedBackgroundColor(m.theme.Muted)
	m.hostSelector.SetSelectedTextColor(m.theme.Primary)

	// Transfer queue
	m.transferList = tview.NewTable()
	m.transferList.SetBorders(false)
	m.transferList.SetBackgroundColor(m.theme.Background)
	m.transferList.SetBorder(true)
	m.transferList.SetBorderColor(m.theme.Border)
	m.transferList.SetTitle(" 📤 Transfer Queue ")
	m.transferList.SetTitleColor(m.theme.Primary)

	// Help bar (permanent)
	m.helpView = tview.NewTextView()
	m.helpView.SetDynamicColors(true)
	m.helpView.SetBackgroundColor(m.theme.Background)
	m.helpView.SetText(m.getHelpText())

	// Status view (for messages)
	m.statusView = tview.NewTextView()
	m.statusView.SetDynamicColors(true)
	m.statusView.SetBackgroundColor(m.theme.Muted)
	m.statusView.SetText(fmt.Sprintf("[%s]Ready. Select files with Space, then copy/paste.[white]", colorToTag(m.theme.Success)))

	// File panels side by side
	filePanels := tview.NewFlex()
	filePanels.AddItem(m.srcPanel, 0, 1, true)
	filePanels.AddItem(m.dstPanel, 0, 1, false)

	// Pages for main view and host selector
	m.pages = tview.NewPages()

	mainView := tview.NewFlex().SetDirection(tview.FlexRow)
	mainView.AddItem(filePanels, 0, 3, true)
	mainView.AddItem(m.transferList, 8, 0, false)
	mainView.AddItem(m.helpView, 1, 0, false)
	mainView.AddItem(m.statusView, 1, 0, false)

	m.pages.AddPage("main", mainView, true, true)
	m.pages.AddPage("hostselect", m.hostSelector, true, false)

	// Main layout
	m.view = tview.NewFlex().SetDirection(tview.FlexRow)
	m.view.AddItem(m.pages, 0, 1, true)
	m.view.SetBackgroundColor(m.theme.Background)

	// Load local files initially
	m.loadLocalDirectory(true, "/")
}

// createFileTable creates a file listing table
func (m *Manager) createFileTable(name string) *tview.Table {
	table := tview.NewTable()
	table.SetBorders(false)
	table.SetSelectable(true, false)
	table.SetBackgroundColor(m.theme.Background)
	table.SetSelectedStyle(tcell.StyleDefault.
		Background(m.theme.Muted).
		Foreground(m.theme.Primary))

	// Headers
	headers := []string{"", "NAME", "SIZE", "MODIFIED"}
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

	if name == "source" {
		table.SetInputCapture(m.handleSrcInput)
	} else {
		table.SetInputCapture(m.handleDstInput)
	}

	return table
}

// getHelpText returns help text
func (m *Manager) getHelpText() string {
	highlight := colorToTag(m.theme.Highlight)
	return fmt.Sprintf(
		"  [%s]Tab[white]=Switch  [%s]h[white]=Host  [%s]Enter[white]=Open  [%s]Space[white]=Select  [%s]c[white]=Copy  [%s]x[white]=Cut  [%s]p[white]=Paste  [%s]d[white]=Delete  [%s]r[white]=Refresh",
		highlight, highlight, highlight, highlight, highlight, highlight, highlight, highlight, highlight)
}

// handleSrcInput handles source panel input
func (m *Manager) handleSrcInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyTab:
		m.switchFocus()
		return nil
	case tcell.KeyEnter:
		m.handleEnter(true)
		return nil
	}

	switch event.Rune() {
	case ' ':
		m.toggleSelection(true)
		return nil
	case 'c':
		m.copyToClipboard(true, false)
		return nil
	case 'x':
		m.copyToClipboard(true, true)
		return nil
	case 'p':
		m.pasteFromClipboard(true)
		return nil
	case 'd':
		m.deleteSelected(true)
		return nil
	case 'n':
		m.createFolder(true)
		return nil
	case 'r':
		m.refreshPanel(true)
		return nil
	case 'h':
		m.showHostSelector(true)
		return nil
	case 'j':
		return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
	case 'k':
		return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
	case 'u':
		m.goUp(true)
		return nil
	}

	return event
}

// handleDstInput handles destination panel input
func (m *Manager) handleDstInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyTab:
		m.switchFocus()
		return nil
	case tcell.KeyEnter:
		m.handleEnter(false)
		return nil
	}

	switch event.Rune() {
	case ' ':
		m.toggleSelection(false)
		return nil
	case 'c':
		m.copyToClipboard(false, false)
		return nil
	case 'x':
		m.copyToClipboard(false, true)
		return nil
	case 'p':
		m.pasteFromClipboard(false)
		return nil
	case 'd':
		m.deleteSelected(false)
		return nil
	case 'n':
		m.createFolder(false)
		return nil
	case 'r':
		m.refreshPanel(false)
		return nil
	case 'h':
		m.showHostSelector(false)
		return nil
	case 'j':
		return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
	case 'k':
		return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
	case 'u':
		m.goUp(false)
		return nil
	}

	return event
}

// switchFocus switches between source and destination panels
func (m *Manager) switchFocus() {
	m.focusLeft = !m.focusLeft
	if m.focusLeft {
		m.srcPanel.SetBorderColor(m.theme.Primary)
		m.dstPanel.SetBorderColor(m.theme.Border)
		if m.app != nil {
			m.app.SetFocus(m.srcList)
		}
	} else {
		m.srcPanel.SetBorderColor(m.theme.Border)
		m.dstPanel.SetBorderColor(m.theme.Primary)
		if m.app != nil {
			m.app.SetFocus(m.dstList)
		}
	}
}

// SetHosts sets available hosts for selection
func (m *Manager) SetHosts(hosts []ssh.HostEntry) {
	m.hosts = hosts
}

// SetApp sets the tview application reference
func (m *Manager) SetApp(app *tview.Application) {
	m.app = app
}

// showHostSelector shows the host selection overlay
func (m *Manager) showHostSelector(isSource bool) {
	m.selectingFor = "source"
	if !isSource {
		m.selectingFor = "dest"
	}

	m.hostSelector.Clear()

	// Add LOCAL option first
	m.hostSelector.AddItem("🖥️  LOCAL (this machine)", "Local filesystem", 'l', func() {
		m.selectLocalHost(isSource)
	})

	// Add connected hosts
	for _, host := range m.hosts {
		h := host // Capture for closure
		name := h.Name
		if name == "" {
			name = h.Hostname
		}
		secondary := h.Hostname
		if h.User != "" {
			secondary = h.User + "@" + h.Hostname
		}
		m.hostSelector.AddItem("🌐 "+name, secondary, 0, func() {
			m.selectRemoteHost(isSource, &h)
		})
	}

	m.hostSelector.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyBackspace || event.Key() == tcell.KeyBackspace2 {
			m.pages.SwitchToPage("main")
			if m.focusLeft {
				if m.app != nil {
					m.app.SetFocus(m.srcList)
				}
			} else {
				if m.app != nil {
					m.app.SetFocus(m.dstList)
				}
			}
			return nil
		}
		return event
	})

	m.pages.SwitchToPage("hostselect")
	if m.app != nil {
		m.app.SetFocus(m.hostSelector)
	}
}

// selectLocalHost selects local machine for a panel
func (m *Manager) selectLocalHost(isSource bool) {
	m.pages.SwitchToPage("main")

	if isSource {
		m.srcIsLocal = true
		m.srcHost = nil
		m.srcPanel.SetTitle(" 📁 Source: LOCAL ")
		m.loadLocalDirectory(true, m.srcCurrentDir)
		if m.app != nil {
			m.app.SetFocus(m.srcList)
		}
	} else {
		m.dstIsLocal = true
		m.dstHost = nil
		m.dstPanel.SetTitle(" 📁 Destination: LOCAL ")
		m.loadLocalDirectory(false, m.dstCurrentDir)
		if m.app != nil {
			m.app.SetFocus(m.dstList)
		}
	}
}

// selectRemoteHost selects a remote host for a panel
func (m *Manager) selectRemoteHost(isSource bool, host *ssh.HostEntry) {
	m.pages.SwitchToPage("main")

	name := host.Name
	if name == "" {
		name = host.Hostname
	}

	// Check if connected, if not - try to connect
	if !m.sshClient.IsConnected(*host) {
		m.statusView.SetText(fmt.Sprintf("[%s]Connecting to %s...[white]",
			colorToTag(m.theme.Warning), name))

		// Try to connect in background
		go func() {
			_, err := m.sshClient.Connect(*host)
			if m.app != nil {
				m.app.QueueUpdateDraw(func() {
					if err != nil {
						m.statusView.SetText(fmt.Sprintf("[%s]Connection failed: %v[white]",
							colorToTag(m.theme.Error), err))
						return
					}

					m.statusView.SetText(fmt.Sprintf("[%s]Connected to %s[white]",
						colorToTag(m.theme.Success), name))

					// Now load the directory
					if isSource {
						m.srcIsLocal = false
						m.srcHost = host
						m.srcPanel.SetTitle(fmt.Sprintf(" 📁 Source: %s ", name))
						m.srcCurrentDir = "/"
						m.loadRemoteDirectory(true, "/")
						m.app.SetFocus(m.srcList)
					} else {
						m.dstIsLocal = false
						m.dstHost = host
						m.dstPanel.SetTitle(fmt.Sprintf(" 📁 Destination: %s ", name))
						m.dstCurrentDir = "/"
						m.loadRemoteDirectory(false, "/")
						m.app.SetFocus(m.dstList)
					}
				})
			}
		}()
		return
	}

	// Already connected - just load directory
	if isSource {
		m.srcIsLocal = false
		m.srcHost = host
		m.srcPanel.SetTitle(fmt.Sprintf(" 📁 Source: %s ", name))
		m.srcCurrentDir = "/"
		m.loadRemoteDirectory(true, "/")
		if m.app != nil {
			m.app.SetFocus(m.srcList)
		}
	} else {
		m.dstIsLocal = false
		m.dstHost = host
		m.dstPanel.SetTitle(fmt.Sprintf(" 📁 Destination: %s ", name))
		m.dstCurrentDir = "/"
		m.loadRemoteDirectory(false, "/")
		if m.app != nil {
			m.app.SetFocus(m.dstList)
		}
	}
}

// SetSourceHost sets the source host (for external use)
func (m *Manager) SetSourceHost(host *ssh.HostEntry) {
	if host != nil {
		m.selectRemoteHost(true, host)
	}
}

// SetDestHost sets the destination host (for external use)
func (m *Manager) SetDestHost(host *ssh.HostEntry) {
	if host != nil {
		m.selectRemoteHost(false, host)
	}
}

// handleEnter handles enter key on file list
func (m *Manager) handleEnter(isSource bool) {
	var files []FileEntry
	var table *tview.Table
	var currentDir *string
	var isLocal bool

	if isSource {
		files = m.srcFiles
		table = m.srcList
		currentDir = &m.srcCurrentDir
		isLocal = m.srcIsLocal
	} else {
		files = m.dstFiles
		table = m.dstList
		currentDir = &m.dstCurrentDir
		isLocal = m.dstIsLocal
	}

	row, _ := table.GetSelection()
	if row <= 0 || row > len(files) {
		return
	}

	file := files[row-1]

	if file.Name == ".." {
		*currentDir = filepath.Dir(*currentDir)
		if *currentDir == "" {
			*currentDir = "/"
		}
	} else if file.IsDir {
		*currentDir = filepath.Join(*currentDir, file.Name)
	} else {
		m.toggleSelection(isSource)
		return
	}

	if isLocal {
		m.loadLocalDirectory(isSource, *currentDir)
	} else {
		m.loadRemoteDirectory(isSource, *currentDir)
	}
}

// goUp navigates to parent directory
func (m *Manager) goUp(isSource bool) {
	var currentDir *string
	var isLocal bool

	if isSource {
		currentDir = &m.srcCurrentDir
		isLocal = m.srcIsLocal
	} else {
		currentDir = &m.dstCurrentDir
		isLocal = m.dstIsLocal
	}

	*currentDir = filepath.Dir(*currentDir)
	if *currentDir == "" {
		*currentDir = "/"
	}

	if isLocal {
		m.loadLocalDirectory(isSource, *currentDir)
	} else {
		m.loadRemoteDirectory(isSource, *currentDir)
	}
}

// loadLocalDirectory loads local directory listing
func (m *Manager) loadLocalDirectory(isSource bool, path string) {
	var files *[]FileEntry
	var pathField *tview.InputField

	if isSource {
		files = &m.srcFiles
		pathField = m.srcPath
		m.srcCurrentDir = path
	} else {
		files = &m.dstFiles
		pathField = m.dstPath
		m.dstCurrentDir = path
	}

	pathField.SetText(path)

	// Read directory using os package
	entries, err := os.ReadDir(path)
	if err != nil {
		m.statusView.SetText(fmt.Sprintf("[%s]Error: %v[white]", colorToTag(m.theme.Error), err))
		return
	}

	*files = make([]FileEntry, 0)

	// Add parent directory
	if path != "/" {
		*files = append(*files, FileEntry{
			Name:  "..",
			Path:  filepath.Dir(path),
			IsDir: true,
		})
	}

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		file := FileEntry{
			Name:    entry.Name(),
			Path:    filepath.Join(path, entry.Name()),
			IsDir:   entry.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime().Format("Jan 02 15:04"),
			Perms:   info.Mode().String(),
		}
		*files = append(*files, file)
	}

	// Sort: directories first
	sort.Slice(*files, func(i, j int) bool {
		if (*files)[i].Name == ".." {
			return true
		}
		if (*files)[j].Name == ".." {
			return false
		}
		if (*files)[i].IsDir != (*files)[j].IsDir {
			return (*files)[i].IsDir
		}
		return strings.ToLower((*files)[i].Name) < strings.ToLower((*files)[j].Name)
	})

	m.updateFileTable(isSource)
}

// loadRemoteDirectory loads remote directory listing via SSH
func (m *Manager) loadRemoteDirectory(isSource bool, path string) {
	var host *ssh.HostEntry
	var files *[]FileEntry
	var pathField *tview.InputField

	if isSource {
		host = m.srcHost
		files = &m.srcFiles
		pathField = m.srcPath
		m.srcCurrentDir = path
	} else {
		host = m.dstHost
		files = &m.dstFiles
		pathField = m.dstPath
		m.dstCurrentDir = path
	}

	if host == nil {
		m.statusView.SetText(fmt.Sprintf("[%s]No host selected. Press 'h' to select.[white]",
			colorToTag(m.theme.Warning)))
		return
	}

	pathField.SetText(path)

	// Check if connected
	if !m.sshClient.IsConnected(*host) {
		m.statusView.SetText(fmt.Sprintf("[%s]Not connected. Connect from sidebar first.[white]",
			colorToTag(m.theme.Warning)))
		return
	}

	// Get directory listing via SSH
	cmd := fmt.Sprintf(`ls -la '%s' 2>/dev/null | tail -n +2`, path)
	output, err := m.sshClient.RunCommand(*host, cmd)
	if err != nil {
		m.statusView.SetText(fmt.Sprintf("[%s]Error: %v[white]", colorToTag(m.theme.Error), err))
		return
	}

	*files = m.parseDirectoryListing(output, path)
	m.updateFileTable(isSource)
}

// parseDirectoryListing parses ls -la output
func (m *Manager) parseDirectoryListing(output, currentPath string) []FileEntry {
	var files []FileEntry

	if currentPath != "/" {
		files = append(files, FileEntry{
			Name:  "..",
			Path:  filepath.Dir(currentPath),
			IsDir: true,
		})
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "total") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue
		}

		perms := fields[0]
		name := strings.Join(fields[8:], " ")

		if name == "." || name == ".." {
			continue
		}

		var size int64
		fmt.Sscanf(fields[4], "%d", &size)
		modTime := strings.Join(fields[5:8], " ")

		file := FileEntry{
			Name:    name,
			Path:    filepath.Join(currentPath, name),
			IsDir:   perms[0] == 'd',
			Size:    size,
			ModTime: modTime,
			Perms:   perms,
		}
		files = append(files, file)
	}

	sort.Slice(files, func(i, j int) bool {
		if files[i].Name == ".." {
			return true
		}
		if files[j].Name == ".." {
			return false
		}
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})

	return files
}

// updateFileTable updates the file table display
func (m *Manager) updateFileTable(isSource bool) {
	var files []FileEntry
	var table *tview.Table
	var selected map[int]bool

	if isSource {
		files = m.srcFiles
		table = m.srcList
		selected = m.srcSelected
	} else {
		files = m.dstFiles
		table = m.dstList
		selected = m.dstSelected
	}

	// Clear existing rows
	for row := table.GetRowCount() - 1; row > 0; row-- {
		table.RemoveRow(row)
	}

	for i, file := range files {
		row := i + 1

		checkbox := "☐"
		if selected[i] {
			checkbox = "☑"
		}
		table.SetCell(row, 0, tview.NewTableCell(checkbox).
			SetTextColor(m.theme.Secondary).
			SetExpansion(0))

		icon := "📄"
		color := m.theme.Foreground
		if file.IsDir {
			icon = "📁"
			color = m.theme.Primary
		}
		if file.Name == ".." {
			icon = "⬆️"
		}
		table.SetCell(row, 1, tview.NewTableCell(fmt.Sprintf("%s %s", icon, file.Name)).
			SetTextColor(color).
			SetExpansion(1))

		sizeStr := ""
		if !file.IsDir {
			sizeStr = formatBytes(uint64(file.Size))
		}
		table.SetCell(row, 2, tview.NewTableCell(sizeStr).
			SetTextColor(m.theme.Muted).
			SetExpansion(0))

		table.SetCell(row, 3, tview.NewTableCell(file.ModTime).
			SetTextColor(m.theme.Muted).
			SetExpansion(1))
	}

	if table.GetRowCount() > 1 {
		table.Select(1, 0)
	}
}

// toggleSelection toggles file selection
func (m *Manager) toggleSelection(isSource bool) {
	var table *tview.Table
	var selected *map[int]bool

	if isSource {
		table = m.srcList
		selected = &m.srcSelected
	} else {
		table = m.dstList
		selected = &m.dstSelected
	}

	row, _ := table.GetSelection()
	if row <= 0 {
		return
	}

	idx := row - 1
	if (*selected)[idx] {
		delete(*selected, idx)
	} else {
		(*selected)[idx] = true
	}

	m.updateFileTable(isSource)
	table.Select(row, 0)
}

// copyToClipboard copies selected files to clipboard
func (m *Manager) copyToClipboard(isSource bool, move bool) {
	var files []FileEntry
	var selected map[int]bool
	var isLocal bool
	var host *ssh.HostEntry
	var currentDir string
	var table *tview.Table
	var panelName string

	if isSource {
		files = m.srcFiles
		selected = m.srcSelected
		isLocal = m.srcIsLocal
		host = m.srcHost
		currentDir = m.srcCurrentDir
		table = m.srcList
		panelName = "source"
	} else {
		files = m.dstFiles
		selected = m.dstSelected
		isLocal = m.dstIsLocal
		host = m.dstHost
		currentDir = m.dstCurrentDir
		table = m.dstList
		panelName = "destination"
	}

	// Collect selected files (marked with Space)
	var selectedFiles []FileEntry
	for idx := range selected {
		if idx < len(files) && files[idx].Name != ".." {
			selectedFiles = append(selectedFiles, files[idx])
		}
	}

	// If nothing selected with Space, use currently highlighted file
	if len(selectedFiles) == 0 && table != nil {
		row, _ := table.GetSelection()
		// row 0 is header, files start at row 1
		fileIdx := row - 1
		if fileIdx >= 0 && fileIdx < len(files) {
			file := files[fileIdx]
			if file.Name != ".." {
				selectedFiles = append(selectedFiles, file)
			}
		}
	}

	if len(selectedFiles) == 0 {
		row, _ := table.GetSelection()
		m.statusView.SetText(fmt.Sprintf(
			"[%s]No file to copy from %s (row=%d, files=%d). Select with Space or navigate to file.[white]",
			colorToTag(m.theme.Warning), panelName, row, len(files)))
		return
	}

	// Store in clipboard
	m.clipboard = selectedFiles
	m.clipboardHost = host
	m.clipboardIsLocal = isLocal
	m.clipboardDir = currentDir
	m.clipboardMove = move

	action := "📋 Copied"
	if move {
		action = "✂️ Cut"
	}

	hostName := "LOCAL"
	if host != nil {
		hostName = host.Name
		if hostName == "" {
			hostName = host.Hostname
		}
	}

	m.statusView.SetText(fmt.Sprintf(
		"[%s]%s %d file(s) from %s (%s). Tab to switch, 'p' to paste.[white]",
		colorToTag(m.theme.Success), action, len(selectedFiles), panelName, hostName))
}

// pasteFromClipboard pastes files from clipboard to current panel
func (m *Manager) pasteFromClipboard(isSource bool) {
	if len(m.clipboard) == 0 {
		m.statusView.SetText(fmt.Sprintf(
			"[%s]Clipboard empty. Select file in other panel, press 'c' to copy first.[white]",
			colorToTag(m.theme.Warning)))
		return
	}

	// Use the passed isSource parameter - paste TO the panel where 'p' was pressed

	var dstDir string
	var dstIsLocal bool
	var dstHost *ssh.HostEntry

	if isSource {
		dstDir = m.srcCurrentDir
		dstIsLocal = m.srcIsLocal
		dstHost = m.srcHost
	} else {
		dstDir = m.dstCurrentDir
		dstIsLocal = m.dstIsLocal
		dstHost = m.dstHost
	}

	srcIsLocal := m.clipboardIsLocal
	srcHost := m.clipboardHost
	files := m.clipboard
	move := m.clipboardMove

	action := "Copying"
	if move {
		action = "Moving"
	}

	totalFiles := len(files)
	m.statusView.SetText(fmt.Sprintf("[%s]%s %d file(s)...[white]",
		colorToTag(m.theme.Warning), action, totalFiles))

	// Helper to update transfer queue progress
	// completed = number of files fully transferred, inProgress = current file being transferred
	updateProgress := func(completed int, fileName string, status string, isComplete bool) {
		if m.app != nil {
			m.app.QueueUpdateDraw(func() {
				m.transferList.Clear()

				// Calculate progress percentage
				var percent float64
				if isComplete {
					percent = 100
				} else if totalFiles == 1 {
					// For single file, show animated progress based on status
					switch status {
					case "Starting...":
						percent = 10
					case "Downloading...":
						percent = 30
					case "Uploading...":
						percent = 60
					case "Remote copy...":
						percent = 50
					case "Transferring...":
						percent = 50
					default:
						percent = 25
					}
				} else {
					// Multiple files: show completed/total
					percent = float64(completed) / float64(totalFiles) * 100
				}

				// Extra wide progress bar (80 chars) to fill the space
				barWidth := 80
				filled := int(percent / 100 * float64(barWidth))
				if filled < 1 && percent > 0 {
					filled = 1
				}

				// Create stylish progress bar with gradient colors
				var bar string
				if isComplete {
					bar = fmt.Sprintf("[%s]%s[white] 100%%",
						colorToTag(m.theme.Success),
						strings.Repeat("█", barWidth))
				} else {
					// Animated style with different block characters
					progressChars := strings.Repeat("█", filled)
					emptyChars := strings.Repeat("░", barWidth-filled)
					bar = fmt.Sprintf("[%s]%s[%s]%s[white] %.0f%%",
						colorToTag(m.theme.Primary), progressChars,
						colorToTag(m.theme.Muted), emptyChars, percent)
				}

				// Compact layout: FILE | WIDE PROGRESS BAR WITH STATUS
				fileCell := tview.NewTableCell(truncate(fileName, 15)).
					SetTextColor(m.theme.Foreground).
					SetExpansion(0)
				
				// Combine bar and status in one cell for maximum width
				statusIcon := "⏳"
				statusColor := m.theme.Warning
				if isComplete {
					statusIcon = "✅"
					statusColor = m.theme.Success
				}
				combinedCell := tview.NewTableCell(fmt.Sprintf("%s  [%s]%s[white]", bar, colorToTag(statusColor), statusIcon)).
					SetExpansion(1)

				m.transferList.SetCell(0, 0, fileCell)
				m.transferList.SetCell(0, 1, combinedCell)

				// Status bar with file count
				if isComplete {
					m.statusView.SetText(fmt.Sprintf("[%s]✓ %s complete! %d/%d files transferred.[white]",
						colorToTag(m.theme.Success), action, completed, totalFiles))
				} else {
					m.statusView.SetText(fmt.Sprintf("[%s]%s %s [%d/%d][white]",
						colorToTag(m.theme.Warning), status, fileName, completed+1, totalFiles))
				}
			})
		}
	}

	// Perform transfer in background
	go func() {
		var err error

		if srcIsLocal && dstIsLocal {
			// Local to local
			for i, file := range files {
				updateProgress(i, file.Name, "Transferring...", false)
				dst := filepath.Join(dstDir, file.Name)
				var cmd *exec.Cmd
				if move {
					cmd = exec.Command("mv", file.Path, dst)
				} else {
					cmd = exec.Command("cp", "-r", file.Path, dst)
				}
				if e := cmd.Run(); e != nil {
					err = e
					break
				}
			}
		} else if srcIsLocal && !dstIsLocal {
			// Local to remote (upload)
			if dstHost == nil {
				err = fmt.Errorf("no destination host")
			} else {
				for i, file := range files {
					updateProgress(i, file.Name, "Uploading...", false)
					dst := filepath.Join(dstDir, file.Name)
					target := fmt.Sprintf("%s@%s:%s", dstHost.User, dstHost.Hostname, dst)

					args := []string{"-r", "-o", "StrictHostKeyChecking=no"}
					if dstHost.KeyFile != "" {
						args = append(args, "-i", dstHost.KeyFile)
					}
					if dstHost.Port != 0 && dstHost.Port != 22 {
						args = append(args, "-P", fmt.Sprintf("%d", dstHost.Port))
					}
					args = append(args, file.Path, target)

					cmd := exec.Command("scp", args...)
					if e := cmd.Run(); e != nil {
						err = e
						break
					}

					if move {
						os.RemoveAll(file.Path)
					}
				}
			}
		} else if !srcIsLocal && dstIsLocal {
			// Remote to local (download)
			if srcHost == nil {
				err = fmt.Errorf("no source host")
			} else {
				for i, file := range files {
					updateProgress(i, file.Name, "Downloading...", false)
					dst := filepath.Join(dstDir, file.Name)
					source := fmt.Sprintf("%s@%s:%s", srcHost.User, srcHost.Hostname, file.Path)

					args := []string{"-r", "-o", "StrictHostKeyChecking=no"}
					if srcHost.KeyFile != "" {
						args = append(args, "-i", srcHost.KeyFile)
					}
					if srcHost.Port != 0 && srcHost.Port != 22 {
						args = append(args, "-P", fmt.Sprintf("%d", srcHost.Port))
					}
					args = append(args, source, dst)

					cmd := exec.Command("scp", args...)
					if e := cmd.Run(); e != nil {
						err = e
						break
					}

					if move && srcHost != nil {
						m.sshClient.RunCommand(*srcHost, fmt.Sprintf("rm -rf '%s'", file.Path))
					}
				}
			}
		} else {
			// Remote to remote
			if srcHost == nil || dstHost == nil {
				err = fmt.Errorf("both hosts must be selected")
			} else if srcHost.Hostname == dstHost.Hostname {
				// Same host - use cp/mv on remote
				for i, file := range files {
					updateProgress(i, file.Name, "Remote copy...", false)
					dst := filepath.Join(dstDir, file.Name)
					var cmd string
					if move {
						cmd = fmt.Sprintf("mv '%s' '%s'", file.Path, dst)
					} else {
						cmd = fmt.Sprintf("cp -r '%s' '%s'", file.Path, dst)
					}
					if _, e := m.sshClient.RunCommand(*srcHost, cmd); e != nil {
						err = e
						break
					}
				}
			} else {
				// Cross-host transfer: download to temp, then upload
				tempDir := os.TempDir()
				for i, file := range files {
					tempPath := filepath.Join(tempDir, file.Name)
					dst := filepath.Join(dstDir, file.Name)

					// Step 1: Download from source to local temp
					updateProgress(i, file.Name, "Downloading...", false)
					source := fmt.Sprintf("%s@%s:%s", srcHost.User, srcHost.Hostname, file.Path)
					downloadArgs := []string{"-r", "-o", "StrictHostKeyChecking=no"}
					if srcHost.KeyFile != "" {
						downloadArgs = append(downloadArgs, "-i", srcHost.KeyFile)
					}
					if srcHost.Port != 0 && srcHost.Port != 22 {
						downloadArgs = append(downloadArgs, "-P", fmt.Sprintf("%d", srcHost.Port))
					}
					downloadArgs = append(downloadArgs, source, tempPath)

					downloadCmd := exec.Command("scp", downloadArgs...)
					if e := downloadCmd.Run(); e != nil {
						err = fmt.Errorf("download failed: %v", e)
						break
					}

					// Step 2: Upload from local temp to destination
					updateProgress(i, file.Name, "Uploading...", false)
					target := fmt.Sprintf("%s@%s:%s", dstHost.User, dstHost.Hostname, dst)
					uploadArgs := []string{"-r", "-o", "StrictHostKeyChecking=no"}
					if dstHost.KeyFile != "" {
						uploadArgs = append(uploadArgs, "-i", dstHost.KeyFile)
					}
					if dstHost.Port != 0 && dstHost.Port != 22 {
						uploadArgs = append(uploadArgs, "-P", fmt.Sprintf("%d", dstHost.Port))
					}
					uploadArgs = append(uploadArgs, tempPath, target)

					uploadCmd := exec.Command("scp", uploadArgs...)
					if e := uploadCmd.Run(); e != nil {
						os.RemoveAll(tempPath)
						err = fmt.Errorf("upload failed: %v", e)
						break
					}

					// Cleanup temp file
					os.RemoveAll(tempPath)

					// If move, delete source
					if move {
						m.sshClient.RunCommand(*srcHost, fmt.Sprintf("rm -rf '%s'", file.Path))
					}
				}
			}
		}

		// Show final status
		if err == nil {
			updateProgress(totalFiles, files[len(files)-1].Name, "Complete!", true)
		}

		if m.app != nil {
			m.app.QueueUpdateDraw(func() {
				if err != nil {
					m.transferList.Clear()
					m.transferList.SetCell(0, 0, tview.NewTableCell(fmt.Sprintf(" ❌ Error: %v", err)).SetTextColor(m.theme.Error))
					m.statusView.SetText(fmt.Sprintf("[%s]Transfer error: %v[white]",
						colorToTag(m.theme.Error), err))
				} else {
					// Clear clipboard after successful move
					if move {
						m.clipboard = nil
					}
					// Clear selections
					m.srcSelected = make(map[int]bool)
					m.dstSelected = make(map[int]bool)
					// Refresh panels
					m.refreshPanel(true)
					m.refreshPanel(false)
				}
			})
		}
	}()
}

// startTransfer initiates file transfer
func (m *Manager) startTransfer(move bool) {
	// Get selected files from source
	var selectedFiles []FileEntry
	for idx := range m.srcSelected {
		if idx < len(m.srcFiles) && m.srcFiles[idx].Name != ".." {
			selectedFiles = append(selectedFiles, m.srcFiles[idx])
		}
	}

	if len(selectedFiles) == 0 {
		m.statusView.SetText(fmt.Sprintf(
			"[%s]No files selected. Press Space to select files.[white]",
			colorToTag(m.theme.Warning)))
		return
	}

	action := "Copying"
	if move {
		action = "Moving"
	}

	// Determine transfer type
	if m.srcIsLocal && m.dstIsLocal {
		// Local to local
		m.localToLocalTransfer(selectedFiles, move)
	} else if m.srcIsLocal && !m.dstIsLocal {
		// Local to remote (upload)
		m.localToRemoteTransfer(selectedFiles, move)
	} else if !m.srcIsLocal && m.dstIsLocal {
		// Remote to local (download)
		m.remoteToLocalTransfer(selectedFiles, move)
	} else {
		// Remote to remote
		m.remoteToRemoteTransfer(selectedFiles, move)
	}

	m.statusView.SetText(fmt.Sprintf("[%s]%s %d file(s)...[white]",
		colorToTag(m.theme.Warning), action, len(selectedFiles)))
}

// localToLocalTransfer handles local to local transfers
func (m *Manager) localToLocalTransfer(files []FileEntry, move bool) {
	for _, file := range files {
		dst := filepath.Join(m.dstCurrentDir, file.Name)
		var cmd *exec.Cmd
		if move {
			cmd = exec.Command("mv", file.Path, dst)
		} else {
			cmd = exec.Command("cp", "-r", file.Path, dst)
		}
		if err := cmd.Run(); err != nil {
			m.statusView.SetText(fmt.Sprintf("[%s]Error: %v[white]", colorToTag(m.theme.Error), err))
			return
		}
	}
	m.srcSelected = make(map[int]bool)
	m.refreshPanel(true)
	m.refreshPanel(false)
	m.statusView.SetText(fmt.Sprintf("[%s]✓ Transfer complete![white]", colorToTag(m.theme.Success)))
}

// localToRemoteTransfer handles upload (local to remote)
func (m *Manager) localToRemoteTransfer(files []FileEntry, move bool) {
	if m.dstHost == nil {
		m.statusView.SetText(fmt.Sprintf("[%s]No destination host selected[white]", colorToTag(m.theme.Error)))
		return
	}

	for _, file := range files {
		dst := filepath.Join(m.dstCurrentDir, file.Name)
		// Use scp for upload
		host := m.dstHost
		target := fmt.Sprintf("%s@%s:%s", host.User, host.Hostname, dst)

		args := []string{"-r"}
		if host.KeyFile != "" {
			args = append(args, "-i", host.KeyFile)
		}
		if host.Port != 0 && host.Port != 22 {
			args = append(args, "-P", fmt.Sprintf("%d", host.Port))
		}
		args = append(args, file.Path, target)

		cmd := exec.Command("scp", args...)
		if err := cmd.Run(); err != nil {
			m.statusView.SetText(fmt.Sprintf("[%s]Upload error: %v[white]", colorToTag(m.theme.Error), err))
			return
		}

		if move {
			os.RemoveAll(file.Path)
		}
	}
	m.srcSelected = make(map[int]bool)
	m.refreshPanel(true)
	m.refreshPanel(false)
	m.statusView.SetText(fmt.Sprintf("[%s]✓ Upload complete![white]", colorToTag(m.theme.Success)))
}

// remoteToLocalTransfer handles download (remote to local)
func (m *Manager) remoteToLocalTransfer(files []FileEntry, move bool) {
	if m.srcHost == nil {
		m.statusView.SetText(fmt.Sprintf("[%s]No source host selected[white]", colorToTag(m.theme.Error)))
		return
	}

	for _, file := range files {
		dst := filepath.Join(m.dstCurrentDir, file.Name)
		host := m.srcHost
		source := fmt.Sprintf("%s@%s:%s", host.User, host.Hostname, file.Path)

		args := []string{"-r"}
		if host.KeyFile != "" {
			args = append(args, "-i", host.KeyFile)
		}
		if host.Port != 0 && host.Port != 22 {
			args = append(args, "-P", fmt.Sprintf("%d", host.Port))
		}
		args = append(args, source, dst)

		cmd := exec.Command("scp", args...)
		if err := cmd.Run(); err != nil {
			m.statusView.SetText(fmt.Sprintf("[%s]Download error: %v[white]", colorToTag(m.theme.Error), err))
			return
		}

		if move {
			m.sshClient.RunCommand(*host, fmt.Sprintf("rm -rf '%s'", file.Path))
		}
	}
	m.srcSelected = make(map[int]bool)
	m.refreshPanel(true)
	m.refreshPanel(false)
	m.statusView.SetText(fmt.Sprintf("[%s]✓ Download complete![white]", colorToTag(m.theme.Success)))
}

// remoteToRemoteTransfer handles remote to remote transfers
func (m *Manager) remoteToRemoteTransfer(files []FileEntry, move bool) {
	if m.srcHost == nil || m.dstHost == nil {
		m.statusView.SetText(fmt.Sprintf("[%s]Both hosts must be selected[white]", colorToTag(m.theme.Error)))
		return
	}

	if m.srcHost.Hostname == m.dstHost.Hostname {
		// Same host - use cp/mv
		for _, file := range files {
			dst := filepath.Join(m.dstCurrentDir, file.Name)
			var cmd string
			if move {
				cmd = fmt.Sprintf("mv '%s' '%s'", file.Path, dst)
			} else {
				cmd = fmt.Sprintf("cp -r '%s' '%s'", file.Path, dst)
			}
			if _, err := m.sshClient.RunCommand(*m.srcHost, cmd); err != nil {
				m.statusView.SetText(fmt.Sprintf("[%s]Error: %v[white]", colorToTag(m.theme.Error), err))
				return
			}
		}
	} else {
		m.statusView.SetText(fmt.Sprintf(
			"[%s]Cross-host transfer: Use scp on terminal[white]\n"+
				"Example: scp user@src:/path user@dst:/path",
			colorToTag(m.theme.Warning)))
		return
	}

	m.srcSelected = make(map[int]bool)
	m.refreshPanel(true)
	m.refreshPanel(false)
	m.statusView.SetText(fmt.Sprintf("[%s]✓ Transfer complete![white]", colorToTag(m.theme.Success)))
}

// deleteSelected deletes selected files
func (m *Manager) deleteSelected(isSource bool) {
	var files []FileEntry
	var selected map[int]bool
	var isLocal bool

	if isSource {
		files = m.srcFiles
		selected = m.srcSelected
		isLocal = m.srcIsLocal
	} else {
		files = m.dstFiles
		selected = m.dstSelected
		isLocal = m.dstIsLocal
	}

	for idx := range selected {
		if idx < len(files) && files[idx].Name != ".." {
			if isLocal {
				os.RemoveAll(files[idx].Path)
			} else {
				host := m.srcHost
				if !isSource {
					host = m.dstHost
				}
				if host != nil {
					m.sshClient.RunCommand(*host, fmt.Sprintf("rm -rf '%s'", files[idx].Path))
				}
			}
		}
	}

	if isSource {
		m.srcSelected = make(map[int]bool)
	} else {
		m.dstSelected = make(map[int]bool)
	}
	m.refreshPanel(isSource)
}

// createFolder creates a new folder
func (m *Manager) createFolder(isSource bool) {
	m.statusView.SetText(fmt.Sprintf(
		"[%s]Create folder: Use command :run mkdir <foldername>[white]",
		colorToTag(m.theme.Warning)))
}

// refreshPanel refreshes the file listing
func (m *Manager) refreshPanel(isSource bool) {
	if isSource {
		if m.srcIsLocal {
			m.loadLocalDirectory(true, m.srcCurrentDir)
		} else if m.srcHost != nil {
			m.loadRemoteDirectory(true, m.srcCurrentDir)
		}
	} else {
		if m.dstIsLocal {
			m.loadLocalDirectory(false, m.dstCurrentDir)
		} else if m.dstHost != nil {
			m.loadRemoteDirectory(false, m.dstCurrentDir)
		}
	}
}

// Refresh refreshes both panels
func (m *Manager) Refresh() error {
	m.refreshPanel(true)
	m.refreshPanel(false)
	return nil
}

// View returns the manager view
func (m *Manager) View() *tview.Flex {
	return m.view
}

// Focus sets focus to the source list
func (m *Manager) Focus() {
	if m.app != nil {
		m.app.SetFocus(m.srcList)
	}
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
