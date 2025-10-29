package ui

import (
	"log"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
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
	}
	ret.w.SetMainMenu(ret.mainMenu())
	ret.w.SetContent(
		canvas.NewImageFromImage(data.BackgroundImage),
	)
	ret.w.Resize(fyne.NewSize(1024, 576))
	ret.w.SetFixedSize(true)
	hDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("error getting home dir: %v\n", err)
	}
	projectPath := ret.app.Preferences().StringWithFallback(
		"last-open-project",
		filepath.Join(hDir, "default.gen-magic"),
	)
	if err := ret.LoadProject(projectPath); err != nil {
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
				fileSave := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
					if err != nil {
						log.Printf("error in new project file: %v\n", err)
					}
					m.LoadProject(writer.URI().Path())
				}, m.w)
				fileSave.SetConfirmText("Create Project")
				fileSave.SetDismissText("Cancel")
				fileSave.SetFilter(storage.NewExtensionFileFilter([]string{
					".gen-magic",
				}))
				fileSave.SetTitleText("Create New Project")
				fileSave.Show()
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

// LoadProject loads a project by filename.
func (m *Main) LoadProject(p string) error {
	m.p.Close()
	return m.p.Load("sqlite", p)
}
