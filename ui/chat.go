package ui

import (
	"errors"
	"log"
	"strconv"
	"unicode"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/qbradq/gen-magic/llm"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Max number of messages in history.
const maxHistory int = 100

// Chat implements the Chat chat interface.
type Chat struct {
	w fyne.Window
	m *Main
	def llm.Definition
	root *fyne.Container
	scroll *container.Scroll
	chat *fyne.Container
	prompt *PromptEntry
	progress *widget.ProgressBarInfinite
	submit *widget.Button
	stop *widget.Button
	ctxLengthEntry *widget.Entry
	llms []LLMName
	llmSelect *IndexedSelect
	history []*llm.Turn
	cancelCompletion func()
}

// NewChat returns a new Chat UI.
func NewChat(m *Main, onClose func()) *Chat {
	ret := &Chat{
		w: fyne.CurrentApp().NewWindow("Chat"),
		m: m,
	}
	ret.w.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName: fyne.KeyEnter,
		Modifier: fyne.KeyModifierControl,
	}, nil)
	ret.w.SetOnClosed(func() {
		if onClose != nil {
			onClose()
		}
		ret.Close()
	})
	ret.chat = container.NewVBox()
	ret.scroll = container.NewVScroll(ret.chat)
	ret.scroll.SetMinSize(fyne.NewSize(480, 320))
	ret.prompt = NewPromptEntry()
	ret.prompt.MultiLine = true;
	ret.prompt.SetMinRowsVisible(5)
	ret.prompt.SetPlaceHolder("LLM Chat Prompt")
	ret.prompt.OnCtrlEnter = ret.Submit
	ret.progress = widget.NewProgressBarInfinite()
	ret.progress.Hide()
	ret.submit = widget.NewButtonWithIcon("", theme.Icon(theme.IconNameMediaPlay), ret.Submit)
	ret.ctxLengthEntry = widget.NewEntry()
	ret.ctxLengthEntry.OnChanged = func(s string) {
		if s == "" {
			ret.ctxLengthEntry.SetText("0")
			s = "0"
		}
		v, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return
		}
		if v < 0 {
			v = 0
		}
		if int(v) > maxHistory {
			v = int64(maxHistory)
		}
		ret.ctxLengthEntry.SetText(strconv.FormatInt(v, 10))
	}
	ret.ctxLengthEntry.Validator = func(s string) error {
		for _, r := range s {
			if !unicode.IsDigit(r) {
				return errors.New("only numbers are allowed")
			}
		}
		return nil
	}
	ret.ctxLengthEntry.SetText("5")
	ret.stop = widget.NewButtonWithIcon("", theme.Icon(theme.IconNameMediaStop), func() {
		if ret.cancelCompletion != nil {
			ret.cancelCompletion()
		}
	})
	ret.stop.Disable()
	ret.llmSelect = NewIndexedSelect(nil, func(idx int) {
		llm := ret.llms[idx]
		ret.def = *ret.m.p.GetLLM(llm.ID)
	})
	ret.OnLLMsUpdated()
	ret.root = container.NewPadded(
		container.NewBorder(
			nil,
			container.NewVBox(
				container.NewStack(
					ret.prompt,
					container.NewCenter(
						container.NewGridWrap(
							fyne.NewSize(
								240,
								ret.progress.MinSize().Height,
							),
							ret.progress,
						),
					),
				),
				container.NewBorder(nil, nil, nil,
					container.NewHBox(
						widget.NewLabel("CTX"),
						container.New(
							layout.NewGridWrapLayout(fyne.NewSize(
								80,
								ret.ctxLengthEntry.MinSize().Height,
							)),
							ret.ctxLengthEntry,
						),
						ret.stop,
						ret.submit,
					),
					ret.llmSelect,
				),
			),
			nil,
			nil,
			ret.scroll,
		),
	)
	ret.w.SetContent(ret.root)
	ret.w.RequestFocus()
	ret.Focus()
	ret.w.Show()
	ret.m.AddChild(ret)
	return ret
}

// Close closes the window.
func (l *Chat) Close() {
	l.w.Close()
	l.m.RemoveChild(l)
}

// Submit submits the current prompt if able.
func (l *Chat) Submit() {
	promptText := l.prompt.Text
	if promptText == "" {
		return
	}
	v, err := strconv.ParseInt(l.ctxLengthEntry.Text, 10, 32)
	if err != nil {
		log.Printf("error while parsing chat context length: %v", err)
		v = 0
		fyne.Do(func() {
			l.ctxLengthEntry.SetText("0")
		})
	}
	ctxLength := int(v)
	promptBubble := l.LogPrompt()
	l.prompt.SetText("")
	lh := len(l.history)
	ctxBegin := lh-(ctxLength+1)
	if ctxBegin < 0 {
		ctxBegin = 0
	}
	ctx := l.history[ctxBegin:]
	turn := &llm.Turn{
		Definition: l.def,
		System: &llm.Message{
			Role: "system",
			Content: "You are a helpful AI assistant.",
		},
		Prompt: &llm.Message{
			Role: "user",
			Content: promptText,
		},
	}
	msgs, cancel, err := llm.ChatCompletion(&turn.Definition, turn.System, turn.Prompt, ctx)
	if err != nil {
		l.chat.Remove(promptBubble)
		dialog.ShowInformation(
			"Completion Error",
			err.Error(),
			l.w,
		)
		return
	}
	go func() {
		fyne.Do(func() {
			l.cancelCompletion = cancel
			l.stop.Enable()
			l.submit.Disable()
			l.prompt.Disable()
			l.progress.Show()
		})
		var bubble *ChatBubble
		for msg := range msgs {
			if !msg.Delta || bubble == nil {
				turn.Response = append(turn.Response, msg)
				bubble = l.LogResponse(msg)
			} else {
				turn.Response[len(turn.Response)-1].Content += msg.Content;
				bubble.AppendText(msg.Content)
				l.scroll.ScrollToBottom()
			}
		}
		fyne.Do(func() {
			l.cancelCompletion = nil
			l.stop.Disable()
			l.submit.Enable()
			l.prompt.Enable()
			l.progress.Hide()
		})
	}()
	l.history = append(l.history, turn)
	if lh > maxHistory {
		l.history = l.history[lh - maxHistory:]
	}
}

// LogPrompt adds the prompt to the chat log.
func (l *Chat) LogPrompt() *ChatBubble {
	s := l.prompt.Text
	bubble := NewChatBubble(
		cases.Title(language.AmericanEnglish).String("user"),
		s,
		theme.Color(theme.ColorNameInputBackground),
		true,
	)
	l.chat.Add(bubble)
	l.scroll.ScrollToBottom()
	return bubble
}

// LogResponse adds the response message to the chat log.
func (l *Chat) LogResponse(msg *llm.Message) *ChatBubble {
	bubble := NewChatBubble(
		cases.Title(language.AmericanEnglish).String(msg.Role),
		msg.Content,
		theme.Color(theme.ColorNameBackground),
		false,
	)
	l.chat.Add(bubble)
	l.scroll.ScrollToBottom()
	return bubble
}

// Root returns the root container.
func (l *Chat) Root() *fyne.Container {
	return l.root
}

// Focus focuses the prompt entry.
func (l *Chat) Focus() {
	l.w.Canvas().Focus(l.prompt)
}

// OnLLMsUpdated is called when the LLM list is updated.
func (l *Chat) OnLLMsUpdated() {
	s := l.llmSelect.Selected
	list := []string{}
	l.llms = l.m.p.ListLLMs()
	for _, llm := range l.llms {
		list = append(list, llm.Name)
	}
	l.llmSelect.SetOptions(list)
	if s != "" {
		l.llmSelect.SetSelected(s)
	} else {
		l.llmSelect.SetSelectedIndex(0)
	}
}
