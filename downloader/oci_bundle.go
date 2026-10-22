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

const (
	dockerManifestMediaType     = "application/vnd.docker.distribution.manifest.v2+json"
	dockerManifestListMediaType = "application/vnd.docker.distribution.manifest.list.v2+json"

	// maxIndexDepth bounds how far extractBundleLayers descends through
	// nested image indexes, matching OPA's own OCI downloader.
	maxIndexDepth = 4
)

// extractBundleLayers unpacks the OPA bundles found in the image described
// by root into dst.
//
// OPA bundles pushed to a registry (for example with "opa build" followed by
// "oras push") store the bundle tarball as a single layer with the
// "application/vnd.oci.image.layer.v1.tar+gzip" media type, the same layer
// type OPA's own OCI downloader looks for. The ORAS file store writes such a
// layer to disk as-is, which leaves only the tarball in the policy directory.
// This unpacks each of those tarballs into dst and removes the tarball, so
// the policies and data in the bundle can be loaded like any other files.
// When root is an image index, every manifest it references is inspected.
//
// Layers that ORAS already unpacked itself (directories pushed with ORAS)
// and layers with other media types, such as the ones written by
// "conftest push", are left untouched.
func extractBundleLayers(ctx context.Context, store content.Fetcher, root ocispec.Descriptor, dst string) error {
	return extractBundleLayersDepth(ctx, store, root, dst, 0, map[string]struct{}{})
}

// extractBundleLayersDepth walks desc, descending into image indexes. visited
// holds the digests already handled: a layer referenced by several manifests
// of an index is only on disk once, so it must only be extracted once.
func extractBundleLayersDepth(ctx context.Context, store content.Fetcher, desc ocispec.Descriptor, dst string, depth int, visited map[string]struct{}) error {
	if _, ok := visited[desc.Digest.String()]; ok {
		return nil
	}
	visited[desc.Digest.String()] = struct{}{}

	switch desc.MediaType {
	case ocispec.MediaTypeImageIndex, dockerManifestListMediaType:
		if depth >= maxIndexDepth {
			return fmt.Errorf("image index %s is nested more than %d levels deep", desc.Digest, maxIndexDepth)
		}

		manifests, err := content.Successors(ctx, store, desc)
		if err != nil {
			return fmt.Errorf("fetch image index: %w", err)
		}

		for _, manifest := range manifests {
			if err := extractBundleLayersDepth(ctx, store, manifest, dst, depth+1, visited); err != nil {
				return err
			}
		}

		return nil

	case ocispec.MediaTypeImageManifest, dockerManifestMediaType:
		return extractManifestBundleLayers(ctx, store, desc, dst, visited)

	default:
		return nil
	}
}

func extractManifestBundleLayers(ctx context.Context, store content.Fetcher, desc ocispec.Descriptor, dst string, visited map[string]struct{}) error {
	manifestBytes, err := content.FetchAll(ctx, store, desc)
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

		if _, ok := visited[layer.Digest.String()]; ok {
			continue
		}
		visited[layer.Digest.String()] = struct{}{}

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
