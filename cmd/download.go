package cmd

import (
	"fmt"
	"gh2/core"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/urfave/cli/v2"
)

var downloadCommand = &cli.Command{
	Name:      "download",
	Aliases:   []string{"dl", "d"},
	Usage:     "Download releases from GitHub",
	UsageText: "gh2 download [repo] ...",
	Args:      true,
	Action:    downloadAction,
}

func downloadAction(c *cli.Context) error {
	baseConfig, err := core.LoadConfig()

	if err != nil {
		return err
	}
	config := baseConfig.Download

	// 检查 output 目录是否存在，如果不存在则创建
	if _, err := os.Stat(config.OutputDir); os.IsNotExist(err) {
		err = os.Mkdir(config.OutputDir, 0755)
		if err != nil {
			return err
		}
	}

	// 首先获取输入数组，如果输入为空则从缓存加载。
	// （缓存就是以前输入过的仓库）
	// 无论是现输入还是从缓存加载，都是不可信任的，需要后续解析。
	args := c.Args().Slice()
	if len(args) == 0 {
		args, err = core.LoadRepos(config.OutputDir)
		if err != nil {
			return err
		}
		if len(args) == 0 {
			return fmt.Errorf("no input")
		}
		args, err = core.SurveyCache(args)
		if err != nil {
			return err
		}
	}

	// 解析输入数组，生成仓库数组
	var repos []*core.Repo
	for _, arg := range args {
		repo := core.ParseRepo(arg)
		if repo != nil {
			repos = append(repos, repo)
		}
	}
	// 如果没有仓库，则退出
	if len(repos) == 0 {
		return fmt.Errorf("no repos found")
	}

	var wg sync.WaitGroup

	type Pair struct {
		repo     *core.Repo
		releases []core.Release
	}
	// 首先获取所有仓库的 releases
	pairChan := make(chan Pair, len(repos))
	wg.Add(len(repos))
	for _, repo := range repos {
		go func(repo *core.Repo) {
			defer wg.Done()
			releases, err := repo.GetReleases()
			if err != nil {
				return
			}
			pairChan <- Pair{repo, releases}
		}(repo)
	}

	// 参考豆包AI，将 Wait 放在一个 goroutine 里面
	// go func() {
	wg.Wait()
	close(pairChan)
	// }()

	// 清空 repos 数组，因为可能部分错误，或者404等
	repos = []*core.Repo{}
	// 交互选取 assets
	var allAssets []core.Asset
	for pair := range pairChan {
		repos = append(repos, pair.repo)
		assets, err := core.SurveyReleases(pair.repo, pair.releases)
		if err != nil {
			continue
		}
		allAssets = append(allAssets, assets...)
	}

	// 如果没有 assets，则退出
	if len(allAssets) == 0 {
		return fmt.Errorf("no assets to download")
	}

	// 获取代理列表
	proxies := core.GetProxies(config.Mirrors)
	proxy := core.TestProxies(proxies)

	// 下载 assets
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = core.DownloadAssets(allAssets, config.OutputDir, proxy)
		if err != nil {
			return
		}
	}()
	wg.Wait()

	// 更新缓存
	cache, err := core.LoadRepos(config.OutputDir)
	// 这里一般不会返回 err ，如果返回 err 则直接退出
	if err != nil {
		return err
	}

	// 使用哈希表去重（用 set 更合适，但是没有标准库支持）
	hashtable := make(map[string]bool)
	for _, repo := range cache {
		hashtable[repo] = true
	}

	for _, repo := range repos {
		r := repo.String()
		if hashtable[r] {
			continue
		}
		cache = append(cache, r)
		hashtable[r] = true
	}

	// 按照字母表顺序排序，忽略大小写
	sort.Slice(cache, func(i, j int) bool {
		return strings.ToLower(cache[i]) < strings.ToLower(cache[j])
	})
	err = core.SaveRepos(config.OutputDir, cache)

	return err
}
