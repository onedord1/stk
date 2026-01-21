package app

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/systask/systask/internal/config"
	"github.com/systask/systask/internal/modules/batch"
	"github.com/systask/systask/internal/modules/disk"
	"github.com/systask/systask/internal/modules/docker"
	"github.com/systask/systask/internal/modules/health"
	"github.com/systask/systask/internal/modules/installer"
	"github.com/systask/systask/internal/modules/logs"
	"github.com/systask/systask/internal/modules/processes"
	"github.com/systask/systask/internal/modules/services"
	"github.com/systask/systask/internal/modules/sftp"
	"github.com/systask/systask/internal/modules/users"
	"github.com/systask/systask/internal/ssh"
	"github.com/systask/systask/internal/ui"
)

// App represents the main application
type App struct {
	config    *config.Config
	theme     *config.Theme
	tviewApp  *tview.Application
	layout    *ui.Layout
	sshClient *ssh.Client
	executor  *ssh.Executor

	// UI components
	commandPalette *ui.CommandPalette
	searchPalette  *ui.SearchPalette
	helpOverlay    *ui.HelpOverlay

	// Modules
	healthModule    *health.Dashboard
	servicesModule  *services.Manager
	processesModule *processes.Manager
	logsModule      *logs.Viewer
	diskModule      *disk.Manager
	batchModule     *batch.Executor
	usersModule     *users.Manager
	dockerModule    *docker.Manager
	installerModule *installer.Manager
	sftpModule      *sftp.Manager

	// State
	hosts         []ssh.HostEntry
	selectedHost  *ssh.HostEntry
	currentModule string
	helpVisible   bool
	hasFzf        bool
}

// New creates a new application
func New(cfg *config.Config) *App {
	theme := config.GetTheme(cfg.Theme)

	app := &App{
		config:        cfg,
		theme:         theme,
		tviewApp:      tview.NewApplication(),
		sshClient:     ssh.NewClient(time.Duration(cfg.SSHTimeout) * time.Second),
		currentModule: "welcome",
	}

	app.executor = ssh.NewExecutor(app.sshClient, cfg.MaxParallel)
	app.hasFzf = app.checkFzf()

	return app
}

// checkFzf checks if fzf is installed
func (a *App) checkFzf() bool {
	_, err := exec.LookPath("fzf")
	return err == nil
}

// Run starts the application
func (a *App) Run() error {
	// Initialize UI
	a.layout = ui.NewLayout(a.tviewApp, a.theme)
	a.commandPalette = ui.NewCommandPalette(a.theme)
	a.searchPalette = ui.NewSearchPalette(a.theme)
	a.helpOverlay = ui.NewHelpOverlay(a.theme)

	// Initialize modules
	a.healthModule = health.NewDashboard(a.theme, a.sshClient)
	a.servicesModule = services.NewManager(a.theme, a.sshClient)
	a.processesModule = processes.NewManager(a.theme, a.sshClient)
	a.logsModule = logs.NewViewer(a.theme, a.sshClient)
	a.diskModule = disk.NewManager(a.theme, a.sshClient)
	a.batchModule = batch.NewExecutor(a.theme, a.sshClient, a.executor)
	a.usersModule = users.NewManager(a.theme, a.sshClient)
	a.dockerModule = docker.NewManager(a.theme, a.sshClient)
	a.installerModule = installer.NewManager(a.theme, a.sshClient)
	a.sftpModule = sftp.NewManager(a.theme, a.sshClient)

	// Add module pages
	a.layout.MainView().AddPage("health", a.healthModule.View(), true, false)
	a.layout.MainView().AddPage("services", a.servicesModule.View(), true, false)
	a.layout.MainView().AddPage("processes", a.processesModule.View(), true, false)
	a.layout.MainView().AddPage("logs", a.logsModule.View(), true, false)
	a.layout.MainView().AddPage("disks", a.diskModule.View(), true, false)
	a.layout.MainView().AddPage("batch", a.batchModule.View(), true, false)
	a.layout.MainView().AddPage("users", a.usersModule.View(), true, false)
	a.layout.MainView().AddPage("docker", a.dockerModule.View(), true, false)
	a.layout.MainView().AddPage("installer", a.installerModule.View(), true, false)
	a.layout.MainView().AddPage("sftp", a.sftpModule.View(), true, false)

	// Discover SSH hosts
	a.discoverHosts()

	// Set up callbacks
	a.setupCallbacks()

	// Set up global key handlers
	a.setupKeyBindings()

	// Start status bar clock
	a.layout.StatusBar().StartClock(a.tviewApp)

	// Set module app references
	a.batchModule.SetApp(a.tviewApp)
	a.dockerModule.SetApp(a.tviewApp)
	a.servicesModule.SetApp(a.tviewApp)
	a.sftpModule.SetApp(a.tviewApp)
	a.processesModule.SetApp(a.tviewApp)
	a.logsModule.SetApp(a.tviewApp)
	a.diskModule.SetApp(a.tviewApp)
	a.usersModule.SetApp(a.tviewApp)
	a.installerModule.SetApp(a.tviewApp)

	// Show fzf status
	if a.hasFzf {
		a.layout.StatusBar().SetSuccess("fzf detected - enhanced menus available")
	}

	// Run the application
	a.tviewApp.SetRoot(a.layout.Root(), true)
	a.tviewApp.EnableMouse(false) // Keyboard-only

	return a.tviewApp.Run()
}

// discoverHosts loads hosts from SSH config
func (a *App) discoverHosts() {
	hosts, err := ssh.DiscoverHosts(a.config.SSHConfigPath, a.config.KnownHostsPath)
	if err != nil {
		a.layout.StatusBar().SetAlert("Could not load SSH hosts", true)
		return
	}

	// Add manually configured hosts (with PEM key support)
	for _, h := range a.config.Hosts {
		hosts = append(hosts, ssh.HostEntry{
			Name:     h.Name,
			Hostname: h.Hostname,
			Port:     h.Port,
			User:     h.User,
			KeyFile:  h.KeyPath,
			Source:   h.Provider,
			Group:    h.Group,
		})
	}

	a.hosts = hosts
	a.layout.ServerList().SetServers(hosts)
	a.batchModule.SetHosts(hosts)
	a.sftpModule.SetHosts(hosts)
	a.layout.StatusBar().SetStatus(fmt.Sprintf("Found %d hosts | fzf: %v", len(hosts), a.hasFzf))
}

// setupCallbacks configures UI callbacks
func (a *App) setupCallbacks() {
	// Server selection
	a.layout.ServerList().OnSelect(func(host ssh.HostEntry) {
		a.selectedHost = &host
		name := host.Name
		if name == "" {
			name = host.Hostname
		}
		a.layout.StatusBar().SetStatus(fmt.Sprintf("Selected: %s", name))
	})

	// Server connection
	a.layout.ServerList().OnConnect(func(host ssh.HostEntry) {
		a.connectToHost(host)
	})

	// Add server button
	a.layout.ServerList().OnAddHost(func() {
		a.showAddHostForm()
	})

	// Delete server
	a.layout.ServerList().OnDeleteHost(func(host ssh.HostEntry) {
		a.deleteHost(host)
	})

	// Edit server
	a.layout.ServerList().OnEditHost(func(host ssh.HostEntry) {
		a.editHost(host)
	})

	// Command execution
	a.commandPalette.OnExecute(func(cmd ui.Command) {
		a.executeCommand(cmd)
		a.layout.StatusBar().SetMode("NORMAL")
		a.tviewApp.SetFocus(a.layout.ServerList().View())
	})

	a.commandPalette.OnCancel(func() {
		a.layout.StatusBar().SetMode("NORMAL")
		a.tviewApp.SetRoot(a.layout.Root(), true)
		a.tviewApp.SetFocus(a.layout.ServerList().View())
	})

	// Search
	a.searchPalette.OnSearch(func(query string) {
		a.filterServers(query)
	})

	a.searchPalette.OnCancel(func() {
		a.layout.StatusBar().SetMode("NORMAL")
		a.layout.ServerList().SetServers(a.hosts)
		a.tviewApp.SetRoot(a.layout.Root(), true)
		a.tviewApp.SetFocus(a.layout.ServerList().View())
	})
}

// showAddHostForm displays the add host form
func (a *App) showAddHostForm() {
	form := ui.NewAddHostForm(a.theme)

	form.OnSubmit(func(host config.Host) {
		// Save to config
		a.config.AddHost(host)
		a.config.Save()

		// Add to hosts list
		a.hosts = append(a.hosts, ssh.HostEntry{
			Name:     host.Name,
			Hostname: host.Hostname,
			Port:     host.Port,
			User:     host.User,
			KeyFile:  host.KeyPath,
			Source:   "manual", // Mark as manually added for edit/delete
			Group:    host.Group,
		})

		// Refresh server list
		a.layout.ServerList().SetServers(a.hosts)
		a.sftpModule.SetHosts(a.hosts)
		a.batchModule.SetHosts(a.hosts)

		// Return to main view
		a.layout.MainView().RemovePage("addhost")
		a.layout.MainView().SwitchToPage("welcome")
		a.tviewApp.SetFocus(a.layout.ServerList().View())
		a.layout.StatusBar().SetSuccess(fmt.Sprintf("Added server: %s", host.Name))
	})

	form.OnCancel(func() {
		a.layout.MainView().RemovePage("addhost")
		a.layout.MainView().SwitchToPage("welcome")
		a.tviewApp.SetFocus(a.layout.ServerList().View())
	})

	a.layout.MainView().AddPage("addhost", form.View(), true, true)
	form.Focus(a.tviewApp)
}

// deleteHost removes a host from the config
func (a *App) deleteHost(host ssh.HostEntry) {
	// Only allow deleting manually added hosts (not from ~/.ssh/config)
	// Manually added hosts have Source = "manual", "config", "custom", "Custom", "Cloud", "SSH Config"
	allowedSources := []string{"manual", "config", "custom", "Custom", "Cloud", "SSH Config"}
	canDelete := false
	for _, s := range allowedSources {
		if host.Source == s {
			canDelete = true
			break
		}
	}
	// Also allow if it's in our config file
	if a.config.GetHost(host.Name) != nil {
		canDelete = true
	}

	if !canDelete {
		a.layout.StatusBar().SetAlert("Can only delete manually added hosts", true)
		return
	}

	// Remove from config
	a.config.RemoveHost(host.Name)
	a.config.Save()

	// Remove from hosts list
	for i, h := range a.hosts {
		if h.Name == host.Name || h.Hostname == host.Hostname {
			a.hosts = append(a.hosts[:i], a.hosts[i+1:]...)
			break
		}
	}

	// Refresh displays
	a.layout.ServerList().SetServers(a.hosts)
	a.sftpModule.SetHosts(a.hosts)
	a.batchModule.SetHosts(a.hosts)
	a.layout.StatusBar().SetSuccess(fmt.Sprintf("Deleted: %s", host.Name))
}

// editHost shows the edit form for a host
func (a *App) editHost(host ssh.HostEntry) {
	// Only allow editing manually added hosts (check if in config file)
	if a.config.GetHost(host.Name) == nil {
		a.layout.StatusBar().SetAlert("Can only edit manually added hosts", true)
		return
	}

	form := ui.NewEditHostForm(a.theme, host)
	originalName := host.Name

	form.OnSubmit(func(updatedHost config.Host) {
		// Remove old host from config
		a.config.RemoveHost(originalName)
		// Add updated host
		a.config.AddHost(updatedHost)
		a.config.Save()

		// Update hosts list
		for i, h := range a.hosts {
			if h.Name == originalName || h.Hostname == host.Hostname {
				a.hosts[i] = ssh.HostEntry{
					Name:     updatedHost.Name,
					Hostname: updatedHost.Hostname,
					Port:     updatedHost.Port,
					User:     updatedHost.User,
					KeyFile:  updatedHost.KeyPath,
					Source:   "manual",
					Group:    updatedHost.Group,
				}
				break
			}
		}

		// Refresh displays
		a.layout.ServerList().SetServers(a.hosts)
		a.sftpModule.SetHosts(a.hosts)
		a.batchModule.SetHosts(a.hosts)

		// Return to main view
		a.layout.MainView().RemovePage("addhost")
		a.layout.MainView().SwitchToPage("welcome")
		a.tviewApp.SetFocus(a.layout.ServerList().View())
		a.layout.StatusBar().SetSuccess(fmt.Sprintf("Updated server: %s", updatedHost.Name))
	})

	form.OnCancel(func() {
		a.layout.MainView().RemovePage("addhost")
		a.layout.MainView().SwitchToPage("welcome")
		a.tviewApp.SetFocus(a.layout.ServerList().View())
	})

	a.layout.MainView().AddPage("addhost", form.View(), true, true)
	form.Focus(a.tviewApp)
}

// setupKeyBindings configures global key handlers
func (a *App) setupKeyBindings() {
	a.tviewApp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Check if we're in add host form - let form handle all input
		currentPage, _ := a.layout.MainView().GetFrontPage()
		if currentPage == "addhost" {
			return event // Let form handle everything
		}

		// Check if help is visible
		if a.helpVisible {
			if event.Key() == tcell.KeyEsc || event.Rune() == '?' {
				a.helpOverlay.Hide(a.layout.MainView())
				a.helpVisible = false
				return nil
			}
			return event
		}

		// Handle command mode
		if a.commandPalette.IsActive() {
			return event
		}

		// Handle search mode
		if a.searchPalette.IsActive() {
			return event
		}

		// Global shortcuts
		switch event.Key() {
		case tcell.KeyEsc:
			// ALWAYS go back to main menu/welcome screen
			// But not from add host form
			currentPage, _ := a.layout.MainView().GetFrontPage()
			if currentPage == "addhost" {
				return event // Let form handle ESC
			}
			a.goToMainMenu()
			return nil

		case tcell.KeyTab, tcell.KeyBacktab:
			// Only handle Tab on welcome screen
			// Let modules handle their own Tab navigation
			if a.currentModule == "welcome" {
				a.layout.NextFocus()
				return nil
			}
			// Let the focused component handle Tab
			return event
		}

		switch event.Rune() {
		case 'q':
			a.tviewApp.Stop()
			return nil

		case ':':
			a.enterCommandMode()
			return nil

		case '/':
			a.enterSearchMode()
			return nil

		case '?':
			a.toggleHelp()
			return nil

		case 'r':
			a.refresh()
			return nil

		case 'c':
			// Skip global 'c' handler in SFTP module (uses 'c' for copy)
			if a.currentModule == "sftp" {
				return event
			}
			if a.selectedHost != nil {
				a.connectToHost(*a.selectedHost)
			}
			return nil

		// Module shortcuts (1-9, 0)
		case '1':
			a.switchModule("health")
			return nil
		case '2':
			a.switchModule("services")
			return nil
		case '3':
			a.switchModule("processes")
			return nil
		case '4':
			a.switchModule("logs")
			return nil
		case '5':
			a.switchModule("disks")
			return nil
		case '6':
			a.switchModule("batch")
			return nil
		case '7':
			a.switchModule("users")
			return nil
		case '8':
			a.switchModule("docker")
			return nil
		case '9':
			a.switchModule("installer")
			return nil
		case '0':
			a.goToMainMenu()
			return nil
		case 'F':
			a.switchModule("sftp")
			return nil
		case 't':
			a.launchTerminal()
			return nil
		}

		return event
	})
}

// goToMainMenu returns to the welcome/main menu screen
func (a *App) goToMainMenu() {
	a.currentModule = "welcome"
	a.layout.MainView().SwitchToPage("welcome")
	a.layout.StatusBar().SetMode("NORMAL")
	a.layout.StatusBar().ClearAlert()
	a.layout.StatusBar().SetStatus("Main Menu | 1-9=Modules F=SFTP ?=Help")
	a.tviewApp.SetFocus(a.layout.ServerList().View())
}

// enterCommandMode activates command input
func (a *App) enterCommandMode() {
	a.layout.StatusBar().SetMode("COMMAND")
	a.commandPalette.Activate()

	// Create a temporary layout with command input
	cmdLayout := tview.NewFlex().SetDirection(tview.FlexRow)
	cmdLayout.AddItem(a.layout.Root(), 0, 1, false)

	commandBar := tview.NewFlex()
	commandBar.AddItem(a.commandPalette.View(), 0, 1, true)
	commandBar.SetBackgroundColor(a.theme.Background)

	cmdLayout.AddItem(commandBar, 1, 0, true)

	a.tviewApp.SetRoot(cmdLayout, true)
	a.tviewApp.SetFocus(a.commandPalette.View())
}

// enterSearchMode activates search input
func (a *App) enterSearchMode() {
	a.layout.StatusBar().SetMode("SEARCH")
	a.searchPalette.Activate()

	// Create a temporary layout with search input
	searchLayout := tview.NewFlex().SetDirection(tview.FlexRow)
	searchLayout.AddItem(a.layout.Root(), 0, 1, false)

	searchBar := tview.NewFlex()
	searchBar.AddItem(a.searchPalette.View(), 0, 1, true)
	searchBar.SetBackgroundColor(a.theme.Background)

	searchLayout.AddItem(searchBar, 1, 0, true)

	a.tviewApp.SetRoot(searchLayout, true)
	a.tviewApp.SetFocus(a.searchPalette.View())
}

// toggleHelp shows/hides help overlay
func (a *App) toggleHelp() {
	if a.helpVisible {
		a.helpOverlay.Hide(a.layout.MainView())
		a.helpVisible = false
	} else {
		a.helpOverlay.SetContext(a.currentModule)
		a.helpOverlay.Show(a.layout.MainView(), a.tviewApp)
		a.helpVisible = true
	}
}

// executeCommand handles command execution
func (a *App) executeCommand(cmd ui.Command) {
	// Restore main layout
	a.tviewApp.SetRoot(a.layout.Root(), true)

	switch cmd.Name {
	case "q", "quit":
		a.tviewApp.Stop()

	case "home", "menu":
		a.goToMainMenu()

	case "theme":
		if len(cmd.Args) > 0 {
			a.setTheme(cmd.Args[0])
		} else {
			a.layout.StatusBar().SetAlert("Usage: :theme <name>", true)
		}

	case "connect":
		if len(cmd.Args) > 0 {
			a.connectByName(cmd.Args[0])
		} else if a.selectedHost != nil {
			a.connectToHost(*a.selectedHost)
		}

	case "disconnect":
		a.disconnect()

	case "health":
		a.switchModule("health")
	case "services":
		a.switchModule("services")
	case "processes", "ps":
		a.switchModule("processes")
	case "logs":
		a.switchModule("logs")
	case "disks", "df":
		a.switchModule("disks")
	case "batch":
		a.switchModule("batch")
	case "users":
		a.switchModule("users")
	case "docker":
		a.switchModule("docker")
	case "install":
		a.switchModule("installer")
	case "sftp", "files":
		a.switchModule("sftp")

	case "run", "exec":
		if len(cmd.Args) > 0 {
			command := strings.Join(cmd.Args, " ")
			a.runCommand(command)
		}

	case "refresh":
		a.refresh()

	case "help":
		a.toggleHelp()

	default:
		a.layout.StatusBar().SetAlert(fmt.Sprintf("Unknown command: %s", cmd.Name), true)
	}
}

// isConnected checks if a host is connected
func (a *App) isConnected() bool {
	return a.selectedHost != nil && a.sshClient.IsConnected(*a.selectedHost)
}

// setTheme changes the color scheme
func (a *App) setTheme(name string) {
	theme := config.GetTheme(name)
	a.theme = theme
	a.config.Theme = name

	// Rebuild UI with new theme
	a.layout = ui.NewLayout(a.tviewApp, a.theme)
	a.commandPalette = ui.NewCommandPalette(a.theme)
	a.searchPalette = ui.NewSearchPalette(a.theme)
	a.helpOverlay = ui.NewHelpOverlay(a.theme)

	// Reinitialize modules with new theme
	a.healthModule = health.NewDashboard(a.theme, a.sshClient)
	a.servicesModule = services.NewManager(a.theme, a.sshClient)
	a.processesModule = processes.NewManager(a.theme, a.sshClient)
	a.logsModule = logs.NewViewer(a.theme, a.sshClient)
	a.diskModule = disk.NewManager(a.theme, a.sshClient)
	a.batchModule = batch.NewExecutor(a.theme, a.sshClient, a.executor)
	a.usersModule = users.NewManager(a.theme, a.sshClient)
	a.dockerModule = docker.NewManager(a.theme, a.sshClient)
	a.installerModule = installer.NewManager(a.theme, a.sshClient)
	a.sftpModule = sftp.NewManager(a.theme, a.sshClient)

	a.batchModule.SetApp(a.tviewApp)
	a.dockerModule.SetApp(a.tviewApp)
	a.servicesModule.SetApp(a.tviewApp)
	a.sftpModule.SetApp(a.tviewApp)
	a.processesModule.SetApp(a.tviewApp)
	a.logsModule.SetApp(a.tviewApp)
	a.diskModule.SetApp(a.tviewApp)
	a.usersModule.SetApp(a.tviewApp)
	a.installerModule.SetApp(a.tviewApp)

	// Re-add module pages
	a.layout.MainView().AddPage("health", a.healthModule.View(), true, false)
	a.layout.MainView().AddPage("services", a.servicesModule.View(), true, false)
	a.layout.MainView().AddPage("processes", a.processesModule.View(), true, false)
	a.layout.MainView().AddPage("logs", a.logsModule.View(), true, false)
	a.layout.MainView().AddPage("disks", a.diskModule.View(), true, false)
	a.layout.MainView().AddPage("batch", a.batchModule.View(), true, false)
	a.layout.MainView().AddPage("users", a.usersModule.View(), true, false)
	a.layout.MainView().AddPage("docker", a.dockerModule.View(), true, false)
	a.layout.MainView().AddPage("installer", a.installerModule.View(), true, false)
	a.layout.MainView().AddPage("sftp", a.sftpModule.View(), true, false)

	a.layout.ServerList().SetServers(a.hosts)
	a.batchModule.SetHosts(a.hosts)
	a.sftpModule.SetHosts(a.hosts)
	a.setupCallbacks()

	// Restart the status bar clock
	a.layout.StatusBar().StartClock(a.tviewApp)

	a.tviewApp.SetRoot(a.layout.Root(), true)
	a.layout.StatusBar().SetSuccess(fmt.Sprintf("Theme: %s", name))
}

// connectToHost establishes connection to a host
func (a *App) connectToHost(host ssh.HostEntry) {
	name := host.Name
	if name == "" {
		name = host.Hostname
	}

	a.layout.ServerList().SetStatusByHost(host, ui.StatusConnecting, "Connecting...")
	a.layout.StatusBar().SetStatus(fmt.Sprintf("Connecting to %s...", name))

	go func() {
		_, err := a.sshClient.Connect(host)
		a.tviewApp.QueueUpdateDraw(func() {
			if err != nil {
				a.layout.ServerList().SetStatusByHost(host, ui.StatusError, err.Error())
				a.layout.StatusBar().SetAlert(fmt.Sprintf("Connection failed: %v", err), true)
			} else {
				a.layout.ServerList().SetStatusByHost(host, ui.StatusConnected, "Connected")
				a.layout.StatusBar().SetSuccess(fmt.Sprintf("Connected to %s", name))
				a.selectedHost = &host

				// Update modules with new host
				a.healthModule.SetHost(&host)
				a.servicesModule.SetHost(&host)
				a.processesModule.SetHost(&host)
				a.logsModule.SetHost(&host)
				a.diskModule.SetHost(&host)
				a.usersModule.SetHost(&host)
				a.dockerModule.SetHost(&host)
				a.installerModule.SetHost(&host)
				a.sftpModule.SetSourceHost(&host)
			}
		})
	}()
}

// connectByName connects to a host by name
func (a *App) connectByName(name string) {
	for _, host := range a.hosts {
		if host.Name == name || host.Hostname == name {
			a.connectToHost(host)
			return
		}
	}
	a.layout.StatusBar().SetAlert(fmt.Sprintf("Host not found: %s", name), true)
}

// disconnect closes current connection
func (a *App) disconnect() {
	if a.selectedHost != nil {
		a.sshClient.Disconnect(*a.selectedHost)
		a.layout.ServerList().SetStatusByHost(*a.selectedHost, ui.StatusDisconnected, "")
		a.layout.StatusBar().SetStatus("Disconnected")
		a.selectedHost = nil
		a.goToMainMenu()
	}
}

// refresh reloads current view
func (a *App) refresh() {
	a.layout.StatusBar().SetStatus("Refreshing...")

	// Check connection for modules that require it
	if a.currentModule != "welcome" && a.currentModule != "batch" && a.currentModule != "installer" && !a.isConnected() {
		a.layout.StatusBar().SetAlert("Not connected to any server", true)
		return
	}

	switch a.currentModule {
	case "health":
		go a.refreshModule(a.healthModule.Refresh, "Health metrics")
	case "services":
		go a.refreshModule(a.servicesModule.Refresh, "Services")
	case "processes":
		go a.refreshModule(a.processesModule.Refresh, "Processes")
	case "logs":
		go a.refreshModule(a.logsModule.Refresh, "Logs")
	case "disks":
		go a.refreshModule(a.diskModule.Refresh, "Disk info")
	case "users":
		go a.refreshModule(a.usersModule.Refresh, "Users")
	case "docker":
		go a.refreshModule(a.dockerModule.Refresh, "Docker")
	case "sftp":
		go a.refreshModule(a.sftpModule.Refresh, "Files")
	default:
		a.discoverHosts()
	}
}

// refreshModule is a helper to refresh a module
func (a *App) refreshModule(refreshFn func() error, name string) {
	err := refreshFn()
	a.tviewApp.QueueUpdateDraw(func() {
		if err != nil {
			a.layout.StatusBar().SetAlert(err.Error(), true)
		} else {
			a.layout.StatusBar().SetSuccess(fmt.Sprintf("%s refreshed", name))
		}
	})
}

// filterServers filters the server list
func (a *App) filterServers(query string) {
	if query == "" {
		a.layout.ServerList().SetServers(a.hosts)
		return
	}

	query = strings.ToLower(query)
	var filtered []ssh.HostEntry
	for _, host := range a.hosts {
		if strings.Contains(strings.ToLower(host.Name), query) ||
			strings.Contains(strings.ToLower(host.Hostname), query) {
			filtered = append(filtered, host)
		}
	}
	a.layout.ServerList().SetServers(filtered)
}

// switchModule changes the active module view
func (a *App) switchModule(module string) {
	// Batch, installer, and sftp don't require connection for viewing
	noConnectionRequired := module == "batch" || module == "installer" || module == "sftp"

	if !noConnectionRequired && !a.isConnected() {
		a.layout.StatusBar().SetAlert("⚠ Connect to a server first (press Enter on a host)", true)
		return
	}

	a.currentModule = module
	a.helpOverlay.SetContext(module)
	a.layout.StatusBar().SetStatus(fmt.Sprintf("Loading %s...", module))

	// For modules not requiring connection, just switch (but still refresh if connected)
	if noConnectionRequired {
		a.layout.MainView().SwitchToPage(module)
		a.layout.StatusBar().SetStatus(fmt.Sprintf("Module: %s | Tab=Switch ESC=Home", module))
		// Set focus to the module's main view
		a.setModuleFocus(module)

		// For installer, try to refresh if connected to detect package manager
		if module == "installer" && a.isConnected() {
			go func() {
				if a.installerModule != nil {
					a.installerModule.Refresh()
					if a.tviewApp != nil {
						a.tviewApp.QueueUpdateDraw(func() {})
					}
				}
			}()
		}
		return
	}

	// Refresh the module data
	go func() {
		var err error
		switch module {
		case "health":
			err = a.healthModule.Refresh()
		case "services":
			err = a.servicesModule.Refresh()
		case "processes":
			err = a.processesModule.Refresh()
		case "logs":
			err = a.logsModule.Refresh()
		case "disks":
			err = a.diskModule.Refresh()
		case "users":
			err = a.usersModule.Refresh()
		case "docker":
			err = a.dockerModule.Refresh()
		case "installer":
			err = a.installerModule.Refresh()
		}

		a.tviewApp.QueueUpdateDraw(func() {
			if err != nil {
				a.layout.StatusBar().SetAlert(err.Error(), true)
			} else {
				a.layout.MainView().SwitchToPage(module)
				a.layout.StatusBar().SetStatus(fmt.Sprintf("Module: %s | Tab=Switch ESC=Home", module))
				// Set focus to the module's main view
				a.setModuleFocus(module)
			}
		})
	}()
}

// setModuleFocus sets focus to the module's primary interactive component
func (a *App) setModuleFocus(module string) {
	switch module {
	case "health":
		a.tviewApp.SetFocus(a.healthModule.View()) // Health has no Focus method
	case "services":
		a.servicesModule.Focus()
	case "processes":
		a.processesModule.Focus()
	case "logs":
		a.logsModule.Focus()
	case "disks":
		a.diskModule.Focus()
	case "users":
		a.usersModule.Focus()
	case "docker":
		a.dockerModule.Focus()
	case "installer":
		a.installerModule.Focus()
	case "sftp":
		a.sftpModule.Focus()
	case "batch":
		a.batchModule.Focus()
	}
}

// runCommand executes a command on selected server(s)
func (a *App) runCommand(command string) {
	if !a.isConnected() {
		a.layout.StatusBar().SetAlert("No server connected", true)
		return
	}

	a.layout.StatusBar().SetStatus(fmt.Sprintf("Running: %s", command))

	go func() {
		output, err := a.sshClient.RunCommand(*a.selectedHost, command)
		a.tviewApp.QueueUpdateDraw(func() {
			if err != nil {
				a.layout.StatusBar().SetAlert(fmt.Sprintf("Command failed: %v", err), true)
			} else {
				a.layout.StatusBar().SetSuccess("Command completed")
				// Show output in a text view
				outputView := tview.NewTextView()
				outputView.SetText(output)
				outputView.SetDynamicColors(true)
				outputView.SetScrollable(true)
				outputView.SetBorder(true)
				outputView.SetTitle(" Command Output (Esc to close) ")
				outputView.SetTitleColor(a.theme.Primary)
				outputView.SetBorderColor(a.theme.Border)
				outputView.SetBackgroundColor(a.theme.Background)

				outputView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
					if event.Key() == tcell.KeyEsc {
						a.layout.MainView().RemovePage("output")
						a.goToMainMenu()
						return nil
					}
					return event
				})

				a.layout.MainView().AddPage("output", outputView, true, true)
			}
		})
	}()
}

// launchTerminal opens interactive SSH terminal
func (a *App) launchTerminal() {
	if !a.isConnected() {
		a.layout.StatusBar().SetAlert("⚠ Connect to a server first", true)
		return
	}

	host := a.selectedHost
	name := host.Name
	if name == "" {
		name = host.Hostname
	}

	a.layout.StatusBar().SetStatus(fmt.Sprintf("Launching terminal to %s... (Ctrl+D to exit)", name))

	// Stop tview temporarily
	a.tviewApp.Suspend(func() {
		// Build SSH command
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

		cmd.Run()
	})

	a.layout.StatusBar().SetSuccess("Terminal session ended")
}

// Stop gracefully shuts down the application
func (a *App) Stop() {
	a.sshClient.DisconnectAll()
	a.tviewApp.Stop()
}
