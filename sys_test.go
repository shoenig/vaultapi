// Author hoenig

package vaultapi

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type tokenerFunc func() Tokener

func getClient(t *testing.T, f tokenerFunc) Client {
	opts := devOpts()
	tokener := f()
	client, err := New(opts, tokener)
	require.NoError(t, err)
	return client
}

func Test_Client_TokenCapabilities(t *testing.T) {
	client := getClient(t, rootTokener)
	token, err := rootTokener().Token()
	require.NoError(t, err)
	caps, err := client.TokenCapabilities("/", token)
	require.NoError(t, err)
	require.Equal(t, "root", caps[0])
}

func Test_Client_SelfCapabilities(t *testing.T) {
	client := getClient(t, rootTokener)
	caps, err := client.SelfCapabilities("/")
	require.NoError(t, err)
	require.Equal(t, "root", caps[0])
}

func Test_Client_Health(t *testing.T) {
	client := getClient(t, rootTokener)
	health, err := client.Health()
	require.NoError(t, err)
	require.False(t, health.Sealed)
	t.Log("health:", health)
}

func Test_Client_Leader(t *testing.T) {
	client := getClient(t, rootTokener)
	leader, err := client.Leader()
	require.NoError(t, err)
	require.False(t, leader.HAEnabled)
	// no content if not in ha, of course
}

func Test_Client_ListMounts(t *testing.T) {
	client := getClient(t, rootTokener)
	mounts, err := client.ListMounts()
	require.NoError(t, err)
	require.Equal(t, "per-token private secret storage", mounts["cubbyhole/"].Description)
	require.Equal(t, "generic secret storage", mounts["secret/"].Description)
}

const pol1 = `
# Allow a token to manage secret/foo/bar/* (no deletes)
path "secret/foo/bar/*" {
	capabilities = ["create", "read", "update", "list"]
}`

func Test_Client_Policies(t *testing.T) {
	client := getClient(t, rootTokener)
	policies, err := client.ListPolicies()
	require.NoError(t, err)
	t.Log("listPolicies:", policies)
	require.Equal(t, 3, len(policies))

	content, err := client.GetPolicy("default")
	require.NoError(t, err)
	require.Contains(t, content, "# Allow tokens to look up their own properties")

	err = client.SetPolicy("foobar", pol1)
	require.NoError(t, err)

	content, err = client.GetPolicy("foobar")
	require.NoError(t, err)
	t.Log("foobar policy content:", content)
	require.Contains(t, content, pol1)

	err = client.DeletePolicy("foobar")
	require.NoError(t, err)

	_, err = client.GetPolicy("foobar")
	require.Error(t, err)
}

func Test_Client_SealStatus(t *testing.T) {
	client := getClient(t, rootTokener)
	status, err := client.SealStatus()
	require.NoError(t, err)
	require.Equal(t, 1, status.Shares)
	require.False(t, status.Sealed)
}

func Test_Client_StepDown(t *testing.T) {
	client := getClient(t, rootTokener)
	err := client.StepDown()
	t.Log("step down error:", err)
}
