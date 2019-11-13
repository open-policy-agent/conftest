package push

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	auth "github.com/deislabs/oras/pkg/auth/docker"
	"github.com/deislabs/oras/pkg/content"
	"github.com/deislabs/oras/pkg/oras"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
)

const (
	openPolicyAgentConfigMediaType      = "application/vnd.cncf.openpolicyagent.config.v1+json"
	openPolicyAgentPolicyLayerMediaType = "application/vnd.cncf.openpolicyagent.policy.layer.v1+rego"
	openPolicyAgentDataLayerMediaType   = "application/vnd.cncf.openpolicyagent.data.layer.v1+json"
)

// NewPushCommand creates a new push command which allows users to push
// bundles to an OCI registry
func NewPushCommand(ctx context.Context, logger *log.Logger) *cobra.Command {
	cmd := cobra.Command{
		Use:   "push <repository> [filepath]",
		Short: "Upload OPA bundles to an OCI registry",
		Long:  `Upload Open Policy Agent bundles to an OCI registry`,
		Args:  cobra.RangeArgs(1, 2),

		RunE: func(cmd *cobra.Command, args []string) error {
			var path string
			if len(args) == 2 {
				path = args[1]
			} else {
				var err error
				path, err = os.Getwd()
				if err != nil {
					return fmt.Errorf("get working directory: %w", err)
				}
			}

			repository := args[0]

			logger.Printf("pushing bundle to: %s\n", repository)
			manifest, err := pushBundle(ctx, repository, path)
			if err != nil {
				return fmt.Errorf("push bundle: %w", err)
			}
			logger.Printf("pushed bundle with digest: %s\n", manifest.Digest)

			return nil
		},
	}

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
	layers, err := buildLayers(memoryStore, path)
	if err != nil {
		return nil, fmt.Errorf("building layers: %w", err)
	}

	var repositoryWithTag string
	if strings.Contains(repository, ":") {
		repositoryWithTag = repository
	} else {
		repositoryWithTag = repository + ":latest"
	}

	extraOpts := []oras.PushOpt{oras.WithConfigMediaType(openPolicyAgentConfigMediaType)}
	manifest, err := oras.Push(ctx, resolver, repositoryWithTag, memoryStore, layers, extraOpts...)
	if err != nil {
		return nil, fmt.Errorf("pushing manifest: %w", err)
	}

	return &manifest, nil
}

func buildLayers(memoryStore *content.Memorystore, path string) ([]ocispec.Descriptor, error) {
	root, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("get abs path: %w", err)
	}

	var policy []string
	var data []string
	err = filepath.Walk(root, func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walk path: %w", err)
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(currentPath) == ".rego" {
			policy = append(policy, currentPath)
		}

		if filepath.Ext(currentPath) == ".json" {
			data = append(data, currentPath)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	policyLayers, err := buildLayer(policy, root, memoryStore, openPolicyAgentPolicyLayerMediaType)
	if err != nil {
		return nil, fmt.Errorf("build policy layer: %w", err)
	}

	dataLayers, err := buildLayer(data, root, memoryStore, openPolicyAgentDataLayerMediaType)
	if err != nil {
		return nil, fmt.Errorf("build data layer: %w", err)
	}

	layers := append(policyLayers, dataLayers...)
	return layers, nil
}

func buildLayer(paths []string, root string, memoryStore *content.Memorystore, mediaType string) ([]ocispec.Descriptor, error) {
	var layer ocispec.Descriptor
	var layers []ocispec.Descriptor

	for _, file := range paths {
		contents, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("read file: %w", err)
		}

		relative, err := filepath.Rel(root, file)
		if err != nil {
			return nil, fmt.Errorf("get relative filepath: %w", err)
		}

		path := filepath.ToSlash(relative)

		layer = memoryStore.Add(path, mediaType, contents)
		layers = append(layers, layer)
	}

	return layers, nil
}
