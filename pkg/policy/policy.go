package policy

import (
	"context"
	"os"
	"path/filepath"

	"github.com/instrumenta/conftest/pkg/util"

	"github.com/containerd/containerd/log"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/deislabs/oras/pkg/content"
	"github.com/deislabs/oras/pkg/oras"
	"github.com/spf13/viper"

	auth "github.com/deislabs/oras/pkg/auth/docker"
)

type Policy struct {
	Repository string
	Tag        string
}

func DownloadPolicy(ctx context.Context, policies []Policy) {
	policyDir := filepath.Join(".", viper.GetString("policy"))
	os.MkdirAll(policyDir, os.ModePerm)

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
		var ref string
		if util.RepositoryNameContainsTag(policy.Repository) {
			ref = policy.Repository
		} else if policy.Tag == "" {
			ref = policy.Repository + ":latest"
		} else {
			ref = policy.Repository + ":" + policy.Tag
		}
		log.G(ctx).Infof("Downloading: %s\n", ref)
		_, _, err = oras.Pull(ctx, resolver, ref, fileStore)
		if err != nil {
			log.G(ctx).Fatalf("Downloading policy failed: %v\n", err)
		}
	}
}
