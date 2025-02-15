package plugin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"sigs.k8s.io/yaml"
)

func TestParseCommand(t *testing.T) {
	executable, arguments, err := parseCommand("plugin.sh arg1", []string{"arg2"})
	if err != nil {
		t.Fatal("parse command:", err)
	}

	if executable != "plugin.sh" {
		t.Errorf("Unexpected executable. expected %v, actual %v", "plugin.sh", executable)
	}
	if arguments[0] != "arg1" {
		t.Errorf("Unexpected argument. expected %v, actual %v", "arg1", arguments[0])
	}
	if arguments[1] != "arg2" {
		t.Errorf("Unexpected argument. expected %v, actual %v", "arg2", arguments[1])
	}
}

func TestParseCommand_WindowsShell(t *testing.T) {
	executable, arguments, err := parseWindowsCommand("plugin.sh arg1", []string{"arg2"})
	if err != nil {
		t.Fatal("parse command:", err)
	}

	if executable != "sh" {
		t.Errorf("Unexpected executable. expected %v, actual %v", "sh", executable)
	}
	if arguments[0] != "plugin.sh" {
		t.Errorf("Unexpected argument. expected %v, actual %v", "plugin.sh", arguments[0])
	}
	if arguments[1] != "arg1" {
		t.Errorf("Unexpected argument. expected %v, actual %v", "arg1", arguments[1])
	}
	if arguments[2] != "arg2" {
		t.Errorf("Unexpected argument. expected %v, actual %v", "arg2", arguments[2])
	}
}

func TestLoad(t *testing.T) {
	t.Run("valid plugin", func(t *testing.T) {
		dir := createTestPlugin(t, &Plugin{
			Name:    "test-plugin",
			Version: "1.0",
			Command: "echo hello",
		})

		plugin, err := Load("test-plugin")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if plugin.Directory() != dir {
			t.Errorf("expected directory %s, got %s", dir, plugin.Directory())
		}
	})

	t.Run("non-existent plugin", func(t *testing.T) {
		_, err := Load("non-existent")
		if err == nil {
			t.Fatal("expected error but got none")
		}
	})
}

func TestFindAll(t *testing.T) {
	// Set up isolated cache directory
	tmpDir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tmpDir)
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// Create plugins directly in the expected cache location
	pluginDir := filepath.Join(tmpDir, ".conftest", "plugins")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create 3 test plugins
	plugins := []*Plugin{
		{Name: "plugin1", Command: "echo 1"},
		{Name: "plugin2", Command: "echo 2"},
		{Name: "plugin3", Command: "echo 3"},
	}

	for _, p := range plugins {
		dir := filepath.Join(pluginDir, p.Name)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}

		data, err := yaml.Marshal(p)
		if err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(filepath.Join(dir, "plugin.yaml"), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}

	// Create invalid plugin directory
	invalidDir := filepath.Join(pluginDir, "invalid-plugin")
	if err := os.MkdirAll(invalidDir, 0755); err != nil {
		t.Fatal(err)
	}

	found, err := FindAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(found) != 3 {
		t.Errorf("expected 3 plugins, found %d", len(found))
	}

	// Verify invalid plugin was removed
	if _, err := os.Stat(invalidDir); !os.IsNotExist(err) {
		t.Error("expected invalid plugin directory to be removed")
	}
}

func TestPluginExec(t *testing.T) {
	t.Run("basic command", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "output.txt")

		p := &Plugin{
			Name:    "exec-test",
			Command: fmt.Sprintf("touch %s", testFile),
		}
		createTestPlugin(t, p)

		plugin, err := Load("exec-test")
		if err != nil {
			t.Fatal(err)
		}

		if err := plugin.Exec(context.Background(), nil); err != nil {
			t.Fatal(err)
		}

		content, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatal(err)
		}

		if string(content) != "" {
			t.Errorf("unexpected output file content: %q", string(content))
		}
	})

	t.Run("with arguments", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "args.txt")

		p := &Plugin{
			Name:    "args-test",
			Command: "echo",
		}
		createTestPlugin(t, p)

		plugin, err := Load("args-test")
		if err != nil {
			t.Fatal(err)
		}

		originalStdout := os.Stdout
		defer func() { os.Stdout = originalStdout }()

		f, err := os.Create(testFile)
		if err != nil {
			t.Fatal(err)
		}
		os.Stdout = f
		defer f.Close()

		if err := plugin.Exec(context.Background(), []string{"arg1", "arg2"}); err != nil {
			t.Fatal(err)
		}

		content, _ := os.ReadFile(testFile)
		expected := "arg1 arg2\n"
		if string(content) != expected {
			t.Errorf("expected %q, got %q", expected, string(content))
		}
	})

	t.Run("command error handling", func(t *testing.T) {
		p := &Plugin{
			Name:    "error-test",
			Command: "exit 42",
		}
		createTestPlugin(t, p)

		plugin, err := Load("error-test")
		if err != nil {
			t.Fatal(err)
		}

		err = plugin.Exec(context.Background(), nil)
		if err == nil {
			t.Fatal("expected error but got none")
		}
	})
}

// Helper to create a test plugin in the cache directory
func createTestPlugin(t *testing.T, plugin *Plugin) string {
	t.Helper()
	tmpDir := t.TempDir()

	// Create a specific plugins directory structure
	pluginsDir := filepath.Join(tmpDir, ".conftest", "plugins")
	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Set both XDG variables to use our temp directory
	t.Setenv("XDG_CACHE_HOME", tmpDir)
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// Create the plugin directory
	resolveDir := filepath.Join(CacheDirectory(), plugin.Name)
	if err := os.MkdirAll(resolveDir, 0755); err != nil {
		t.Fatal(err)
	}

	data, err := yaml.Marshal(plugin)
	if err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(resolveDir, "plugin.yaml")
	if err := os.WriteFile(configPath, data, 0o600); err != nil {
		t.Fatal(err)
	}

	return resolveDir
}
