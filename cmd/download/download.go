package download

import (
	"fmt"
	"os"
	"sync"

	cfg "gh2/pkg/config"
	"gh2/pkg/rest"
)

type DownloadCmd struct {
	Repo []string `arg:"" optional:""`
}

func (c DownloadCmd) Run() error {
	baseConfig, err := cfg.LoadConfig()
	if err != nil {
		return err
	}

	config := baseConfig.Download
	urls := c.Repo

	logger := cfg.GetLogger()
	logger.Debug(
		"downloadCommand",
		"urls", urls)

	// 检查 output 目录是否存在，如果不存在则创建
	if _, err := os.Stat(config.OutputDir); os.IsNotExist(err) {
		err = os.Mkdir(config.OutputDir, 0o755)
		if err != nil {
			return err
		}
	}

	// 如果输入为空则从缓存加载。
	// （缓存就是以前输入过的仓库）
	// 无论是现输入还是从缓存加载，都是不可信任的，需要后续解析。
	if len(urls) == 0 {
		tmp, err := cfg.LoadRepos()
		if err != nil {
			return err
		}

		// 如果仍然为空，则退出
		if len(tmp) == 0 {
			return fmt.Errorf("no input")
		}

		// 否则 survey 交互
		urls, err = SurveyCache(tmp)
		if err != nil {
			return err
		}
	}

	logger.Debug("downloadCommand", "urls", urls)

	// 此时 urls 不为空
	// 解析 urls 获取 repos
	var repos []rest.Repo
	for _, url := range urls {
		repo, err := rest.ParseRepo(url)
		if err == nil {
			repos = append(repos, repo)
		}
	}

	logger.Debug("downloadCommand", "repos", repos)

	// 如果没有 repo 则退出
	if len(repos) == 0 {
		return fmt.Errorf("no repos found")
	}

	var wg sync.WaitGroup

	type Pair struct {
		err      error
		repo     rest.Repo
		releases []rest.Release
	}
	// 首先获取所有仓库的 releases
	pairChan := make(chan Pair, len(repos))
	wg.Add(len(repos))
	for _, repo := range repos {
		go func(repo rest.Repo) {
			defer wg.Done()
			releases, err := GetReleases(repo)
			pairChan <- Pair{err, repo, releases}
		}(repo)
	}

	// 参考豆包AI，将 Wait 放在一个 goroutine 里面
	// go func() {
	wg.Wait()
	close(pairChan)
	// }()

	// 清空 repos 数组，因为可能部分错误，或者404等
	repos = []rest.Repo{}
	// 交互选取 assets
	var allAssets []rest.Asset
	for pair := range pairChan {
		if pair.err != nil {
			fmt.Printf("Error on %s: %s\n", pair.repo.String(), pair.err)
			continue
		}
		repos = append(repos, pair.repo)
		assets, err := SurveyReleases(pair.repo, pair.releases)
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
	proxy, err := GetProxy(config.Mirrors)
	if err != nil {
		return err
	}

	cfg.GetLogger().Debug("downloadCommand", "proxy", proxy)

	// 下载 assets
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := DownloadAssets(allAssets, config.OutputDir, proxy)
		if err != nil {
			return
		}
	}()
	wg.Wait()

	cache, err := cfg.UpdateRepos(repos)
	if err != nil {
		return err
	}
	err = cfg.SaveRepos(cache)

	return err
}
