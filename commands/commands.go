package commands

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/instrumenta/conftest/commands/parse"
	"github.com/instrumenta/conftest/commands/pull"
	"github.com/instrumenta/conftest/commands/push"
	"github.com/instrumenta/conftest/commands/test"
	"github.com/instrumenta/conftest/commands/update"
	"github.com/instrumenta/conftest/commands/verify"
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

	cmd.PersistentFlags().StringP("policy", "p", "policy", "path to the Rego policy files directory. For the test command, specifying a specific .rego file is allowed.")
	cmd.PersistentFlags().BoolP("trace", "", false, "enable more verbose trace output for rego queries")
	cmd.PersistentFlags().StringP("namespace", "", "main", "namespace in which to find deny and warn rules")
	cmd.PersistentFlags().BoolP("no-color", "", false, "disable color when printing")

	cmd.SetVersionTemplate(`{{.Version}}`)

	viper.BindPFlag("policy", cmd.PersistentFlags().Lookup("policy"))
	viper.BindPFlag("trace", cmd.PersistentFlags().Lookup("trace"))
	viper.BindPFlag("namespace", cmd.PersistentFlags().Lookup("namespace"))
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

	cmd.AddCommand(test.NewTestCommand(ctx))
	cmd.AddCommand(parse.NewParseCommand(ctx))
	cmd.AddCommand(update.NewUpdateCommand(ctx))
	cmd.AddCommand(push.NewPushCommand(ctx, logger))
	cmd.AddCommand(pull.NewPullCommand(ctx))
	cmd.AddCommand(verify.NewVerifyCommand(
		test.GetOutputManager,
	))

	return &cmd
}
