package ui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/systask/systask/internal/config"
)

// Layout manages the main application layout
type Layout struct {
	root       *tview.Flex
	serverList *ServerList
	mainView   *tview.Pages
	statusBar  *StatusBar
	theme      *config.Theme
	app        *tview.Application
	focusIndex int
}

// NewLayout creates a new application layout
func NewLayout(app *tview.Application, theme *config.Theme) *Layout {
	l := &Layout{
		theme: theme,
		app:   app,
	}
	l.build()
	return l
}

// build constructs the layout
func (l *Layout) build() {
	// Apply theme colors
	tview.Styles.PrimitiveBackgroundColor = l.theme.Background
	tview.Styles.ContrastBackgroundColor = l.theme.Muted
	tview.Styles.MoreContrastBackgroundColor = l.theme.Muted
	tview.Styles.BorderColor = l.theme.Border
	tview.Styles.TitleColor = l.theme.Primary
	tview.Styles.GraphicsColor = l.theme.Muted
	tview.Styles.PrimaryTextColor = l.theme.Foreground
	tview.Styles.SecondaryTextColor = l.theme.Muted

	// Create server list (sidebar)
	l.serverList = NewServerList(l.theme)

	// Create main view area (tabbed pages)
	l.mainView = tview.NewPages()
	l.mainView.SetBackgroundColor(l.theme.Background)

	// Create welcome screen
	welcome := l.createWelcomeScreen()
	l.mainView.AddPage("welcome", welcome, true, true)

	// Main content area with border
	mainContent := tview.NewFlex()
	mainContent.SetDirection(tview.FlexColumn)
	mainContent.AddItem(l.serverList.View(), 35, 0, true)
	mainContent.AddItem(l.mainView, 0, 1, false)
	mainContent.SetBackgroundColor(l.theme.Background)

	// Create status bar
	l.statusBar = NewStatusBar(l.theme)

	// Root layout
	l.root = tview.NewFlex()
	l.root.SetDirection(tview.FlexRow)
	l.root.AddItem(mainContent, 0, 1, true)
	l.root.AddItem(l.statusBar.View(), 1, 0, false)
	l.root.SetBackgroundColor(l.theme.Background)
}

// createWelcomeScreen creates the welcome/home screen
func (l *Layout) createWelcomeScreen() *tview.Flex {
	// ASCII art logo with gradient effect
	logo := `
   ███████╗██╗   ██╗███████╗████████╗ █████╗ ███████╗██╗  ██╗
   ██╔════╝╚██╗ ██╔╝██╔════╝╚══██╔══╝██╔══██╗██╔════╝██║ ██╔╝
   ███████╗ ╚████╔╝ ███████╗   ██║   ███████║███████╗█████╔╝ 
   ╚════██║  ╚██╔╝  ╚════██║   ██║   ██╔══██║╚════██║██╔═██╗ 
   ███████║   ██║   ███████║   ██║   ██║  ██║███████║██║  ██╗
   ╚══════╝   ╚═╝   ╚══════╝   ╚═╝   ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝`

	primary := colorToTag(l.theme.Primary)
	secondary := colorToTag(l.theme.Secondary)
	muted := colorToTag(l.theme.Muted)
	success := colorToTag(l.theme.Success)
	highlight := colorToTag(l.theme.Highlight)

	// Logo with color
	logoView := tview.NewTextView()
	logoView.SetDynamicColors(true)
	logoView.SetTextAlign(tview.AlignCenter)
	logoView.SetText(fmt.Sprintf("[%s]%s[white]", primary, logo))
	logoView.SetBackgroundColor(l.theme.Background)

	// Subtitle
	subtitle := tview.NewTextView()
	subtitle.SetDynamicColors(true)
	subtitle.SetTextAlign(tview.AlignCenter)
	subtitle.SetText(fmt.Sprintf("\n[%s]Terminal System Manager for Linux Servers[white]",
		secondary))
	subtitle.SetBackgroundColor(l.theme.Background)

	// Module grid - left-aligned content for consistency
	moduleGrid := fmt.Sprintf(`[%s]── MODULES ──[white]

  [%s]1[white] [%s]Health[white]        [%s]2[white] [%s]Services[white]       [%s]3[white] [%s]Processes[white]
  [%s]4[white] [%s]Logs[white]          [%s]5[white] [%s]Disks[white]          [%s]6[white] [%s]Batch[white]
  [%s]7[white] [%s]Users[white]         [%s]8[white] [%s]Docker[white]         [%s]9[white] [%s]Installer[white]
  [%s]F[white] [%s]SFTP[white]          [%s]t[white] [%s]Terminal[white]`,
		muted,
		highlight, success, highlight, success, highlight, success,
		highlight, success, highlight, success, highlight, success,
		highlight, success, highlight, success, highlight, success,
		highlight, success, highlight, success)

	modulesView := tview.NewTextView()
	modulesView.SetDynamicColors(true)
	modulesView.SetTextAlign(tview.AlignCenter)
	modulesView.SetText(moduleGrid)
	modulesView.SetBackgroundColor(l.theme.Background)

	// Keyboard shortcuts - left-aligned content for consistency
	shortcuts := fmt.Sprintf(`[%s]── KEYBOARD ──[white]

  [%s]Enter[white]  Connect to server     [%s]ESC[white]   Return to home
  [%s]:[white]      Command mode          [%s]/[white]     Search/filter
  [%s]?[white]      Help overlay          [%s]q[white]     Quit
  [%s]Tab[white]    Switch panels         [%s]r[white]     Refresh
  [%s]a[white]      Add server            [%s]e/d[white]   Edit/Delete`,
		muted,
		highlight, highlight,
		highlight, highlight,
		highlight, highlight,
		highlight, highlight,
		highlight, highlight)

	shortcutsView := tview.NewTextView()
	shortcutsView.SetDynamicColors(true)
	shortcutsView.SetTextAlign(tview.AlignCenter)
	shortcutsView.SetText(shortcuts)
	shortcutsView.SetBackgroundColor(l.theme.Background)

	// Themes info
	themes := fmt.Sprintf(`[%s]🎨 Themes:[white] [%s]:theme <name>[white]
[%s]catppuccin[white] │ [%s]dracula[white] │ [%s]nord[white] │ [%s]gruvbox[white] │ [%s]solarized[white] │ [%s]tokyo-night[white]`,
		muted, highlight,
		primary, primary, primary, primary, primary, primary)

	themesView := tview.NewTextView()
	themesView.SetDynamicColors(true)
	themesView.SetTextAlign(tview.AlignCenter)
	themesView.SetText(themes)
	themesView.SetBackgroundColor(l.theme.Background)

	// Tip
	tip := tview.NewTextView()
	tip.SetDynamicColors(true)
	tip.SetTextAlign(tview.AlignCenter)
	tip.SetText(fmt.Sprintf("\n[%s]💡 Select a server and press [white][%s]Enter[white][%s] to connect[white]",
		muted, highlight, muted))
	tip.SetBackgroundColor(l.theme.Background)

	// Version / timestamp
	version := tview.NewTextView()
	version.SetDynamicColors(true)
	version.SetTextAlign(tview.AlignCenter)
	version.SetText(fmt.Sprintf("[%s]v3.0 │ %s[white]", muted, time.Now().Format("2006-01-02")))
	version.SetBackgroundColor(l.theme.Background)

	// Layout
	content := tview.NewFlex().SetDirection(tview.FlexRow)
	content.AddItem(tview.NewBox().SetBackgroundColor(l.theme.Background), 1, 0, false)
	content.AddItem(logoView, 7, 0, false)
	content.AddItem(subtitle, 4, 0, false)
	content.AddItem(modulesView, 10, 0, false)
	content.AddItem(shortcutsView, 12, 0, false)
	content.AddItem(themesView, 3, 0, false)
	content.AddItem(tip, 2, 0, false)
	content.AddItem(tview.NewBox().SetBackgroundColor(l.theme.Background), 0, 1, false)
	content.AddItem(version, 1, 0, false)
	content.SetBackgroundColor(l.theme.Background)

	// Wrapper with title
	wrapper := tview.NewFlex()
	wrapper.AddItem(content, 0, 1, false)
	wrapper.SetBorder(true)
	wrapper.SetBorderColor(l.theme.Border)
	wrapper.SetTitle(" Main View ")
	wrapper.SetTitleColor(l.theme.Primary)
	wrapper.SetBackgroundColor(l.theme.Background)

	flex := tview.NewFlex()
	flex.AddItem(wrapper, 0, 1, false)
	flex.SetBackgroundColor(l.theme.Background)

	return flex
}

// Root returns the root layout
func (l *Layout) Root() *tview.Flex {
	return l.root
}

// ServerList returns the server list
func (l *Layout) ServerList() *ServerList {
	return l.serverList
}

// MainView returns the main view pages
func (l *Layout) MainView() *tview.Pages {
	return l.mainView
}

// StatusBar returns the status bar
func (l *Layout) StatusBar() *StatusBar {
	return l.statusBar
}

// NextFocus moves focus to the next panel
func (l *Layout) NextFocus() {
	l.focusIndex = (l.focusIndex + 1) % 2
	l.applyFocus()
}

// PrevFocus moves focus to the previous panel
func (l *Layout) PrevFocus() {
	l.focusIndex = (l.focusIndex + 1) % 2
	l.applyFocus()
}

func (l *Layout) applyFocus() {
	switch l.focusIndex {
	case 0:
		l.app.SetFocus(l.serverList.View())
	case 1:
		l.app.SetFocus(l.mainView)
	}
}

// Helper to convert color to tview tag
func colorToTag(c tcell.Color) string {
	r, g, b := c.RGB()
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}
