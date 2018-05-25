// Author hoenig

package vaultapi

import (
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
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

func Test_TokenRole(t *testing.T) {
	roleName := "provisioner-role"
	clientWithPerm := getClient(t, rootTokener)
	clientWithoutPerm := getClient(t, renewableTokener)
	roleOpts := TokenRoleOptions{
		Name:               roleName,
		AllowedPolicies:    "p1,p2",
		DisallowedPolicies: "p3,p4",
		Orphan:             true,
		Period:             "10s",
		Renewable:          true,
		ExplicitMaxTTL:     12,
		PathSuffix:         "suffix",
		BoundCIDRs:         []string{"10,0,0,0/8"},
	}

	// Delete the role, in case it exists
	require.NoError(t, clientWithPerm.DeleteTokenRole(roleOpts.Name))

	// Can't create role without permission
	require.Error(t, clientWithoutPerm.CreateTokenRole(roleOpts))
	_, err := clientWithPerm.LookupTokenRole(roleOpts.Name)
	require.Equal(t, ErrPathNotFound, errors.Cause(err))

	// Can create role with permission
	require.NoError(t, clientWithPerm.CreateTokenRole(roleOpts))
	lookedUpTokenRole, err := clientWithPerm.LookupTokenRole(roleOpts.Name)
	require.NoError(t, err)

	// Can't look up or delete the role without permission
	_, err = clientWithoutPerm.LookupTokenRole(roleOpts.Name)
	require.Error(t, err)
	require.Error(t, clientWithoutPerm.DeleteTokenRole(roleOpts.Name))

	// Check that the correct role details came back
	require.Equal(t, roleOpts.AllowedPolicies, strings.Join(lookedUpTokenRole.AllowedPolicies, ","))
	require.Equal(t, roleOpts.DisallowedPolicies, strings.Join(lookedUpTokenRole.DisallowedPolicies, ","))
	require.Equal(t, roleOpts.ExplicitMaxTTL, lookedUpTokenRole.ExplicitMaxTTL)
	require.Equal(t, roleOpts.Name, lookedUpTokenRole.Name)
	require.Equal(t, roleOpts.Orphan, lookedUpTokenRole.Orphan)
	require.Equal(t, roleOpts.PathSuffix, lookedUpTokenRole.PathSuffix)
	period, err := time.ParseDuration(roleOpts.Period)
	require.NoError(t, err)
	require.Equal(t, int(period.Seconds()), lookedUpTokenRole.Period)
	require.Equal(t, roleOpts.Renewable, lookedUpTokenRole.Renewable)

	// Delete the role
	require.NoError(t, clientWithPerm.DeleteTokenRole(roleOpts.Name))

	// Make sure it's really gone
	_, err = clientWithPerm.LookupTokenRole(roleOpts.Name)
	require.Equal(t, ErrPathNotFound, errors.Cause(err))
}
