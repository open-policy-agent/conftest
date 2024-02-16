package downloader

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	reg "github.com/open-policy-agent/conftest/internal/registry"

	getter "github.com/hashicorp/go-getter"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
)

// OCIGetter is responsible for handling OCI repositories
type OCIGetter struct {
	client *getter.Client
}

// ClientMode returns the client mode directory
func (g *OCIGetter) ClientMode(_ *url.URL) (getter.ClientMode, error) {
	return getter.ClientModeDir, nil
}

// Get gets the repository as the specified url
func (g *OCIGetter) Get(path string, u *url.URL) error {
	ctx := g.Context()

	repository := ociURL(u)
	ref, err := registry.ParseReference(repository)
	if err != nil {
		return fmt.Errorf("reference: %w", err)
	}

	if ref.Reference == "" {
		ref.Reference = "latest"
		repository = ref.String()
	}

	src, err := remote.NewRepository(repository)
	if err != nil {
		return fmt.Errorf("repository: %w", err)
	}

	if err := reg.SetupClient(src); err != nil {
		return fmt.Errorf("registry client setup: %w", err)
	}

	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return fmt.Errorf("make policy directory: %w", err)
	}

	fileStore, err := file.New(path)
	if err != nil {
		return fmt.Errorf("file store: %w", err)
	}
	defer fileStore.Close()

	_, err = oras.Copy(ctx, src, repository, fileStore, "", oras.DefaultCopyOptions)
	if err != nil {
		return fmt.Errorf("pulling policy: %w", err)
	}

	return nil
}

// GetFile is currently a NOOP
func (g *OCIGetter) GetFile(_ string, _ *url.URL) error {
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

// ociURL returns the string representation of the url that is an acceptable
// OCI URL. In short, it strips off the scheme, e.g. "https://", from the URL.
func ociURL(u *url.URL) string {
	scheme, url, found := strings.Cut(u.String(), "://")
	if !found {
		url = scheme
	}
	return url
}
