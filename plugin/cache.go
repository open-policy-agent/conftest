package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	ConftestDir     = ".conftest"
	PluginsCacheDir = "plugins"
)

func fetchHomeDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("Could not fetch home directory: %w", err)
	}

	return home, nil
}

func createPluginCacheDir(basePath string) (string, error) {
	pluginsCacheDirPath := filepath.Join(basePath, ConftestDir, PluginsCacheDir)

	if _, err := os.Stat(pluginsCacheDirPath); os.IsExist(err) {
		// No need to create the directory if it exists
		return pluginsCacheDirPath, nil
	}

	err := os.MkdirAll(pluginsCacheDirPath, 0700)
	if err != nil {
		return "", fmt.Errorf("Could not create conftest plugin cache: %w", err)
	}

	return pluginsCacheDirPath, nil
}

func checkIfURLInCache(pluginDirPath string) bool {
	if _, err := os.Stat(pluginDirPath); os.IsExist(err) {
		return true
	}

	return false
}

func getPluginDirPath(cacheDir string, url string) string {
	pluginDir := convertURLToValidPath(url)
	return filepath.Join(cacheDir, pluginDir)
}

func convertURLToValidPath(url string) string {
	url = strings.ReplaceAll(url, "://", "-")
	url = strings.ReplaceAll(url, "/", "-")
	url = strings.ReplaceAll(url, ".", "-")
	return url
}
