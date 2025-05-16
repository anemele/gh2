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

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

var client = http.Client{
	Timeout: 30 * time.Second,
}

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

const workerNumber = 10

func DownloadAssets(assets []Asset, dir string, proxy Proxy) error {
	var wg sync.WaitGroup
	type empty struct{}
	var sem = make(chan empty, workerNumber)

	pb := mpb.New(mpb.WithWaitGroup(&wg))
	wg.Add(len(assets))

	for _, asset := range assets {
		sem <- empty{}

		bar := pb.AddBar(int64(asset.Size),
			mpb.BarWidth(20),
			mpb.PrependDecorators(
				decor.OnComplete(
					decor.Spinner([]string{}),
					"✔",
				),
				decor.Name(" ["),
				decor.Elapsed(decor.ET_STYLE_MMSS),
				decor.Name("]"),
			),
			mpb.AppendDecorators(
				decor.CountersKibiByte("% .2f / % .2f", decor.WC{W: 25}),
				decor.Name(fmt.Sprintf(" | %s", asset.Name)),
			),
		)

		go func(asset Asset, sem chan empty) {
			defer func() {
				<-sem
				wg.Done()
			}()
			DownloadAsset(asset, dir, proxy, bar)
		}(asset, sem)
	}

	pb.Wait()

	return nil
}

const chunkSize = 1024 * 1024 * 2

func DownloadAsset(asset Asset, dir string, proxy Proxy, bar *mpb.Bar) error {
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

		bar.IncrInt64(end - start + 1)

		start = end + 1
	}

	return nil
}
