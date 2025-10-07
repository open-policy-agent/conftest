package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// XDGDataHome is the directory to search for data files in the XDG spec
	XDGDataHome = "XDG_DATA_HOME"

	// XDGDataDirs defines an additional list of directories which can be searched for data files
	XDGDataDirs = "XDG_DATA_DIRS"
)

type xdgPath string

// Preferred returns the preferred path according to the XDG specification
func (p xdgPath) Preferred(path string) string {
	return p.preferred(path, os.Getenv(XDGDataHome), os.Getenv(XDGDataDirs))
}

func (p xdgPath) preferred(path, xdgDataHome, xdgDataDirs string) string {
	if xdgDataHome != "" && p.writable(xdgDataHome) {
		return filepath.ToSlash(filepath.Join(xdgDataHome, string(p), path))
	}

	if xdgDataDirs != "" {
		// Pick the first dir that is writable.
		for dir := range strings.SplitSeq(xdgDataDirs, ":") {
			if p.writable(dir) {
				return filepath.ToSlash(filepath.Join(dir, string(p), path))
			}
		}
	}

	return p.homeDir(path)
}

func (p xdgPath) homeDir(path string) string {
	dir, _ := os.UserHomeDir()
	return filepath.ToSlash(filepath.Join(dir, string(p), path))
}

func (p xdgPath) writable(path string) bool {
	// The easiest cross-platform way to check if it is writable is
	// to just create a directory and then remove it.
	tempDir, err := os.MkdirTemp(path, ".conftestcheck-")
	if err != nil {
		return false
	}
	os.RemoveAll(tempDir)
	return true
}

// Find verifies whether the file exists somewhere in the expected XDG
// preference order. If no error is returned, the given string indicates
// where the file was found.
func (p xdgPath) Find(path string) (string, error) {
	return p.find(path, os.Getenv(XDGDataHome), os.Getenv(XDGDataDirs))
}

func (p xdgPath) find(path, xdgDataHome, xdgDataDirs string) (string, error) {
	if xdgDataHome != "" {
		dir := filepath.ToSlash(filepath.Join(xdgDataHome, string(p), path))
		_, err := os.Stat(dir)
		if err != nil && !os.IsNotExist(err) {
			return "", fmt.Errorf("get data home directory: %w", err)
		}
		if err == nil {
			return dir, nil
		}
	}

	if xdgDataDirs != "" {
		for dataDir := range strings.SplitSeq(xdgDataDirs, ":") {
			dir := filepath.ToSlash(filepath.Join(dataDir, string(p), path))
			_, err := os.Stat(dir)
			if err != nil && !os.IsNotExist(err) {
				return "", fmt.Errorf("get data dirs directory: %w", err)
			}
			if err == nil {
				return dir, nil
			}
		}
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}

	dir := filepath.ToSlash(filepath.Join(homeDir, string(p), path))
	_, err = os.Stat(dir)
	if err != nil {
		return "", fmt.Errorf("get data dirs directory: %w", err)
	}

	return dir, nil
}
