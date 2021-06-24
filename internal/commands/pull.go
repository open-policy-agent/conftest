package commands

import (
	"context"
	"fmt"
	"path/filepath"

	orascontext "github.com/deislabs/oras/pkg/context"
	"github.com/open-policy-agent/conftest/downloader"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const pullDesc = `
This command downloads individual policies from a remote location.

Several locations are supported by the pull command. Under the hood
conftest leverages go-getter (https://github.com/hashicorp/go-getter).
The following protocols are supported for downloading policies:

	- OCI Registries
	- Local Files
	- Git
	- HTTP/HTTPS
	- Mercurial
	- Amazon S3
	- Google Cloud Storage

The location of the policies is specified by passing an URL, e.g.:

	$ conftest pull http://<my-policy-url>

Based on the protocol a different mechanism will be used to download the policy.
The pull command will also try to infer the protocol based on the URL if the 
URL does not contain a protocol. For example, the OCI mechanism will be used if
an azure registry URL is passed, e.g.

	$ conftest pull instrumenta.azurecr.io/my-registry

The policy location defaults to the policy directory in the local folder.
The location can be overridden with the '--policy' flag, e.g.:

	$ conftest pull --policy <my-directory> <oci-url>
`

// NewPullCommand creates a new pull command to allow users
// to download individual policies.
func NewPullCommand(ctx context.Context) *cobra.Command {
	cmd := cobra.Command{
		Use:   "pull <repository>",
		Short: "Download individual policies",
		Long:  pullDesc,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlag("policy", cmd.Flags().Lookup("policy")); err != nil {
				return fmt.Errorf("bind flag: %w", err)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				cmd.Usage() //nolint
				return fmt.Errorf("missing required arguments")
			}

			ctx = orascontext.Background()

			policyDir := filepath.Join(".", viper.GetString("policy"))

			if err := downloader.Download(ctx, policyDir, args); err != nil {
				return fmt.Errorf("download policies: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringP("policy", "p", "policy", "Path to download the policies to")

	return &cmd
}
