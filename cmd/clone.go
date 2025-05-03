package cmd

import "github.com/urfave/cli/v2"

var cloneCommand = &cli.Command{
	Name:      "clone",
	Aliases:   []string{"cl", "c"},
	Usage:     "Clone repository from GitHub",
	UsageText: "gh2 clone [repo] ...",
	Args:      true,
	Action:    cloneAction,
}

func cloneAction(c *cli.Context) error {

	return nil
}
