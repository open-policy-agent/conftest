package downloader

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"runtime"

	getter "github.com/hashicorp/go-getter"
)

// CopyFileGetter is a custom FileGetter that copies directories instead of
// creating symlinks. This is needed on Windows where symlinks require
// administrator privileges.
type CopyFileGetter struct {
	getter.FileGetter
}

// Get implements getter.Getter. On Windows, it copies the directory instead
// of creating a symlink/junction.
func (g *CopyFileGetter) Get(dst string, u *url.URL) error {
	// On non-Windows systems, use the standard behavior
	if runtime.GOOS != "windows" {
		return g.FileGetter.Get(dst, u)
	}

	path := u.Path
	if u.RawPath != "" {
		path = u.RawPath
	}

	// Handle Windows file:// URLs per RFC 8089
	// When using file:///C:/path format, u.Path is "/C:/path"
	// We need to strip the leading "/" to get the actual Windows path
	if len(path) > 2 && path[0] == '/' && filepath.VolumeName(path[1:]) != "" {
		path = path[1:]
	}

	// The source path must exist and be a directory
	fi, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("source path error: %s", err)
	}
	if !fi.IsDir() {
		return fmt.Errorf("source path must be a directory")
	}

	// Remove destination if it exists
	if _, err := os.Lstat(dst); err == nil {
		if err := os.RemoveAll(dst); err != nil {
			return fmt.Errorf("failed to remove existing destination: %w", err)
		}
	}

	// Copy the directory
	return copyDir(path, dst)
}

// copyDir recursively copies a directory from src to dst.
func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source: %w", err)
	}

	// Create the destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file from src to dst.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}

	dstFile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	return nil
}
