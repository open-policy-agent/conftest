package commands

import (
	"context"
	"fmt"

	"github.com/open-policy-agent/conftest/plugin"
	"github.com/spf13/cobra"
)

func loadPlugins(ctx context.Context) ([]*cobra.Command, error) {
	plugins, err := plugin.FindPlugins()
	if err != nil {
		return nil, fmt.Errorf("loading plugins: %v", err)
	}

	var cmds []*cobra.Command
	for _, plugin := range plugins {
		plugin := plugin
		metaData := plugin.MetaData
		cmd := &cobra.Command{
			Use:   metaData.Name,
			Short: metaData.Usage,
			Long:  metaData.Description,
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := plugin.Exec(ctx, args); err != nil {
					return fmt.Errorf("execute plugin: %v", err)
				}

				return nil
			},
			DisableFlagParsing: true,
		}

		cmds = append(cmds, cmd)
	}

	return cmds, nil
}
