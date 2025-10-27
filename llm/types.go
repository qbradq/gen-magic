package llm

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	"image/png"
	"strings"
)

// Definition contains all of the data needed to define and communicate with a
// language model.
type Definition struct {
	API string `json:"api"` // LLM API in use
	APIEndpoint string `json:"api_endpoint"` // Network endpoint for the API, if any
	APIKey string `json:"api_key"` // Network endpoint API key, if any
	Model string `json:"model"` // Model to use
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
	Role string `json:"role"`
	Content string `json:"content"`
	Images []*Image `json:"images"`
	Delta bool `json:"delta"`
}

// Turn holds the data of a complete turn of LLM exchanges.
type Turn struct {
	Definition Definition
	System *Message
	Prompt *Message
	Response []*Message
}
