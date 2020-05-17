package plugin

import (
	"context"
	"fmt"
	"os"

	"github.com/open-policy-agent/conftest/downloader"
)

// Download downloads the given plugin into the cache
func Download(ctx context.Context, url string) error {
	homePath, err := fetchHomeDir()
	if err != nil {
		return fmt.Errorf("fetch home path: %w", err)
	}

	cacheDirPath, err := createPluginCacheDir(homePath)
	if err != nil {
		return fmt.Errorf("create plugin cache: %w", err)
	}

	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working dir: %w", err)
	}

	sourcedURL, err := downloader.Detect(url, pwd)
	if err != nil {
		return fmt.Errorf("detect plugin: %w", err)
	}

	pluginDirPath := getPluginDirPath(cacheDirPath, sourcedURL)
	if checkIfURLInCache(pluginDirPath) {
		return nil
	}

	if err = downloader.Download(ctx, pluginDirPath, []string{sourcedURL}); err != nil {
		return fmt.Errorf("download plugin: %w", err)
	}

	return nil
}
