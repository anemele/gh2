package cmd

import (
	"fmt"
	"gh2/core"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
)

var cloneCommand = &cli.Command{
	Name:      "clone",
	Aliases:   []string{"cl"},
	Usage:     "Clone repository from GitHub",
	UsageText: "gh2 clone [repo] ...",
	Args:      true,
	Action:    cloneAction,
}

func cloneRepo(url string, config *core.CloneConfig) error {
	repo := core.ParseRepo(url)
	if repo == nil {
		return fmt.Errorf("invalid url: %s", url)
	}

	repoUrl := fmt.Sprintf("%s%s.git", config.MirrorUrl, repo.String())
	destDir := filepath.Join(config.OutputDir, repo.String())

	args := []string{
		"clone", repoUrl, destDir,
	}
	args = append(args, config.GitConfig...)

	cmd := exec.Command("git", args...)

	fmt.Printf("Running: %s\n", strings.Join(args, " "))
	output, err := cmd.CombinedOutput()
	fmt.Println(string(output))

	return err
}

func cloneAction(c *cli.Context) error {
	baseConfig, err := core.LoadConfig()

	if err != nil {
		return err
	}
	config := baseConfig.Clone

	if c.NArg() == 0 {
		return fmt.Errorf("required at least one repo")
	}

	for _, arg := range c.Args().Slice() {
		err = cloneRepo(arg, &config)
		if err != nil {
			fmt.Println(err)
		}
	}

	return nil
}
