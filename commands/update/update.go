package update

import (
	"context"

	"github.com/instrumenta/conftest/policy"

	"github.com/containerd/containerd/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Config struct {
	Policy    string
	Namespace string
	Policies  []policy.Policy
}

// NewUpdateCommand creates a new update command
func NewUpdateCommand() *cobra.Command {

	command := &cobra.Command{
		Use:   "update",
		Short: "Download policy from registry",
		Long:  `Download latest policy files according to configuration file`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			var config Config

			if err := viper.Unmarshal(&config); err != nil {
				log.G(ctx).Fatal(err)
			}

			policy.DownloadPolicy(ctx, config.Policies)
		},
	}

	return command
}
