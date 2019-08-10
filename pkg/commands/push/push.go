package push

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/instrumenta/conftest/pkg/constants"

	"github.com/containerd/containerd/log"
	"github.com/containerd/containerd/remotes/docker"
	auth "github.com/deislabs/oras/pkg/auth/docker"
	"github.com/deislabs/oras/pkg/content"
	"github.com/deislabs/oras/pkg/oras"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
)

// NewPushCommand creates a new push command
func NewPushCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "push <repository> [filepath]",
		Short: "Upload OPA bundles to an OCI registry",
		Long:  `Upload Open Policy Agent bundles to an OCI registry`,
		Args:  cobra.RangeArgs(1, 2),

		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			var path string
			if len(args) == 2 {
				path = args[1]
			} else {
				var err error
				path, err = os.Getwd()
				if err != nil {
					log.G(ctx).Fatal(err)
				}
			}

			uploadBundle(ctx, args[0], path)
		},
	}

	return cmd
}

func uploadBundle(ctx context.Context, repository string, root string) {

	cli, err := auth.NewClient()
	if err != nil {
		log.G(ctx).Warnf("Error loading auth file: %v\n", err)
	}

	resolver, err := cli.Resolver(ctx)
	if err != nil {
		log.G(ctx).Warnf("Error loading resolver: %v\n", err)
		resolver = docker.NewResolver(docker.ResolverOptions{})
	}

	var ref string
	if strings.Contains(repository, ":") {
		ref = repository
	} else {
		ref = repository + ":latest"
	}

	layers, memoryStore := buildLayers(ctx, root)

	log.G(ctx).Infof("Pushing bundle to %s\n", ref)
	extraOpts := []oras.PushOpt{oras.WithConfigMediaType(constants.OpenPolicyAgentConfigMediaType)}

	manifest, err := oras.Push(ctx, resolver, ref, memoryStore, layers, extraOpts...)
	if err != nil {
		log.G(ctx).Fatal(err)
	}

	log.G(ctx).Infof("Pushed bundle to %s with digest %s\n", ref, manifest.Digest)
}

func buildLayers(ctx context.Context, root string) ([]ocispec.Descriptor, *content.Memorystore) {
	var data []string
	var policy []string
	var layers []ocispec.Descriptor
	var err error

	root, err = filepath.Abs(root)
	if err != nil {
		log.G(ctx).Fatal(err)
	}

	info, err := os.Stat(root)
	if err != nil {
		log.G(ctx).Fatal(err)
	}

	if !info.IsDir() {
		log.G(ctx).Fatalf("%s isn't a directory", root)
	}

	memoryStore := content.NewMemoryStore()

	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".rego" {
			policy = append(policy, path)
		}
		if filepath.Ext(path) == ".json" {
			data = append(data, path)
		}
		return nil
	})
	if err != nil {
		log.G(ctx).Fatal(err)
	}

	policyLayers := buildLayer(ctx, policy, root, memoryStore, constants.OpenPolicyAgentPolicyLayerMediaType)
	dataLayers := buildLayer(ctx, data, root, memoryStore, constants.OpenPolicyAgentDataLayerMediaType)
	layers = append(policyLayers, dataLayers...)

	return layers, memoryStore
}

func buildLayer(ctx context.Context, paths []string, root string, memoryStore *content.Memorystore, mediaType string) []ocispec.Descriptor {
	var layer ocispec.Descriptor
	var layers []ocispec.Descriptor
	for _, file := range paths {
		contents, err := ioutil.ReadFile(file)
		if err != nil {
			log.G(ctx).Fatal(err)
		}
		relative, err := filepath.Rel(root, file)
		if err != nil {
			log.G(ctx).Fatal(err)
		}

		path := filepath.ToSlash(relative)

		layer = memoryStore.Add(path, constants.OpenPolicyAgentPolicyLayerMediaType, contents)
		layers = append(layers, layer)
	}
	return layers
}
