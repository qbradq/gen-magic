package ui

import (
	"fmt"
	"log"
	"maps"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"github.com/qbradq/gen-magic/data"
)

// Closer implementations offer a Close() method.
type Closer interface {
	Close()
}

// Main implements the main UI.
type Main struct {
	app fyne.App
	w fyne.Window
	p Project
	children map[Closer]struct{}
}

// NewMain returns a new Main UI and returns it.
func NewMain() *Main {
	app := app.NewWithID("github.com/qbradq/gen-magic")
	w := app.NewWindow("Gen Magic")
	ret := &Main{
		app: app,
		w: w,
		children: map[Closer]struct{}{},
	}
	ret.w.SetOnClosed(func() {
		ret.CloseChildren()  
	})
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
					if writer == nil {
						return
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
				fileOpen := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
					if err != nil {
						log.Printf("error in open project file: %v\n", err)
					}
					if reader == nil {
						return
					}
					m.LoadProject(reader.URI().Path())
				}, m.w)
				fileOpen.SetConfirmText("Open Project")
				fileOpen.SetDismissText("Cancel")
				fileOpen.SetFilter(storage.NewExtensionFileFilter([]string{
					".gen-magic",
				}))
				fileOpen.SetTitleText("Open Project")
				fileOpen.Show()
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
				ShowAgentSettings(m)
			}),
		),
		fyne.NewMenu("Start",
			fyne.NewMenuItem("Chat", func() {
				var chat Closer
				chat = NewChat(m, func() {
					delete(m.children, chat)
				})
				m.children[chat] = struct{}{}
			}),
		),
	)
}

// AddChild adds a Closer to the app as a child.
func (m *Main) AddChild(child Closer) {
	m.children[child] = struct{}{}
}

// RemoveChild removes a Closer from the app.
func (m *Main) RemoveChild(child Closer) {
	delete(m.children, child)
}

// CloseChildren closes all open child windows.
func (m *Main) CloseChildren() {
	closers := []Closer{}
	for closer := range maps.Keys(m.children) {
		closers = append(closers, closer)
	}
	for _, closer := range closers {
		closer.Close()
		delete(m.children, closer)
	}
}

// LoadProject loads a project by filename.
func (m *Main) LoadProject(p string) error {
	m.CloseChildren()
	m.p.Close()
	if err := m.p.Load("sqlite", p); err != nil {
		return err
	}
	m.w.SetTitle(fmt.Sprintf("Gen Magic \"%s\"", p))
	m.app.Preferences().SetString("last-open-project", p)
	return nil
}

// FireOnLLMsUpdated fires the OnLLMsUpdated method on all open windows that
// implement it.
func (m *Main) FireOnLLMsUpdated() {
	for child := range maps.Keys(m.children) {
		iChild, ok := child.(interface{OnLLMsUpdated()})
		if !ok {
			continue
		}
		iChild.OnLLMsUpdated()
	}
}
