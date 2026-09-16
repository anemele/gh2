package main

import (
	"log/slog"

	"gh2/core"

	"github.com/alecthomas/kong"
)

var CLI struct {
	Debug    bool        `help:"Enable debug mode." hidden:""`
	Clone    CloneCmd    `cmd:"" aliases:"cl" help:"Clone repository from GitHub"`
	Download DownloadCmd `cmd:"" aliases:"dl" help:"Download releases from GitHub."`
	Config   ConfigCmd   `cmd:"" help:"Configure gh2."`
}

func main() {
	ctx := kong.Parse(&CLI, kong.UsageOnError())

	if CLI.Debug {
		core.InitLogger(slog.LevelDebug)
	} else {
		core.InitLogger(slog.LevelInfo)
	}

	err := ctx.Run()
	ctx.FatalIfErrorf(err)

	core.GetLogger().Info("exit program")
}
