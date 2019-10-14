package policy

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/containerd/containerd/log"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/deislabs/oras/pkg/content"
	"github.com/deislabs/oras/pkg/oras"
	"github.com/spf13/viper"

	auth "github.com/deislabs/oras/pkg/auth/docker"
)

// Policy represents a policy
type Policy struct {
	Repository string
	Tag        string
}

// DownloadPolicy downloads the given policies
func DownloadPolicy(ctx context.Context, policies []Policy) {
	policyDir := filepath.Join(".", viper.GetString("policy"))
	err := os.MkdirAll(policyDir, os.ModePerm)
	if err != nil {
		log.G(ctx).Warnf("Error creating policy directory %q: %v\n", policyDir, err)
	}

	cli, err := auth.NewClient()
	if err != nil {
		log.G(ctx).Warnf("Error loading auth file: %v\n", err)
	}

	resolver, err := cli.Resolver(ctx)
	if err != nil {
		log.G(ctx).Warnf("Error loading resolver: %v\n", err)
		resolver = docker.NewResolver(docker.ResolverOptions{})
	}

	fileStore := content.NewFileStore(policyDir)
	defer fileStore.Close()

	for _, policy := range policies {
		repository := getRepositoryFromPolicy(policy)

		log.G(ctx).Infof("Downloading: %s\n", repository)
		_, _, err = oras.Pull(ctx, resolver, repository, fileStore)
		if err != nil {
			log.G(ctx).Fatalf("Downloading policy failed: %v\n", err)
		}
	}
}

func getRepositoryFromPolicy(policy Policy) string {
	var repository string
	if repositoryContainsTag(policy.Repository) {
		repository = policy.Repository
	} else if policy.Tag == "" {
		repository = policy.Repository + ":latest"
	} else {
		repository = policy.Repository + ":" + policy.Tag
	}

	return repository
}

func repositoryContainsTag(repository string) bool {
	split := strings.Split(repository, "/")
	return strings.Contains(split[len(split)-1], ":")
}
