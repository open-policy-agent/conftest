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

	opaversion "github.com/open-policy-agent/opa/v1/version"

	// Load the custom builtins.
	_ "github.com/open-policy-agent/conftest/builtins"
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
		Version:      createVersionString(),
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if err := viper.BindPFlag("config-file", cmd.Flags().Lookup("config-file")); err != nil {
				return fmt.Errorf("bind flag: %s", err)
			}
			if err := readInConfig(); err != nil {
				return fmt.Errorf("read in config: %s", err)
			}
			return nil
		},
	}

	cmd.SetVersionTemplate(`{{.Version}}`)

	viper.SetEnvPrefix("CONFTEST")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.SetConfigName("conftest")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	logger := log.New(os.Stdout, "", log.LstdFlags)
	ctx := context.Background()

	cmd.AddCommand(NewTestCommand(ctx))
	cmd.AddCommand(NewParseCommand())
	cmd.AddCommand(NewPushCommand(ctx, logger))
	cmd.AddCommand(NewPullCommand(ctx))
	cmd.AddCommand(NewVerifyCommand(ctx))
	cmd.AddCommand(NewPluginCommand(ctx))
	cmd.AddCommand(NewFormatCommand())
	cmd.AddCommand(NewReformatCommand())
	cmd.AddCommand(NewDocumentCommand())

	plugins, err := plugin.FindAll()
	if err != nil {
		logger.Fatalf("find all plugins: %s", err)
	}

	for p := range plugins {
		cmd.AddCommand(newCommandFromPlugin(ctx, plugins[p]))
	}

	cmd.PersistentFlags().StringP("config-file", "c", "", "path to configuration file")

	return &cmd
}

func newCommandFromPlugin(ctx context.Context, p *plugin.Plugin) *cobra.Command {
	pluginCommand := cobra.Command{
		Use:   p.Name,
		Short: p.Usage,
		Long:  p.Description,
		RunE: func(_ *cobra.Command, args []string) error {
			if err := p.Exec(ctx, args); err != nil {
				return fmt.Errorf("execute plugin: %v", err)
			}

			return nil
		},

		DisableFlagParsing: true,
	}

	return &pluginCommand
}

func createVersionString() string {
	return fmt.Sprintf("Conftest: %s\nOPA: %s\n", version, opaversion.Version)
}

func readInConfig() error {
	viper.SetConfigFile(viper.GetString("config-file"))
	if err := viper.ReadInConfig(); err != nil {
		var e viper.ConfigFileNotFoundError
		if !errors.As(err, &e) {
			return fmt.Errorf("error reading config: %s", err)
		}
	}
	return nil
}
