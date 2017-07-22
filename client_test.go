// Author hoenig

package vaultapi

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// Running these tests requires running the vault defined in the hack/docker-compose.yaml file.
// With docker and docker-compose installed, simply run "docker-compose up" in the hack directory
// to get setup to run these tests.

func cleanup(t *testing.T, client Client) {
	t.Log("-- cleaning up vault keyspace --")

	// show keys at root before cleanup
	keys, err := client.Keys("/")
	require.NoError(t, err) // this will fail if no tests were run
	t.Log("cleanup will recursively delete keys:", keys)

	err = client.Delete("/")
	t.Logf("error of deleting root key: %v", err)
	require.NoError(t, err)

	// assert no keys to list after cleaning up
	_, err = client.Keys("/")
	require.Error(t, err)
}

// a tokener for the root token to the dev vault
func rootTokener() Tokener {
	return NewFileToken("/tmp/root.token")
}

func nonRenewableTokener() Tokener {
	return NewFileToken("/tmp/t1.token")
}

func renewableTokener() Tokener {
	return NewFileToken("/tmp/t2.token")
}

func devOpts() ClientOptions {
	return ClientOptions{
		Servers:             []string{"http://localhost:8200"},
		SkipTLSVerification: true,
		Logger:              log.New(os.Stdout, "[vaultapi] ", log.LstdFlags),
	}
}
