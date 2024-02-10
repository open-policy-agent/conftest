package plugin

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPreferred(t *testing.T) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	const (
		conftestDir = ".conftest"
		pluginsDir  = "plugins"
	)

	tests := []struct {
		name  string
		xdg   xdgPath
		path  string
		setup func()
		want  string
	}{
		{
			"should return homeDir if no XDG path is set.",
			xdgPath(conftestDir),
			pluginsDir,
			func() {
				os.Unsetenv(XDGDataHome)
				os.Unsetenv(XDGDataDirs)
			},
			filepath.Join(userHome, conftestDir, pluginsDir),
		},
		{
			"should return XDG_DATA_HOME if both XDG_DATA_HOME and XDG_DATA_DIRS is set",
			xdgPath(conftestDir),
			pluginsDir,
			func() {
				os.Unsetenv(XDGDataHome)
				os.Unsetenv(XDGDataDirs)
				os.Setenv(XDGDataHome, "/tmp")
				os.Setenv(XDGDataDirs, "/tmp2:/tmp3")
			},
			filepath.Join("/tmp", conftestDir, pluginsDir),
		},
		{
			"should return first XDG_DATA_DIRS if only XDG_DATA_DIRS is set",
			xdgPath(conftestDir),
			pluginsDir,
			func() {
				os.Unsetenv(XDGDataHome)
				os.Unsetenv(XDGDataDirs)
				os.Setenv(XDGDataDirs, "/tmp2:/tmp3")
			},
			filepath.Join("/tmp2", conftestDir, pluginsDir),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			if got := tt.xdg.Preferred(tt.path); got != tt.want {
				t.Errorf("xdgPath.Preferred() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFind(t *testing.T) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	const (
		cacheDir      = ".conftestcache"
		pluginsDir    = "plugins"
		xdgDataHome   = "xdgDataHome"
		xdgDataDirOne = "xdgDataDirOne"
		xdgDataDirTwo = "xdgDataDirTwo"
	)

	tests := []struct {
		name     string
		path     string
		setup    func() (string, error)
		teardown func(string) error
		want     func(string) string
		wantErr  bool
	}{
		{
			"should return error if dir does not exist.",
			pluginsDir,
			func() (string, error) {
				os.Unsetenv(XDGDataHome)
				os.Unsetenv(XDGDataDirs)
				return "/does/not/exist", nil
			},
			func(_ string) error {
				return nil
			},
			func(_ string) string {
				return ""
			},
			true,
		},
		{
			"should return homeDir if no XDG path is set.",
			pluginsDir,
			func() (string, error) {
				os.Unsetenv(XDGDataHome)
				os.Unsetenv(XDGDataDirs)

				homeDir, err := os.UserHomeDir()
				if err != nil {
					return "", err
				}

				dir := filepath.ToSlash(filepath.Join(homeDir))
				cache := filepath.Join(dir, cacheDir)

				err = os.MkdirAll(filepath.Join(cache, pluginsDir), os.ModePerm)
				if err != nil {
					return "", err
				}

				return cache, nil
			},
			func(path string) error {
				return os.RemoveAll(path)
			},
			func(_ string) string {
				return filepath.Join(userHome, cacheDir, pluginsDir)
			},
			false,
		},
		{
			"should return XDG_DATA_HOME if XDG_DATA_HOME is set.",
			pluginsDir,
			func() (string, error) {
				os.Unsetenv(XDGDataHome)
				os.Unsetenv(XDGDataDirs)

				homeDir, err := os.UserHomeDir()
				if err != nil {
					return "", err
				}

				dir := filepath.ToSlash(filepath.Join(homeDir))
				tmp, err := os.MkdirTemp(dir, "")
				if err != nil {
					return "", err
				}

				tmpXdg := filepath.Join(tmp, xdgDataHome)
				err = os.Mkdir(tmpXdg, os.ModePerm)
				if err != nil {
					return "", err
				}
				os.Setenv(XDGDataHome, tmpXdg)

				err = os.MkdirAll(filepath.Join(tmpXdg, cacheDir, pluginsDir), os.ModePerm)
				if err != nil {
					return "", err
				}

				return tmp, nil
			},
			func(path string) error {
				return os.RemoveAll(path)
			},
			func(path string) string {
				return filepath.Join(path, xdgDataHome, cacheDir, pluginsDir)
			},
			false,
		},
		{
			"should return Data Dir with cache if XDG_DATA_DIRS is set.",
			pluginsDir,
			func() (string, error) {
				os.Unsetenv(XDGDataHome)
				os.Unsetenv(XDGDataDirs)

				homeDir, err := os.UserHomeDir()
				if err != nil {
					return "", err
				}

				dir := filepath.ToSlash(filepath.Join(homeDir))
				tmp, err := os.MkdirTemp(dir, "")
				if err != nil {
					return "", err
				}

				tmpXdg1 := filepath.Join(tmp, xdgDataDirOne)
				err = os.Mkdir(tmpXdg1, os.ModePerm)
				if err != nil {
					return "", err
				}

				tmpXdg2 := filepath.Join(tmp, xdgDataDirTwo)
				err = os.Mkdir(tmpXdg2, os.ModePerm)
				if err != nil {
					return "", err
				}
				os.Setenv(XDGDataDirs, tmpXdg1+":"+tmpXdg2)

				err = os.MkdirAll(filepath.Join(tmpXdg2, cacheDir, pluginsDir), os.ModePerm)
				if err != nil {
					return "", err
				}

				return tmp, nil
			},
			func(path string) error {
				return os.RemoveAll(path)
			},
			func(path string) string {
				return filepath.Join(path, xdgDataDirTwo, cacheDir, pluginsDir)
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, err := tt.setup()
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			defer func() {
				err := tt.teardown(dir)
				if err != nil {
					t.Fatalf("unexpected error %v", err)
				}
			}()

			xdg := xdgPath(cacheDir)
			got, err := xdg.Find(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("xdgPath.Find() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			want := tt.want(dir)
			if got != want {
				t.Errorf("xdgPath.Find() = %v, want %v", got, want)
			}
		})
	}
}
