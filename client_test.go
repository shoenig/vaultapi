// Author hoenig

package vaultapi

import (
	"github.com/stretchr/testify/require"

	"io/ioutil"
	"log"
	"os"
	"testing"
)

// Running these tests requires running the vault defined in the hack/docker-compose.yaml file.
// With docker and docker-compose installed, simply run "docker-compose up" in the hack directory
// to get setup to run these tests.

func cleanup(t *testing.T, client Client) {
	t.Log("-- cleaning up vault keyspace --")
	//keys, err := client.Keys("/")
	//require.NoError(t, err)
	//
	////for _, key := range keys {
	//// err := client.Delete(key)
	////require.NoError(t, err)
	////}
}

// a tokener that reads the token from /tmp/dev-vault.token
func devTokener(t *testing.T) Tokener {
	bs, err := ioutil.ReadFile("/tmp/dev-vault.token")
	require.NoError(t, err)
	token := string(bs)
	return NewStaticToken(token)
}

func devOpts() ClientOptions {
	return ClientOptions{
		Servers:             []string{"http://localhost:8200"},
		SkipTLSVerification: true,
		Logger:              log.New(os.Stdout, "[vaultapi] ", log.LstdFlags),
	}
}
