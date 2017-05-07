// Author hoenig

package vaultapi

import (
	"github.com/stretchr/testify/require"

	"testing"
)

func Test_Client_KV(t *testing.T) {
	opts := devOpts()
	tokener := devTokener(t)
	client, err := New(opts, tokener)
	require.NoError(t, err)

	keys, err := client.Keys("/")
	require.NoError(t, err)
	t.Log("keys", keys)

	value, err := client.Get("/foo")
	require.NoError(t, err)
	t.Log("value:", value)

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
