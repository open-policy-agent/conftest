package pull

import (
	"context"

	"github.com/instrumenta/conftest/policy"

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
			RunPullCommand(args)
		},
	}

	return cmd
}

// RunPullCommand runs the pull command
func RunPullCommand(repositories []string) {
	policies := getPolicies(repositories)

	ctx := context.Background()
	policy.Download(ctx, policies)
}

func getPolicies(repositories []string) []policy.Policy {
	policies := []policy.Policy{}
	for _, ref := range repositories {
		policies = append(policies, policy.Policy{Repository: ref})
	}

	return policies
}
