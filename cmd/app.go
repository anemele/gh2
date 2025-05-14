package cmd

import (
	"fmt"
	"gh2/core"
	"log/slog"

	"github.com/alecthomas/kong"
)

var CLI struct {
	Debug bool `help:"Enable debug mode." hidden:""`
	Clone struct {
		Repo []string `arg:"" required:""`
	} `cmd:"" aliases:"cl" help:"Clone repository from GitHub"`
	Download struct {
		Repo []string `arg:"" optional:""`
	} `cmd:"" aliases:"dl" help:"Download releases from GitHub."`
	Config struct {
	} `cmd:"" help:"Configure gh2."`
}

func Run() error {
	baseConfig, err := core.LoadConfig()
	if err != nil {
		return err
	}

	ctx := kong.Parse(&CLI, kong.UsageOnError())

	if CLI.Debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	switch ctx.Command() {

	case "clone <repo>":
		for _, repo := range CLI.Clone.Repo {
			err = cloneCommand(repo, baseConfig.Clone)
			if err != nil {
				break
			}
		}

	case "download":
		fallthrough
	case "download <repo>":
		err = downloadCommand(CLI.Download.Repo, baseConfig.Download)

	case "config":
		err = configCommand()

	default:
		return fmt.Errorf("unknown command: %s", ctx.Command())
	}

	return err
}
