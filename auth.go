// Author hoenig

package vaultapi

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/pkg/errors"
)

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

type Auth interface {
	ListAccessors() ([]string, error)
	CreateToken(opts TokenOptions) (CreatedToken, error)
	LookupToken(id string) (LookedUpToken, error)
}

func (c *client) ListAccessors() ([]string, error) {
	var m map[string]map[string][]string
	if err := c.list("/auth/tokens/accessors", &m); err != nil {
		return nil, errors.Wrap(err, "failed to list accessors")
	}
	accessors := m["data"]["keys"]
	sort.Strings(accessors)
	return accessors, nil
}

type createdToken struct {
	Data CreatedToken `json:"auth"`
}

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

type LookedUpToken struct {
	CreatedToken
	Path string `json:"path"`
}

func (c *client) LookupToken(id string) (LookedUpToken, error) {
	var tok LookedUpToken
	if err := c.get("/v1/auth/token/lookup/"+id, &tok); err != nil {
		// do not provide token id anywhere
		return LookedUpToken{}, errors.Wrapf(err, "failed to lookup token")
	}
	return tok, nil
}
