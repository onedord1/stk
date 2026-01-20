package docker

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/systask/systask/internal/config"
	"github.com/systask/systask/internal/ssh"
)

// Container represents a Docker container
type Container struct {
	ID      string
	Name    string
	Image   string
	Status  string
	Ports   string
	Created string
	Running bool
}

// Image represents a Docker image
type Image struct {
	ID      string
	Repo    string
	Tag     string
	Size    string
	Created string
}

// Manager manages Docker containers and images
type Manager struct {
	view         *tview.Flex
	tabs         *tview.TextView
	containerTbl *tview.Table
	imageTbl     *tview.Table
	detailView   *tview.TextView
	actionView   *tview.TextView
	pages        *tview.Pages
	theme        *config.Theme
	sshClient    *ssh.Client
	host         *ssh.HostEntry
	app          *tview.Application

	containers  []Container
	images      []Image
	currentTab  string
	selectedIdx int
}

// NewManager creates a new Docker manager
func NewManager(theme *config.Theme, client *ssh.Client) *Manager {
	m := &Manager{
		theme:      theme,
		sshClient:  client,
		currentTab: "containers",
	}
	m.build()
	return m
}

// build constructs the Docker manager view
func (m *Manager) build() {
	// Tab bar
	m.tabs = tview.NewTextView()
	m.tabs.SetDynamicColors(true)
	m.tabs.SetBackgroundColor(m.theme.Muted)
	m.tabs.SetTextAlign(tview.AlignCenter)
	m.updateTabs()

	// Container table
	m.containerTbl = tview.NewTable()
	m.containerTbl.SetBorders(false)
	m.containerTbl.SetSelectable(true, false)
	m.containerTbl.SetBackgroundColor(m.theme.Background)
	m.containerTbl.SetSelectedStyle(tcell.StyleDefault.
		Background(m.theme.Muted).
		Foreground(m.theme.Primary))

	containerHeaders := []string{"STATUS", "NAME", "IMAGE", "PORTS", "UPTIME"}
	for i, h := range containerHeaders {
		cell := tview.NewTableCell(h).
			SetTextColor(m.theme.Secondary).
			SetSelectable(false).
			SetExpansion(1)
		if i == 0 {
			cell.SetExpansion(0)
		}
		m.containerTbl.SetCell(0, i, cell)
	}
	m.containerTbl.SetFixed(1, 0)
	m.containerTbl.SetInputCapture(m.handleContainerInput)

	m.containerTbl.SetSelectionChangedFunc(func(row, col int) {
		if row > 0 && row <= len(m.containers) {
			m.showContainerDetails(m.containers[row-1])
		}
	})

	// Image table
	m.imageTbl = tview.NewTable()
	m.imageTbl.SetBorders(false)
	m.imageTbl.SetSelectable(true, false)
	m.imageTbl.SetBackgroundColor(m.theme.Background)
	m.imageTbl.SetSelectedStyle(tcell.StyleDefault.
		Background(m.theme.Muted).
		Foreground(m.theme.Primary))

	imageHeaders := []string{"REPOSITORY", "TAG", "IMAGE ID", "SIZE", "CREATED"}
	for i, h := range imageHeaders {
		cell := tview.NewTableCell(h).
			SetTextColor(m.theme.Secondary).
			SetSelectable(false).
			SetExpansion(1)
		m.imageTbl.SetCell(0, i, cell)
	}
	m.imageTbl.SetFixed(1, 0)
	m.imageTbl.SetInputCapture(m.handleImageInput)

	// Pages for tabs
	m.pages = tview.NewPages()
	m.pages.AddPage("containers", m.containerTbl, true, true)
	m.pages.AddPage("images", m.imageTbl, true, false)

	// Content with border
	content := tview.NewFlex().SetDirection(tview.FlexRow)
	content.AddItem(m.pages, 0, 1, true)
	content.SetBorder(true)
	content.SetBorderColor(m.theme.Border)
	content.SetTitle(" 🐳 Docker Containers & Images ")
	content.SetTitleColor(m.theme.Primary)
	content.SetBackgroundColor(m.theme.Background)

	// Detail view
	m.detailView = tview.NewTextView()
	m.detailView.SetDynamicColors(true)
	m.detailView.SetScrollable(true)
	m.detailView.SetBorder(true)
	m.detailView.SetBorderColor(m.theme.Border)
	m.detailView.SetTitle(" Details / Logs ")
	m.detailView.SetTitleColor(m.theme.Primary)
	m.detailView.SetBackgroundColor(m.theme.Background)

	// Action bar
	m.actionView = tview.NewTextView()
	m.actionView.SetDynamicColors(true)
	m.actionView.SetBackgroundColor(m.theme.Muted)
	m.actionView.SetText(m.getActionsText())

	// Right panel
	rightPanel := tview.NewFlex().SetDirection(tview.FlexRow)
	rightPanel.AddItem(m.detailView, 0, 1, false)
	rightPanel.AddItem(m.actionView, 3, 0, false)

	// Main content
	mainContent := tview.NewFlex()
	mainContent.AddItem(content, 0, 2, true)
	mainContent.AddItem(rightPanel, 0, 1, false)

	// Layout
	m.view = tview.NewFlex().SetDirection(tview.FlexRow)
	m.view.AddItem(m.tabs, 1, 0, false)
	m.view.AddItem(mainContent, 0, 1, true)
	m.view.SetBackgroundColor(m.theme.Background)
}

// updateTabs updates the tab display
func (m *Manager) updateTabs() {
	containerStyle := "[white]"
	imageStyle := "[white]"

	if m.currentTab == "containers" {
		containerStyle = fmt.Sprintf("[%s::b]", colorToTag(m.theme.Primary))
	} else {
		imageStyle = fmt.Sprintf("[%s::b]", colorToTag(m.theme.Primary))
	}

	m.tabs.SetText(fmt.Sprintf("  %s📦 Containers (1)[white]  │  %s🖼 Images (2)[white]  ",
		containerStyle, imageStyle))
}

// getActionsText returns the actions help text
func (m *Manager) getActionsText() string {
	highlight := colorToTag(m.theme.Highlight)
	return fmt.Sprintf(
		"  [%s]s[white]=Start  [%s]S[white]=Stop  [%s]r[white]=Restart  [%s]d[white]=Delete  [%s]l[white]=Logs  [%s]e[white]=Exec Shell  [%s]i[white]=Inspect  [%s]Tab[white]=Switch",
		highlight, highlight, highlight, highlight, highlight, highlight, highlight, highlight)
}

// handleContainerInput handles container table input
func (m *Manager) handleContainerInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyTab:
		m.currentTab = "images"
		m.updateTabs()
		m.pages.SwitchToPage("images")
		return nil
	}

	if len(m.containers) == 0 {
		return event
	}

	row, _ := m.containerTbl.GetSelection()
	if row <= 0 || row > len(m.containers) {
		return event
	}
	container := m.containers[row-1]

	switch event.Rune() {
	case 's': // Start
		m.containerAction("start", container.Name)
		return nil
	case 'S': // Stop
		m.containerAction("stop", container.Name)
		return nil
	case 'r': // Restart
		m.containerAction("restart", container.Name)
		return nil
	case 'd': // Delete/Remove
		m.containerAction("rm -f", container.Name)
		return nil
	case 'l': // Logs
		m.showContainerLogs(container.Name)
		return nil
	case 'e': // Exec shell
		m.execContainer(container.Name)
		return nil
	case 'i': // Inspect
		m.inspectContainer(container.Name)
		return nil
	case 'p': // Pause/Unpause
		if container.Running {
			m.containerAction("pause", container.Name)
		} else {
			m.containerAction("unpause", container.Name)
		}
		return nil
	case 'j':
		return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
	case 'k':
		return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
	case '1':
		m.currentTab = "containers"
		m.updateTabs()
		m.pages.SwitchToPage("containers")
		return nil
	case '2':
		m.currentTab = "images"
		m.updateTabs()
		m.pages.SwitchToPage("images")
		return nil
	}

	return event
}

// handleImageInput handles image table input
func (m *Manager) handleImageInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyTab:
		m.currentTab = "containers"
		m.updateTabs()
		m.pages.SwitchToPage("containers")
		return nil
	}

	if len(m.images) == 0 {
		return event
	}

	row, _ := m.imageTbl.GetSelection()
	if row <= 0 || row > len(m.images) {
		return event
	}
	image := m.images[row-1]

	switch event.Rune() {
	case 'd': // Delete
		m.imageAction("rmi", image.ID)
		return nil
	case 'p': // Pull
		m.pullImage()
		return nil
	case 'i': // Inspect
		m.inspectImage(image.ID)
		return nil
	case 'R': // Run container from image
		m.runFromImage(image.Repo, image.Tag)
		return nil
	case 'j':
		return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
	case 'k':
		return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
	case '1':
		m.currentTab = "containers"
		m.updateTabs()
		m.pages.SwitchToPage("containers")
		return nil
	case '2':
		m.currentTab = "images"
		m.updateTabs()
		m.pages.SwitchToPage("images")
		return nil
	}

	return event
}

// SetApp sets the tview application reference
func (m *Manager) SetApp(app *tview.Application) {
	m.app = app
}

// SetHost sets the current host
func (m *Manager) SetHost(host *ssh.HostEntry) {
	m.host = host
}

// Refresh updates container and image lists
func (m *Manager) Refresh() error {
	if m.host == nil {
		return fmt.Errorf("no host set")
	}

	// Check if Docker is installed
	_, err := m.sshClient.RunCommand(*m.host, "which docker")
	if err != nil {
		m.detailView.SetText(fmt.Sprintf("[%s]Docker not installed on this host[white]\n\n"+
			"Install with:\n"+
			"  :install docker\n\n"+
			"Or manually:\n"+
			"  curl -fsSL https://get.docker.com | sh",
			colorToTag(m.theme.Warning)))
		return nil
	}

	// Get containers
	containerOutput, err := m.sshClient.RunCommand(*m.host,
		`docker ps -a --format "{{.ID}}|{{.Names}}|{{.Image}}|{{.Status}}|{{.Ports}}|{{.CreatedAt}}"`)
	if err != nil {
		return err
	}
	m.containers = m.parseContainers(containerOutput)
	m.updateContainerTable()

	// Get images
	imageOutput, _ := m.sshClient.RunCommand(*m.host,
		`docker images --format "{{.ID}}|{{.Repository}}|{{.Tag}}|{{.Size}}|{{.CreatedAt}}"`)
	m.images = m.parseImages(imageOutput)
	m.updateImageTable()

	// Show stats summary
	running := 0
	for _, c := range m.containers {
		if c.Running {
			running++
		}
	}
	m.detailView.SetText(fmt.Sprintf("[%s]Docker Status[white]\n\n"+
		"Containers: %d total, [%s]%d running[white]\n"+
		"Images: %d\n\n"+
		"[%s]Select a container for details[white]",
		colorToTag(m.theme.Primary),
		len(m.containers), colorToTag(m.theme.Success), running,
		len(m.images),
		colorToTag(m.theme.Muted)))

	return nil
}

// parseContainers parses docker ps output
func (m *Manager) parseContainers(output string) []Container {
	var containers []Container
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Split(line, "|")
		if len(fields) < 6 {
			continue
		}

		container := Container{
			ID:      fields[0],
			Name:    fields[1],
			Image:   fields[2],
			Status:  fields[3],
			Ports:   fields[4],
			Created: fields[5],
			Running: strings.Contains(strings.ToLower(fields[3]), "up"),
		}

		containers = append(containers, container)
	}

	return containers
}

// parseImages parses docker images output
func (m *Manager) parseImages(output string) []Image {
	var images []Image
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

		image := Image{
			ID:      fields[0][:12],
			Repo:    fields[1],
			Tag:     fields[2],
			Size:    fields[3],
			Created: fields[4],
		}

		images = append(images, image)
	}

	return images
}

// updateContainerTable updates the container table
func (m *Manager) updateContainerTable() {
	for row := m.containerTbl.GetRowCount() - 1; row > 0; row-- {
		m.containerTbl.RemoveRow(row)
	}

	for i, c := range m.containers {
		row := i + 1

		// Status icon
		icon := "○"
		iconColor := m.theme.Muted
		if c.Running {
			icon = "●"
			iconColor = m.theme.Success
		} else if strings.Contains(strings.ToLower(c.Status), "exited") {
			icon = "■"
			iconColor = m.theme.Error
		}
		m.containerTbl.SetCell(row, 0, tview.NewTableCell(icon).
			SetTextColor(iconColor).
			SetExpansion(0))

		// Name
		m.containerTbl.SetCell(row, 1, tview.NewTableCell(c.Name).
			SetTextColor(m.theme.Foreground).
			SetExpansion(1))

		// Image
		m.containerTbl.SetCell(row, 2, tview.NewTableCell(truncate(c.Image, 30)).
			SetTextColor(m.theme.Muted).
			SetExpansion(1))

		// Ports
		ports := c.Ports
		if len(ports) > 25 {
			ports = ports[:22] + "..."
		}
		m.containerTbl.SetCell(row, 3, tview.NewTableCell(ports).
			SetTextColor(m.theme.Secondary).
			SetExpansion(1))

		// Uptime/Status
		m.containerTbl.SetCell(row, 4, tview.NewTableCell(truncate(c.Status, 20)).
			SetTextColor(m.theme.Muted).
			SetExpansion(1))
	}

	if m.containerTbl.GetRowCount() > 1 {
		m.containerTbl.Select(1, 0)
	}
}

// updateImageTable updates the image table
func (m *Manager) updateImageTable() {
	for row := m.imageTbl.GetRowCount() - 1; row > 0; row-- {
		m.imageTbl.RemoveRow(row)
	}

	for i, img := range m.images {
		row := i + 1

		m.imageTbl.SetCell(row, 0, tview.NewTableCell(truncate(img.Repo, 30)).
			SetTextColor(m.theme.Foreground).
			SetExpansion(1))

		m.imageTbl.SetCell(row, 1, tview.NewTableCell(img.Tag).
			SetTextColor(m.theme.Secondary).
			SetExpansion(1))

		m.imageTbl.SetCell(row, 2, tview.NewTableCell(img.ID).
			SetTextColor(m.theme.Muted).
			SetExpansion(1))

		m.imageTbl.SetCell(row, 3, tview.NewTableCell(img.Size).
			SetTextColor(m.theme.Muted).
			SetExpansion(1))

		created := img.Created
		if len(created) > 10 {
			created = created[:10]
		}
		m.imageTbl.SetCell(row, 4, tview.NewTableCell(created).
			SetTextColor(m.theme.Muted).
			SetExpansion(1))
	}

	if m.imageTbl.GetRowCount() > 1 {
		m.imageTbl.Select(1, 0)
	}
}

// showContainerDetails shows details for a container
func (m *Manager) showContainerDetails(container Container) {
	primary := colorToTag(m.theme.Primary)
	success := colorToTag(m.theme.Success)
	muted := colorToTag(m.theme.Muted)

	status := "Stopped"
	statusColor := m.theme.Error
	if container.Running {
		status = "Running"
		statusColor = m.theme.Success
	}

	details := fmt.Sprintf(
		"[%s]Container: %s[white]\n\n"+
			"[%s]Status:[white] [%s]%s[white]\n"+
			"[%s]Image:[white] %s\n"+
			"[%s]ID:[white] %s\n"+
			"[%s]Ports:[white] %s\n"+
			"[%s]Created:[white] %s\n\n"+
			"[%s]Press:[white]\n"+
			"  s=Start  S=Stop  r=Restart\n"+
			"  l=Logs   e=Exec  i=Inspect",
		primary, container.Name,
		primary, colorToTag(statusColor), status,
		primary, container.Image,
		primary, container.ID,
		primary, container.Ports,
		primary, container.Created,
		muted,
	)

	// Get container stats if running
	if container.Running {
		statsOutput, _ := m.sshClient.RunCommand(*m.host,
			fmt.Sprintf("docker stats --no-stream --format '{{.CPUPerc}}|{{.MemUsage}}' %s 2>/dev/null", container.Name))
		if statsOutput != "" {
			stats := strings.Split(strings.TrimSpace(statsOutput), "|")
			if len(stats) >= 2 {
				details += fmt.Sprintf("\n\n[%s]Resources:[white]\n  CPU: %s\n  Memory: %s",
					success, stats[0], stats[1])
			}
		}
	}

	m.detailView.SetTitle(fmt.Sprintf(" Container: %s ", container.Name))
	m.detailView.SetText(details)
}

// Container actions
func (m *Manager) containerAction(action, name string) {
	cmd := fmt.Sprintf("docker %s %s", action, name)
	output, err := m.sshClient.RunCommand(*m.host, cmd)
	if err != nil {
		m.detailView.SetText(fmt.Sprintf("[%s]Error: %v[white]\n\n%s",
			colorToTag(m.theme.Error), err, output))
	} else {
		m.detailView.SetText(fmt.Sprintf("[%s]✓ %s %s: success[white]",
			colorToTag(m.theme.Success), action, name))
		m.Refresh()
	}
}

func (m *Manager) showContainerLogs(name string) {
	output, _ := m.sshClient.RunCommand(*m.host,
		fmt.Sprintf("docker logs --tail 100 %s 2>&1", name))
	m.detailView.SetTitle(fmt.Sprintf(" Logs: %s ", name))
	m.detailView.SetText(output)
}

func (m *Manager) execContainer(name string) {
	// Show command to exec into container
	m.detailView.SetTitle(fmt.Sprintf(" Exec: %s ", name))
	m.detailView.SetText(fmt.Sprintf(
		"[%s]Execute Shell in Container[white]\n\n"+
			"To get a shell in this container, run:\n\n"+
			"  docker exec -it %s /bin/bash\n\n"+
			"Or:\n\n"+
			"  docker exec -it %s /bin/sh\n\n"+
			"[%s]Note: Interactive shells require a terminal.\n"+
			"Use the SSH terminal to run these commands.[white]",
		colorToTag(m.theme.Primary), name, name, colorToTag(m.theme.Muted)))
}

func (m *Manager) inspectContainer(name string) {
	output, _ := m.sshClient.RunCommand(*m.host,
		fmt.Sprintf("docker inspect %s 2>&1 | head -80", name))
	m.detailView.SetTitle(fmt.Sprintf(" Inspect: %s ", name))
	m.detailView.SetText(output)
}

func (m *Manager) imageAction(action, id string) {
	cmd := fmt.Sprintf("docker %s %s", action, id)
	_, err := m.sshClient.RunCommand(*m.host, cmd)
	if err != nil {
		m.detailView.SetText(fmt.Sprintf("[%s]Error: %v[white]", colorToTag(m.theme.Error), err))
	} else {
		m.Refresh()
	}
}

func (m *Manager) pullImage() {
	m.detailView.SetText(fmt.Sprintf(
		"[%s]Pull Image[white]\n\nUse command:\n\n  :run docker pull <image:tag>\n\nExample:\n  :run docker pull nginx:latest",
		colorToTag(m.theme.Primary)))
}

func (m *Manager) runFromImage(repo, tag string) {
	image := repo
	if tag != "" && tag != "<none>" {
		image = repo + ":" + tag
	}
	m.detailView.SetText(fmt.Sprintf(
		"[%s]Run Container from Image[white]\n\n"+
			"Use command:\n\n"+
			"  :run docker run -d --name mycontainer %s\n\n"+
			"With port mapping:\n\n"+
			"  :run docker run -d -p 8080:80 --name mycontainer %s",
		colorToTag(m.theme.Primary), image, image))
}

func (m *Manager) inspectImage(id string) {
	output, _ := m.sshClient.RunCommand(*m.host,
		fmt.Sprintf("docker image inspect %s 2>&1 | head -60", id))
	m.detailView.SetTitle(fmt.Sprintf(" Inspect Image: %s ", id))
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
