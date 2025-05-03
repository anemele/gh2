package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"sync"
	"time"

	"github.com/briandowns/spinner"
)

var client http.Client

func (repo Repo) GetReleases() ([]Release, error) {
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

// 应该并发下载，但是没有合适的 multi progressbar 库，暂且单一下载
func DownloadAssets(assets []Asset, dir string, proxies []Proxy) error {
	var wg sync.WaitGroup
	wg.Add(len(assets))
	var done = make(chan bool)

	for _, asset := range assets {
		go func(asset Asset) {
			defer wg.Done()
			err := DownloadAsset(asset, dir, proxies)
			done <- err == nil
		}(asset)
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

func DownloadAsset(asset Asset, dir string, proxies []Proxy) error {
	filepath := path.Join(dir, asset.Name)
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	url := asset.DownloadUrl
	resp, err := client.Head(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		for _, proxy := range proxies {
			url = proxy(url)
			resp, err = client.Head(url)
			if err != nil || resp.StatusCode != http.StatusOK {
				continue
			}
			break
		}
	}

	start := 0
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
