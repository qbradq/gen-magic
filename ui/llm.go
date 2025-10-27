package ui

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"unicode"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/qbradq/gen-magic/llm"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Max number of messages in history.
const maxHistory int = 100

// LLM implements the LLM chat interface.
type LLM struct {
	id string
	def llm.Definition
	root *fyne.Container
	scroll *container.Scroll
	chat *fyne.Container
	prompt *PromptEntry
	progress *widget.ProgressBarInfinite
	submit *widget.Button
	stop *widget.Button
	ctxLengthEntry *widget.Entry
	history []*llm.Turn
	cancelCompletion func()
}

// NewLLM returns a new LLM UI.
func NewLLM(id string) *LLM {
	ret := &LLM{
		id: id,
	}
	ds := fyneApp.Preferences().String("llm." + id + ".def")
	if len(ds) > 0 {
		err := json.Unmarshal([]byte(ds), &ret.def)
		if err != nil {
			log.Printf("error loading llm def: %v", err)
		}
	}
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
	ret.ctxLengthEntry.SetText(
		strconv.Itoa(fyneApp.Preferences().Int("llm.context-length")),
	)
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
	ret.stop = widget.NewButtonWithIcon("", theme.Icon(theme.IconNameMediaStop), func() {
		if ret.cancelCompletion != nil {
			ret.cancelCompletion()
		}
	})
	ret.stop.Disable()
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
				container.NewHBox(
					layout.NewSpacer(),
					widget.NewLabel("Context Turns"),
					container.New(
						layout.NewGridWrapLayout(fyne.NewSize(
							80,
							ret.ctxLengthEntry.MinSize().Height,
						)),
						ret.ctxLengthEntry,
					),
					widget.NewButtonWithIcon(
						"Configure LLM",
						theme.SettingsIcon(),
						func() {
							ShowLLMSettings(&ret.def, func(def llm.Definition) {
								ret.def = def
								d, err := json.Marshal(def)
								if err != nil {
									log.Printf("error marshaling llm def: %v", err)
								}
								fyneApp.Preferences().SetString(
									"llm." + ret.id + ".def",
									string(d),
								)
							})
						},
					),
					ret.stop,
					ret.submit,
				),
			),
			nil,
			nil,
			ret.scroll,
		),
	)
	return ret
}

// Submit submits the current prompt if able.
func (l *LLM) Submit() {
	promptText := l.prompt.Text
	if promptText == "" {
		return
	}
	v, err := strconv.ParseInt(l.ctxLengthEntry.Text, 10, 32)
	if err != nil {
		log.Fatalf("Error while parsing llm.context-length: %v", err)
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
			window,
		)	
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
func (l *LLM) LogPrompt() *ChatBubble {
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
func (l *LLM) LogResponse(msg *llm.Message) *ChatBubble {
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
func (l *LLM) Root() *fyne.Container {
	return l.root
}

// Focus focuses the prompt entry.
func (l *LLM) Focus() {
	window.Canvas().Focus(l.prompt)
}
