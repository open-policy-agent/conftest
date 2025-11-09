package plugin

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPreferred(t *testing.T) {
	t.Parallel()

	userHome, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	tempDir := t.TempDir()

	const (
		conftestDir = ".conftestcache-test"
		pluginsDir  = "plugins"
		nonExistant = "/doesnotexist-349857042375"
	)

	tests := []struct {
		name        string
		path        string
		xdgDataHome string
		xdgDataDirs []string
		want        string
	}{
		{
			name: "should return homeDir if no XDG path is set.",
			path: pluginsDir,
			want: filepath.Join(userHome, conftestDir, pluginsDir),
		},
		{
			name:        "unwritble XDG_DATA_HOME also returns homeDir",
			path:        pluginsDir,
			xdgDataHome: nonExistant,
			want:        filepath.Join(userHome, conftestDir, pluginsDir),
		},
		{
			name:        "should return XDG_DATA_HOME if both XDG_DATA_HOME and XDG_DATA_DIRS is set",
			path:        pluginsDir,
			xdgDataHome: tempDir,
			xdgDataDirs: []string{"/tmp2", "/tmp3"},
			want:        filepath.Join(tempDir, conftestDir, pluginsDir),
		},
		{
			name:        "should return first XDG_DATA_DIRS that exists if only XDG_DATA_DIRS is set",
			path:        pluginsDir,
			xdgDataDirs: []string{nonExistant, tempDir},
			want:        filepath.Join(tempDir, conftestDir, pluginsDir),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			want := filepath.ToSlash(tt.want)

			xdg := xdgPath(conftestDir)
			if got := xdg.preferred(tt.path, tt.xdgDataHome, tt.xdgDataDirs); got != want {
				t.Errorf("xdgPath.Preferred() = %v, want %v", got, want)
			}
		})
	}
}

func TestFind(t *testing.T) {
	t.Parallel()

	const (
		cacheDir    = ".conftestcache-test"
		pluginsDir  = "plugins"
		nonExistant = "/doesnotexist-349857042375"
	)

	userHome, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	userTempDir := filepath.Join(userHome, cacheDir, pluginsDir)
	if err := os.MkdirAll(userTempDir, os.ModePerm); err != nil {
		t.Fatalf("create cache dir under homeDir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(userTempDir) })

	tempDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tempDir, cacheDir, pluginsDir), os.ModePerm); err != nil {
		t.Fatalf("create cache dir under temp dir: %v", err)
	}

	tests := []struct {
		name        string
		path        string
		xdgDataHome string
		xdgDataDirs []string
		want        string
	}{
		{
			name: "should return error if dir does not exist",
			path: nonExistant,
		},
		{
			name: "should use homeDir if no XDG path is set.",
			path: pluginsDir,
			want: filepath.Join(userHome, cacheDir, pluginsDir),
		},
		{
			name:        "should use XDG_DATA_HOME if set",
			path:        pluginsDir,
			xdgDataHome: tempDir,
			want:        filepath.Join(tempDir, cacheDir, pluginsDir),
		},
		{
			name:        "should use first existing XDG_DATA_DIRS if set",
			path:        pluginsDir,
			xdgDataDirs: []string{nonExistant, tempDir},
			want:        filepath.Join(tempDir, cacheDir, pluginsDir),
		},
		{
			name:        "fall back to homeDir if XDG dirs point at non-existent paths",
			path:        pluginsDir,
			xdgDataHome: nonExistant,
			xdgDataDirs: []string{nonExistant},
			want:        filepath.Join(userHome, cacheDir, pluginsDir),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			want := filepath.ToSlash(tt.want)

			xdg := xdgPath(cacheDir)
			got, err := xdg.find(tt.path, tt.xdgDataHome, tt.xdgDataDirs)
			gotErr := err != nil
			wantErr := tt.want == ""
			if gotErr != wantErr {
				t.Fatalf("xdgPath.Find() error = %v, wantErr %v", err, wantErr)
			}
			if got != want {
				t.Errorf("xdgPath.Find() = %s, want %s", got, want)
			}
		})
	}
}
