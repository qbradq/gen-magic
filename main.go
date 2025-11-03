package main

import (
	"io"
	"log/slog"
	"os"

	"github.com/qbradq/gen-magic/ui"
)

func main() {
	lf, err := os.Create("gen-magic.log")
	if err != nil {
		panic(err)
	}
	opts := &slog.HandlerOptions{
		AddSource: true,
		Level: slog.LevelInfo,
	}
	if os.Getenv("DEBUG") != "" {
		opts.Level = slog.LevelDebug
	}
	logger := slog.NewTextHandler(
		io.MultiWriter(
			os.Stderr,
			lf,
		),
		opts,
	)
	slog.SetDefault(slog.New(logger))
	m := ui.NewMain()
	m.Run()
}
