package policy

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/containerd/containerd/log"
	auth "github.com/deislabs/oras/pkg/auth/docker"
	"github.com/deislabs/oras/pkg/content"
	"github.com/deislabs/oras/pkg/oras"
	getter "github.com/hashicorp/go-getter"
)

type OCIGetter struct {
	client *getter.Client
}

func (g *OCIGetter) ClientMode(u *url.URL) (getter.ClientMode, error) {
	return getter.ClientModeDir, nil
}

func (g *OCIGetter) Get(path string, u *url.URL) error {
	ctx := g.Context()

	if !pathContainsTag(u.Path) {
		u.Path = u.Path + ":latest"
	}

	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return fmt.Errorf("make policy directory: %w", err)
	}

	cli, err := auth.NewClient()
	if err != nil {
		return fmt.Errorf("new auth client: %w", err)
	}

	resolver, err := cli.Resolver(ctx, http.DefaultClient, false)
	if err != nil {
		return fmt.Errorf("new resolver: %w", err)
	}

	fileStore := content.NewFileStore(path)
	defer fileStore.Close()

	repository := u.Host + u.Path
	log.G(ctx).Infof("Downloading: %s\n", repository)
	_, _, err = oras.Pull(ctx, resolver, repository, fileStore)
	if err != nil {
		return fmt.Errorf("pulling policy: %w", err)
	}

	return nil
}

func (g *OCIGetter) GetFile(dst string, u *url.URL) error {
	// NOOP for now
	return nil
}

//These methods are normally handled by the base getter in go-getter but
// the base getter is not exported
func (g *OCIGetter) SetClient(c *getter.Client) { g.client = c }

// Context tries to returns the Contex from the getter's
// client. otherwise context.Background() is returned.
func (g *OCIGetter) Context() context.Context {
	if g == nil || g.client == nil {
		return context.Background()
	}
	return g.client.Ctx
}
