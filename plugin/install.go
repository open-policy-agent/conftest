package plugin

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"unicode"

	"github.com/open-policy-agent/conftest/downloader"
)

// Install installs the plugin to the host machine from either a
// path on the file system, or a URL. A configuration file must be
// present and valid at the source in order for the installation to
// complete successfully.
//
// If the installation is successful, the plugin will be saved to the
// plugin cache inside of a folder named the same name of the plugin
// as defined in the plugins configuration file.
func Install(ctx context.Context, source string) error {
	if _, err := os.Stat(CacheDirectory()); os.IsNotExist(err) {
		if err := os.MkdirAll(CacheDirectory(), os.ModePerm); err != nil {
			return fmt.Errorf("make plugin dir: %w", err)
		}
	}

	sourceIsDirectory, err := isDirectory(source)
	if err != nil {
		return fmt.Errorf("detect source type: %w", err)
	}

	if sourceIsDirectory {
		if err := installFromDirectory(ctx, source); err != nil {
			return fmt.Errorf("install from dir: %w", err)
		}
	} else {
		if err := installFromURL(ctx, source); err != nil {
			return fmt.Errorf("install from url: %w", err)
		}
	}

	return nil
}

func isDirectory(source string) (bool, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return false, fmt.Errorf("get working dir: %w", err)
	}

	detect, err := downloader.Detect(source, workingDir)
	if err != nil {
		return false, fmt.Errorf("detect: %w", err)
	}

	parsed, err := url.Parse(detect)
	if err != nil {
		return false, fmt.Errorf("parse detected: %w", err)
	}

	// A file on the file system will have a scheme of file, set
	// by the detector.
	//
	// The detector also sets the scheme to file if none of the other
	// detectors were able to detect the source. This could lead to
	// false positives when the source input is unknown, but will
	// ultimately get caught when attempting to read the file from disk.
	if parsed.Scheme == "file" {
		return true, nil
	}

	// The detector will not perform any formatting if the source was successfully
	// parsed as a URL and contains a scheme. When the source is an absolute path
	// on Windows, the scheme will be the drive letter.
	if len(parsed.Scheme) == 1 && unicode.IsLetter(rune(parsed.Scheme[0])) {
		return true, nil
	}

	return false, nil
}

func installFromURL(ctx context.Context, url string) error {

	// Before allowing the plugin to be stored in the plugin directory,
	// make sure the plugin has a valid configuration file by first downloading
	// the plugin into a temporary directory and validating its configuration.
	tempDirectory, err := ioutil.TempDir(CacheDirectory(), "conftest-plugin-*")
	if err != nil {
		return fmt.Errorf("create tmp dir: %w", err)
	}
	defer os.RemoveAll(tempDirectory)

	if err := downloader.Download(ctx, tempDirectory, []string{url}); err != nil {
		return fmt.Errorf("download plugin: %w", err)
	}

	plugin, err := FromDirectory(tempDirectory)
	if err != nil {
		return fmt.Errorf("loading plugin from dir: %w", err)
	}
	if err := os.RemoveAll(plugin.Directory()); err != nil {
		return fmt.Errorf("remove old plugin: %w", err)
	}
	if err := os.Rename(tempDirectory, plugin.Directory()); err != nil {
		return fmt.Errorf("rename temp dir: %w", err)
	}

	return nil
}

func installFromDirectory(ctx context.Context, sourceDirectory string) error {
	sourceDirectory, err := filepath.Abs(sourceDirectory)
	if err != nil {
		return fmt.Errorf("abs source: %w", err)
	}

	plugin, err := FromDirectory(sourceDirectory)
	if err != nil {
		return fmt.Errorf("loading plugin from dir: %w", err)
	}
	if err := os.RemoveAll(plugin.Directory()); err != nil {
		return fmt.Errorf("remove old plugin: %w", err)
	}
	if err := downloader.Download(ctx, plugin.Directory(), []string{sourceDirectory}); err != nil {
		return fmt.Errorf("download plugin: %w", err)
	}

	return nil
}
