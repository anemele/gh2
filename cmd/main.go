package main

import (
	"log/slog"

	ccl "gh2/cmd/clone"
	ccfg "gh2/cmd/config"
	cdl "gh2/cmd/download"
	"gh2/pkg/config"

	"github.com/alecthomas/kong"
)

var CLI struct {
	Debug    bool            `help:"Enable debug mode." hidden:""`
	Clone    ccl.CloneCmd    `cmd:"" aliases:"cl" help:"Clone repository from GitHub"`
	Download cdl.DownloadCmd `cmd:"" aliases:"dl" help:"Download releases from GitHub."`
	Config   ccfg.ConfigCmd  `cmd:"" help:"Configure gh2."`
}

func main() {
	ctx := kong.Parse(&CLI, kong.UsageOnError())

	if CLI.Debug {
		config.InitLogger(slog.LevelDebug)
	} else {
		config.InitLogger(slog.LevelInfo)
	}

	err := ctx.Run()
	ctx.FatalIfErrorf(err)

	config.GetLogger().Info("exit program")
}
