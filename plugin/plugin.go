package plugin

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/ghodss/yaml"
)

const (
	cacheDir       = ".conftest"
	cacheDirectory = xdgPath(cacheDir)
)

// Plugin represents a plugin.
type Plugin struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Usage       string `yaml:"usage"`
	Description string `yaml:"description"`
	Command     string `yaml:"command"`
}

// Load loads a plugin given the name of the plugin.
// The name of the plugin is defined in the plugin
// configuration and is stored in a folder with the name
// of the plugin.
func Load(name string) (*Plugin, error) {
	plugin := Plugin{
		Name: name,
	}

	loadedPlugin, err := FromDirectory(plugin.Directory())
	if err != nil {
		return nil, fmt.Errorf("from directory: %w", err)
	}

	return loadedPlugin, nil
}

// FindAll finds all of the plugins available on the
// local file system.
func FindAll() ([]*Plugin, error) {
	if _, err := os.Stat(CacheDirectory()); os.IsNotExist(err) {
		return []*Plugin{}, nil
	}

	files, err := ioutil.ReadDir(CacheDirectory())
	if err != nil {
		return nil, fmt.Errorf("read plugin cache: %w", err)
	}

	var plugins []*Plugin
	for _, file := range files {
		plugin := Plugin{
			Name: file.Name(),
		}

		// While it should not be possible for invalid plugins to be added to
		// the cache, if it does occur, remove the plugin from the cache so it
		// does not prevent valid plugins from being loaded.
		foundPlugin, err := FromDirectory(plugin.Directory())
		if err != nil {
			os.RemoveAll(plugin.Directory())
			continue
		}

		plugins = append(plugins, foundPlugin)
	}

	return plugins, nil
}

// Exec executes the command defined by the plugin along with any
// arguments.
//
// Arguments that are passed into Exec will be added after
// any arguments that are defined in the plugins configuration.
func (p *Plugin) Exec(ctx context.Context, args []string) error {

	// Plugin configurations reference the CONFTEST_PLUGIN_DIR
	// environment to be able to call the plugin.
	os.Setenv("CONFTEST_PLUGIN_DIR", p.Directory())
	expandedCommand := os.ExpandEnv(p.Command)

	var command string
	var arguments []string
	var err error
	if runtime.GOOS == "windows" {
		command, arguments, err = parseWindowsCommand(expandedCommand, args)
	} else {
		command, arguments, err = parseCommand(expandedCommand, args)
	}
	if err != nil {
		return fmt.Errorf("parse command: %w", err)
	}

	cmd := exec.CommandContext(ctx, command, arguments...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	// If an error is found during the execution of the plugin, figure
	// out if the error was from not being able to execute the plugin or
	// an error set by the plugin itself.
	if err := cmd.Run(); err != nil {
		exiterr, ok := err.(*exec.ExitError)
		if !ok {
			return fmt.Errorf("exit: %w", err)
		}

		status, ok := exiterr.Sys().(syscall.WaitStatus)
		if !ok {
			return fmt.Errorf("status: %w", err)
		}

		// Conftest can either return 1 or 2 for an error. If Conftest
		// returns an error, let it handle its own error.
		if status.ExitStatus() == 1 || status.ExitStatus() == 2 {
			return nil
		}

		return fmt.Errorf("plugin exec: %w", err)
	}

	return nil
}

// Directory returns the full path of the directory where the
// plugin is stored in the plugin cache.
func (p *Plugin) Directory() string {
	return filepath.Join(CacheDirectory(), p.Name)
}

// CacheDirectory returns the full path to the
// cache directory where all of the plugins are stored.
func CacheDirectory() string {
	dir, _ := cacheDirectory.Find("plugins")
	return dir
}

// FromDirectory returns a plugin from a specific directory.
//
// The given directory must contain a plugin configuration file
// in order to return successfully.
func FromDirectory(directory string) (*Plugin, error) {
	const configurationFileName = "plugin.yaml"

	configPath := filepath.Join(directory, configurationFileName)
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var plugin Plugin
	if err := yaml.Unmarshal(data, &plugin); err != nil {
		return nil, fmt.Errorf("unmarshal plugin: %w", err)
	}

	return &plugin, nil
}

func parseCommand(command string, extraArgs []string) (string, []string, error) {
	args := strings.Split(command, " ")
	if len(args) == 0 || args[0] == "" {
		return "", nil, fmt.Errorf("prepare plugin command: no command found")
	}

	executable := args[0]

	var configArguments []string
	if len(args) > 1 {
		configArguments = args[1:]
	}

	if len(extraArgs) > 0 {
		configArguments = append(configArguments, extraArgs...)
	}

	return executable, configArguments, nil
}

func parseWindowsCommand(command string, extraArgs []string) (string, []string, error) {
	executable, arguments, err := parseCommand(command, extraArgs)
	if err != nil {
		return "", nil, fmt.Errorf("parse command: %w", err)
	}

	// When executing shell scripts on Windows, the sh
	// program needs to be used to run the script.
	if strings.HasSuffix(executable, ".sh") {
		arguments = append([]string{executable}, arguments...)
		return "sh", arguments, nil
	}

	return executable, arguments, nil
}
