// Author hoenig

package vaultapi

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
)

// Auth provides a way to manage what may be
// authenticated to vault.
//
// For now, this API
// supports only the token authentication
// mechanism that is built into vault. Support
// for additional types of authentication may
// be added in future releases.
//
// More information about managing tokens via
// the auth backend can be found here:
// https://www.vaultproject.io/docs/auth/token.html
type Auth interface {
	CreateToken(opts TokenOptions) (CreatedToken, error)
	LookupToken(id string) (LookedUpToken, error)
	LookupSelfToken() (LookedUpToken, error)
	RenewToken(id string, increment time.Duration) (RenewedToken, error)
}

// TokenOptions are used to define properties
// of a token being created. More information
// about the different options can be found in
// the token documentation at:
// https://www.vaultproject.io/docs/concepts/tokens.html
type TokenOptions struct {
	Policies        []string      `json:"policies,omitempty"`
	NoDefaultPolicy bool          `json:"no_default_policy,omitempty"`
	Orphan          bool          `json:"no_parent,omitempty"`
	Renewable       bool          `json:"renewable,omitempty"`
	DisplayName     string        `json:"display_name,omitempty"`
	MaxUses         int           `json:"num_uses,omitempty"`
	TTL             time.Duration `json:"ttl,omitempty"`
	MaxTTL          time.Duration `json:"explicit_max_ttl,omitempty"`
	Period          time.Duration `json:"period,omitmempty"`
}

type createdToken struct {
	Data CreatedToken `json:"auth"`
}

// A CreatedToken represents information returned from
// vault after creating a token. The ID attribute is
// the token itself; this is the value used to authenticate
// with vault later on.
type CreatedToken struct {
	ID            string            `json:"client_token"`
	Policies      []string          `json:"policies"`
	Metadata      map[string]string `json:"metadata"`
	LeaseDuration int               `json:"lease_duration"`
	Renewable     bool              `json:"renewable"`
}

func (c *client) CreateToken(opts TokenOptions) (CreatedToken, error) {
	bs, err := json.Marshal(opts)
	if err != nil {
		return CreatedToken{}, err
	}
	tokenRequest := string(bs)
	c.opts.Logger.Printf("token create request: %v", tokenRequest)

	var ct createdToken
	if err := c.post("/v1/auth/token/create", string(bs), &ct); err != nil {
		return CreatedToken{}, err
	}

	if ct.Data.ID == "" {
		// most likely parse errors on our part
		return CreatedToken{}, errors.Errorf("create token returned empty id")
	}

	return ct.Data, nil
}

// A LookedUpToken represents information returned from
// vault after making a request for information about
// a particular token.
type LookedUpToken struct {
	ID           string   `json:"id"`
	Accessor     string   `json:"accessor"`
	CreationTime int      `json:"creation_time"`
	CreationTTL  int      `json:"creation_ttl"`
	DisplayName  string   `json:"display_name"`
	MaxTTL       int      `json:"explicit_max_ttl"`
	NumUses      int      `json:"num_uses"`
	Orphan       bool     `json:"orphan"`
	Path         string   `json:"path"`
	Policies     []string `json:"policies"`
	TTL          int      `json:"ttl"`
}

type lookedUpTokenWrapper struct {
	Data LookedUpToken `json:"data"`
}

type lookupToken struct {
	Token string `json:"token"`
}

func (c *client) LookupToken(id string) (LookedUpToken, error) {
	var tok lookedUpTokenWrapper
	bs, err := json.Marshal(lookupToken{Token: id})
	if err != nil {
		return LookedUpToken{}, err
	}

	if err := c.post("/v1/auth/token/lookup", string(bs), &tok); err != nil {
		// do not provide token id anywhere
		return LookedUpToken{}, errors.Wrapf(err, "failed to lookup token")
	}

	return tok.Data, nil
}

func (c *client) LookupSelfToken() (LookedUpToken, error) {
	var tok lookedUpTokenWrapper
	if err := c.get("/v1/auth/token/lookup-self", &tok); err != nil {
		// do not provide token id anywhere
		return LookedUpToken{}, errors.Wrapf(err, "failed to lookup self token")
	}
	return tok.Data, nil
}

// A RenewedToken represents information returned from
// vault after making a request to renew a periodic
// token.
type RenewedToken struct {
	ClientToken   string   `json:"client_token"`
	Accessor      string   `json:"accessor"`
	Policies      []string `json:"policies"`
	LeaseDuration int      `json:"lease_duration"`
	Renewable     bool     `json:"renewable"`
}

type wrappedRenewedToken struct {
	Auth RenewedToken `json:"auth"`
}

func (c *client) RenewToken(id string, increment time.Duration) (RenewedToken, error) {
	var tok wrappedRenewedToken
	bs, err := json.Marshal(lookupToken{Token: id})
	if err != nil {
		return RenewedToken{}, err
	}

	if err := c.post("/v1/auth/token/renew", string(bs), &tok); err != nil {
		return RenewedToken{}, errors.Wrapf(err, "failed to renew token")
	}

	return tok.Auth, nil
}
