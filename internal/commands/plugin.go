package commands

import (
	"context"

	"github.com/spf13/cobra"
)

// NewPluginCommand creates a new plugin command
func NewPluginCommand(ctx context.Context) *cobra.Command {
	cmd := cobra.Command{
		Use:   "plugin",
		Short: "manage conftest plugins",
		Long:  "This command manages conftest plugins",
	}

	cmd.AddCommand(NewPluginInstallCommand(ctx))

	return &cmd
}
