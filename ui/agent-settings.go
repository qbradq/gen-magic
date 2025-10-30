package ui

import (
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/qbradq/gen-magic/llm"
)

// ShowAgentSettings shows the agent editor dialog.
func ShowAgentSettings(m *Main) {
	// Variables
	var agent *llm.Agent
	var agents []AgentName
	var llms []LLMName
	var agentSelect *IndexedSelect
	var btnDelete *widget.Button
	var btnNew *widget.Button
	var nameEntry *widget.Entry
	var llmSelect *IndexedSelect
	var sysEntry *widget.Entry
	lastEditedAgent := m.p.IntSetting("agent.last-edited", 0)
	f := widget.NewForm()
	// Internal functions
	var save = func() {
		m.p.SetAgent(agent)
	}
	var refreshAgentList = func() {
		agents = m.p.ListAgents()
		names := []string{}
		idx := -1
		for i, name := range agents {
			if agent.ID == name.ID {
				idx = i
			}
			names = append(names, name.Name)
		}
		agentSelect.SetOptions(names)
		agentSelect.rawSetSelectedIndex(idx)
	}
	var refreshLLMList = func() {
		llms = m.p.ListLLMs()
		llmStrs := []string{}
		idx := -1
		for i, llmName := range llms {
			llmStrs = append(llmStrs, llmName.Name)
			if llmName.ID == agent.ID {
				idx = i
			}
		}
		llmSelect.SetOptions(llmStrs)
		llmSelect.rawSetSelectedIndex(idx)
	}
	var updateUI = func() {
		if agent == nil {
			agent = m.p.NewAgent()
		}
		refreshAgentList()
		refreshLLMList()
		nameEntry.SetText(agent.Name)
	}
	var load = func(id int64) {
		agent = m.p.GetAgent(id)
		updateUI()
	}
	// Agent select with delete and new buttons
	lastEditedAgent = m.p.IntSetting("agent.last-edited", 0)
	agentSelect = NewIndexedSelect(nil, func(idx int) {
		name := agents[idx]
		lastEditedAgent = idx
		m.p.SetIntSetting("agent.last-edited", idx)
		load(name.ID)
	})
	btnDelete = widget.NewButtonWithIcon("", theme.Icon(theme.IconNameDelete), func () {
		m.p.DeleteAgent(agent)
		agent = nil
		temp := m.p.ListAgents()
		if len(temp) == 0 {
			agent = m.p.NewAgent()
			lastEditedAgent = 0
		} else {
			lastEditedAgent--
			if lastEditedAgent < 0 {
				lastEditedAgent = 0
			}
			agent = m.p.GetAgent(agents[lastEditedAgent].ID)
		}
		m.p.SetIntSetting("agent.last-edited", lastEditedAgent)
		updateUI()
	})
	btnNew = widget.NewButtonWithIcon("", theme.Icon(theme.IconNameFile), func() {
		save()
		agent = m.p.NewAgent()
		refreshAgentList()
		agentSelect.SetSelectedIndex(len(agentSelect.Options)-1)
		updateUI()
	})
	f.Append("Agent", container.NewBorder(nil, nil, nil,
			container.NewHBox(btnDelete, btnNew), agentSelect,
		),
	)
	// Agent name entry
	nameEntry = widget.NewEntry()
	nameEntry.OnChanged = func(s string) {
		agent.Name = s
	}
	f.Append("Name", nameEntry)
	// LLM select
	llmSelect = NewIndexedSelect(nil, func(idx int) {
		name := llms[idx]
		agent.LLM = m.p.GetLLM(name.ID)
		updateUI()
	})
	f.Append("LLM", llmSelect)
	// System prompt entry area
	sysEntry = widget.NewEntry()
	sysEntry.MultiLine = true
	sysEntry.SetMinRowsVisible(12)
	sysEntry.OnChanged = func(s string) {
		agent.System.Content = s
	}
	f.Append("System", sysEntry)
	// Load last edited agent
	agents = m.p.ListAgents()
	agent = m.p.GetAgent(agents[lastEditedAgent].ID)
	updateUI()
	// Complete and show dialog
	dlg := dialog.NewCustom("Agent Settings", "Done", f, m.w)
	dlg.SetOnClosed(func() {
		save()
		// TODO Fire global OnAgentClosed event
	})
	dlg.Resize(dlg.MinSize().AddWidthHeight(320, 0))
	dlg.Show()
}
