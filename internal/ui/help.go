package ui

import (
	"fmt"

	"github.com/rivo/tview"
	"github.com/systask/systask/internal/config"
)

// HelpOverlay manages the help display
type HelpOverlay struct {
	view  *tview.Flex
	theme *config.Theme

	// Context-specific help
	context string
}

// NewHelpOverlay creates a new help overlay
func NewHelpOverlay(theme *config.Theme) *HelpOverlay {
	h := &HelpOverlay{
		theme:   theme,
		context: "global",
	}
	h.build()
	return h
}

// build constructs the help overlay
func (h *HelpOverlay) build() {
	h.view = tview.NewFlex()
	h.view.SetBackgroundColor(h.theme.Background)
}

// getHelpContent returns help text for current context
func (h *HelpOverlay) getHelpContent() string {
	primary := colorToTag(h.theme.Primary)
	secondary := colorToTag(h.theme.Secondary)
	highlight := colorToTag(h.theme.Highlight)
	muted := colorToTag(h.theme.Muted)

	globalHelp := fmt.Sprintf(`[%s]━━━ Global Navigation ━━━[white]

[%s]Movement:[white]
  [%s]j/↓[white]     Move down              [%s]k/↑[white]     Move up
  [%s]h/←[white]     Move left              [%s]l/→[white]     Move right
  [%s]g[white]       Go to top              [%s]G[white]       Go to bottom
  [%s]Tab[white]     Next panel             [%s]S-Tab[white]   Previous panel

[%s]Modes:[white]
  [%s]:[white]       Command mode           [%s]/[white]       Search mode
  [%s]Esc[white]     Exit mode/cancel       [%s]?[white]       Toggle help

[%s]Actions:[white]
  [%s]Enter[white]   Select/Confirm         [%s]r[white]       Refresh
  [%s]q[white]       Quit                   [%s]c[white]       Connect to server

[%s]━━━ Modules ━━━[white]

  [%s]1[white] Health Dashboard       [%s]2[white] Service Manager
  [%s]3[white] Process Manager        [%s]4[white] Log Viewer
  [%s]5[white] Disk Manager           [%s]6[white] Batch Commands

[%s]━━━ Commands ━━━[white]

  [%s]:theme <name>[white]      Switch color theme
  [%s]:connect <host>[white]    Connect to a server
  [%s]:disconnect[white]        Disconnect from current server
  [%s]:run <cmd>[white]         Run command on selected servers
  [%s]:quit[white] or [%s]:q[white]        Exit application

[%s]━━━ Themes ━━━[white]

  catppuccin  dracula  nord  gruvbox  solarized  tokyo-night

[%s]Press ? or Esc to close this help[white]`,
		primary,
		secondary,
		highlight, highlight,
		highlight, highlight,
		highlight, highlight,
		highlight, highlight,
		secondary,
		highlight, highlight,
		highlight, highlight,
		secondary,
		highlight, highlight,
		highlight, highlight,
		primary,
		highlight, highlight,
		highlight, highlight,
		highlight, highlight,
		primary,
		highlight,
		highlight,
		highlight,
		highlight,
		highlight, highlight,
		primary,
		muted,
	)

	// Add context-specific help based on current view
	switch h.context {
	case "health":
		return globalHelp + h.getHealthHelp()
	case "services":
		return globalHelp + h.getServicesHelp()
	case "processes":
		return globalHelp + h.getProcessesHelp()
	case "logs":
		return globalHelp + h.getLogsHelp()
	default:
		return globalHelp
	}
}

func (h *HelpOverlay) getHealthHelp() string {
	highlight := colorToTag(h.theme.Highlight)
	return fmt.Sprintf(`

[%s]━━━ Health Dashboard ━━━[white]

  [%s]r[white]     Refresh metrics
  [%s]1-4[white]   Select metric panel
`,
		colorToTag(h.theme.Secondary),
		highlight,
		highlight,
	)
}

func (h *HelpOverlay) getServicesHelp() string {
	highlight := colorToTag(h.theme.Highlight)
	return fmt.Sprintf(`

[%s]━━━ Service Manager ━━━[white]

  [%s]s[white]     Start service
  [%s]S[white]     Stop service
  [%s]r[white]     Restart service
  [%s]e[white]     Enable service
  [%s]d[white]     Disable service
  [%s]l[white]     View service logs
`,
		colorToTag(h.theme.Secondary),
		highlight, highlight, highlight,
		highlight, highlight, highlight,
	)
}

func (h *HelpOverlay) getProcessesHelp() string {
	highlight := colorToTag(h.theme.Highlight)
	return fmt.Sprintf(`

[%s]━━━ Process Manager ━━━[white]

  [%s]k[white]     Kill process (SIGTERM)
  [%s]K[white]     Kill process (SIGKILL)
  [%s]t[white]     Sort by CPU
  [%s]m[white]     Sort by memory
  [%s]p[white]     Sort by PID
`,
		colorToTag(h.theme.Secondary),
		highlight, highlight,
		highlight, highlight, highlight,
	)
}

func (h *HelpOverlay) getLogsHelp() string {
	highlight := colorToTag(h.theme.Highlight)
	return fmt.Sprintf(`

[%s]━━━ Log Viewer ━━━[white]

  [%s]f[white]     Follow mode (tail -f)
  [%s]g[white]     Go to beginning
  [%s]G[white]     Go to end
  [%s]/[white]     Filter logs
  [%s]n[white]     Next match
  [%s]N[white]     Previous match
`,
		colorToTag(h.theme.Secondary),
		highlight, highlight, highlight,
		highlight, highlight, highlight,
	)
}

// Show displays the help overlay
func (h *HelpOverlay) Show(pages *tview.Pages, app *tview.Application) {
	content := h.getHelpContent()

	textView := tview.NewTextView()
	textView.SetDynamicColors(true)
	textView.SetText(content)
	textView.SetBackgroundColor(h.theme.Background)
	textView.SetTextColor(h.theme.Foreground)
	textView.SetScrollable(true)

	box := tview.NewFlex().SetDirection(tview.FlexRow)
	box.AddItem(textView, 0, 1, true)
	box.SetBorder(true)
	box.SetBorderColor(h.theme.Primary)
	box.SetTitle(" Help - Press ? or Esc to close ")
	box.SetTitleColor(h.theme.Primary)
	box.SetBackgroundColor(h.theme.Background)

	// Center the help box
	flex := tview.NewFlex()
	flex.AddItem(nil, 0, 1, false)
	flex.AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(box, 0, 4, true).
		AddItem(nil, 0, 1, false), 0, 3, true)
	flex.AddItem(nil, 0, 1, false)

	pages.AddPage("help", flex, true, true)
}

// Hide removes the help overlay
func (h *HelpOverlay) Hide(pages *tview.Pages) {
	pages.RemovePage("help")
}

// SetContext updates the help context
func (h *HelpOverlay) SetContext(context string) {
	h.context = context
}

// View returns the underlying view
func (h *HelpOverlay) View() *tview.Flex {
	return h.view
}
