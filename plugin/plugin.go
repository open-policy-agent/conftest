package plugin

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/ghodss/yaml"
)

const metadataFileName = "plugin.yaml"

// Command is the command to be executed by conftest,
// passed as a single string.
type Command string

// Prepare prepares the plugin command and parses out the
// main and args
func (c Command) Prepare() (string, []string, error) {
	args := strings.Split(os.ExpandEnv(string(c)), " ")
	if len(args) == 0 || args[0] == "" {
		return "", nil, fmt.Errorf("prepare plugin command: no command found")
	}

	main := args[0]
	var cmdArgs []string
	if len(args) > 1 {
		cmdArgs = args[1:]
	}

	return main, cmdArgs, nil
}

// MetaData contains the required metadata for conftest plugins
type MetaData struct {
	// Name is the name of the plugin
	Name string `yaml:"name"`

	// Version is a SemVer 2 version of the plugin.
	Version string `yaml:"version"`

	// Usage provides a short description
	// of what the plugin does
	Usage string `yaml:"usage"`

	// Description provides a long description
	// of what the plugin does
	Description string `yaml:"description"`

	// Command is the command to add to conftest
	Command Command `yaml:"command"`
}

// Plugin represents a conftest plugin
type Plugin struct {
	// Metadata contains the contents of the plugin.yaml metatdata file
	MetaData *MetaData

	// Dir contains the full path to the plugin
	Dir string

	stdIn  io.Reader
	stdOut io.Writer
	stdErr io.Writer
	env    []string
}

// SetStdIn configures where the plugin reads from
// when the command is executed
func (p *Plugin) SetStdIn(r io.Reader) *Plugin {
	p.stdIn = r
	return p
}

// SetStdOut configures to where the plugin writes to
// when the command is executed
func (p *Plugin) SetStdOut(w io.Writer) *Plugin {
	p.stdOut = w
	return p
}

// SetStdErr configures to where the plugin writes errors to
// when the command is executed
func (p *Plugin) SetStdErr(w io.Writer) *Plugin {
	p.stdErr = w
	return p
}

// Exec executes the plugin command
func (p *Plugin) Exec(ctx context.Context, args []string) error {
	// Prepare env so plugin has Dir available
	p.setDirInEnv()
	main, cmdArgs, err := p.MetaData.Command.Prepare()
	if err != nil {
		return fmt.Errorf("plugin exec prepare: %w", err)
	}

	execArgs := append(cmdArgs, args...)
	cmd := exec.CommandContext(ctx, main, execArgs...)
	cmd.Stdin = p.stdIn
	cmd.Stdout = p.stdOut
	cmd.Stderr = p.stdErr
	cmd.Env = p.env

	if err = cmd.Run(); err != nil {
		// Check for a 1 exit status to check if it is a conftest test failure
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				if status.ExitStatus() == 1 {
					return nil
				}

				return fmt.Errorf("plugin exec: %w", err)
			}
		}

		return fmt.Errorf("plugin exec: %w", err)
	}

	return nil
}

func (p *Plugin) setDirInEnv() {
	// Use os.SetEnv as the plugin needs access to this environment variable
	os.Setenv("CONFTEST_PLUGIN_DIR", p.Dir)
}

// LoadPlugin loads the plugin.yaml from the given path
// and parses the metadata into a Plugin struct
func LoadPlugin(path string) (*Plugin, error) {
	pluginPath := filepath.Join(path, metadataFileName)
	data, err := ioutil.ReadFile(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("read plugin file: %w", err)
	}

	plugin := &Plugin{
		Dir:    path,
		stdIn:  os.Stdin,
		stdOut: os.Stdout,
		stdErr: os.Stderr,
		env:    os.Environ(),
	}

	if err := yaml.Unmarshal(data, &plugin.MetaData); err != nil {
		return nil, fmt.Errorf("parse plugin file: %w", err)
	}

	return plugin, nil
}

// FindPlugins returns a list of all plugins available on the local file system.
func FindPlugins() ([]*Plugin, error) {
	var plugins []*Plugin
	homePath, err := fetchHomeDir()
	if err != nil {
		return nil, fmt.Errorf("fetch home path: %w", err)
	}

	pluginsCacheDirPath := filepath.Join(homePath, ConftestDir, PluginsCacheDir)
	if _, err := os.Stat(pluginsCacheDirPath); os.IsNotExist(err) {
		// No plugins, so just return the empty slice
		return plugins, nil
	}

	files, err := ioutil.ReadDir(pluginsCacheDirPath)
	if err != nil {
		return nil, fmt.Errorf("read plugin cache directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			path := filepath.Join(pluginsCacheDirPath, file.Name())
			plugin, err := LoadPlugin(path)
			if err != nil {
				return nil, fmt.Errorf("load plugin from cache directory: %w", err)
			}

			plugins = append(plugins, plugin)
		} else if file.Mode()&os.ModeSymlink != 0 {
			// go-getter symlinks if it is a plugin on the local file system.
			symlinkPath := filepath.Join(pluginsCacheDirPath, file.Name())
			path, err := os.Readlink(symlinkPath)
			if err != nil {
				return nil, fmt.Errorf("resolve plugin symlink: %w", err)
			}

			plugin, err := LoadPlugin(path)
			if err != nil {
				return nil, fmt.Errorf("load plugin from cache directory: %w", err)
			}

			plugins = append(plugins, plugin)
		}
	}

	return plugins, nil
}
