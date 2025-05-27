package core

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

var transport = &http.Transport{
	MaxIdleConns:        100,              // 最大空闲连接数
	MaxIdleConnsPerHost: 10,               // 每个主机的最大空闲连接数
	IdleConnTimeout:     30 * time.Second, // 空闲连接超时时间
	DialContext: (&net.Dialer{
		Timeout:   10 * time.Second, // 连接超时
		KeepAlive: 30 * time.Second, // 保持连接活跃时间
	}).DialContext,
	TLSHandshakeTimeout:   10 * time.Second, // TLS握手超时
	ResponseHeaderTimeout: 10 * time.Second, // 响应头超时
}

// 创建HTTP客户端，设置总超时和自定义传输层
var client = &http.Client{
	Timeout:   30 * time.Second, // 整个请求的总超时
	Transport: transport,
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
			err := DownloadAsset(asset, dir, proxy, bar)
			if err != nil {
				logger.Error(err.Error())
			}
		}(asset, sem)
	}

	pb.Wait()

	return nil
}

func DownloadAsset(asset Asset, dir string, proxy Proxy, bar *mpb.Bar) error {
	url := asset.DownloadUrl
	if proxy != nil {
		url = proxy(url)
	}

	// HEAD 检验文件大小是否相符
	resp, err := client.Head(url)
	if err != nil {
		return err
	}
	size := resp.ContentLength
	if size != asset.Size {
		return fmt.Errorf("mismatched file size: expected %d, got %d", asset.Size, size)
	}

	// 打开本地文件，准备接收字节流
	file, err := os.OpenFile(filepath.Join(dir, asset.Name), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// 应该优化下载算法，而非简单地分段下载。

	// 如果文件小于 5MB，可以直接全部请求，无需分段。
	if size <= 1024*1024*5 {
		resp, err = client.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		n, err := file.ReadFrom(resp.Body)
		if err != nil {
			return err
		}
		if n != size {
			return fmt.Errorf("expected %d, got %d", size, n)
		}
		bar.IncrInt64(size)
		return nil
	}

	// 假设带宽可达 2MBps ，设置基础块大小为 2MB 。
	// 如果文件大于 5MB，需要分段请求，首次请求的块大小为 2MB 。
	// 依据请求时间二分法动态调整块大小，找到合适并以此大小进行后续请求。
	var chunkSize int64 = 1024 * 1024 * 2

	var start int64 = 0
	for start < size {
		end := start + chunkSize
		if end >= size {
			end = size - 1
		}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return err
		}

		// Range header 格式为 bytes=start-end
		// 包含首尾字节
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
		realChunkSize := end - start + 1
		if n != realChunkSize {
			return fmt.Errorf("expected %d, got %d", realChunkSize, n)
		}

		bar.IncrInt64(realChunkSize)

		start = end + 1
	}

	return nil
}
