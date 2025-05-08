package core

import (
	"fmt"
	"regexp"
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

func ParseRepo(s string) (Repo, error) {
	pattern1 := regexp.MustCompile(`^(?:https://[\w\.\-]+/)|(?:git@[\w\.\-]+:)`)

	i := pattern1.FindStringIndex(s)
	if i != nil {
		s = s[i[1]:]
	}

	s = strings.TrimPrefix(s, "/")

	pattern2 := regexp.MustCompile(`^([\w-]+)/([\w\.-]+)`)

	i = pattern2.FindStringIndex(s)
	if i == nil {
		return Repo{}, fmt.Errorf("invalid repo: %s", s)
	}

	matches := pattern2.FindStringSubmatch(s)
	owner := matches[1]
	name := matches[2]
	grp0 := matches[0]
	if len(grp0) == len(s) || (len(s) > len(grp0) && s[len(grp0)] != '/') {
		name = strings.TrimSuffix(name, ".git")
	}

	return Repo{
		Owner: owner,
		Name:  name,
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
	Size        int64     `json:"size"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DownloadUrl string    `json:"browser_download_url"`
}

func humanSize(size int64) string {
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
