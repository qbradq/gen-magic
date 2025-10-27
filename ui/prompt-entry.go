package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// PromptEntry is an Entry customized for AI prompts.
type PromptEntry struct {
	widget.Entry
	OnCtrlEnter func()
}

// NewPromptEntry returns a new PromptEntry.
func NewPromptEntry() *PromptEntry {
	ret := &PromptEntry{}
	ret.ExtendBaseWidget(ret)
	return ret
}

// TypedShortcut handles keyboard shortcuts.
func (e *PromptEntry) TypedShortcut(s fyne.Shortcut) {
	cs, ok := s.(*desktop.CustomShortcut)
	if ok && cs.Modifier == fyne.KeyModifierControl {
		switch cs.KeyName {
		case fyne.KeyEnter:
			fallthrough
		case fyne.KeyReturn:
			if e.OnCtrlEnter != nil {
				e.OnCtrlEnter()
			}
			return
		}
	}
	e.Entry.TypedShortcut(s)
}
