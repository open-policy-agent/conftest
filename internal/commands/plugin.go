package commands

import (
	"context"

	"github.com/spf13/cobra"
)

const pluginDesc = `
	This command manages conftest plugins
`

func NewPluginCommand(ctx context.Context) *cobra.Command {
	cmd := cobra.Command{
		Use:   "plugin",
		Short: "manage conftest plugins",
		Long:  pluginDesc,
	}

	cmd.AddCommand(NewPluginInstallCommand(ctx))

	return &cmd
}
