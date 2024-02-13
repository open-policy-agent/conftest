package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/open-policy-agent/conftest/internal/registry"
	"github.com/open-policy-agent/conftest/policy"
	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry/remote"
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
be stored in compatible OCI registries. Currently Open Policy Agent bundles are supported by
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
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			if err := viper.BindPFlag("policy", cmd.Flags().Lookup("policy")); err != nil {
				return fmt.Errorf("bind flag: %w", err)
			}
			if err := viper.BindPFlag("data", cmd.Flags().Lookup("data")); err != nil {
				return fmt.Errorf("bind flag: %w", err)
			}
			if err := viper.BindPFlag("tls", cmd.Flags().Lookup("tls")); err != nil {
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
			if policyPath == "" && dataPath == "" {
				return errors.New("either policy or data must be set")
			}
			if dataPath == "" {
				dataPath = policyPath
			}
			manifest, err := pushBundle(ctx, repository, policyPath, dataPath)
			if err != nil {
				return fmt.Errorf("push bundle: %w", err)
			}

			logger.Printf("pushed bundle with digest: %s", manifest.Digest)
			return nil
		},
	}

	cmd.Flags().StringP("policy", "p", "policy", "Directory to push as a bundle")
	cmd.Flags().StringP("data", "d", "", "Directory containing data to include in the bundle, defaults to the value of the policy flag")
	cmd.Flags().BoolP("tls", "s", true, "Use TLS to access the registry")

	return &cmd
}

func pushBundle(ctx context.Context, repository, policyPath, dataPath string) (*ocispec.Descriptor, error) {
	dest, err := remote.NewRepository(repository)
	if err != nil {
		return nil, fmt.Errorf("constructing repository: %w", err)
	}

	if err := registry.SetupClient(dest); err != nil {
		return nil, fmt.Errorf("setting up the registry client: %w", err)
	}

	layers, err := pushLayers(ctx, dest, policyPath, dataPath)
	if err != nil {
		return nil, fmt.Errorf("pushing layers: %w", err)
	}

	configBytes := []byte("{}")
	configDesc := content.NewDescriptorFromBytes(oras.MediaTypeUnknownConfig, configBytes)
	if err != nil {
		return nil, fmt.Errorf("serializing manifest conifg: %w", err)
	}

	if err := dest.Push(ctx, configDesc, bytes.NewReader(configBytes)); err != nil && !errors.Is(err, errdef.ErrAlreadyExists) {
		return nil, fmt.Errorf("pushing manifest conifg: %w", err)
	}

	manifest := ocispec.Manifest{
		Config:    configDesc,
		Layers:    layers,
		Versioned: specs.Versioned{SchemaVersion: 2},
	}
	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		return nil, fmt.Errorf("serializing manifest: %w", err)
	}

	manifestDesc := content.NewDescriptorFromBytes(ocispec.MediaTypeImageManifest, manifestBytes)
	if err := dest.Push(ctx, manifestDesc, bytes.NewReader(manifestBytes)); err != nil && !errors.Is(err, errdef.ErrAlreadyExists) {
		return nil, fmt.Errorf("pushing manifest conifg: %w", err)
	}

	afterLastSlash := repository[strings.LastIndex(repository, "/")+1:]
	tag := afterLastSlash[strings.Index(afterLastSlash, ":")+1:]

	if err := dest.Tag(ctx, manifestDesc, tag); err != nil && !errors.Is(err, errdef.ErrAlreadyExists) {
		return nil, fmt.Errorf("tagging: %w", err)
	}

	return &manifestDesc, nil
}

func pushLayers(ctx context.Context, pusher content.Pusher, policyPath, dataPath string) ([]ocispec.Descriptor, error) {
	var policyPaths []string
	if policyPath != "" {
		policyPaths = append(policyPaths, policyPath)
	}
	var dataPaths []string
	if dataPath != "" {
		dataPaths = append(dataPaths, dataPath)
	}
	engine, err := policy.LoadWithData(policyPaths, dataPaths, "", false)
	if err != nil {
		return nil, fmt.Errorf("load: %w", err)
	}

	var layers []ocispec.Descriptor
	for path, contents := range engine.Policies() {
		data := []byte(contents)
		desc := content.NewDescriptorFromBytes(openPolicyAgentPolicyLayerMediaType, data)
		desc.Annotations = map[string]string{
			ocispec.AnnotationTitle: path,
		}
		if err := pusher.Push(ctx, desc, bytes.NewReader(data)); err != nil && !errors.Is(err, errdef.ErrAlreadyExists) {
			return nil, fmt.Errorf("pushing policy layer: %w", err)
		}
		layers = append(layers, desc)
	}

	for path, contents := range engine.Documents() {
		data := []byte(contents)
		desc := content.NewDescriptorFromBytes(openPolicyAgentDataLayerMediaType, data)
		desc.Annotations = map[string]string{
			ocispec.AnnotationTitle: path,
		}
		if err := pusher.Push(ctx, desc, bytes.NewReader(data)); err != nil && !errors.Is(err, errdef.ErrAlreadyExists) {
			return nil, fmt.Errorf("pushing data layer: %w", err)
		}
		layers = append(layers, desc)
	}

	return layers, nil
}
