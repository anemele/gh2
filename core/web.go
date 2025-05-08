package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/briandowns/spinner"
)

var client http.Client

// GitHub REST API 请求频繁会被限流，但是携带身份请求可以提高请求频率
// 这里加一个方法，使用 GitHub CLI 的 api 命令发送请求，获取 releases
func getReleases_gh_api(repo Repo) ([]Release, error) {
	url := fmt.Sprintf("repos/%s/releases", repo.String())
	cmd := exec.Command("gh", "api", url)
	resp, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var releases []Release
	err = json.Unmarshal(resp, &releases)
	if err != nil {
		return nil, err
	}
	return releases, nil
}

func getReleases_base(repo Repo) ([]Release, error) {
	url := getReleasesUrl(repo)
	resp, err := client.Get(url)

	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get releases: %s", resp.Status)
	}

	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)

	var releases []Release
	err = decoder.Decode(&releases)
	if err != nil {
		return nil, err
	}

	return releases, nil
}

func (repo Repo) GetReleases() ([]Release, error) {
	if res, err := getReleases_base(repo); err == nil {
		return res, nil
	}
	return getReleases_gh_api(repo)
}

const workers = 10

func DownloadAssets(assets []Asset, dir string, proxy Proxy) error {
	var wg sync.WaitGroup
	type empty struct{}
	var sem = make(chan empty, workers)
	var done = make(chan bool)

	for _, asset := range assets {
		wg.Add(1)
		sem <- empty{}
		go func(asset Asset, sem chan empty) {
			defer func() {
				<-sem
				wg.Done()
			}()
			err := DownloadAsset(asset, dir, proxy)
			done <- err == nil
		}(asset, sem)
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	sp := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	sp.Start()
	defer sp.Stop()

	succ := 0
	fail := 0
	sp.Prefix = fmt.Sprintf("Downloading %d assets ", len(assets))

	for ok := range done {
		if ok {
			succ++
		} else {
			fail++
		}
		// \x1b[2K 控制字符：清除当前行
		sp.Suffix = fmt.Sprintf(" Success: %d | Failure: %d", succ, fail)
	}
	sp.FinalMSG = fmt.Sprintf("All: %d | Success: %d | Failure: %d\n", len(assets), succ, fail)

	return nil
}

const chunkSize = 1024 * 1024 * 2

func DownloadAsset(asset Asset, dir string, proxy Proxy) error {
	file, err := os.OpenFile(filepath.Join(dir, asset.Name), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	url := asset.DownloadUrl
	if proxy != nil {
		url = proxy(url)
	}

	var start int64 = 0
	for start < asset.Size {
		end := start + chunkSize
		if end >= asset.Size {
			end = asset.Size - 1
		}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusPartialContent {
			return fmt.Errorf("failed to download asset: %s", resp.Status)
		}
		defer resp.Body.Close()

		n, err := file.ReadFrom(resp.Body)
		if err != nil {
			return err
		}
		if n != int64(end-start+1) {
			return fmt.Errorf("expected %d, got %d", end-start+1, n)
		}

		start = end + 1
	}

	return nil
}
