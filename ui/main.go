package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/driver/desktop"
)

// Hacky globals
var window fyne.Window
var fyneApp fyne.App

// Main implements the main UI.
type Main struct {
	app fyne.App // Application pointer
	w fyne.Window // Single window instance
}

// NewMain returns a new Main UI and returns it.
func NewMain() *Main {
	fyneApp = app.NewWithID("github.com/qbradq/gen-magic")
	window = fyneApp.NewWindow("Gen Magic")
	window.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName: fyne.KeyEnter,
		Modifier: fyne.KeyModifierControl,
	}, nil)
	llm := NewLLM("cnc")
	window.SetContent(llm.Root())
	llm.Focus()
	return &Main{
		app: fyneApp,
		w: window,
	}
}

// Run runs the app.
func (m *Main) Run() {
	m.w.ShowAndRun()
}
