package commands

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/open-policy-agent/conftest/plugin"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// These values are set at build time
var (
	version = ""
)

// NewDefaultCommand creates the default command
func NewDefaultCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:          "conftest <subcommand>",
		Short:        "Test your configuration files using Open Policy Agent",
		Version:      fmt.Sprintf("Version: %s\n", version),
		SilenceUsage: true,
	}

	cmd.SetVersionTemplate(`{{.Version}}`)

	viper.SetEnvPrefix("CONFTEST")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.SetConfigName("conftest")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	logger := log.New(os.Stdout, "", log.LstdFlags)
	ctx := context.Background()

	if err := viper.ReadInConfig(); err != nil {
		var e viper.ConfigFileNotFoundError
		if !errors.As(err, &e) {
			logger.Fatalf("error reading config: %s", err)
		}
	}

	cmd.AddCommand(NewTestCommand(ctx))
	cmd.AddCommand(NewParseCommand(ctx))
	cmd.AddCommand(NewPushCommand(ctx, logger))
	cmd.AddCommand(NewPullCommand(ctx))
	cmd.AddCommand(NewVerifyCommand(ctx))
	cmd.AddCommand(NewPluginCommand(ctx))
	cmd.AddCommand(NewFormatCommand(ctx))

	plugins, err := plugin.FindAll()
	if err != nil {
		logger.Fatalf("find all plugins: %s", err)
	}

	for p := range plugins {
		cmd.AddCommand(newCommandFromPlugin(ctx, plugins[p]))
	}

	return &cmd
}

func newCommandFromPlugin(ctx context.Context, p *plugin.Plugin) *cobra.Command {
	pluginCommand := cobra.Command{
		Use:   p.Name,
		Short: p.Usage,
		Long:  p.Description,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := p.Exec(ctx, args); err != nil {
				return fmt.Errorf("execute plugin: %v", err)
			}

			return nil
		},

		DisableFlagParsing: true,
	}

	return &pluginCommand
}
