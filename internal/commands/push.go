package commands

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/open-policy-agent/conftest/policy"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"oras.land/oras-go/pkg/auth"
	dockerauth "oras.land/oras-go/pkg/auth/docker"
	"oras.land/oras-go/pkg/content"
	orascontext "oras.land/oras-go/pkg/context"
	"oras.land/oras-go/pkg/oras"
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
			if err := viper.BindPFlag("data", cmd.Flags().Lookup("data")); err != nil {
				return fmt.Errorf("bind flag: %w", err)
			}
			return nil
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				cmd.Usage() //nolint
				return fmt.Errorf("missing required arguments")
			}

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
			policyPath := viper.GetString("policy")
			dataPath := viper.GetString("data")
			if dataPath == "" {
				dataPath = policyPath
			}
			manifest, err := pushBundle(orascontext.Background(), repository, policyPath, dataPath)
			if err != nil {
				return fmt.Errorf("push bundle: %w", err)
			}

			logger.Printf("pushed bundle with digest: %s", manifest.Digest)
			return nil
		},
	}

	cmd.Flags().StringP("policy", "p", "policy", "Directory to push as a bundle")
	cmd.Flags().StringP("data", "d", "", "Directory containing data to include in the bundle")

	return &cmd
}

func pushBundle(ctx context.Context, repository, policyPath, dataPath string) (*ocispec.Descriptor, error) {
	cli, err := dockerauth.NewClient()
	if err != nil {
		return nil, fmt.Errorf("get auth client: %w", err)
	}
	opts := []auth.ResolverOption{auth.WithResolverClient(http.DefaultClient)}
	resolver, err := cli.ResolverWithOpts(opts...)
	if err != nil {
		return nil, fmt.Errorf("docker resolver: %w", err)
	}
	registry := content.Registry{Resolver: resolver}

	memoryStore := content.NewMemory()
	layers, err := buildLayers(ctx, memoryStore, policyPath, dataPath)
	if err != nil {
		return nil, fmt.Errorf("building layers: %w", err)
	}
	manifestData, manifest, cfgData, cfg, err := content.GenerateManifestAndConfig(nil, nil, layers...)
	if err != nil {
		return nil, fmt.Errorf("generate manifest: %w", err)
	}
	memoryStore.Set(cfg, cfgData)
	err = memoryStore.StoreManifest(repository, manifest, manifestData)
	if err != nil {
		return nil, fmt.Errorf("store manifest: %w", err)
	}

	_, err = oras.Copy(ctx, memoryStore, repository, registry, "")
	if err != nil {
		return nil, fmt.Errorf("pushing manifest: %w", err)
	}

	return &manifest, nil
}

func buildLayers(ctx context.Context, memoryStore *content.Memory, policyPath, dataPath string) ([]ocispec.Descriptor, error) {
	engine, err := policy.LoadWithData(ctx, []string{policyPath}, []string{dataPath}, "")
	if err != nil {
		return nil, fmt.Errorf("load: %w", err)
	}

	var layers []ocispec.Descriptor
	for path, contents := range engine.Policies() {
		desc, err := memoryStore.Add(path, openPolicyAgentPolicyLayerMediaType, []byte(contents))
		if err != nil {
			return nil, fmt.Errorf("add policy layer to store: %w", err)
		}
		layers = append(layers, desc)
	}

	for path, contents := range engine.Documents() {
		desc, err := memoryStore.Add(path, openPolicyAgentDataLayerMediaType, []byte(contents))
		if err != nil {
			return nil, fmt.Errorf("add data layer to store: %w", err)
		}
		layers = append(layers, desc)
	}

	return layers, nil
}
