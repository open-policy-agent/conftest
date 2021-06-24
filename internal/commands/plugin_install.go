package commands

import (
	"context"
	"fmt"

	"github.com/open-policy-agent/conftest/plugin"
	"github.com/spf13/cobra"
)

const installDesc = `
This command installs a plugin from the given path or url

Several locations are supported by the plugin install command. Under the hood
conftest leverages go-getter (https://github.com/hashicorp/go-getter).
The following protocols are supported for downloading plugins:

	- Local Files
	- Git
	- HTTP/HTTPS
	- Mercurial
	- Amazon S3
	- Google Cloud GCP

The location of the plugins is specified by passing a path or URL, e.g.:

	$ conftest plugin install github.com/open-policy-agent/conftest/examples/plugins/kubectl
	$ conftest plugin install contrib/plugins/kubectl

Based on the protocol a different mechanism will be used to download the plugin.
The pull command will also try to infer the protocol based on the URL if the
URL does not contain a protocol.

The plugins will be installed on disk in ~/.conftest/plugins.
`

// NewPluginInstallCommand creates the install plugin subcommand
func NewPluginInstallCommand(ctx context.Context) *cobra.Command {
	cmd := cobra.Command{
		Use:   "install <path|url>",
		Short: "Install a plugin from the given path or url",
		Long:  installDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				cmd.Usage() //nolint
				return fmt.Errorf("missing required arguments")
			}

			if err := plugin.Install(ctx, args[0]); err != nil {
				return fmt.Errorf("install: %v", err)
			}

			return nil
		},
	}

	return &cmd
}
