// Author hoenig

package vaultapi

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_AuthToken(t *testing.T) {
	client := getClient(t, rootTokener)
	opts := TokenOptions{
		Policies:    []string{"default"},
		Orphan:      true,
		Renewable:   false,
		DisplayName: "test-token1",
		MaxUses:     1000,
		MaxTTL:      1 * time.Hour,
	}
	token, err := client.CreateToken(opts)
	require.NoError(t, err)
	require.Equal(t, 36, len(token.ID))
	t.Log("created token:", token.ID)
	t.Log("policies:", token.Policies)

	lookedUp, err := client.LookupToken(token.ID)
	require.NoError(t, err)
	t.Log("token lookup:", lookedUp)

	selfLookedUp, err := client.LookupSelfToken()
	require.NoError(t, err)
	t.Log("self token lookup:", selfLookedUp)
}

func Test_Renew_NonRenewable(t *testing.T) {
	client := getClient(t, nonRenewableTokener)
	token, err := nonRenewableTokener().Token()
	require.NoError(t, err)
	lookedUp, err := client.LookupSelfToken()
	require.NoError(t, err)
	t.Log("non-renewable max ttl:", lookedUp.MaxTTL)
	t.Log("non-renewable ttl:", lookedUp.TTL)

	// this token does not have permission to use the
	// general token renewal endpoint
	_, err = client.RenewToken(token, 1*time.Second)
	require.Error(t, err)

	// this token was not created from a role that
	// supports creating rewnewable tokens, and thus
	// cannot be used to renew itself - however, vault
	// does not return an error for trying to self renew
	// a token that is not renewable
	selfRenewed, err := client.RenewSelfToken(1 * time.Second)
	require.NoError(t, err)
	t.Log("non-renewable self token renewed lease duration:", selfRenewed.LeaseDuration)
}

func Test_Renew_Renewable(t *testing.T) {
	client := getClient(t, renewableTokener)
	token, err := renewableTokener().Token()
	require.NoError(t, err)
	lookedUp, err := client.LookupSelfToken()
	require.NoError(t, err)
	t.Log("renewable max ttl:", lookedUp.MaxTTL)
	t.Log("renewable ttl:", lookedUp.TTL)

	// this token does not have permission to use the
	// general token renewal endpoint
	_, err = client.RenewToken(token, 1*time.Second)
	require.Error(t, err)

	selfRenewed, err := client.RenewSelfToken(1 * time.Second)
	require.NoError(t, err)
	t.Log("renewable self token renewed lease duration:", selfRenewed.LeaseDuration)
}
