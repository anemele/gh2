package core

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	logger.Info("get releases with base api", "repo", repo.String())
	if res, err := getReleases_base(repo); err == nil {
		return res, nil
	}
	logger.Info("base api failed, use gh cli", "repo", repo.String())
	return getReleases_gh_api(repo)
}

const workerNumber = 10

func DownloadAssets(assets []Asset, dir string, proxy Proxy) error {
	logger.Info("download assets", "dir", dir, "length of assets", len(assets))
	if proxy != nil {
		logger.Info("use proxy")
	} else {
		logger.Info("no proxy")
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, workerNumber)

	pb := mpb.New(mpb.WithWaitGroup(&wg))
	wg.Add(len(assets))

	taskBar := pb.AddBar(
		int64(len(assets)),
		mpb.BarWidth(20),
		mpb.PrependDecorators(
			decor.Name("Tasks"),
		),
		mpb.AppendDecorators(
			decor.CountersNoUnit("%d/%d", decor.WCSyncWidth),
		),
	)

	fails := make(chan string, len(assets))

	for _, asset := range assets {
		sem <- struct{}{}

		bar := pb.AddBar(
			asset.Size,
			mpb.BarWidth(20),
			mpb.PrependDecorators(
				decor.Elapsed(decor.ET_STYLE_MMSS),
			),
			mpb.AppendDecorators(
				decor.CountersKibiByte("% .2f / % .2f", decor.WC{W: 25}),
				decor.Name(fmt.Sprintf(" | %s", asset.Name)),
			),
			mpb.BarRemoveOnComplete(),
		)

		go func(asset Asset, bar *mpb.Bar) {
			defer func() {
				taskBar.Increment()
				<-sem
				wg.Done()
			}()
			err := DownloadAsset(asset, dir, proxy, bar)
			if err != nil {
				fails <- asset.DownloadUrl
				logger.Error(err.Error())
				bar.Abort(true)
			} else {
				logger.Info("done", "name", asset.Name)
			}
		}(asset, bar)
	}

	pb.Wait()

	close(fails)
	var failedTasks []string
	for fail := range fails {
		failedTasks = append(failedTasks, fail)
	}
	if len(failedTasks) == 0 {
		fmt.Println(" all tasks completed")
	} else {
		fmt.Println(" failed tasks:\n ", strings.Join(failedTasks, "\n  "))
	}

	return nil
}

// ref: https://zhuanlan.zhihu.com/p/40819486
func handleNetError(err error) error {
	netErr, ok := err.(net.Error)
	if !ok {
		return nil
	}

	if netErr.Timeout() {
		return nil
	}

	opErr, ok := netErr.(*net.OpError)
	if !ok {
		return nil
	}

	switch opErr.Err.(type) {
	case *net.DNSError:
		return err
	case *os.SyscallError:
		return err
	default:
		return nil
	}
}

const oneMegaByte = 1 << 20

func DownloadAsset(asset Asset, dir string, proxy Proxy, bar *mpb.Bar) error {
	logger.Info("download asset", "name", asset.Name, "size", asset.Size)

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
	if size <= oneMegaByte*5 {
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

	// 假设带宽可达 2MB/s ，设置基础块大小为 2MB 。
	// 如果文件大于 5MB，需要分段请求，首次请求的块大小为 2MB 。
	// 依据请求时间二分法动态调整块大小，找到合适并以此大小进行后续请求。
	var chunkSize int64 = oneMegaByte * 2

	// 上次请求的块增量
	var lastAccChunkSize int64 = chunkSize
	// 上次请求的带宽，初始 2MB/s
	var lastSpeed float64 = oneMegaByte * 2.0
	var speed float64

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

		// 记录请求到返回的时间
		beginTime := time.Now()
		resp, err := client.Do(req)
		costTime := time.Since(beginTime)

		if err != nil {
			if handleNetError(err) != nil {
				return err
			}
			continue
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

		// 调整块大小
		speed = float64(realChunkSize) / costTime.Seconds()
		if speed > lastSpeed {
			chunkSize += lastAccChunkSize
		} else {
			chunkSize -= lastAccChunkSize
			chunkSize = max(chunkSize, oneMegaByte)
		}
		lastSpeed = speed
	}

	return nil
}
