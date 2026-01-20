package installer

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/systask/systask/internal/config"
	"github.com/systask/systask/internal/ssh"
)

// Package represents an installable package
type Package struct {
	Name        string
	Description string
	Category    string
	Installed   bool
	Icon        string
}

// DefaultPackages returns common server packages
func DefaultPackages() []Package {
	return []Package{
		// Web Servers
		{Name: "nginx", Description: "High-performance HTTP server", Category: "Web Servers", Icon: "🌐"},
		{Name: "apache2", Description: "Apache HTTP Server", Category: "Web Servers", Icon: "🌐"},
		{Name: "caddy", Description: "Fast, easy HTTPS server", Category: "Web Servers", Icon: "🌐"},

		// Databases
		{Name: "postgresql", Description: "PostgreSQL database server", Category: "Databases", Icon: "🗄️"},
		{Name: "mysql-server", Description: "MySQL database server", Category: "Databases", Icon: "🗄️"},
		{Name: "mariadb-server", Description: "MariaDB database server", Category: "Databases", Icon: "🗄️"},
		{Name: "redis", Description: "In-memory data store", Category: "Databases", Icon: "🗄️"},
		{Name: "mongodb", Description: "Document database", Category: "Databases", Icon: "🗄️"},

		// Containers
		{Name: "docker", Description: "Container runtime", Category: "Containers", Icon: "🐳"},
		{Name: "docker-compose", Description: "Docker Compose tool", Category: "Containers", Icon: "🐳"},
		{Name: "podman", Description: "Daemonless container engine", Category: "Containers", Icon: "🐳"},
		{Name: "containerd", Description: "Container runtime", Category: "Containers", Icon: "🐳"},

		// Dev Tools
		{Name: "git", Description: "Version control system", Category: "Dev Tools", Icon: "🔧"},
		{Name: "nodejs", Description: "JavaScript runtime", Category: "Dev Tools", Icon: "🔧"},
		{Name: "python3", Description: "Python 3 interpreter", Category: "Dev Tools", Icon: "🔧"},
		{Name: "golang", Description: "Go programming language", Category: "Dev Tools", Icon: "🔧"},
		{Name: "rustc", Description: "Rust compiler", Category: "Dev Tools", Icon: "🔧"},

		// Security
		{Name: "ufw", Description: "Uncomplicated Firewall", Category: "Security", Icon: "🔒"},
		{Name: "fail2ban", Description: "Intrusion prevention", Category: "Security", Icon: "🔒"},
		{Name: "certbot", Description: "Let's Encrypt client", Category: "Security", Icon: "🔒"},
		{Name: "wireguard", Description: "VPN solution", Category: "Security", Icon: "🔒"},

		// Monitoring
		{Name: "htop", Description: "Interactive process viewer", Category: "Monitoring", Icon: "📊"},
		{Name: "netdata", Description: "Real-time monitoring", Category: "Monitoring", Icon: "📊"},
		{Name: "prometheus", Description: "Monitoring system", Category: "Monitoring", Icon: "📊"},
		{Name: "grafana", Description: "Visualization platform", Category: "Monitoring", Icon: "📊"},

		// Utilities
		{Name: "curl", Description: "URL transfer tool", Category: "Utilities", Icon: "⚡"},
		{Name: "wget", Description: "Network downloader", Category: "Utilities", Icon: "⚡"},
		{Name: "tmux", Description: "Terminal multiplexer", Category: "Utilities", Icon: "⚡"},
		{Name: "vim", Description: "Text editor", Category: "Utilities", Icon: "⚡"},
		{Name: "neovim", Description: "Hyperextensible Vim", Category: "Utilities", Icon: "⚡"},
		{Name: "fzf", Description: "Fuzzy finder", Category: "Utilities", Icon: "⚡"},
		{Name: "ripgrep", Description: "Fast grep alternative", Category: "Utilities", Icon: "⚡"},
		{Name: "jq", Description: "JSON processor", Category: "Utilities", Icon: "⚡"},
	}
}

// Manager manages package installation
type Manager struct {
	view         *tview.Flex
	categoryList *tview.List
	packageList  *tview.List
	detailView   *tview.TextView
	statusView   *tview.TextView
	theme        *config.Theme
	sshClient    *ssh.Client
	host         *ssh.HostEntry
	app          *tview.Application

	packages        []Package
	categories      []string
	currentCat      string
	pkgManager      string
	focusOnPackages bool
}

// NewManager creates a new installer manager
func NewManager(theme *config.Theme, client *ssh.Client) *Manager {
	m := &Manager{
		theme:     theme,
		sshClient: client,
		packages:  DefaultPackages(),
	}
	m.buildCategories()
	m.build()
	return m
}

// buildCategories extracts unique categories
func (m *Manager) buildCategories() {
	seen := make(map[string]bool)
	for _, pkg := range m.packages {
		if !seen[pkg.Category] {
			m.categories = append(m.categories, pkg.Category)
			seen[pkg.Category] = true
		}
	}
	if len(m.categories) > 0 {
		m.currentCat = m.categories[0]
	}
}

// build constructs the installer view
func (m *Manager) build() {
	// Category list
	m.categoryList = tview.NewList()
	m.categoryList.ShowSecondaryText(false)
	m.categoryList.SetBackgroundColor(m.theme.Background)
	m.categoryList.SetBorder(true)
	m.categoryList.SetBorderColor(m.theme.Border)
	m.categoryList.SetTitle(" 📦 Categories ")
	m.categoryList.SetTitleColor(m.theme.Primary)
	m.categoryList.SetMainTextColor(m.theme.Foreground)
	m.categoryList.SetSelectedBackgroundColor(m.theme.Muted)
	m.categoryList.SetSelectedTextColor(m.theme.Primary)

	for _, cat := range m.categories {
		m.categoryList.AddItem(cat, "", 0, nil)
	}

	m.categoryList.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		m.currentCat = mainText
		m.updatePackageList()
	})

	// Tab handler for category list
	m.categoryList.SetInputCapture(m.handleCategoryInput)

	// Package list
	m.packageList = tview.NewList()
	m.packageList.ShowSecondaryText(true)
	m.packageList.SetBackgroundColor(m.theme.Background)
	m.packageList.SetBorder(true)
	m.packageList.SetBorderColor(m.theme.Border)
	m.packageList.SetTitle(" Packages ")
	m.packageList.SetTitleColor(m.theme.Primary)
	m.packageList.SetMainTextColor(m.theme.Foreground)
	m.packageList.SetSecondaryTextColor(m.theme.Muted)
	m.packageList.SetSelectedBackgroundColor(m.theme.Muted)
	m.packageList.SetSelectedTextColor(m.theme.Primary)
	m.packageList.SetInputCapture(m.handlePackageInput)

	// Detail view
	m.detailView = tview.NewTextView()
	m.detailView.SetDynamicColors(true)
	m.detailView.SetScrollable(true)
	m.detailView.SetBorder(true)
	m.detailView.SetBorderColor(m.theme.Border)
	m.detailView.SetTitle(" Installation Guide ")
	m.detailView.SetTitleColor(m.theme.Primary)
	m.detailView.SetBackgroundColor(m.theme.Background)
	m.detailView.SetText(m.getWelcomeText())

	// Status view
	m.statusView = tview.NewTextView()
	m.statusView.SetDynamicColors(true)
	m.statusView.SetBackgroundColor(m.theme.Muted)
	m.statusView.SetText(m.getStatusText())

	// Left panel
	leftPanel := tview.NewFlex().SetDirection(tview.FlexRow)
	leftPanel.AddItem(m.categoryList, 0, 1, true)

	// Center panel
	centerPanel := tview.NewFlex().SetDirection(tview.FlexRow)
	centerPanel.AddItem(m.packageList, 0, 1, false)

	// Right panel
	rightPanel := tview.NewFlex().SetDirection(tview.FlexRow)
	rightPanel.AddItem(m.detailView, 0, 1, false)
	rightPanel.AddItem(m.statusView, 2, 0, false)

	// Layout
	m.view = tview.NewFlex()
	m.view.AddItem(leftPanel, 16, 0, true) // Reduced for small screens
	m.view.AddItem(centerPanel, 0, 1, false)
	m.view.AddItem(rightPanel, 0, 1, false)
	m.view.SetBackgroundColor(m.theme.Background)

	m.updatePackageList()
}

// getWelcomeText returns the welcome text
func (m *Manager) getWelcomeText() string {
	primary := colorToTag(m.theme.Primary)
	highlight := colorToTag(m.theme.Highlight)

	return fmt.Sprintf(`[%s]🚀 Service Installer[white]

Quick installation of common server packages.

[%s]Navigation:[white]
  [%s]Tab[white]      Switch panels
  [%s]j/k[white]      Move up/down
  [%s]Enter[white]    Install package
  [%s]u[white]        Uninstall package
  [%s]i[white]        Show package info
  [%s]r[white]        Refresh status

[%s]Supported Package Managers:[white]
  • apt (Debian/Ubuntu)
  • dnf/yum (RHEL/Fedora)
  • pacman (Arch)
  • zypper (openSUSE)

Connect to a server first to enable installation.`,
		primary, primary,
		highlight, highlight, highlight,
		highlight, highlight, highlight,
		primary)
}

// getStatusText returns the status bar text
func (m *Manager) getStatusText() string {
	highlight := colorToTag(m.theme.Highlight)
	return fmt.Sprintf("  [%s]Enter[white]=Install  [%s]u[white]=Uninstall  [%s]i[white]=Info  [%s]Tab[white]=Switch  [%s]ESC[white]=Home",
		highlight, highlight, highlight, highlight, highlight)
}

// handlePackageInput handles package list input
func (m *Manager) handlePackageInput(event *tcell.EventKey) *tcell.EventKey {
	idx := m.packageList.GetCurrentItem()
	pkg := m.getPackageByIndex(idx)
	if pkg == nil {
		return event
	}

	switch event.Key() {
	case tcell.KeyEnter:
		m.installPackage(pkg.Name)
		return nil
	case tcell.KeyTab:
		// Switch back to categories
		m.focusOnPackages = false
		if m.app != nil {
			m.app.SetFocus(m.categoryList)
		}
		return nil
	}

	switch event.Rune() {
	case 'u':
		m.uninstallPackage(pkg.Name)
		return nil
	case 'i':
		m.showPackageInfo(pkg.Name)
		return nil
	case 'j':
		return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
	case 'k':
		return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
	}

	return event
}

// handleCategoryInput handles category list input
func (m *Manager) handleCategoryInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyTab:
		// Switch to packages
		m.focusOnPackages = true
		if m.app != nil {
			m.app.SetFocus(m.packageList)
		}
		return nil
	}

	switch event.Rune() {
	case 'j':
		return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
	case 'k':
		return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
	}

	return event
}

// getPackageByIndex gets package from current category by index
func (m *Manager) getPackageByIndex(idx int) *Package {
	count := 0
	for i := range m.packages {
		if m.packages[i].Category == m.currentCat {
			if count == idx {
				return &m.packages[i]
			}
			count++
		}
	}
	return nil
}

// updatePackageList updates the package list for current category
func (m *Manager) updatePackageList() {
	m.packageList.Clear()
	m.packageList.SetTitle(fmt.Sprintf(" %s ", m.currentCat))

	for _, pkg := range m.packages {
		if pkg.Category != m.currentCat {
			continue
		}

		icon := pkg.Icon
		status := ""
		if pkg.Installed {
			status = " ✓"
		}

		mainText := fmt.Sprintf("%s %s%s", icon, pkg.Name, status)
		m.packageList.AddItem(mainText, pkg.Description, 0, nil)
	}
}

// SetHost sets the current host
func (m *Manager) SetHost(host *ssh.HostEntry) {
	m.host = host
}

// SetApp sets the tview application reference
func (m *Manager) SetApp(app *tview.Application) {
	m.app = app
}

// Focus sets focus to the active list
func (m *Manager) Focus() {
	if m.app != nil {
		if m.focusOnPackages {
			m.app.SetFocus(m.packageList)
		} else {
			m.app.SetFocus(m.categoryList)
		}
	}
}

// Refresh detects package manager and checks installed packages
func (m *Manager) Refresh() error {
	if m.host == nil {
		return fmt.Errorf("no host set")
	}

	// Detect package manager
	m.pkgManager = m.detectPackageManager()

	// Check installed status for each package
	for i := range m.packages {
		installed := m.isPackageInstalled(m.packages[i].Name)
		m.packages[i].Installed = installed
	}

	m.updatePackageList()
	m.detailView.SetText(fmt.Sprintf("[%s]Package Manager:[white] %s\n\n%s",
		colorToTag(m.theme.Primary), m.pkgManager, m.getWelcomeText()))

	return nil
}

// detectPackageManager detects the system's package manager
func (m *Manager) detectPackageManager() string {
	// Simple which/type command checks - return immediately on first match
	managers := []string{"apt-get", "apt", "dnf", "yum", "pacman", "yay", "zypper", "apk"}

	for _, mgr := range managers {
		cmd := fmt.Sprintf("which %s 2>/dev/null || type %s 2>/dev/null || command -v %s 2>/dev/null", mgr, mgr, mgr)
		if output, err := m.sshClient.RunCommand(*m.host, cmd); err == nil && strings.TrimSpace(output) != "" {
			// Return the canonical manager name
			if mgr == "apt-get" || mgr == "apt" {
				return "apt"
			}
			return mgr
		}
	}

	// Fallback: check for package manager binaries directly
	fallbackChecks := []struct {
		path    string
		manager string
	}{
		{"/usr/bin/apt-get", "apt"},
		{"/usr/bin/apt", "apt"},
		{"/usr/bin/dnf", "dnf"},
		{"/usr/bin/yum", "yum"},
		{"/usr/bin/pacman", "pacman"},
		{"/usr/bin/zypper", "zypper"},
	}

	for _, check := range fallbackChecks {
		if _, err := m.sshClient.RunCommand(*m.host, "test -x "+check.path); err == nil {
			return check.manager
		}
	}

	// Last resort: check /etc files for distro
	if output, err := m.sshClient.RunCommand(*m.host, "cat /etc/os-release 2>/dev/null"); err == nil {
		lower := strings.ToLower(output)
		// Debian/Ubuntu family
		if strings.Contains(lower, "ubuntu") || strings.Contains(lower, "debian") || strings.Contains(lower, "mint") {
			return "apt"
		}
		// Amazon Linux uses yum/dnf
		if strings.Contains(lower, "amzn") || strings.Contains(lower, "amazon") {
			// Amazon Linux 2023+ uses dnf, older uses yum
			if strings.Contains(lower, "2023") {
				return "dnf"
			}
			return "yum"
		}
		// Red Hat family
		if strings.Contains(lower, "fedora") || strings.Contains(lower, "rhel") || strings.Contains(lower, "centos") || strings.Contains(lower, "rocky") || strings.Contains(lower, "alma") {
			return "dnf"
		}
		// Arch family
		if strings.Contains(lower, "arch") || strings.Contains(lower, "manjaro") || strings.Contains(lower, "endeavour") {
			return "pacman"
		}
		// SUSE family
		if strings.Contains(lower, "opensuse") || strings.Contains(lower, "suse") {
			return "zypper"
		}
		// Alpine
		if strings.Contains(lower, "alpine") {
			return "apk"
		}
	}

	return "unknown"
}

// isPackageInstalled checks if a package is installed
func (m *Manager) isPackageInstalled(name string) bool {
	if m.host == nil {
		return false
	}

	var cmd string
	switch m.pkgManager {
	case "apt":
		cmd = fmt.Sprintf("dpkg -l %s 2>/dev/null | grep -q '^ii'", name)
	case "dnf", "yum":
		cmd = fmt.Sprintf("rpm -q %s 2>/dev/null", name)
	case "pacman", "yay":
		cmd = fmt.Sprintf("pacman -Qi %s 2>/dev/null", name)
	case "zypper":
		cmd = fmt.Sprintf("rpm -q %s 2>/dev/null", name)
	case "apk":
		cmd = fmt.Sprintf("apk info -e %s 2>/dev/null", name)
	default:
		cmd = fmt.Sprintf("command -v %s 2>/dev/null", name)
	}

	_, err := m.sshClient.RunCommand(*m.host, cmd)
	return err == nil
}

// installPackage installs a package
func (m *Manager) installPackage(name string) {
	if m.host == nil {
		m.detailView.SetText(fmt.Sprintf("[%s]Not connected to any server[white]",
			colorToTag(m.theme.Error)))
		return
	}

	var cmd string
	switch m.pkgManager {
	case "apt":
		cmd = fmt.Sprintf("sudo apt-get update && sudo apt-get install -y %s", name)
	case "dnf":
		cmd = fmt.Sprintf("sudo dnf install -y %s", name)
	case "yum":
		cmd = fmt.Sprintf("sudo yum install -y %s", name)
	case "pacman":
		cmd = fmt.Sprintf("sudo pacman -S --noconfirm %s", name)
	case "yay":
		cmd = fmt.Sprintf("yay -S --noconfirm %s", name)
	case "zypper":
		cmd = fmt.Sprintf("sudo zypper install -y %s", name)
	case "apk":
		cmd = fmt.Sprintf("sudo apk add --no-cache %s", name)
	default:
		m.detailView.SetText(fmt.Sprintf("[%s]Unknown package manager: %s[white]\n\n"+
			"Detected: %s\n\n"+
			"Please install manually with your package manager.",
			colorToTag(m.theme.Error), m.pkgManager, m.pkgManager))
		return
	}

	m.detailView.SetText(fmt.Sprintf("[%s]Installing %s...[white]\n\nCommand: %s",
		colorToTag(m.theme.Warning), name, cmd))

	go func() {
		output, err := m.sshClient.RunCommand(*m.host, cmd)
		if err != nil {
			m.detailView.SetText(fmt.Sprintf("[%s]Installation failed:[white]\n\n%v\n\n%s",
				colorToTag(m.theme.Error), err, output))
		} else {
			m.detailView.SetText(fmt.Sprintf("[%s]✓ %s installed successfully![white]\n\n%s",
				colorToTag(m.theme.Success), name, truncate(output, 500)))
			m.Refresh()
		}
	}()
}

// uninstallPackage removes a package
func (m *Manager) uninstallPackage(name string) {
	if m.host == nil {
		return
	}

	var cmd string
	switch m.pkgManager {
	case "apt":
		cmd = fmt.Sprintf("sudo apt-get remove -y %s", name)
	case "dnf":
		cmd = fmt.Sprintf("sudo dnf remove -y %s", name)
	case "yum":
		cmd = fmt.Sprintf("sudo yum remove -y %s", name)
	case "pacman":
		cmd = fmt.Sprintf("sudo pacman -Rs --noconfirm %s", name)
	case "yay":
		cmd = fmt.Sprintf("yay -Rs --noconfirm %s", name)
	case "zypper":
		cmd = fmt.Sprintf("sudo zypper remove -y %s", name)
	case "apk":
		cmd = fmt.Sprintf("sudo apk del %s", name)
	default:
		return
	}

	m.detailView.SetText(fmt.Sprintf("[%s]Uninstalling %s...[white]",
		colorToTag(m.theme.Warning), name))

	go func() {
		output, err := m.sshClient.RunCommand(*m.host, cmd)
		if err != nil {
			m.detailView.SetText(fmt.Sprintf("[%s]Uninstall failed:[white]\n\n%v",
				colorToTag(m.theme.Error), err))
		} else {
			m.detailView.SetText(fmt.Sprintf("[%s]✓ %s uninstalled[white]\n\n%s",
				colorToTag(m.theme.Success), name, truncate(output, 500)))
			m.Refresh()
		}
	}()
}

// showPackageInfo shows package information
func (m *Manager) showPackageInfo(name string) {
	if m.host == nil {
		return
	}

	var cmd string
	switch m.pkgManager {
	case "apt":
		cmd = fmt.Sprintf("apt-cache show %s 2>/dev/null | head -30", name)
	case "dnf", "yum":
		cmd = fmt.Sprintf("dnf info %s 2>/dev/null || yum info %s 2>/dev/null | head -30", name, name)
	case "pacman", "yay":
		cmd = fmt.Sprintf("pacman -Si %s 2>/dev/null || pacman -Qi %s 2>/dev/null | head -30", name, name)
	case "zypper":
		cmd = fmt.Sprintf("zypper info %s 2>/dev/null | head -30", name)
	case "apk":
		cmd = fmt.Sprintf("apk info %s 2>/dev/null | head -30", name)
	default:
		cmd = fmt.Sprintf("command -v %s; %s --version 2>/dev/null || %s -v 2>/dev/null", name, name, name)
	}

	output, _ := m.sshClient.RunCommand(*m.host, cmd)
	m.detailView.SetTitle(fmt.Sprintf(" Info: %s ", name))
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
	lines := strings.Split(s, "\n")
	if len(lines) > 20 {
		return strings.Join(lines[:20], "\n") + "\n..."
	}
	return s[:max-3] + "..."
}
