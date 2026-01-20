package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/systask/systask/internal/config"
)

// Modal represents a modal dialog
type Modal struct {
	view  *tview.Modal
	theme *config.Theme

	onDone func(buttonIndex int, buttonLabel string)
}

// NewModal creates a new modal dialog
func NewModal(theme *config.Theme) *Modal {
	m := &Modal{
		theme: theme,
	}
	m.build()
	return m
}

// build constructs the modal
func (m *Modal) build() {
	m.view = tview.NewModal()
	m.view.SetBackgroundColor(m.theme.Background)
	m.view.SetTextColor(m.theme.Foreground)
	m.view.SetButtonBackgroundColor(m.theme.Primary)
	m.view.SetButtonTextColor(m.theme.Background)

	m.view.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if m.onDone != nil {
			m.onDone(buttonIndex, buttonLabel)
		}
	})
}

// Show displays the modal
func (m *Modal) Show(text string, buttons []string) {
	m.view.SetText(text)
	m.view.ClearButtons()
	for _, btn := range buttons {
		m.view.AddButtons([]string{btn})
	}
}

// OnDone sets the completion callback
func (m *Modal) OnDone(fn func(buttonIndex int, buttonLabel string)) {
	m.onDone = fn
}

// View returns the underlying view
func (m *Modal) View() *tview.Modal {
	return m.view
}

// ConfirmModal shows a confirmation dialog
func ConfirmModal(app *tview.Application, pages *tview.Pages, theme *config.Theme,
	title, message string, onConfirm func()) {

	modal := tview.NewModal()
	modal.SetBackgroundColor(theme.Background)
	modal.SetTextColor(theme.Foreground)
	modal.SetButtonBackgroundColor(theme.Primary)
	modal.SetButtonTextColor(theme.Background)
	modal.SetText(fmt.Sprintf("[%s]%s[-]\n\n%s", colorToTag(theme.Warning), title, message))
	modal.AddButtons([]string{"Cancel", "Confirm"})

	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		pages.RemovePage("confirm-modal")
		if buttonLabel == "Confirm" && onConfirm != nil {
			onConfirm()
		}
	})

	pages.AddPage("confirm-modal", modal, true, true)
}

// InputModal shows an input dialog
func InputModal(app *tview.Application, pages *tview.Pages, theme *config.Theme,
	title, label string, onSubmit func(string)) {

	input := tview.NewInputField()
	input.SetBackgroundColor(theme.Muted)
	input.SetFieldBackgroundColor(theme.Background)
	input.SetFieldTextColor(theme.Foreground)
	input.SetLabelColor(theme.Primary)
	input.SetLabel(label + ": ")
	input.SetFieldWidth(40)

	form := tview.NewFlex().SetDirection(tview.FlexRow)
	form.SetBackgroundColor(theme.Background)
	form.SetBorder(true)
	form.SetBorderColor(theme.Border)
	form.SetTitle(fmt.Sprintf(" %s ", title))
	form.SetTitleColor(theme.Primary)

	form.AddItem(tview.NewBox().SetBackgroundColor(theme.Background), 1, 0, false)
	form.AddItem(input, 1, 0, true)
	form.AddItem(tview.NewBox().SetBackgroundColor(theme.Background), 1, 0, false)

	// Create centered container
	flex := tview.NewFlex()
	flex.AddItem(nil, 0, 1, false)
	flex.AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(form, 5, 0, true).
		AddItem(nil, 0, 1, false), 50, 0, true)
	flex.AddItem(nil, 0, 1, false)

	input.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			text := input.GetText()
			pages.RemovePage("input-modal")
			if onSubmit != nil {
				onSubmit(text)
			}
		case tcell.KeyEsc:
			pages.RemovePage("input-modal")
		}
	})

	pages.AddPage("input-modal", flex, true, true)
	app.SetFocus(input)
}

// ErrorModal shows an error message
func ErrorModal(app *tview.Application, pages *tview.Pages, theme *config.Theme,
	title, message string) {

	modal := tview.NewModal()
	modal.SetBackgroundColor(theme.Background)
	modal.SetTextColor(theme.Foreground)
	modal.SetButtonBackgroundColor(theme.Error)
	modal.SetButtonTextColor(theme.Foreground)
	modal.SetText(fmt.Sprintf("[%s]⚠ %s[-]\n\n%s", colorToTag(theme.Error), title, message))
	modal.AddButtons([]string{"OK"})

	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		pages.RemovePage("error-modal")
	})

	pages.AddPage("error-modal", modal, true, true)
}

// SuccessModal shows a success message
func SuccessModal(app *tview.Application, pages *tview.Pages, theme *config.Theme,
	title, message string) {

	modal := tview.NewModal()
	modal.SetBackgroundColor(theme.Background)
	modal.SetTextColor(theme.Foreground)
	modal.SetButtonBackgroundColor(theme.Success)
	modal.SetButtonTextColor(theme.Background)
	modal.SetText(fmt.Sprintf("[%s]✓ %s[-]\n\n%s", colorToTag(theme.Success), title, message))
	modal.AddButtons([]string{"OK"})

	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		pages.RemovePage("success-modal")
	})

	pages.AddPage("success-modal", modal, true, true)
}
