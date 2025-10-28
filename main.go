package main

import (
	"log"

	"github.com/qbradq/gen-magic/ui"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	m := ui.NewMain()
	m.Run()
}
