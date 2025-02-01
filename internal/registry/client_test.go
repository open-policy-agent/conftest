package registry

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/open-policy-agent/conftest/internal/network"
	"github.com/spf13/viper"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

func mustParseReference(ref string) *remote.Repository {
	r, err := remote.NewRepository(ref)

	if err == nil {
		return r
	}

	panic(fmt.Sprintf("Unable to parse reference: %s: %v", ref, err))
}

func TestSetupClient(t *testing.T) {
	cases := []struct {
		ref          string
		tlsViaParams bool
		plaintext    bool
		credentials  auth.Credential
	}{
		{ref: "localhost/test:tag", plaintext: true},
		{ref: "127.0.0.1/test:tag", plaintext: true},
		{ref: "127.0.0.1.nip.io/test:tag", plaintext: true},
		{ref: "localhost/test:tag", tlsViaParams: true},
		{ref: "127.0.0.1/test:tag", tlsViaParams: true},
		{ref: "test.127.0.0.1.nip.io/test:tag", tlsViaParams: true},
		{ref: "test.registry.io/test:tag", tlsViaParams: true, credentials: auth.Credential{Username: "test", Password: "supersecret"}},
	}

	for _, c := range cases {
		t.Run(c.ref, func(t *testing.T) {
			repository := mustParseReference(c.ref)
			if c.tlsViaParams {
				viper.Set("tls", true)
				t.Cleanup(func() {
					viper.Set("tls", false)
				})
			}

			t.Setenv("DOCKER_CONFIG", ".")

			if err := SetupClient(repository); err != nil {
				t.Fatal(err)
			}

			if repository.PlainHTTP != c.plaintext {
				t.Errorf(`expecting repository.PlainHTTP == %v, but it was %v`, c.plaintext, repository.PlainHTTP)
			}

			client, ok := repository.Client.(*auth.Client)
			if !ok {
				t.Error("expecting repository.Client to be instance of `auth.Client`")
			}

			if got := client.Header.Get("User-Agent"); got != "conftest" {
				t.Errorf(`expecting client.Header['User-Agent'] == "conftest", but it was %v`, got)
			}

			if got, err := client.Credential(context.Background(), network.Hostname(c.ref)); err == nil {
				if got != c.credentials {
					t.Errorf(`expecting client.Credential(_, "%v") == "%v", but it was %v`, c.ref, c.credentials, got)
				}
			} else {
				t.Errorf(`unexpected error while fetching credentials: %v`, err)
			}
		})
	}
}

func TestSetupClientCredentialsError(t *testing.T) {
	tmpDir := t.TempDir()
	configFilePath := filepath.Join(tmpDir, "config.json")

	// Write file with no permissions
	err := os.WriteFile(configFilePath, []byte(`{"auths":{}}`), 0o000)
	if err != nil {
		t.Fatalf("failed to write to config file: %v", err)
	}

	// Ensure permissions are restored for cleanup
	t.Cleanup(func() {
		if err := os.Chmod(configFilePath, 0o600); err != nil {
			t.Errorf("failed to restore permissions during cleanup: %v", err)
		}
	})

	repository := mustParseReference("local-test-registry/image:tag")
	t.Setenv("DOCKER_CONFIG", configFilePath)

	if err := SetupClient(repository); err == nil {
		t.Error("expected error when credentials store initialization fails, got nil")
	}
}
