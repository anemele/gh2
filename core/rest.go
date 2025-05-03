package core

import (
	"fmt"
	"strings"
	"time"
)

type Repo struct {
	Owner string
	Name  string
}

func (r Repo) String() string {
	return fmt.Sprintf("%s/%s", r.Owner, r.Name)
}

// 该解析函数很简单，只是按照斜线分割，后续应该优化
func ParseRepo(s string) (*Repo, error) {
	ss := strings.Split(s, "/")
	if len(ss) != 2 {
		return nil, fmt.Errorf("invalid repo: %s", s)
	}
	return &Repo{
		Owner: ss[0],
		Name:  ss[1],
	}, nil
}

func getReleasesUrl(repo Repo) string {
	return fmt.Sprintf("https://api.github.com/repos/%s/releases", repo.String())
}

type Release struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	CreatedAt   time.Time `json:"created_at"`
	PublishedAt time.Time `json:"published_at"`
	Assets      []Asset   `json:"assets"`
}

func (r Release) Title() string {
	// 这里要使用 PublishedAt 而不是 CreatedAt，因为 release 可以更新，
	// 例如 x64dbg/x64dbg 以 snapshot 发布，它的 created_at 不会变
	date := r.PublishedAt.Format("2006-01-02")
	return fmt.Sprintf("%s (%s)", r.TagName, date)
}

type Asset struct {
	Name        string    `json:"name"`
	Label       string    `json:"label"`
	ContentType string    `json:"content_type"`
	Size        int       `json:"size"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DownloadUrl string    `json:"browser_download_url"`
}

func humanSize(size int) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}
	sizef := float64(size)
	for _, unit := range []string{"KB", "MB", "GB"} {
		sizef /= 1024
		if sizef < 1024 {
			return fmt.Sprintf("%.2f %s", sizef, unit)
		}
	}
	return fmt.Sprintf("%.2f TB", sizef)
}

func (a Asset) Title() string {
	// 这里是否改用 updated_at 更好？
	date := a.CreatedAt.Format("2006-01-02")
	size := humanSize(a.Size)
	return fmt.Sprintf("%s (%s, %s)", a.Name, date, size)
}
