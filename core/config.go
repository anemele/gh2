package core

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
)

type CloneConfig struct {
	OutputDir string   `toml:"output_dir"`
	MirrorUrl string   `toml:"mirror_url"`
	GitConfig []string `toml:"git_config"`
}

type DownloadConfig struct {
	OutputDir string   `toml:"output_dir"`
	Mirrors   []string `toml:"mirrors"`
}

type Config struct {
	Clone    CloneConfig    `toml:"clone"`
	Download DownloadConfig `toml:"download"`
}

func DefaultConfig() Config {
	return Config{
		Clone: CloneConfig{
			OutputDir: ".",
			MirrorUrl: "https://github.com/",
			GitConfig: []string{},
		},
		Download: DownloadConfig{
			OutputDir: ".",
			Mirrors:   []string{},
		},
	}
}

func LoadConfig() (Config, error) {
	config := DefaultConfig()

	homeDir, _ := os.UserHomeDir()
	configFilePath := filepath.Join(homeDir, ".gh2rc")

	fp, err := os.Open(configFilePath)
	if err != nil {
		logger.Warn(
			"config file not found, creating default config",
			"path", configFilePath,
			"error", err,
		)
		// 如果打开文件出错，可能是不存在，则创建默认配置，并写入文件
		// 尝试写入文件，如果出错则不理会，直接返回默认配置
		fp, err = os.OpenFile(configFilePath, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return config, nil
		}
		defer fp.Close()
		// 以 TOML 格式写入文件
		fmt.Println("writing default config to", configFilePath, "please edit it")
		err = toml.NewEncoder(fp).Encode(config)
		// 如果出错则打印错误，不影响后续
		if err != nil {
			fmt.Println("failed to write default config, please check it")
			logger.Error(err.Error())
		}
		return config, nil
	}
	defer fp.Close()

	decoder := toml.NewDecoder(fp)
	_, err = decoder.Decode(&config)
	if err != nil {
		fmt.Println("failed to parse config file, please check it")
		logger.Error(err.Error())
		return DefaultConfig(), err
	}

	return config, nil
}

const repoCacheFileName = "gh-repos"

// repo cache file format:
// owner1/name1
// owner2/name2
// ...
func LoadRepos(dir string) ([]string, error) {
	filename := filepath.Join(dir, repoCacheFileName)

	// not exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		logger.Error(
			"not found repo cache file",
			"path", filename,
		)
		return nil, nil
	}

	fp, err := os.Open(filename)
	if err != nil {
		logger.Error(
			"failed to open repo cache file",
			"path", filename,
			"error", err,
		)
		return nil, err
	}
	defer fp.Close()

	// read all string from file and split by \n
	buf, err := io.ReadAll(fp)
	if err != nil {
		logger.Error(
			"failed to read repo cache file",
			"path", filename,
			"error", err,
		)
		return nil, err
	}
	repos := strings.Split(string(buf), "\n")
	return repos, nil
}

func UpdateRepos(dir string, repos []Repo) ([]string, error) {
	cache, err := LoadRepos(dir)
	// 这里一般不会返回 err ，如果返回 err 则直接退出
	if err != nil {
		return nil, err
	}
	logger.Debug("length of repo cache", "length", len(cache))

	// 使用哈希表去重（用 set 更合适，但是没有标准库支持）
	type empty struct{}
	hashtable := make(map[string]empty)
	for _, repo := range cache {
		hashtable[repo] = empty{}
	}

	logger.Debug("length of repos", "length", len(repos))

	for _, repo := range repos {
		r := repo.String()
		_, ok := hashtable[r]
		if ok {
			continue
		}
		cache = append(cache, r)
		hashtable[r] = empty{}
	}

	// 按照字母表顺序排序，忽略大小写
	sort.Slice(cache, func(i, j int) bool {
		return strings.ToLower(cache[i]) < strings.ToLower(cache[j])
	})

	return cache, nil
}

func SaveRepos(dir string, repos []string) error {
	filename := filepath.Join(dir, repoCacheFileName)

	fp, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Error(
			"failed to open repo cache file",
			"path", filename,
			"error", err,
		)
		return err
	}
	defer fp.Close()

	_, err = io.WriteString(fp, strings.Join(repos, "\n"))
	if err != nil {
		logger.Error(
			"failed to write repo cache file",
			"path", filename,
			"error", err,
		)
		return err
	}

	return nil
}
