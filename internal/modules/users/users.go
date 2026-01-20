package users

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/systask/systask/internal/config"
	"github.com/systask/systask/internal/ssh"
)

// User represents a system user
type User struct {
	Name   string
	UID    string
	GID    string
	Home   string
	Shell  string
	Groups []string
	IsSudo bool
	Locked bool
}

// Group represents a system group
type Group struct {
	Name    string
	GID     string
	Members []string
}

// Manager manages system users and groups
type Manager struct {
	view       *tview.Flex
	tabs       *tview.TextView
	userTable  *tview.Table
	groupTable *tview.Table
	detailView *tview.TextView
	actionView *tview.TextView
	pages      *tview.Pages
	theme      *config.Theme
	sshClient  *ssh.Client
	host       *ssh.HostEntry

	users       []User
	groups      []Group
	currentTab  string
	selectedIdx int
}

// NewManager creates a new user manager
func NewManager(theme *config.Theme, client *ssh.Client) *Manager {
	m := &Manager{
		theme:      theme,
		sshClient:  client,
		currentTab: "users",
	}
	m.build()
	return m
}

// build constructs the user manager view
func (m *Manager) build() {
	// Tab bar
	m.tabs = tview.NewTextView()
	m.tabs.SetDynamicColors(true)
	m.tabs.SetBackgroundColor(m.theme.Muted)
	m.tabs.SetTextAlign(tview.AlignCenter)
	m.updateTabs()

	// User table
	m.userTable = tview.NewTable()
	m.userTable.SetBorders(false)
	m.userTable.SetSelectable(true, false)
	m.userTable.SetBackgroundColor(m.theme.Background)
	m.userTable.SetSelectedStyle(tcell.StyleDefault.
		Background(m.theme.Muted).
		Foreground(m.theme.Primary))

	userHeaders := []string{"", "USER", "UID", "HOME", "SHELL", "SUDO", "STATUS"}
	for i, h := range userHeaders {
		cell := tview.NewTableCell(h).
			SetTextColor(m.theme.Secondary).
			SetSelectable(false).
			SetExpansion(1)
		if i == 0 || i == 2 || i == 5 || i == 6 {
			cell.SetExpansion(0)
		}
		m.userTable.SetCell(0, i, cell)
	}
	m.userTable.SetFixed(1, 0)
	m.userTable.SetInputCapture(m.handleUserInput)

	m.userTable.SetSelectionChangedFunc(func(row, col int) {
		if row > 0 && row <= len(m.users) {
			m.selectedIdx = row - 1
			m.showUserDetails(m.users[row-1])
		}
	})

	// Group table
	m.groupTable = tview.NewTable()
	m.groupTable.SetBorders(false)
	m.groupTable.SetSelectable(true, false)
	m.groupTable.SetBackgroundColor(m.theme.Background)
	m.groupTable.SetSelectedStyle(tcell.StyleDefault.
		Background(m.theme.Muted).
		Foreground(m.theme.Primary))

	groupHeaders := []string{"GROUP", "GID", "MEMBERS"}
	for i, h := range groupHeaders {
		cell := tview.NewTableCell(h).
			SetTextColor(m.theme.Secondary).
			SetSelectable(false).
			SetExpansion(1)
		if i == 1 {
			cell.SetExpansion(0)
		}
		m.groupTable.SetCell(0, i, cell)
	}
	m.groupTable.SetFixed(1, 0)
	m.groupTable.SetInputCapture(m.handleGroupInput)

	// Pages
	m.pages = tview.NewPages()
	m.pages.AddPage("users", m.userTable, true, true)
	m.pages.AddPage("groups", m.groupTable, true, false)

	// Content box
	contentBox := tview.NewFlex().SetDirection(tview.FlexRow)
	contentBox.AddItem(m.pages, 0, 1, true)
	contentBox.SetBorder(true)
	contentBox.SetBorderColor(m.theme.Border)
	contentBox.SetTitle(" 👤 Users & Groups ")
	contentBox.SetTitleColor(m.theme.Primary)
	contentBox.SetBackgroundColor(m.theme.Background)

	// Detail view
	m.detailView = tview.NewTextView()
	m.detailView.SetDynamicColors(true)
	m.detailView.SetScrollable(true)
	m.detailView.SetBorder(true)
	m.detailView.SetBorderColor(m.theme.Border)
	m.detailView.SetTitle(" Details ")
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
	mainContent.AddItem(contentBox, 0, 2, true)
	mainContent.AddItem(rightPanel, 0, 1, false)

	// Layout
	m.view = tview.NewFlex().SetDirection(tview.FlexRow)
	m.view.AddItem(m.tabs, 1, 0, false)
	m.view.AddItem(mainContent, 0, 1, true)
	m.view.SetBackgroundColor(m.theme.Background)
}

// updateTabs updates the tab display
func (m *Manager) updateTabs() {
	userStyle := "[white]"
	groupStyle := "[white]"

	if m.currentTab == "users" {
		userStyle = fmt.Sprintf("[%s::b]", colorToTag(m.theme.Primary))
	} else {
		groupStyle = fmt.Sprintf("[%s::b]", colorToTag(m.theme.Primary))
	}

	m.tabs.SetText(fmt.Sprintf("  %s👤 Users (1)[white]  │  %s👥 Groups (2)[white]  ",
		userStyle, groupStyle))
}

// getActionsText returns the actions help text
func (m *Manager) getActionsText() string {
	highlight := colorToTag(m.theme.Highlight)
	return fmt.Sprintf(
		"  [%s]a[white]=Add  [%s]d[white]=Delete  [%s]p[white]=Password  [%s]s[white]=Toggle Sudo  [%s]l[white]=Lock/Unlock  [%s]g[white]=Groups  [%s]Tab[white]=Switch",
		highlight, highlight, highlight, highlight, highlight, highlight, highlight)
}

// handleUserInput handles user table input
func (m *Manager) handleUserInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyTab:
		m.currentTab = "groups"
		m.updateTabs()
		m.pages.SwitchToPage("groups")
		return nil
	}

	if len(m.users) == 0 {
		return event
	}

	row, _ := m.userTable.GetSelection()
	if row <= 0 || row > len(m.users) {
		return event
	}
	user := m.users[row-1]

	switch event.Rune() {
	case 'a': // Add user
		m.showAddUser()
		return nil
	case 'd': // Delete user
		m.deleteUser(user.Name)
		return nil
	case 'p': // Change password
		m.showChangePassword(user.Name)
		return nil
	case 's': // Toggle sudo
		m.toggleSudo(user.Name, !user.IsSudo)
		return nil
	case 'l': // Lock/unlock
		m.toggleLock(user.Name, !user.Locked)
		return nil
	case 'g': // Manage groups
		m.showUserGroups(user.Name)
		return nil
	case 'j':
		return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
	case 'k':
		return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
	case '1':
		m.currentTab = "users"
		m.updateTabs()
		m.pages.SwitchToPage("users")
		return nil
	case '2':
		m.currentTab = "groups"
		m.updateTabs()
		m.pages.SwitchToPage("groups")
		return nil
	}

	return event
}

// handleGroupInput handles group table input
func (m *Manager) handleGroupInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyTab:
		m.currentTab = "users"
		m.updateTabs()
		m.pages.SwitchToPage("users")
		return nil
	}

	if len(m.groups) == 0 {
		return event
	}

	row, _ := m.groupTable.GetSelection()
	if row <= 0 || row > len(m.groups) {
		return event
	}
	group := m.groups[row-1]

	switch event.Rune() {
	case 'a': // Add group
		m.showAddGroup()
		return nil
	case 'd': // Delete group
		m.deleteGroup(group.Name)
		return nil
	case 'm': // Manage members
		m.showGroupMembers(group.Name)
		return nil
	case 'j':
		return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
	case 'k':
		return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
	case '1':
		m.currentTab = "users"
		m.updateTabs()
		m.pages.SwitchToPage("users")
		return nil
	case '2':
		m.currentTab = "groups"
		m.updateTabs()
		m.pages.SwitchToPage("groups")
		return nil
	}

	return event
}

// SetHost sets the current host
func (m *Manager) SetHost(host *ssh.HostEntry) {
	m.host = host
}

// Refresh updates the user and group lists
func (m *Manager) Refresh() error {
	if m.host == nil {
		return fmt.Errorf("no host set")
	}

	// Get users from /etc/passwd (filter system users)
	userOutput, err := m.sshClient.RunCommand(*m.host,
		`awk -F: '$3 >= 1000 || $1 == "root" {print $1":"$3":"$4":"$6":"$7}' /etc/passwd`)
	if err != nil {
		return err
	}

	// Get sudoers
	sudoers, _ := m.sshClient.RunCommand(*m.host,
		`getent group sudo wheel 2>/dev/null | cut -d: -f4 | tr ',' '\n'`)
	sudoersList := strings.Split(strings.TrimSpace(sudoers), "\n")
	sudoersMap := make(map[string]bool)
	for _, s := range sudoersList {
		sudoersMap[strings.TrimSpace(s)] = true
	}

	// Get locked users
	lockedOutput, _ := m.sshClient.RunCommand(*m.host,
		`cat /etc/shadow 2>/dev/null | awk -F: '$2 ~ /^!/ || $2 ~ /^\*/ {print $1}'`)
	lockedList := strings.Split(strings.TrimSpace(lockedOutput), "\n")
	lockedMap := make(map[string]bool)
	for _, l := range lockedList {
		lockedMap[strings.TrimSpace(l)] = true
	}

	// Parse users
	m.users = m.parseUsers(userOutput, sudoersMap, lockedMap)
	m.updateUserTable()

	// Get groups
	groupOutput, _ := m.sshClient.RunCommand(*m.host,
		`awk -F: '$3 >= 1000 || $1 == "root" || $1 == "sudo" || $1 == "wheel" || $1 == "docker" {print $1":"$3":"$4}' /etc/group`)
	m.groups = m.parseGroups(groupOutput)
	m.updateGroupTable()

	return nil
}

// parseUsers parses passwd output
func (m *Manager) parseUsers(output string, sudoers, locked map[string]bool) []User {
	var users []User
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Split(line, ":")
		if len(fields) < 5 {
			continue
		}

		user := User{
			Name:   fields[0],
			UID:    fields[1],
			GID:    fields[2],
			Home:   fields[3],
			Shell:  fields[4],
			IsSudo: sudoers[fields[0]],
			Locked: locked[fields[0]],
		}

		users = append(users, user)
	}

	// Sort: root first, then by name
	sort.Slice(users, func(i, j int) bool {
		if users[i].Name == "root" {
			return true
		}
		if users[j].Name == "root" {
			return false
		}
		return users[i].Name < users[j].Name
	})

	return users
}

// parseGroups parses group output
func (m *Manager) parseGroups(output string) []Group {
	var groups []Group
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Split(line, ":")
		if len(fields) < 3 {
			continue
		}

		members := []string{}
		if len(fields) > 2 && fields[2] != "" {
			members = strings.Split(fields[2], ",")
		}

		group := Group{
			Name:    fields[0],
			GID:     fields[1],
			Members: members,
		}

		groups = append(groups, group)
	}

	// Sort by name
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Name < groups[j].Name
	})

	return groups
}

// updateUserTable updates the user table display
func (m *Manager) updateUserTable() {
	for row := m.userTable.GetRowCount() - 1; row > 0; row-- {
		m.userTable.RemoveRow(row)
	}

	for i, user := range m.users {
		row := i + 1

		// Icon
		icon := "👤"
		if user.Name == "root" {
			icon = "👑"
		}
		m.userTable.SetCell(row, 0, tview.NewTableCell(icon).
			SetExpansion(0))

		// Name
		m.userTable.SetCell(row, 1, tview.NewTableCell(user.Name).
			SetTextColor(m.theme.Foreground).
			SetExpansion(1))

		// UID
		m.userTable.SetCell(row, 2, tview.NewTableCell(user.UID).
			SetTextColor(m.theme.Muted).
			SetExpansion(0))

		// Home
		m.userTable.SetCell(row, 3, tview.NewTableCell(truncate(user.Home, 20)).
			SetTextColor(m.theme.Foreground).
			SetExpansion(1))

		// Shell
		shell := user.Shell
		if idx := strings.LastIndex(shell, "/"); idx >= 0 {
			shell = shell[idx+1:]
		}
		m.userTable.SetCell(row, 4, tview.NewTableCell(shell).
			SetTextColor(m.theme.Muted).
			SetExpansion(1))

		// Sudo
		sudoText := ""
		sudoColor := m.theme.Muted
		if user.IsSudo {
			sudoText = "✓"
			sudoColor = m.theme.Success
		}
		m.userTable.SetCell(row, 5, tview.NewTableCell(sudoText).
			SetTextColor(sudoColor).
			SetExpansion(0))

		// Status
		status := "Active"
		statusColor := m.theme.Success
		if user.Locked {
			status = "Locked"
			statusColor = m.theme.Error
		}
		m.userTable.SetCell(row, 6, tview.NewTableCell(status).
			SetTextColor(statusColor).
			SetExpansion(0))
	}

	if m.userTable.GetRowCount() > 1 {
		m.userTable.Select(1, 0)
	}
}

// updateGroupTable updates the group table display
func (m *Manager) updateGroupTable() {
	for row := m.groupTable.GetRowCount() - 1; row > 0; row-- {
		m.groupTable.RemoveRow(row)
	}

	for i, group := range m.groups {
		row := i + 1

		// Name
		m.groupTable.SetCell(row, 0, tview.NewTableCell(group.Name).
			SetTextColor(m.theme.Foreground).
			SetExpansion(1))

		// GID
		m.groupTable.SetCell(row, 1, tview.NewTableCell(group.GID).
			SetTextColor(m.theme.Muted).
			SetExpansion(0))

		// Members
		members := strings.Join(group.Members, ", ")
		m.groupTable.SetCell(row, 2, tview.NewTableCell(truncate(members, 40)).
			SetTextColor(m.theme.Secondary).
			SetExpansion(1))
	}

	if m.groupTable.GetRowCount() > 1 {
		m.groupTable.Select(1, 0)
	}
}

// showUserDetails shows details for a user
func (m *Manager) showUserDetails(user User) {
	primary := colorToTag(m.theme.Primary)
	muted := colorToTag(m.theme.Muted)

	// Get user groups
	groupsOutput, _ := m.sshClient.RunCommand(*m.host,
		fmt.Sprintf("groups %s 2>/dev/null | cut -d: -f2", user.Name))

	// Get last login
	lastLogin, _ := m.sshClient.RunCommand(*m.host,
		fmt.Sprintf("lastlog -u %s 2>/dev/null | tail -1 | awk '{print $4,$5,$6,$7,$9}'", user.Name))

	status := "Active"
	if user.Locked {
		status = "Locked (Suspended)"
	}

	sudo := "No"
	if user.IsSudo {
		sudo = "Yes (Admin)"
	}

	details := fmt.Sprintf(
		"[%s]User: %s[white]\n\n"+
			"[%s]UID:[white] %s\n"+
			"[%s]GID:[white] %s\n"+
			"[%s]Home:[white] %s\n"+
			"[%s]Shell:[white] %s\n"+
			"[%s]Status:[white] %s\n"+
			"[%s]Sudo:[white] %s\n\n"+
			"[%s]Groups:[white]\n%s\n\n"+
			"[%s]Last Login:[white]\n%s",
		primary, user.Name,
		primary, user.UID,
		primary, user.GID,
		primary, user.Home,
		primary, user.Shell,
		primary, status,
		primary, sudo,
		primary, strings.TrimSpace(groupsOutput),
		muted, strings.TrimSpace(lastLogin),
	)

	m.detailView.SetTitle(fmt.Sprintf(" User: %s ", user.Name))
	m.detailView.SetText(details)
}

// User action methods
func (m *Manager) showAddUser() {
	m.detailView.SetTitle(" Add User ")
	m.detailView.SetText(fmt.Sprintf(
		"[%s]Add New User[white]\n\n"+
			"Use command:\n\n"+
			"  :run sudo useradd -m -s /bin/bash <username>\n\n"+
			"With password:\n\n"+
			"  :run sudo useradd -m -s /bin/bash <username> && sudo passwd <username>",
		colorToTag(m.theme.Primary)))
}

func (m *Manager) deleteUser(username string) {
	if username == "root" {
		m.detailView.SetText(fmt.Sprintf("[%s]Cannot delete root user![white]",
			colorToTag(m.theme.Error)))
		return
	}

	m.detailView.SetTitle(fmt.Sprintf(" Delete User: %s ", username))
	m.detailView.SetText(fmt.Sprintf(
		"[%s]Deleting user: %s[white]\n\n"+
			"This will remove the user and their home directory.",
		colorToTag(m.theme.Warning), username))

	// Execute delete
	cmd := fmt.Sprintf("sudo userdel -r %s 2>&1", username)
	output, err := m.sshClient.RunCommand(*m.host, cmd)
	if err != nil {
		m.detailView.SetText(fmt.Sprintf("[%s]Error: %v[white]\n%s",
			colorToTag(m.theme.Error), err, output))
	} else {
		m.detailView.SetText(fmt.Sprintf("[%s]✓ User %s deleted[white]",
			colorToTag(m.theme.Success), username))
		m.Refresh()
	}
}

func (m *Manager) showChangePassword(username string) {
	m.detailView.SetTitle(fmt.Sprintf(" Change Password: %s ", username))
	m.detailView.SetText(fmt.Sprintf(
		"[%s]Change Password[white]\n\n"+
			"Interactive password change requires terminal.\n\n"+
			"SSH to the server and run:\n"+
			"  sudo passwd %s",
		colorToTag(m.theme.Primary), username))
}

func (m *Manager) toggleSudo(username string, enable bool) {
	var cmd string
	var action string

	if enable {
		action = "granted"
		cmd = fmt.Sprintf("sudo usermod -aG sudo %s 2>/dev/null || sudo usermod -aG wheel %s 2>&1", username, username)
	} else {
		action = "revoked"
		cmd = fmt.Sprintf("sudo gpasswd -d %s sudo 2>/dev/null; sudo gpasswd -d %s wheel 2>&1", username, username)
	}

	_, err := m.sshClient.RunCommand(*m.host, cmd)
	if err != nil {
		m.detailView.SetText(fmt.Sprintf("[%s]Error: %v[white]",
			colorToTag(m.theme.Error), err))
	} else {
		m.detailView.SetText(fmt.Sprintf("[%s]✓ Sudo access %s for %s[white]",
			colorToTag(m.theme.Success), action, username))
		m.Refresh()
	}
}

func (m *Manager) toggleLock(username string, lock bool) {
	var cmd string
	var action string

	if lock {
		action = "locked (suspended)"
		cmd = fmt.Sprintf("sudo usermod -L %s", username)
	} else {
		action = "unlocked"
		cmd = fmt.Sprintf("sudo usermod -U %s", username)
	}

	_, err := m.sshClient.RunCommand(*m.host, cmd)
	if err != nil {
		m.detailView.SetText(fmt.Sprintf("[%s]Error: %v[white]",
			colorToTag(m.theme.Error), err))
	} else {
		m.detailView.SetText(fmt.Sprintf("[%s]✓ User %s %s[white]",
			colorToTag(m.theme.Success), username, action))
		m.Refresh()
	}
}

func (m *Manager) showUserGroups(username string) {
	output, _ := m.sshClient.RunCommand(*m.host,
		fmt.Sprintf("groups %s && echo '---' && cat /etc/group | cut -d: -f1 | sort", username))
	m.detailView.SetTitle(fmt.Sprintf(" Groups for %s ", username))
	m.detailView.SetText(fmt.Sprintf(
		"[%s]Current groups for %s:[white]\n%s\n\n"+
			"[%s]Add to group:[white]\n  :run sudo usermod -aG <group> %s\n\n"+
			"[%s]Remove from group:[white]\n  :run sudo gpasswd -d %s <group>",
		colorToTag(m.theme.Primary), username, output,
		colorToTag(m.theme.Secondary), username,
		colorToTag(m.theme.Secondary), username))
}

// Group action methods
func (m *Manager) showAddGroup() {
	m.detailView.SetTitle(" Add Group ")
	m.detailView.SetText(fmt.Sprintf(
		"[%s]Add New Group[white]\n\n"+
			"Use command:\n\n"+
			"  :run sudo groupadd <groupname>",
		colorToTag(m.theme.Primary)))
}

func (m *Manager) deleteGroup(groupname string) {
	protected := map[string]bool{"root": true, "sudo": true, "wheel": true, "docker": true}
	if protected[groupname] {
		m.detailView.SetText(fmt.Sprintf("[%s]Cannot delete protected group: %s[white]",
			colorToTag(m.theme.Error), groupname))
		return
	}

	cmd := fmt.Sprintf("sudo groupdel %s 2>&1", groupname)
	output, err := m.sshClient.RunCommand(*m.host, cmd)
	if err != nil {
		m.detailView.SetText(fmt.Sprintf("[%s]Error: %v[white]\n%s",
			colorToTag(m.theme.Error), err, output))
	} else {
		m.detailView.SetText(fmt.Sprintf("[%s]✓ Group %s deleted[white]",
			colorToTag(m.theme.Success), groupname))
		m.Refresh()
	}
}

func (m *Manager) showGroupMembers(groupname string) {
	output, _ := m.sshClient.RunCommand(*m.host,
		fmt.Sprintf("getent group %s", groupname))
	m.detailView.SetTitle(fmt.Sprintf(" Group: %s ", groupname))
	m.detailView.SetText(fmt.Sprintf(
		"[%s]Group: %s[white]\n\n%s\n\n"+
			"[%s]Add user to group:[white]\n  :run sudo usermod -aG %s <username>\n\n"+
			"[%s]Remove user from group:[white]\n  :run sudo gpasswd -d <username> %s",
		colorToTag(m.theme.Primary), groupname, output,
		colorToTag(m.theme.Secondary), groupname,
		colorToTag(m.theme.Secondary), groupname))
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
