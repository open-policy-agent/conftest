package downloader

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"oras.land/oras-go/pkg/auth"
	dockerauth "oras.land/oras-go/pkg/auth/docker"

	getter "github.com/hashicorp/go-getter"
	"oras.land/oras-go/pkg/content"
	"oras.land/oras-go/pkg/oras"
)

// OCIGetter is responsible for handling OCI repositories
type OCIGetter struct {
	client *getter.Client
}

// ClientMode returns the client mode directory
func (g *OCIGetter) ClientMode(u *url.URL) (getter.ClientMode, error) {
	return getter.ClientModeDir, nil
}

// Get gets the repository as the specified url
func (g *OCIGetter) Get(path string, u *url.URL) error {
	ctx := g.Context()

	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return fmt.Errorf("make policy directory: %w", err)
	}

	cli, err := dockerauth.NewClient()
	if err != nil {
		return fmt.Errorf("new auth client: %w", err)
	}

	opts := []auth.ResolverOption{auth.WithResolverClient(http.DefaultClient)}
	resolver, err := cli.ResolverWithOpts(opts...)
	if err != nil {
		return fmt.Errorf("docker resolver: %w", err)
	}

	registry := content.Registry{Resolver: resolver}

	fileStore := content.NewFile(path)
	defer fileStore.Close()

	repository := getRepositoryFromURL(u.Path)
	pullURL := u.Host + repository

	_, err = oras.Copy(ctx, registry, pullURL, fileStore, "")
	if err != nil {
		return fmt.Errorf("pulling policy: %w", err)
	}

	return nil
}

// GetFile is currently a NOOP
func (g *OCIGetter) GetFile(dst string, u *url.URL) error {
	return nil
}

// SetClient sets the client for the OCIGetter
// NOTE: These methods are normally handled by the base getter in go-getter but
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
