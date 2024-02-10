package registry

import (
	"context"

	"github.com/cpuguy83/dockercfg"
	"github.com/open-policy-agent/conftest/internal/network"
	"github.com/spf13/viper"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

func SetupClient(repository *remote.Repository) {
	registry := repository.Reference.Host()

	// If `--tls=false` was provided or accessing the registry via loopback with
	// `--tls` flag was not provided
	if !viper.GetBool("tls") || (network.IsLoopback(network.Hostname(registry)) && !viper.IsSet("tls")) {
		// Docker by default accesses localhost using plaintext HTTP
		repository.PlainHTTP = true
	}

	client := auth.DefaultClient
	client.SetUserAgent("conftest")
	client.Credential = func(_ context.Context, registry string) (auth.Credential, error) {
		host := dockercfg.ResolveRegistryHost(registry)
		username, password, err := dockercfg.GetRegistryCredentials(host)
		if err != nil {
			return auth.EmptyCredential, err
		}

		return auth.Credential{
			Username: username,
			Password: password,
		}, nil
	}

	repository.Client = client
}
