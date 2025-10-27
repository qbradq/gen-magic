package ui

import (
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/qbradq/gen-magic/llm"
)

// ShowLLMSettings creates and shows a new LLMSettings dialog.
func ShowLLMSettings(in *llm.Definition, onApply func(llm.Definition)) {
	var fn = func(s, d string) string {
		if s == "" {
			return d
		}
		return s
	}
	def := *in
	f := widget.NewForm()
	ats := widget.NewSelect([]string{
		"OpenRouter",
	}, nil)
	ats.SetSelected(fn(def.API, "OpenRouter"))
	def.API = ats.Selected
	ats.OnChanged = func(s string) {
		def.API = s
	}
	f.Append("API Type", ats)
	aee := widget.NewEntry()
	aee.SetText(fn(def.APIEndpoint, "https://openrouter.ai/api/v1"))
	def.APIEndpoint = aee.Text
	aee.PlaceHolder = "LLM API Endpoint URL"
	aee.OnChanged = func(s string) {
		def.APIEndpoint = s
	}
	f.Append("API Endpoint", aee)
	mdl := widget.NewEntry()
	mdl.SetText(fn(def.Model, "meta-llama/llama-3.3-70b-instruct:free"))
	def.Model = mdl.Text
	mdl.OnChanged = func(s string) {
		def.Model = s
	}
	f.Append("Model", mdl)
	ak := widget.NewEntry()
	ak.Password = true
	ak.SetText(def.APIKey)
	ak.OnChanged = func(s string) {
		def.APIKey = s
	}
	f.Append("API Key", ak)
	dlg := dialog.NewCustomConfirm("LLM Definition", "Apply", "Cancel", f, func(b bool) {
		if !b {
			return
		}
		onApply(def)
	}, window)
	dlg.Resize(dlg.MinSize().AddWidthHeight(240, 0))
	dlg.Show()
}
