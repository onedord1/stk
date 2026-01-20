package ui

import (
	"fmt"
	"os/user"
	"strings"

	"github.com/rivo/tview"
	"github.com/systask/systask/internal/config"
	"github.com/systask/systask/internal/ssh"
)

// AddHostForm represents a form for adding/editing servers
type AddHostForm struct {
	form     *tview.Form
	flex     *tview.Flex
	theme    *config.Theme
	onSubmit func(config.Host)
	onCancel func()
	editMode bool
	editHost *ssh.HostEntry
}

// NewAddHostForm creates a new add host form
func NewAddHostForm(theme *config.Theme) *AddHostForm {
	f := &AddHostForm{
		theme: theme,
	}
	f.build("", "", "22", "", "key", "~/.ssh/id_rsa", "", "Custom")
	return f
}

// NewEditHostForm creates a form pre-filled with host data
func NewEditHostForm(theme *config.Theme, host ssh.HostEntry) *AddHostForm {
	f := &AddHostForm{
		theme:    theme,
		editMode: true,
		editHost: &host,
	}

	port := "22"
	if host.Port != 0 {
		port = fmt.Sprintf("%d", host.Port)
	}

	keyPath := host.KeyFile
	if keyPath == "" {
		keyPath = "~/.ssh/id_rsa"
	}

	auth := "key"
	if keyPath != "" {
		auth = "key"
	}

	group := "Custom"
	if host.Source != "" {
		group = host.Source
	}

	f.build(host.Name, host.Hostname, port, host.User, auth, keyPath, "", group)
	return f
}

// build constructs the form
func (f *AddHostForm) build(name, hostname, port, username, authType, keyPath, password, group string) {
	// Get default username if not provided
	if username == "" {
		if u, err := user.Current(); err == nil {
			username = u.Username
		} else {
			username = "root"
		}
	}

	// Create the form
	f.form = tview.NewForm()
	f.form.SetBackgroundColor(f.theme.Background)
	f.form.SetFieldBackgroundColor(f.theme.Muted)
	f.form.SetFieldTextColor(f.theme.Foreground)
	f.form.SetButtonBackgroundColor(f.theme.Primary)
	f.form.SetButtonTextColor(f.theme.Background)
	f.form.SetLabelColor(f.theme.Foreground)
	f.form.SetBorder(true)
	f.form.SetBorderColor(f.theme.Primary)

	title := " ➕ Add New Server "
	if f.editMode {
		title = " ✏️ Edit Server "
	}
	f.form.SetTitle(title)
	f.form.SetTitleColor(f.theme.Primary)

	// Variables to store form data
	var fName, fHostname, fPort, fUsername, fKeyPath, fPassword, fAuthType, fGroup string
	fName = name
	fHostname = hostname
	fPort = port
	fUsername = username
	fKeyPath = keyPath
	fPassword = password
	fAuthType = authType
	fGroup = group

	// Add fields
	f.form.AddInputField("Name", name, 20, nil, func(text string) { fName = text })
	f.form.AddInputField("Hostname", hostname, 20, nil, func(text string) { fHostname = text })
	f.form.AddInputField("Port", port, 6, tview.InputFieldInteger, func(text string) { fPort = text })
	f.form.AddInputField("User", username, 12, nil, func(text string) { fUsername = text })
	f.form.AddInputField("Auth", authType, 10, nil, func(text string) { fAuthType = text })
	f.form.AddInputField("Key Path", keyPath, 20, nil, func(text string) { fKeyPath = text })
	f.form.AddPasswordField("Password", password, 15, '*', func(text string) { fPassword = text })
	f.form.AddInputField("Group", group, 10, nil, func(text string) { fGroup = text })

	// Add buttons
	saveLabel := "Save"
	if f.editMode {
		saveLabel = "Update"
	}

	f.form.AddButton(saveLabel, func() {
		if fName == "" || fHostname == "" {
			return
		}

		portNum := 22
		fmt.Sscanf(fPort, "%d", &portNum)

		// Expand ~
		if strings.HasPrefix(fKeyPath, "~") {
			if u, err := user.Current(); err == nil {
				fKeyPath = strings.Replace(fKeyPath, "~", u.HomeDir, 1)
			}
		}

		auth := strings.ToLower(strings.TrimSpace(fAuthType))
		if auth != "key" && auth != "password" && auth != "agent" {
			auth = "key"
		}

		finalKeyPath := fKeyPath
		finalPassword := ""

		switch auth {
		case "key":
			finalPassword = ""
		case "password":
			finalKeyPath = ""
			finalPassword = fPassword
		case "agent":
			finalKeyPath = ""
			finalPassword = ""
		}

		host := config.Host{
			Name:     fName,
			Hostname: fHostname,
			Port:     portNum,
			User:     fUsername,
			KeyPath:  finalKeyPath,
			Password: finalPassword,
			Group:    fGroup,
			Provider: "custom",
			AuthType: auth,
		}

		if f.onSubmit != nil {
			f.onSubmit(host)
		}
	})

	f.form.AddButton("Cancel", func() {
		if f.onCancel != nil {
			f.onCancel()
		}
	})

	// Create full-screen background with centered form
	f.flex = tview.NewFlex().SetDirection(tview.FlexRow)
	f.flex.SetBackgroundColor(f.theme.Background)

	// Top spacer
	f.flex.AddItem(tview.NewBox().SetBackgroundColor(f.theme.Background), 2, 0, false)

	// Center row
	centerRow := tview.NewFlex()
	centerRow.SetBackgroundColor(f.theme.Background)
	centerRow.AddItem(tview.NewBox().SetBackgroundColor(f.theme.Background), 0, 1, false)
	centerRow.AddItem(f.form, 45, 0, true)
	centerRow.AddItem(tview.NewBox().SetBackgroundColor(f.theme.Background), 0, 1, false)

	f.flex.AddItem(centerRow, 22, 0, true)

	// Bottom spacer
	f.flex.AddItem(tview.NewBox().SetBackgroundColor(f.theme.Background), 0, 1, false)
}

// IsEditMode returns whether this is an edit form
func (f *AddHostForm) IsEditMode() bool {
	return f.editMode
}

// GetEditHost returns the host being edited
func (f *AddHostForm) GetEditHost() *ssh.HostEntry {
	return f.editHost
}

// OnSubmit sets the submit callback
func (f *AddHostForm) OnSubmit(fn func(config.Host)) {
	f.onSubmit = fn
}

// OnCancel sets the cancel callback
func (f *AddHostForm) OnCancel(fn func()) {
	f.onCancel = fn
}

// View returns the modal view
func (f *AddHostForm) View() tview.Primitive {
	return f.flex
}

// Focus sets focus to the form
func (f *AddHostForm) Focus(app *tview.Application) {
	app.SetFocus(f.form)
}
