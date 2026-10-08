package downloader

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/open-policy-agent/opa/v1/bundle"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/file"
)

// extractBundleLayers unpacks the OPA bundles found in the manifest described
// by root into dst.
//
// OPA bundles pushed to a registry (for example with "opa build" followed by
// "oras push") store the bundle tarball as a single layer with the
// "application/vnd.oci.image.layer.v1.tar+gzip" media type, the same layer
// type OPA's own OCI downloader looks for. The ORAS file store writes such a
// layer to disk as-is, which leaves only the tarball in the policy directory.
// This unpacks each of those tarballs next to it and removes the tarball, so
// the policies and data in the bundle can be loaded like any other files.
//
// Layers that ORAS already unpacked itself (directories pushed with ORAS)
// and layers with other media types, such as the ones written by
// "conftest push", are left untouched.
func extractBundleLayers(ctx context.Context, store content.Fetcher, root ocispec.Descriptor, dst string) error {
	if root.MediaType != ocispec.MediaTypeImageManifest {
		return nil
	}

	manifestBytes, err := content.FetchAll(ctx, store, root)
	if err != nil {
		return fmt.Errorf("fetch manifest: %w", err)
	}

	var manifest ocispec.Manifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return fmt.Errorf("parse manifest: %w", err)
	}

	for _, layer := range manifest.Layers {
		if layer.MediaType != ocispec.MediaTypeImageLayerGzip {
			continue
		}

		name := layer.Annotations[ocispec.AnnotationTitle]
		if name == "" || layer.Annotations[file.AnnotationUnpack] == "true" {
			continue
		}

		tarball := filepath.Join(dst, name)
		if err := extractBundle(tarball, dst); err != nil {
			return fmt.Errorf("extract bundle %s: %w", name, err)
		}

		if err := os.Remove(tarball); err != nil {
			return fmt.Errorf("remove bundle archive: %w", err)
		}
	}

	return nil
}

// extractBundle writes the files of the gzipped bundle tarball at path into
// dst. Every file must resolve to a location inside dst.
func extractBundle(path string, dst string) error {
	archive, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open archive: %w", err)
	}
	defer archive.Close()

	root, err := os.OpenRoot(dst)
	if err != nil {
		return fmt.Errorf("open destination: %w", err)
	}
	defer root.Close()

	loader := bundle.NewTarballLoaderWithBaseURL(archive, "").
		WithSizeLimitBytes(bundle.DefaultSizeLimitBytes)

	for {
		f, err := loader.NextFile()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read archive: %w", err)
		}

		if err := writeBundleFile(root, f); err != nil {
			return err
		}
	}
}

func writeBundleFile(root *os.Root, f *bundle.Descriptor) error {
	defer f.Close()

	// Bundle paths are slash separated and usually rooted, e.g. "/main.rego".
	name := filepath.FromSlash(strings.TrimLeft(f.Path(), "/"))
	if !filepath.IsLocal(name) {
		return fmt.Errorf("bundle file %q is outside of the destination directory", f.Path())
	}

	if err := root.MkdirAll(filepath.Dir(name), os.ModePerm); err != nil {
		return fmt.Errorf("create directory for %s: %w", name, err)
	}

	out, err := root.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("create %s: %w", name, err)
	}
	defer out.Close()

	// The loader already rejected files larger than the size limit, so
	// reading up to the limit copies the whole file.
	if _, err := f.Read(out, bundle.DefaultSizeLimitBytes); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("write %s: %w", name, err)
	}

	return out.Close()
}
