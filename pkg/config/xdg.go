package config

import "github.com/adrg/xdg"

var configFilePath,
	repoCacheFilePath,
	logFilePath string

func init() {
	configFilePath, _ = xdg.ConfigFile("gh2/config.toml")
	repoCacheFilePath, _ = xdg.DataFile("gh2/repos")
	logFilePath, _ = xdg.DataFile("gh2/log")
}
