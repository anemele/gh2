package core

import (
	"fmt"
	"net/http"
	"strings"
)

type Proxy func(url string) string

// 代理链接一般是添加前缀，有的保留 github.com 部分，有的不保留，
// 这部分留给用户自定义，这里统一去除 https://github.com 前缀，
// 如果需要保留，则用户自己添加。
// 例如代理地址 https://a.b/https://github.com/
func GetProxies(hosts []string) []Proxy {
	var proxies []Proxy
	for _, host := range hosts {
		proxies = append(proxies, func(url string) string {
			tail := strings.TrimPrefix(url, "https://github.com")
			if strings.HasSuffix(host, "/") {
				tail = strings.TrimPrefix(tail, "/")
			}
			return host + tail
		})
	}
	return proxies
}

// 任意 asset 的下载链接都可以
const testUrl = "https://github.com/cli/cli/releases/download/v2.50.0/gh_2.50.0_windows_arm64.zip"

// 获取第一个可用代理
func TestProxies(proxies []Proxy) (Proxy, error) {
	client := &http.Client{}

	// 策略：首先使用代理
	for _, proxy := range proxies {
		url := proxy(testUrl)
		resp, err := client.Head(url)
		if err != nil || resp.StatusCode != http.StatusOK {
			logger.Debug("test proxy", "url", url)
			continue
		}
		logger.Info("use proxy", "url", url)
		return proxy, nil
	}

	// 如果没有代理，或者代理都失效了，尝试使用默认链接
	resp, err := client.Head(testUrl)
	if err == nil && resp.StatusCode == http.StatusOK {
		logger.Info("default url accessible")

		// 返回 nil 表示不使用代理
		return nil, nil
	}

	logger.Error("no resource usable")
	return nil, fmt.Errorf("no resource usable")
}
