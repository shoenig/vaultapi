// Author hoenig

package vaultapi

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_AuthToken(t *testing.T) {
	client := getClient(t)
	opts := TokenOptions{
		Policies:    []string{"default"},
		Orphan:      true,
		Renewable:   true,
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

	renewed, err := client.RenewToken(token.ID, 1*time.Minute)
	require.NoError(t, err)
	t.Log("renewed:", renewed)
}
