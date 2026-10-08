package downloader

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"testing"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/content/memory"
)

const testPolicy = `package main

deny contains msg if {
	input.kind == "Deployment"
	msg := "no deployments"
}
`

func TestExtractBundleLayers(t *testing.T) {
	archive := bundleArchive(t, map[string]string{
		"/.manifest":         `{"revision":"1"}`,
		"/policy/main.rego":  testPolicy,
		"/policy/data.json":  `{"allowed":["a"]}`,
		"/lib/helpers.rego":  "package lib\n",
		"/nested/a/b/c.rego": "package nested\n",
	})

	dst := pullLayers(t, layer{mediaType: ocispec.MediaTypeImageLayerGzip, title: "bundle.tar.gz", data: archive})

	for name, want := range map[string]string{
		".manifest":         `{"revision":"1"}`,
		"policy/main.rego":  testPolicy,
		"policy/data.json":  `{"allowed":["a"]}`,
		"lib/helpers.rego":  "package lib\n",
		"nested/a/b/c.rego": "package nested\n",
	} {
		got, err := os.ReadFile(filepath.Join(dst, filepath.FromSlash(name)))
		if err != nil {
			t.Fatalf("read extracted file %s: %v", name, err)
		}
		if string(got) != want {
			t.Errorf("extracted file %s = %q, want %q", name, got, want)
		}
	}

	if _, err := os.Stat(filepath.Join(dst, "bundle.tar.gz")); !os.IsNotExist(err) {
		t.Errorf("bundle archive should be removed after extraction, stat err: %v", err)
	}
}

func TestExtractBundleLayersIgnoresOtherLayers(t *testing.T) {
	dst := pullLayers(t, layer{
		mediaType: "application/vnd.cncf.openpolicyagent.policy.layer.v1+rego",
		title:     "main.rego",
		data:      []byte(testPolicy),
	})

	got, err := os.ReadFile(filepath.Join(dst, "main.rego"))
	if err != nil {
		t.Fatalf("read pulled policy: %v", err)
	}
	if string(got) != testPolicy {
		t.Errorf("pulled policy = %q, want %q", got, testPolicy)
	}
}

func TestExtractBundleLayersRejectsPathsOutsideDestination(t *testing.T) {
	for _, name := range []string{"../escape.rego", "/../../escape.rego", "policy/../../escape.rego"} {
		t.Run(name, func(t *testing.T) {
			archive := bundleArchive(t, map[string]string{name: testPolicy})

			ctx := context.Background()
			parent := t.TempDir()
			dst := filepath.Join(parent, "policy")
			if err := os.Mkdir(dst, 0o755); err != nil {
				t.Fatal(err)
			}

			store, root := copyLayers(ctx, t, dst, layer{mediaType: ocispec.MediaTypeImageLayerGzip, title: "bundle.tar.gz", data: archive})
			if err := extractBundleLayers(ctx, store, root, dst); err == nil {
				t.Fatal("expected an error for a bundle file outside of the destination")
			}

			if _, err := os.Stat(filepath.Join(parent, "escape.rego")); !os.IsNotExist(err) {
				t.Errorf("file was written outside of the destination, stat err: %v", err)
			}
		})
	}
}

type layer struct {
	mediaType string
	title     string
	data      []byte
}

// pullLayers copies an image with the given layers into a fresh directory the
// same way OCIGetter does, extracts any bundle layers, and returns the
// directory.
func pullLayers(t *testing.T, layers ...layer) string {
	t.Helper()

	ctx := context.Background()
	dst := t.TempDir()
	store, root := copyLayers(ctx, t, dst, layers...)
	if err := extractBundleLayers(ctx, store, root, dst); err != nil {
		t.Fatalf("extract bundle layers: %v", err)
	}

	return dst
}

func copyLayers(ctx context.Context, t *testing.T, dst string, layers ...layer) (*file.Store, ocispec.Descriptor) {
	t.Helper()

	src := memory.New()
	descs := make([]ocispec.Descriptor, 0, len(layers))
	for _, l := range layers {
		desc := content.NewDescriptorFromBytes(l.mediaType, l.data)
		desc.Annotations = map[string]string{ocispec.AnnotationTitle: l.title}
		if err := src.Push(ctx, desc, bytes.NewReader(l.data)); err != nil {
			t.Fatalf("push layer: %v", err)
		}
		descs = append(descs, desc)
	}

	config := []byte("{}")
	configDesc := content.NewDescriptorFromBytes(ocispec.MediaTypeImageConfig, config)
	if err := src.Push(ctx, configDesc, bytes.NewReader(config)); err != nil {
		t.Fatalf("push config: %v", err)
	}

	manifest, err := oras.PackManifest(ctx, src, oras.PackManifestVersion1_1, "", oras.PackManifestOptions{
		Layers:           descs,
		ConfigDescriptor: &configDesc,
	})
	if err != nil {
		t.Fatalf("pack manifest: %v", err)
	}
	if err := src.Tag(ctx, manifest, "latest"); err != nil {
		t.Fatalf("tag manifest: %v", err)
	}

	store, err := file.New(dst)
	if err != nil {
		t.Fatalf("file store: %v", err)
	}
	t.Cleanup(func() { store.Close() })

	root, err := oras.Copy(ctx, src, "latest", store, "", oras.DefaultCopyOptions)
	if err != nil {
		t.Fatalf("copy: %v", err)
	}

	return store, root
}

func bundleArchive(t *testing.T, files map[string]string) []byte {
	t.Helper()

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	for name, contents := range files {
		header := &tar.Header{
			Name:     name,
			Mode:     0o600,
			Size:     int64(len(contents)),
			Typeflag: tar.TypeReg,
		}
		if err := tw.WriteHeader(header); err != nil {
			t.Fatalf("write tar header: %v", err)
		}
		if _, err := tw.Write([]byte(contents)); err != nil {
			t.Fatalf("write tar contents: %v", err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("close tar writer: %v", err)
	}
	if err := gw.Close(); err != nil {
		t.Fatalf("close gzip writer: %v", err)
	}

	return buf.Bytes()
}
