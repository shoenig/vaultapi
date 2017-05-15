// Author hoenig

package vaultapi

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_CreateToken(t *testing.T) {
	client := getClient(t)
	opts := TokenOptions{
		Policies:    []string{"default", "herp"},
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
	t.Log("looked up:", lookedUp)
}
