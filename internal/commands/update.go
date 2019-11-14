package commands

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/instrumenta/conftest/policy"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewUpdateCommand creates a new update command allowing users
// to update their policies from a given registry
func NewUpdateCommand(ctx context.Context) *cobra.Command {
	cmd := cobra.Command{
		Use:   "update",
		Short: "Download policy from registry",
		Long:  `Download latest policy files according to configuration file`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var policies []policy.Policy

			if err := viper.Unmarshal(&policies); err != nil {
				return fmt.Errorf("unmarshal policies: %w", err)
			}

			policyDir := filepath.Join(".", viper.GetString("policy"))

			if err := policy.Download(ctx, policyDir, policies); err != nil {
				return fmt.Errorf("downloading policies: %w", err)
			}

			return nil
		},
	}

	return &cmd
}
