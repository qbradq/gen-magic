package ui

import (
	"log"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"github.com/qbradq/gen-magic/data"
)

// Main implements the main UI.
type Main struct {
	app fyne.App
	w fyne.Window
	p Project
}

// NewMain returns a new Main UI and returns it.
func NewMain() *Main {
	app := app.NewWithID("github.com/qbradq/gen-magic")
	w := app.NewWindow("Gen Magic")
	ret := &Main{
		app: app,
		w: w,
		p: Project{},
	}
	ret.w.SetMainMenu(ret.mainMenu())
	ret.w.SetContent(
		canvas.NewImageFromImage(data.BackgroundImage),
	)
	ret.w.Resize(fyne.NewSize(1024, 576))
	hDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("error getting home dir: %v\n", err)
	}
	projectPath := ret.app.Preferences().StringWithFallback(
		"last-open-project",
		filepath.Join(hDir, "default.gen-magic"),
	)
	if err := ret.loadProject(projectPath); err != nil {
		log.Fatalf("error loading project: %v\n", err)
	}
	return ret
}

// Run runs the app.
func (m *Main) Run() {
	m.w.ShowAndRun()
}

// mainMenu returns the MainMenu object.
func (m *Main) mainMenu() *fyne.MainMenu {
	return fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem("New Project", func() {
				
			}),
			fyne.NewMenuItem("Open Project", func() {

			}),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Quit", func() {
				m.app.Quit()
			}),
		),
		fyne.NewMenu("Settings",
			fyne.NewMenuItem("LLMs", func() {
				ShowLLMSettings(m)
			}),
			fyne.NewMenuItem("Agents", func() {
			}),
		),
	)
}

// loadProject loads a project by filename.
func (m *Main) loadProject(p string) error {
	return m.p.Load("sqlite", p)
}
