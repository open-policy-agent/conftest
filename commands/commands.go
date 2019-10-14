package commands

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/containerd/containerd/log"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/instrumenta/conftest/commands/pull"
	"github.com/instrumenta/conftest/commands/push"
	"github.com/instrumenta/conftest/commands/test"
	"github.com/instrumenta/conftest/commands/update"
	"github.com/instrumenta/conftest/commands/verify"
	"github.com/instrumenta/conftest/constants"
)

// NewDefaultCommand creates the default command
func NewDefaultCommand() *cobra.Command {
	ctx := context.Background()
	cmd := cobra.Command{
		Use:     "conftest <subcommand>",
		Short:   "Test your configuration files using Open Policy Agent",
		Version: fmt.Sprintf("Version: %s\nCommit: %s\nDate: %s\n", constants.Version, constants.Commit, constants.Date),
	}

	cmd.PersistentFlags().StringP("policy", "p", "policy", "path to the Rego policy files directory. For the test command, specifying a specific .rego file is allowed.")
	cmd.PersistentFlags().BoolP("debug", "", false, "enable more verbose log output")
	cmd.PersistentFlags().BoolP("trace", "", false, "enable more verbose trace output for rego queries")
	cmd.PersistentFlags().StringP("namespace", "", "main", "namespace in which to find deny and warn rules")
	cmd.PersistentFlags().BoolP("no-color", "", false, "disable color when printing")

	cmd.SetVersionTemplate(`{{.Version}}`)

	viper.BindPFlag("policy", cmd.PersistentFlags().Lookup("policy"))
	viper.BindPFlag("debug", cmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("trace", cmd.PersistentFlags().Lookup("trace"))
	viper.BindPFlag("namespace", cmd.PersistentFlags().Lookup("namespace"))
	viper.BindPFlag("no-color", cmd.PersistentFlags().Lookup("no-color"))

	viper.SetEnvPrefix("CONFTEST")
	viper.SetConfigName("conftest")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		var e viper.ConfigFileNotFoundError
		if !errors.As(err, &e) {
			log.G(ctx).Fatal("failed to parse config file: ", err)
		}
	}

	cmd.AddCommand(test.NewTestCommand(
		os.Exit,
		test.GetOutputManager,
	))
	cmd.AddCommand(update.NewUpdateCommand())
	cmd.AddCommand(push.NewPushCommand())
	cmd.AddCommand(pull.NewPullCommand())
	cmd.AddCommand(verify.NewVerifyCommand(
		test.GetOutputManager,
	))

	if viper.GetBool("debug") {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.ErrorLevel)
	}

	return &cmd
}
