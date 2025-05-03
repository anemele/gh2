package cmd

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

var configCommand = &cli.Command{
	Name:   "config",
	Usage:  "Manage configuration",
	Action: configAction,
}

func configAction(c *cli.Context) error {
	fmt.Println("Config command")
	return nil
}
