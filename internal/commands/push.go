package commands

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	auth "github.com/deislabs/oras/pkg/auth/docker"
	"github.com/deislabs/oras/pkg/content"
	orascontext "github.com/deislabs/oras/pkg/context"
	"github.com/deislabs/oras/pkg/oras"
	"github.com/open-policy-agent/conftest/policy"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const pushDesc = `
This command uploads Open Policy Agent bundles to an OCI registry

Storing policies in OCI registries is similar to how Docker containers are stored.
With Conftest, Rego policies are bundled and pushed to the OCI registry e.g.:

	$ conftest push instrumenta.azurecr.io/my-registry

Optionally, a tag can be specified, e.g.:

	$ conftest push instrumenta.azurecr.io/my-registry:v1

Optionally, specific directory can be passed as a second argument, e.g.:

	$ conftest push instrumenta.azurecr.io/my-registry:v1 path/to/dir

Conftest leverages the ORAS library under the hood. This allows arbitrary artifacts to 
be stored in compatible OCI registries. Currently open policy agent bundles are supported by 
the docker/distribution (https://github.com/docker/distribution) registry and by Azure.

The policy location defaults to the policy directory in the local folder.
The location can be overridden with the '--policy' flag, e.g.:

	$ conftest push --policy <my-directory> url
`

const (
	openPolicyAgentConfigMediaType      = "application/vnd.cncf.openpolicyagent.config.v1+json"
	openPolicyAgentPolicyLayerMediaType = "application/vnd.cncf.openpolicyagent.policy.layer.v1+rego"
	openPolicyAgentDataLayerMediaType   = "application/vnd.cncf.openpolicyagent.data.layer.v1+json"
)

// NewPushCommand creates a new push command which allows users to push
// bundles to an OCI registry.
func NewPushCommand(ctx context.Context, logger *log.Logger) *cobra.Command {
	cmd := cobra.Command{
		Use:   "push <repository>",
		Short: "Push OPA bundles to an OCI registry",
		Long:  pushDesc,
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

			repository := args[0]
			if !strings.Contains(repository, "/") {
				return errors.New("destination url missing repository")
			}

			// At the moment, push only supports pushing to OCI registries
			// which makes the oci: prefix redundant and has been known to
			// cause issues.
			repository = strings.ReplaceAll(repository, "oci://", "")

			// When the destination repository to push to does not contain a
			// tag, append the latest tag so the bundle is not pushed without
			// a tag.
			pathParts := strings.Split(repository, "/")
			lastPathPart := pathParts[len(pathParts)-1]
			if !strings.Contains(lastPathPart, ":") {
				repository = repository + ":latest"
			}

			logger.Printf("pushing bundle to: %s", repository)
			manifest, err := pushBundle(ctx, repository, viper.GetString("policy"))
			if err != nil {
				return fmt.Errorf("push bundle: %w", err)
			}

			logger.Printf("pushed bundle with digest: %s", manifest.Digest)
			return nil
		},
	}

	cmd.Flags().StringP("policy", "p", "policy", "Directory to push as a bundle")

	return &cmd
}

func pushBundle(ctx context.Context, repository string, path string) (*ocispec.Descriptor, error) {
	cli, err := auth.NewClient()
	if err != nil {
		return nil, fmt.Errorf("get auth client: %w", err)
	}

	resolver, err := cli.Resolver(ctx, http.DefaultClient, false)
	if err != nil {
		return nil, fmt.Errorf("docker resolver: %w", err)
	}

	memoryStore := content.NewMemoryStore()
	layers, err := buildLayers(ctx, memoryStore, path)
	if err != nil {
		return nil, fmt.Errorf("building layers: %w", err)
	}

	extraOpts := []oras.PushOpt{oras.WithConfigMediaType(openPolicyAgentConfigMediaType)}
	manifest, err := oras.Push(ctx, resolver, repository, memoryStore, layers, extraOpts...)
	if err != nil {
		return nil, fmt.Errorf("pushing manifest: %w", err)
	}

	return &manifest, nil
}

func buildLayers(ctx context.Context, memoryStore *content.Memorystore, path string) ([]ocispec.Descriptor, error) {
	engine, err := policy.LoadWithData(ctx, []string{path}, []string{path})
	if err != nil {
		return nil, fmt.Errorf("load: %w", err)
	}

	var layers []ocispec.Descriptor
	for path, contents := range engine.Policies() {
		layers = append(layers, memoryStore.Add(path, openPolicyAgentPolicyLayerMediaType, []byte(contents)))
	}

	for path, contents := range engine.Documents() {
		layers = append(layers, memoryStore.Add(path, openPolicyAgentDataLayerMediaType, []byte(contents)))
	}

	return layers, nil
}
