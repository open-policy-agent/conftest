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
	dataHome := os.Getenv(XDGDataHome)
	if dataHome != "" {
		return filepath.ToSlash(filepath.Join(dataHome, string(p), path))
	}

	dataDirs := os.Getenv(XDGDataDirs)
	if dataDirs != "" {
		dirs := strings.Split(dataDirs, ":")
		return filepath.ToSlash(filepath.Join(dirs[0], string(p), path))
	}

	homeDir, _ := os.UserHomeDir()
	return filepath.ToSlash(filepath.Join(homeDir, string(p), path))
}

// Find verifies whether the file exists somewhere in the expected XDG
// preference order. If no error is returned, the given string indicates
// where the file was found.
func (p xdgPath) Find(path string) (string, error) {
	dataHome := os.Getenv(XDGDataHome)
	if dataHome != "" {
		dir := filepath.ToSlash(filepath.Join(dataHome, string(p), path))
		_, err := os.Stat(dir)
		if err != nil && !os.IsNotExist(err) {
			return "", fmt.Errorf("get data home directory: %w", err)
		}

		if err == nil {
			return dir, nil
		}
	}

	dataDirs := os.Getenv(XDGDataDirs)
	if dataDirs != "" {
		dirs := strings.Split(dataDirs, ":")
		for _, dataDir := range dirs {
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
