package cmd

import (
	"fmt"
	"gh2/core"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func cloneCommand(url string, config core.CloneConfig) error {
	slog.Debug("cloneCommand", "url", url, "config", config)

	repo, err := core.ParseRepo(url)
	if err != nil {
		return err
	}

	slog.Debug("cloneCommand", "repo", repo)

	repoUrl := fmt.Sprintf("%s%s.git", config.MirrorUrl, repo.String())
	destDir := filepath.Join(config.OutputDir, repo.String())

	args := []string{
		"clone", repoUrl, destDir,
	}
	args = append(args, config.GitConfig...)

	slog.Debug("cloneCommand", "cmd", "git "+strings.Join(args, " "))

	cmd := exec.Command("git", args...)

	// 以下两行是正确打印 git clone 输出关键
	// 尝试过 stdout.Read bufio.Scanner io.MultiWriter 等不管用
	// 看到一篇文章讲 git clone 输出到 stderr 而非 stdout
	// https://deepinout.com/git/git-questions/1048_git_git_clone_writes_to_sderr_fine_but_why_cant_i_redirect_to_stdout.html
	// 虽然不懂，但摸索出来的下面两行代码实现了功能
	cmd.Stdout = io.Writer(os.Stdout)
	cmd.Stderr = io.Writer(os.Stderr)

	err = cmd.Run()

	return err
}
