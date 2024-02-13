package registry

import (
	"net/http"

	"github.com/open-policy-agent/conftest/internal/network"
	"github.com/spf13/viper"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/credentials"
	"oras.land/oras-go/v2/registry/remote/retry"
)

func SetupClient(repository *remote.Repository) error {
	registry := repository.Reference.Host()

	// If `--tls=false` was provided or accessing the registry via loopback with
	// `--tls` flag was not provided
	if !viper.GetBool("tls") || (network.IsLoopback(network.Hostname(registry)) && !viper.IsSet("tls")) {
		// Docker by default accesses localhost using plaintext HTTP
		repository.PlainHTTP = true
	}

	httpClient := &http.Client{
		Transport: retry.NewTransport(http.DefaultTransport),
	}

	store, err := credentials.NewStoreFromDocker(credentials.StoreOptions{
		AllowPlaintextPut:        true,
		DetectDefaultNativeStore: true,
	})
	if err != nil {
		return err
	}

	client := &auth.Client{
		Client:     httpClient,
		Credential: credentials.Credential(store),
		Cache:      auth.NewCache(),
	}
	client.SetUserAgent("conftest")

	repository.Client = client

	return nil
}
