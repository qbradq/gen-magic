package ui

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/qbradq/gen-magic/llm"
)

// ShowLLMSettings creates and shows a new LLMSettings dialog.
func ShowLLMSettings(m *Main) {
	// Setup variables
	var def *llm.Definition
	f := widget.NewForm()
	var llmSelect *widget.Select
	var llmNameEntry *widget.Entry
	var apiSelect *widget.Select
	var urlEntry *widget.Entry
	var modelEntry *widget.Entry
	var apiKeyEntry *widget.Entry
	var llms []LLMName
	var apis []LLMApi
	lastEditedLLM := m.p.IntSetting("llm.last-edited", 0)
	// Internal functions
	var refreshLLMList = func() {
		llms = m.p.ListLLMs()
		llmStrs := []string{}
		for _, llmName := range llms {
			llmStrs = append(llmStrs, llmName.Name)
		}
		llmSelect.SetOptions(llmStrs)
		temp := llmSelect.OnChanged
		llmSelect.OnChanged = nil
		llmSelect.SetSelectedIndex(lastEditedLLM)
		llmSelect.OnChanged = temp
	}
	var updateUI = func() {
		// Set the value of all inputs
		refreshLLMList()
		llmNameEntry.SetText(def.Name)
		apiIdx := -1
		for i, api := range apis {
			if api.ID == def.API {
				apiIdx = i
				break
			}
		}
		apiSelect.SetSelectedIndex(apiIdx)
		urlEntry.SetText(def.APIEndpoint)
		modelEntry.SetText(def.Model)
		apiKeyEntry.SetText(def.APIKey)
	}
	var save = func() {
		if def != nil {
			if err := m.p.SetLLM(def); err != nil {
				log.Printf("error saving LLM: %v\n", err)
			}
		}
	}
	var load = func(id int64) {
		// Load the LLM definition
		def = m.p.GetLLM(id)
		if def == nil {
			dialog.NewError(
				fmt.Errorf("failed to load LLM %d", id),
				m.w,
			)
			m.app.Quit()
		}
		temp := llmSelect.OnChanged
		llmSelect.OnChanged = nil
		updateUI()
		llmSelect.OnChanged = temp
	}
	// LLM select
	llmSelect = widget.NewSelect(nil, func(s string) {
		if def == nil {
			return
		}
		save()
		lastEditedLLM = llmSelect.SelectedIndex()
		m.app.Preferences().SetInt("llm.last-edited", lastEditedLLM)
		load(llms[lastEditedLLM].ID)
	})
	f.Append("LLM", container.NewBorder(nil, nil, nil, container.NewHBox(
				widget.NewButtonWithIcon("", theme.Icon(theme.IconNameDelete), func() {
					m.p.DeleteLLM(def)
					def = nil
					temp := m.p.ListLLMs()
					if len(temp) == 0 {
						def = m.p.NewLLM()
						lastEditedLLM = 0
					} else {
						lastEditedLLM--
						if lastEditedLLM < 0 {
							lastEditedLLM = 0
						}
						def = m.p.GetLLM(llms[lastEditedLLM].ID)
					}
					m.app.Preferences().SetInt("llm.last-edited", lastEditedLLM)
					updateUI()
				}),
				widget.NewButtonWithIcon("", theme.Icon(theme.IconNameFile), func() {
					save()
					d := m.p.NewLLM()
					def = nil
					refreshLLMList()
					llmSelect.SetSelectedIndex(len(llmSelect.Options)-1)
					def = d
					updateUI()
				}),
			),
			llmSelect,
		),
	)
	refreshLLMList()
	// LLM name entry
	llmNameEntry = widget.NewEntry()
	llmNameEntry.SetPlaceHolder("LLM Definition Name")
	llmNameEntry.OnChanged = func(s string) {
		if def == nil {
			return
		}
		def.Name = s
	}
	f.Append("LLM Name", llmNameEntry)
	// API select
	apis = m.p.ListAPIs()
	apiNames := []string{}
	for _, api := range apis {
		apiNames = append(apiNames, api.Name)
	}
	apiSelect = widget.NewSelect(apiNames, nil)
	apiSelect.OnChanged = func(s string) {
		if def == nil {
			return
		}
		def.API = apis[apiSelect.SelectedIndex()].ID
	}
	f.Append("API Type", apiSelect)
	// URL entry
	urlEntry = widget.NewEntry()
	urlEntry.PlaceHolder = "LLM API Endpoint URL"
	urlEntry.OnChanged = func(s string) {
		def.APIEndpoint = s
	}
	f.Append("API Endpoint", urlEntry)
	// Model entry
	modelEntry = widget.NewEntry()
	modelEntry.OnChanged = func(s string) {
		def.Model = s
	}
	f.Append("Model", modelEntry)
	// API key
	apiKeyEntry = widget.NewEntry()
	apiKeyEntry.Password = true
	apiKeyEntry.OnChanged = func(s string) {
		def.APIKey = s
	}
	f.Append("API Key", apiKeyEntry)
	// Load the last edited LLM
	load(llms[llmSelect.SelectedIndex()].ID)
	// Show the dialog
	dlg := dialog.NewCustom("LLM Definitions", "Done", f, m.w)
	dlg.SetOnClosed(func() {
		save()
	})
	dlg.Resize(dlg.MinSize().AddWidthHeight(240, 0))
	dlg.Show()
}
