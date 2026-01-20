package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/systask/systask/internal/config"
	"github.com/systask/systask/internal/ssh"
)

// ServerStatus represents connection status
type ServerStatus int

const (
	StatusDisconnected ServerStatus = iota
	StatusConnecting
	StatusConnected
	StatusError
)

// ServerItem represents a server in the list
type ServerItem struct {
	Host   ssh.HostEntry
	Status ServerStatus
	Info   string
	Group  string
}

// GroupItem represents a group header in the list
type GroupItem struct {
	Name     string
	Icon     string
	Expanded bool
	Count    int
}

// ServerList manages the server sidebar with groups
type ServerList struct {
	list         *tview.List
	theme        *config.Theme
	servers      []ServerItem
	groupedItems []interface{} // Mix of GroupItem and ServerItem

	// Callbacks
	onSelect     func(ssh.HostEntry)
	onConnect    func(ssh.HostEntry)
	onAddHost    func()
	onDeleteHost func(ssh.HostEntry)
	onEditHost   func(ssh.HostEntry)
}

// NewServerList creates a new server list
func NewServerList(theme *config.Theme) *ServerList {
	sl := &ServerList{
		theme:   theme,
		servers: []ServerItem{},
	}
	sl.build()
	return sl
}

// build constructs the server list view
func (sl *ServerList) build() {
	sl.list = tview.NewList()
	sl.list.SetBackgroundColor(sl.theme.Background)
	sl.list.SetBorder(true)
	sl.list.SetBorderColor(sl.theme.Border)
	sl.list.SetTitle(" SERVERS ")
	sl.list.SetTitleColor(sl.theme.Primary)
	sl.list.ShowSecondaryText(true)
	sl.list.SetHighlightFullLine(true)
	sl.list.SetSelectedBackgroundColor(sl.theme.Muted)
	sl.list.SetSelectedTextColor(sl.theme.Primary)
	sl.list.SetMainTextColor(sl.theme.Foreground)
	sl.list.SetSecondaryTextColor(sl.theme.Muted)
	sl.list.SetShortcutColor(sl.theme.Secondary)

	// Selection handler
	sl.list.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		if index < len(sl.groupedItems) {
			switch item := sl.groupedItems[index].(type) {
			case *ServerItem:
				if sl.onConnect != nil {
					sl.onConnect(item.Host)
				}
			case string:
				if item == "add" && sl.onAddHost != nil {
					sl.onAddHost()
				}
			}
		}
	})

	sl.list.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		if index < len(sl.groupedItems) {
			if item, ok := sl.groupedItems[index].(*ServerItem); ok {
				if sl.onSelect != nil {
					sl.onSelect(item.Host)
				}
			}
		}
	})

	// Key handling
	sl.list.SetInputCapture(sl.HandleInput)
}

// SetServers updates the server list
func (sl *ServerList) SetServers(hosts []ssh.HostEntry) {
	sl.servers = make([]ServerItem, len(hosts))

	// Group hosts by source
	groups := map[string][]ssh.HostEntry{
		"SSH Config": {},
		"Cloud":      {},
		"Custom":     {},
	}

	for i, host := range hosts {
		sl.servers[i] = ServerItem{
			Host:   host,
			Status: StatusDisconnected,
		}

		// Determine group
		group := "SSH Config"
		source := strings.ToLower(host.Source)
		if source == "aws" || source == "azure" || source == "gcp" || source == "cloud" {
			group = "Cloud"
		} else if source == "config" || source == "manual" || source == "custom" {
			group = "Custom"
		}

		groups[group] = append(groups[group], host)
	}

	sl.refresh(groups)
}

// refresh rebuilds the list
func (sl *ServerList) refresh(groups map[string][]ssh.HostEntry) {
	sl.list.Clear()
	sl.groupedItems = nil

	// Add host button
	addText := fmt.Sprintf("[%s]+ Add Server[white]", colorToTag(sl.theme.Success))
	sl.list.AddItem(addText, "", 'a', nil)
	sl.groupedItems = append(sl.groupedItems, "add")

	// Group order
	groupOrder := []struct {
		name string
		icon string
	}{
		{"SSH Config", ""},
		{"Cloud", ""},
		{"Custom", ""},
	}

	for _, g := range groupOrder {
		hosts := groups[g.name]
		if len(hosts) == 0 {
			continue
		}

		// Sort hosts
		sort.Slice(hosts, func(i, j int) bool {
			return strings.ToLower(hosts[i].Name) < strings.ToLower(hosts[j].Name)
		})

		// Group header
		groupText := fmt.Sprintf("[%s]%s (%d)[white]",
			colorToTag(sl.theme.Secondary), g.name, len(hosts))
		sl.list.AddItem(groupText, "", 0, nil)
		sl.groupedItems = append(sl.groupedItems, &GroupItem{Name: g.name, Count: len(hosts)})

		// Servers in group
		for _, host := range hosts {
			h := host
			serverItem := &ServerItem{Host: h, Status: StatusDisconnected}

			name := h.Name
			if name == "" {
				name = h.Hostname
			}

			// Truncate long hostnames to fit in sidebar (max 18 chars for name)
			if len(name) > 18 {
				name = name[:15] + "..."
			}

			icon := sl.getStatusIcon(StatusDisconnected)
			keyIcon := ""
			if h.KeyFile != "" {
				keyIcon = " [yellow]KEY[white]"
			}

			mainText := fmt.Sprintf("  %s %s%s", icon, name, keyIcon)

			secondary := ""
			if h.Hostname != "" {
				hostDisplay := h.Hostname
				// Truncate long hostnames in secondary text too
				if len(hostDisplay) > 20 {
					hostDisplay = hostDisplay[:17] + "..."
				}
				secondary = fmt.Sprintf("    %s", hostDisplay)
				if h.Port != 0 && h.Port != 22 {
					secondary += fmt.Sprintf(":%d", h.Port)
				}
			}

			sl.list.AddItem(mainText, secondary, 0, nil)
			sl.groupedItems = append(sl.groupedItems, serverItem)
		}
	}
}

// getStatusIcon returns status indicator
func (sl *ServerList) getStatusIcon(status ServerStatus) string {
	switch status {
	case StatusConnected:
		return fmt.Sprintf("[%s]●[white]", colorToTag(sl.theme.Success))
	case StatusConnecting:
		return fmt.Sprintf("[%s]◐[white]", colorToTag(sl.theme.Warning))
	case StatusError:
		return fmt.Sprintf("[%s]✗[white]", colorToTag(sl.theme.Error))
	default:
		return fmt.Sprintf("[%s]○[white]", colorToTag(sl.theme.Muted))
	}
}

// SetStatus updates server connection status
func (sl *ServerList) SetStatus(index int, status ServerStatus, info string) {
	if index >= 0 && index < len(sl.servers) {
		sl.servers[index].Status = status
		sl.servers[index].Info = info
	}
}

// SetStatusByHost updates status by host entry
func (sl *ServerList) SetStatusByHost(host ssh.HostEntry, status ServerStatus, info string) {
	// Find and update in groupedItems
	for i, item := range sl.groupedItems {
		if server, ok := item.(*ServerItem); ok {
			if server.Host.Name == host.Name || server.Host.Hostname == host.Hostname {
				server.Status = status
				server.Info = info

				// Update display
				name := server.Host.Name
				if name == "" {
					name = server.Host.Hostname
				}

				icon := sl.getStatusIcon(status)
				keyIcon := ""
				if server.Host.KeyFile != "" {
					keyIcon = " [yellow]KEY[white]"
				}

				infoText := ""
				if info != "" && status == StatusConnected {
					infoText = fmt.Sprintf(" [%s](%s)[white]", colorToTag(sl.theme.Muted), truncate(info, 10))
				}

				mainText := fmt.Sprintf("  %s %s%s%s", icon, name, keyIcon, infoText)
				secondary := ""
				if server.Host.Hostname != "" {
					secondary = fmt.Sprintf("    %s", server.Host.Hostname)
				}

				sl.list.SetItemText(i, mainText, secondary)
				break
			}
		}
	}

	// Update servers slice
	for i := range sl.servers {
		if sl.servers[i].Host.Name == host.Name || sl.servers[i].Host.Hostname == host.Hostname {
			sl.servers[i].Status = status
			sl.servers[i].Info = info
			break
		}
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

// OnSelect sets the selection callback
func (sl *ServerList) OnSelect(fn func(ssh.HostEntry)) {
	sl.onSelect = fn
}

// OnConnect sets the connection callback
func (sl *ServerList) OnConnect(fn func(ssh.HostEntry)) {
	sl.onConnect = fn
}

// OnAddHost sets the add host callback
func (sl *ServerList) OnAddHost(fn func()) {
	sl.onAddHost = fn
}

// OnDeleteHost sets the delete host callback
func (sl *ServerList) OnDeleteHost(fn func(ssh.HostEntry)) {
	sl.onDeleteHost = fn
}

// OnEditHost sets the edit host callback
func (sl *ServerList) OnEditHost(fn func(ssh.HostEntry)) {
	sl.onEditHost = fn
}

// View returns the underlying view
func (sl *ServerList) View() tview.Primitive {
	return sl.list
}

// GetSelected returns the currently selected server
func (sl *ServerList) GetSelected() *ServerItem {
	index := sl.list.GetCurrentItem()
	if index >= 0 && index < len(sl.groupedItems) {
		if item, ok := sl.groupedItems[index].(*ServerItem); ok {
			return item
		}
	}
	return nil
}

// AddHost adds a new host to the list
func (sl *ServerList) AddHost(host ssh.HostEntry) {
	sl.servers = append(sl.servers, ServerItem{
		Host:   host,
		Status: StatusDisconnected,
	})
}

// HandleInput handles custom key input
func (sl *ServerList) HandleInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyRune:
		switch event.Rune() {
		case 'a':
			if sl.onAddHost != nil {
				sl.onAddHost()
			}
			return nil
		case 'd':
			// Delete selected host
			if item := sl.GetSelected(); item != nil {
				if sl.onDeleteHost != nil {
					sl.onDeleteHost(item.Host)
				}
			}
			return nil
		case 'e':
			// Edit selected host
			if item := sl.GetSelected(); item != nil {
				if sl.onEditHost != nil {
					sl.onEditHost(item.Host)
				}
			}
			return nil
		case 'j':
			return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
		case 'k':
			return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
		case 'g':
			sl.list.SetCurrentItem(0)
			return nil
		case 'G':
			sl.list.SetCurrentItem(sl.list.GetItemCount() - 1)
			return nil
		}
	}
	return event
}
