package llm

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	"image/png"
	"strings"
)

// LanguageModel contains all of the data needed to define and communicate with
// a language model.
type LanguageModel struct {
	ID int64
	Name string
	API string
	APIEndpoint string
	APIKey string
	Model string
}

// Image wraps an image.Image for the LLM.
type Image struct {
	img image.Image
	b64 string
}

// NewImage returns a new image with the given contents.
func NewImage(img image.Image) *Image {
	ret := &Image{}
	ret.SetImage(img)
	return ret
}

// NewImageBase64 returns a new image with the given contents.
func NewImageBase64(s string) *Image {
	ret := &Image{}
	ret.SetImageBase64(s)
	return ret
}

// MarshalJSON implements json.Marshaler.
func (l *Image) MarshalJSON() ([]byte, error) {
	return []byte("\"" + l.b64 + "\""), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (l *Image) UnmarshalJSON(data []byte) error {
	ld := len(data)
	if(ld < 2) {
		return errors.New("invalid string")
	}
	l.SetImageBase64(string(data[1:ld-1]))
	return nil
}

// SetImageBase64 sets the image data from a Base64-encoded string containing
// the data of a PNG.
func (l *Image) SetImageBase64(s string) error {
	if strings.Contains(s, ",") {
		parts := strings.Split(s, ",")
		s = parts[len(parts)-1]
	}
	imgData, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return err
	}
	img, err := png.Decode(bytes.NewReader(imgData))
	if err != nil {
		return err
	}
	l.b64 = s
	l.img = img
	return nil
}

// SetImage sets the image from a standard image.Image.
func (l *Image) SetImage(img image.Image) error {
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		return err
	}
	l.b64 = base64.StdEncoding.EncodeToString(buf.Bytes())
	l.img = img
	return nil
}

// Message holds the data of a single LLM message.
type Message struct {
	Role string
	Content string
	Images []*Image
	Delta bool
}

// Turn holds the data of a complete turn of LLM exchanges.
type Turn struct {
	Definition LanguageModel
	System *Message
	Prompt *Message
	Response []*Message
}
