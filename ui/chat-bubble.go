package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ChatBubble implements an IM-style chat bubble.
type ChatBubble struct {
	widget.BaseWidget
	Role string
	Text string
	BubbleColor color.Color
	AlignRight bool
	c *fyne.Container
	text *widget.RichText
}

// NewChatBubble returns a new chat bubble with the given data.
func NewChatBubble(role, text string, bubbleColor color.Color, alignRight bool) *ChatBubble {
	ret := &ChatBubble{
		Role: role,
		Text: text,
		BubbleColor: bubbleColor,
		AlignRight: alignRight,
	}
	ret.ExtendBaseWidget(ret)
	bg := canvas.NewRectangle(ret.BubbleColor)
	bg.CornerRadius = 12
	bg.StrokeWidth = 2
	bg.StrokeColor = theme.Current().Color(theme.ColorNameForeground, theme.VariantDark)
	ret.text = widget.NewRichTextFromMarkdown(ret.Text)
	ret.text.Wrapping = fyne.TextWrapWord
	ret.text.Scroll = fyne.ScrollNone
	ret.c = container.NewStack(
		bg,
		container.NewVBox(
			container.NewHBox(
				widget.NewLabelWithStyle(
					ret.Role,
					fyne.TextAlignLeading,
					fyne.TextStyle{
						Bold: true,
					},
				),
				layout.NewSpacer(),
				container.NewPadded(
					widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
						fyne.CurrentApp().Clipboard().SetContent(ret.Text)
					}),
				),
			),
			ret.text,
		),
	)
	return ret
}

// CreateRenderer returns a new renderer for the widget.
func (w *ChatBubble) CreateRenderer() fyne.WidgetRenderer {
	var left, right float32
	if w.AlignRight {
		left = theme.Padding()*11
		right = theme.Padding()*4
	} else {
		left = theme.Padding()
		right = theme.Padding()*14
	}
	return widget.NewSimpleRenderer(container.New(
		layout.NewCustomPaddedLayout(
			theme.Padding(),
			theme.Padding(),
			left,
			right,
		),
		w.c,
	))
}

// AppendText appends the given text to the bubble's markdown.
func (w *ChatBubble) AppendText(text string) {
	w.Text += text
	fyne.Do(func() {
		w.text.ParseMarkdown(w.Text)
		w.Refresh()
	})
}
