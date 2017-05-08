// Author hoenig

package vaultapi

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func getClient(t *testing.T) Client {
	opts := devOpts()
	tokener := devTokener(t)
	client, err := New(opts, tokener)
	require.NoError(t, err)
	return client
}

func Test_Client_TokenCapabilities(t *testing.T) {
	client := getClient(t)
	token, err := devTokener(t).Token()
	require.NoError(t, err)
	caps, err := client.TokenCapabilities("/", token)
	require.NoError(t, err)
	require.Equal(t, "root", caps[0])
}

func Test_Client_SelfCapabilities(t *testing.T) {
	client := getClient(t)
	caps, err := client.SelfCapabilities("/")
	require.NoError(t, err)
	require.Equal(t, "root", caps[0])
}

func Test_Client_Health(t *testing.T) {
	client := getClient(t)
	health, err := client.Health()
	require.NoError(t, err)
	require.False(t, health.Sealed)
	t.Log("health:", health)
}

func Test_Client_Leader(t *testing.T) {
	client := getClient(t)
	leader, err := client.Leader()
	require.NoError(t, err)
	require.False(t, leader.HAEnabled)
	// no content if not in ha, of course
}

func Test_Client_ListMounts(t *testing.T) {
	client := getClient(t)
	mounts, err := client.ListMounts()
	require.NoError(t, err)
	t.Log("mounts:", mounts)
}
