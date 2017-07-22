// Author hoenig

package vaultapi

import (
	"github.com/stretchr/testify/require"

	"testing"
)

func Test_Client_KV(t *testing.T) {
	opts := devOpts()
	client, err := New(opts, rootTokener())
	require.NoError(t, err)
	defer cleanup(t, client)

	// no keys initially (404)
	_, err = client.Keys("/")
	require.Error(t, err)

	err = client.Put("/foo/bar", "baz")
	require.NoError(t, err)

	value, err := client.Get("/foo/bar")
	require.NoError(t, err)
	t.Log("value:", value)
	require.Equal(t, "baz", value)

	keys, err := client.Keys("/")
	require.NoError(t, err)
	require.Equal(t, 1, len(keys))

	_, err = client.Get("/noexist")
	require.Error(t, err)

	err = client.Put("/alpha", "beta")
	require.NoError(t, err)

	value, err = client.Get("/alpha")
	require.NoError(t, err)
	require.Equal(t, "beta", value)
	t.Log("value:", value)

	err = client.Delete("/alpha")
	require.NoError(t, err)

	_, err = client.Get("/alpha")
	t.Log("del error:", err)
	require.Error(t, err)
}
