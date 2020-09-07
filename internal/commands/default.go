package commands

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/open-policy-agent/conftest/plugin"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// These values are set at build time
var (
	version = ""
	commit  = ""
	date    = ""
)

// NewDefaultCommand creates the default command
func NewDefaultCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:     "conftest <subcommand>",
		Short:   "Test your configuration files using Open Policy Agent",
		Version: fmt.Sprintf("Version: %s\nCommit: %s\nDate: %s\n", version, commit, date),
	}

	cmd.SetVersionTemplate(`{{.Version}}`)

	cmd.PersistentFlags().StringSliceP("policy", "p", []string{"policy"}, "path to the Rego policy files directory. For the test command, specifying a specific .rego file is allowed. Can be specified multiple times.")
	cmd.PersistentFlags().Bool("no-color", false, "disable color when printing")

	viper.BindPFlag("policy", cmd.PersistentFlags().Lookup("policy"))
	viper.BindPFlag("no-color", cmd.PersistentFlags().Lookup("no-color"))

	viper.SetEnvPrefix("CONFTEST")
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

	pluginCmds, err := loadPlugins(ctx)
	if err != nil {
		logger.Fatalf("error loading plugins: %v", err)
	}

	cmd.AddCommand(pluginCmds...)

	return &cmd
}

func loadPlugins(ctx context.Context) ([]*cobra.Command, error) {
	plugins, err := plugin.FindPlugins()
	if err != nil {
		return nil, fmt.Errorf("loading plugins: %v", err)
	}

	var cmds []*cobra.Command
	for _, plugin := range plugins {
		plugin := plugin
		metaData := plugin.MetaData
		cmd := &cobra.Command{
			Use:   metaData.Name,
			Short: metaData.Usage,
			Long:  metaData.Description,
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := plugin.Exec(ctx, args); err != nil {
					return fmt.Errorf("execute plugin: %v", err)
				}

				return nil
			},
			DisableFlagParsing: true,
		}

		cmds = append(cmds, cmd)
	}

	return cmds, nil
}
