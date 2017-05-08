// Author hoenig

package vaultapi

import (
	"encoding/json"
	"sort"

	"github.com/pkg/errors"
)

type Sys interface {
	AccessorCapabilities(path, accessor string) ([]string, error)
	TokenCapabilities(path, token string) ([]string, error)
	SelfCapabilities(path string) ([]string, error)

	Health() (Health, error)
	Leader() (Leader, error)

	ListMounts() (Mounts, error)
}

type capabilities struct {
	Capabilities []string `json:"capabilities"`
}

func (c *client) TokenCapabilities(path, token string) ([]string, error) {
	bs, err := json.Marshal(struct {
		Path  string `json:"path"`
		Token string `json:"token"`
	}{Path: path, Token: token})
	if err != nil {
		return nil, err
	}
	var caps capabilities
	if err := c.post("/v1/sys/capabilities", string(bs), &caps); err != nil {
		return nil, errors.Wrap(err, "failed to read token capabilities")
	}
	sort.Strings(caps.Capabilities)
	return caps.Capabilities, nil
}

func (c *client) AccessorCapabilities(path, accessor string) ([]string, error) {
	bs, err := json.Marshal(struct {
		Path     string `json:"path"`
		Accessor string `json:"accessor"`
	}{Path: path, Accessor: accessor})
	if err != nil {
		return nil, err
	}
	var caps capabilities
	if err := c.post("/v1/sys/capabilities-accessor", string(bs), &caps); err != nil {
		return nil, errors.Wrap(err, "failed to read accessor capabilities")
	}
	sort.Strings(caps.Capabilities)
	return caps.Capabilities, nil
}

func (c *client) SelfCapabilities(path string) ([]string, error) {
	bs, err := json.Marshal(struct {
		Path string `json:"path"`
	}{Path: path})
	if err != nil {
		return nil, err
	}
	var caps capabilities
	if err := c.post("/v1/sys/capabilities-self", string(bs), &caps); err != nil {
		return nil, errors.Wrap(err, "failed to read self token capabilities")
	}
	sort.Strings(caps.Capabilities)
	return caps.Capabilities, nil
}

type Health struct {
	Initialized   bool   `json:"initialized"`
	Sealed        bool   `json:"sealed"`
	Standby       bool   `json:"standby"`
	ServerTimeUTC int    `json:"server_time_utc"`
	Version       string `json:"version"`
	ClusterName   string `json:"cluster_name"`
	ClusterID     string `json:"cluster_id"`
}

func (c *client) Health() (Health, error) {
	var health Health
	if err := c.get("/v1/sys/health", &health); err != nil {
		return Health{}, errors.Wrap(err, "failed to read health")
	}
	return health, nil
}

type Leader struct {
	HAEnabled     bool   `json:"ha_enabled"`
	IsSelf        bool   `json:"is_self"`
	LeaderAddress string `json:"leader_address"`
}

func (c *client) Leader() (Leader, error) {
	var leader Leader
	if err := c.get("/v1/sys/leader", &leader); err != nil {
		return Leader{}, errors.Wrap(err, "failed to read leader")
	}
	return leader, nil
}

type mountsWrapper struct {
	Data Mounts `json:"data"`
}

type Mounts map[string]struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Config      struct {
		DefaultLeaseTTL int  `json:"default_lease_ttl"`
		MaxLeaseTTL     int  `json:"max_lease_ttl"`
		ForceNoCache    bool `json:"force_no_cache"`
	} `json:"config"`
}

func (c *client) ListMounts() (Mounts, error) {
	// documentation is incorrect, must use the data field
	// to get to mount information
	var wrapper mountsWrapper
	if err := c.get("/v1/sys/mounts", &wrapper); err != nil {
		return nil, errors.Wrap(err, "failed to read mounts")
	}
	return wrapper.Data, nil
}
