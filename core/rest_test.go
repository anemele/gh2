package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func f(url string) string {
	repo := ParseRepo(url)
	if repo == nil {
		return ""
	}
	return repo.String()
}

func TestParseRepo_1(t *testing.T) {
	samples := []string{
		"x/y",
		"x/y.git",
		"/x/y",
		"/x/y.git",
		"https://github.com/x/y",
		"https://github.com/x/y.git",
		"git@github.com:x/y",
		"git@github.com:x/y.git",
		"https://github.mirror/x/y",
		"https://github.mirror/x/y.git",
		"https://a.b.c/x/y",
		"https://a.b.c/x/y.git",
	}
	for _, sample := range samples {
		assert.Equal(t, "x/y", f(sample))
	}
}

func TestParseRepo_2(t *testing.T) {
	samples := []struct {
		Url  string
		Repo string
	}{
		{"a-b/c_d", "a-b/c_d"},
		{"a-b/c_d.git", "a-b/c_d"},
		{"a-b/c_d.git.git", "a-b/c_d.git"},
		{"a-b/c_d.github.io", "a-b/c_d.github.io"},
		{"a-b/c_d.github.io.git", "a-b/c_d.github.io"},
		{"a-b/c_d.github.io.git.git", "a-b/c_d.github.io.git"},
	}
	for _, sample := range samples {
		assert.Equal(t, sample.Repo, f(sample.Url))
	}
}

func TestParseRepo_3(t *testing.T) {
	samples := []string{
		"xy",
		"https://github.com/xy",
	}
	for _, sample := range samples {
		assert.Equal(t, "", f(sample))
	}
}

func TestParseRepo_4(t *testing.T) {
	samples := []struct {
		Url  string
		Repo string
	}{
		{"a/b/c/x/y", "a/b"},
		{"a/b.git/c/x/y", "a/b.git"},
		{"a/b/c/x/y.git", "a/b"},
		{"a/b.git/c/x/y.git", "a/b.git"},
		{"a/b.git/c.git/x.git/y.git", "a/b.git"},
		{"/a/b.git/c.git/x.git/y.git", "a/b.git"},
		{"https://github.com/x/y/issues", "x/y"},
		{"https://github.com/x/y.git/issues", "x/y.git"},
		{"https://github.com/x/y/issues.git", "x/y"},
		{"https://github.com/x/y.git/issues.git", "x/y.git"},
		{"https://github.com/x/y/releases/tag/v1.0", "x/y"},
		{"https://github.com/x/y.git/releases/tag/v1.0", "x/y.git"},
		{"https://github.com/x/y/releases/tag/v1.0.git", "x/y"},
	}
	for _, sample := range samples {
		assert.Equal(t, sample.Repo, f(sample.Url))
	}
}

func TestParseRepo_5(t *testing.T) {
	samples := []string{
		"https://github.com/[]/{}",
		"git@github.com:[]/{}",
		"git@github.com:/[]/{}",
	}
	for _, sample := range samples {
		assert.Equal(t, "", f(sample))
	}
}

func TestParseRepo_6(t *testing.T) {
	samples := []string{
		"https:///x/y",
		"git@:/x/y",
	}
	for _, sample := range samples {
		assert.Equal(t, "", f(sample))
	}
}
