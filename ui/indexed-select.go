package ui

import "fyne.io/fyne/v2/widget"

// IndexedSelect implements fyne.widget.Select but using ints instead of strings
// to index the list, thereby making duplicate strings no longer an issue.
type IndexedSelect struct {
	widget.Select
	CurrentSelectedIndex int
	OnChangedIndexed func(idx int)
}

// NewIndexedSelect creates a new IndexedSelect object.
func NewIndexedSelect(options []string, onChangedIndexed func(idx int)) *IndexedSelect {
	ret := &IndexedSelect{
		OnChangedIndexed: onChangedIndexed,
	}
	ret.ExtendBaseWidget(ret)
	ret.OnChanged = func(s string) {
		idx := -1
		for i, o := range ret.Options {
			if o == s {
				idx = i
				break
			}
		}
		if idx < 0 {
			return
		}
		ret.CurrentSelectedIndex = idx
		ret.Selected = s
		if ret.OnChangedIndexed != nil {
			ret.OnChangedIndexed(ret.CurrentSelectedIndex)
		}
	}
	ret.SetOptions(options)
	ret.rawSetSelectedIndex(0)	
	return ret
}

func (w *IndexedSelect) SelectedIndex() int {
	return w.CurrentSelectedIndex
}

func (w *IndexedSelect) SetSelected(text string) {
	defer w.Refresh()
	for i, o := range w.Options {
		if text == o {
			w.Selected = text
			w.CurrentSelectedIndex = i
			return
		}
	}
}

func (w *IndexedSelect) rawSetSelectedIndex(index int) {
	defer w.Refresh()
	if len(w.Options) == 0 {
		w.CurrentSelectedIndex = 0
		w.Selected = ""
		return
	}
	if index >= len(w.Options) {
		index = len(w.Options) - 1
	}
	if index < 0 {
		index = 0
	}
	w.CurrentSelectedIndex = index
	w.Selected = w.Options[index]
}

func (w *IndexedSelect) SetSelectedIndex(index int) {
	w.rawSetSelectedIndex(index)
	if w.OnChangedIndexed != nil {
		w.OnChangedIndexed(w.CurrentSelectedIndex)
	}
}
