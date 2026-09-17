package download

import (
	"fmt"
	"net/http"
	"strings"

	cfg "gh2/pkg/config"
)

type Proxy func(url string) string

// 任意 asset 的下载链接都可以
const testURL = "https://github.com/cli/cli/releases/download/v2.50.0/gh_2.50.0_windows_arm64.zip"

// 镜像格式：
// https://mirror.com/xxx/%s
// 其中 %s 是去除 https://github.com/ 剩下的内容
// 例如 cli/cli/releases/download/v2.50.0/gh_2.50.0_windows_arm64.zip
func GetProxy(mirrors []string) (Proxy, error) {
	var proxies []Proxy
	for _, mirror := range mirrors {
		proxies = append(
			proxies,
			func(url string) string {
				tail := strings.TrimPrefix(url, "https://github.com/")
				return fmt.Sprintf(mirror, tail)
			})
	}
	return TestProxies(proxies)
}

// 获取第一个可用代理
func TestProxies(proxies []Proxy) (Proxy, error) {
	client := &http.Client{}

	logger := cfg.GetLogger()
	// 策略：首先使用代理
	for _, proxy := range proxies {
		url := proxy(testURL)
		resp, err := client.Head(url)
		if err != nil || resp.StatusCode != http.StatusOK {
			logger.Debug("test proxy", "url", url)
			continue
		}
		logger.Info("use proxy", "url", url)
		return proxy, nil
	}

	// 如果没有代理，或者代理都失效了，尝试使用默认链接
	resp, err := client.Head(testURL)
	if err == nil && resp.StatusCode == http.StatusOK {
		logger.Info("default url accessible")

		// 返回 nil 表示不使用代理
		return nil, nil
	}

	logger.Error("no resource usable")
	return nil, fmt.Errorf("no resource usable")
}
