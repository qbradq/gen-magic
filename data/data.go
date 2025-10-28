package data

import (
	"bytes"
	_ "embed"
	"image"

	_ "image/jpeg"
)

//go:embed system.md
var DefaultSystem string

//go:embed background.jpeg
var bgImgData []byte

//go:embed schema.sql
var SchemaSQL string

//go:embed static-data.sql
var StaticDataSQL string

var BackgroundImage image.Image

func init() {
	var err error
	BackgroundImage, _, err = image.Decode(bytes.NewReader(bgImgData))
	if err != nil {
		panic(err)
	}
}
