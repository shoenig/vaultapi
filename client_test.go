// Author hoenig

package vaultapi

import (
	"github.com/stretchr/testify/require"
	"log"
	"os"
	"testing"
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

// a tokener that reads the token from /tmp/dev-vault.token
func devTokener(t *testing.T) Tokener {
	return NewFileToken("/tmp/dev-vault.token")
}

func devOpts() ClientOptions {
	return ClientOptions{
		Servers:             []string{"http://localhost:8200"},
		SkipTLSVerification: true,
		Logger:              log.New(os.Stdout, "[vaultapi] ", log.LstdFlags),
	}
}
