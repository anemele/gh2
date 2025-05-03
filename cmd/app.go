package cmd

import (
	"os"

	"github.com/urfave/cli/v2"
)

var app = cli.NewApp()

func init() {
	app.Name = "gh2"
	app.Usage = "Complement for GitHub CLI."
	app.Commands = []*cli.Command{
		cloneCommand,
		downloadCommand,
		configCommand,
	}
}

func Run() error {
	return app.Run(os.Args)
}
