package commands

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/instrumenta/conftest/policy"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewPullCommand creates a new pull command to allow users
// to download individual policies
func NewPullCommand(ctx context.Context) *cobra.Command {
	cmd := cobra.Command{
		Use:   "pull <repository>",
		Short: "Download individual policies",
		Long:  `Download individual policies from a registry`,
		Args:  cobra.MinimumNArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {
			policyDir := filepath.Join(".", viper.GetString("policy"))

			if err := policy.Download(ctx, policyDir, args); err != nil {
				return fmt.Errorf("download policies: %w", err)
			}

			return nil
		},
	}

	return &cmd
}
