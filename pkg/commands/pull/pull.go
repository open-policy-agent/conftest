package pull

import (
	"context"

	"github.com/instrumenta/conftest/pkg/policy"

	"github.com/spf13/cobra"
)

// NewPullCommand creates a new pull command
func NewPullCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pull <repository>",
		Short: "Download individual policies",
		Long:  `Download individual policies from a registry`,
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			policies := []policy.Policy{}
			for _, ref := range args {
				policies = append(policies, policy.Policy{Repository: ref})
			}
			policy.DownloadPolicy(ctx, policies)
		},
	}

	return cmd
}
