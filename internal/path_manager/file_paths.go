package path_manager

import (
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

var registry PathRegistry

func init() {
	// Default Paths
	registry.configFilePath = "./config.toml"
	registry.cacheFolderPath = "./cache"
	registry.zoneFilesFolderPath = "./zone-files"
	registry.pluginFolderPath = "./plugin"
}

type PathRegistry struct {
	configFilePath      string
	cacheFolderPath     string
	zoneFilesFolderPath string
	pluginFolderPath    string
}

func GetConfigFilePath() string {
	return registry.configFilePath
}

func SetConfigFilePath(configFilePath string) {
	absFilePath, err := filepath.Abs(configFilePath)
	if err != nil {
		zap.L().Fatal(
			"Unable to convert to absolute file path.",
			zap.String("file_path", configFilePath),
			zap.String("err", err.Error()),
		)
	}

	registry.configFilePath = absFilePath
}

func GetCacheFolderPath() string {
	return registry.cacheFolderPath
}

func SetCacheFolderPath(cacheFolderPath string) {
	absFilePath, err := filepath.Abs(cacheFolderPath)
	if err != nil {
		zap.L().Fatal(
			"Unable to convert to absolute file path.",
			zap.String("file_path", cacheFolderPath),
			zap.String("err", err.Error()),
		)
	}

	registry.cacheFolderPath = absFilePath
}

func GetZoneFilesFolderPath() string {
	return registry.zoneFilesFolderPath
}

func SetZoneFilesFolderPath(zoneFilesFolderPath string) {
	absFilePath, err := filepath.Abs(zoneFilesFolderPath)
	if err != nil {
		zap.L().Fatal(
			"Unable to convert to absolute file path.",
			zap.String("file_path", zoneFilesFolderPath),
			zap.String("err", err.Error()),
		)
	}

	registry.zoneFilesFolderPath = absFilePath
}

func GetPluginFolderPath() string {
	return registry.pluginFolderPath
}

func SetPluginFolderPath(pluginFolderPath string) {
	absFilePath, err := filepath.Abs(pluginFolderPath)
	if err != nil {
		zap.L().Fatal(
			"Unable to convert to absolute file path.",
			zap.String("file_path", pluginFolderPath),
			zap.String("err", err.Error()),
		)
	}

	registry.pluginFolderPath = absFilePath
}

func GetPreCrawlFilterOutputFilePath(tldName string) string {
	return filepath.Join(GetCacheFolderPath(), "zone", tldName, "pre_crawl_output.txt")
}

func GetPostCrawlFilterOutputFilePath(tldName string) string {
	return filepath.Join(GetCacheFolderPath(), "zone", tldName, "post_crawl_output.txt")
}

func GetIndexerDatabaseFilePath(indexerType string) string {
	// Let's replace dots in the indexerType
	indexerName := strings.Replace(indexerType, ".", "-", -1)

	return filepath.Join(GetCacheFolderPath(), "indexer", indexerName, "index.bleve")
}

func GetRankingFilePath(userQuery string) string {
	// TODO [HP]: We need to Base64 encode the userQuery and cache it as needed.
	// Alternatively, we can just use a metadata field.
	return filepath.Join(GetCacheFolderPath(), "ranking", "output.txt")
}
