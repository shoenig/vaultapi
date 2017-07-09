// Author hoenig

package vaultapi

import (
	"encoding/json"
	"sort"

	"github.com/pkg/errors"
)

// A Sys represents the vault system interface.
//
// For more information about the system backend, visit:
// https://www.vaultproject.io/api/system/index.html.
type Sys interface {
	// Capabilities
	AccessorCapabilities(path, accessor string) ([]string, error)
	TokenCapabilities(path, token string) ([]string, error)
	SelfCapabilities(path string) ([]string, error)

	// Leases
	LookupLease(id string) (Lease, error)

	// Policies
	ListPolicies() ([]string, error)
	GetPolicy(name string) (string, error)
	SetPolicy(name, content string) error
	DeletePolicy(name string) error

	// Vault Status
	Health() (Health, error)
	Leader() (Leader, error)
	StepDown() error
	SealStatus() (SealStatus, error)
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
		return nil, errors.Wrapf(err, "failed to read token capabilities for %q at %q", token, path)
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
		return nil, errors.Wrapf(err, "failed to read accessor capabilities for %q at %q", accessor, path)
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
		return nil, errors.Wrapf(err, "failed to read self token capabilities for %q", path)
	}
	sort.Strings(caps.Capabilities)
	return caps.Capabilities, nil
}

// A Lease is a piece of meta data around something in vault
// which may be designed to expire at some time. A common example
// is that every non-root token is associated with a lease which
// indicates a TTL. Once the lease expires, that token is no longer
// valid and cannot be used to authenticate with vault.
type Lease struct {
	ID              string `json:"id"`
	IssueTime       string `json:"issue_time"`
	ExpireTime      string `json:"expire_time"`
	LastRenewalTime string `json:"last_renewal_time"`
	Renewable       bool   `json:"renewable"`
	TTL             int    `json:"ttl"`
}

func (c *client) LookupLease(id string) (Lease, error) {
	bs, err := json.Marshal(struct {
		ID string `json:"lease_id"`
	}{ID: id})
	if err != nil {
		return Lease{}, err
	}
	var lease Lease
	if err := c.put("/v1/sys/leases/lookup", string(bs)); err != nil {
		return Lease{}, errors.Wrapf(err, "failed to lookup lease for %q", id)
	}
	return lease, nil
}

// A Health is returned upon requesting health status from vault
// and contains some metadata about the vault configuartion.
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

// A Leader is returned upon requesting the leader from vault.
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

func (c *client) StepDown() error {
	if err := c.put("/v1/sys/step-down", ""); err != nil {
		return errors.Wrap(err, "failed to step down")
	}
	return nil
}

type mountsWrapper struct {
	Data Mounts `json:"data"`
}

// Mounts contains information about the mounts currently configured
// with vault. Generally, users of the vaultapi client library are
// likely only concerned with the default backends that come with vault,
// and will not need to be concerned with this API.
//
// More information can be found here:
// https://www.vaultproject.io/docs/internals/architecture.html
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

type listPolicies struct {
	Policies []string `json:"policies"`
}

func (c *client) ListPolicies() ([]string, error) {
	var pols listPolicies
	if err := c.get("/v1/sys/policy", &pols); err != nil {
		return nil, errors.Wrap(err, "failed to list listPolicies")
	}
	sort.Strings(pols.Policies)
	return pols.Policies, nil
}

type getPolicy struct {
	Rules string `json:"rules"`
}

func (c *client) GetPolicy(name string) (string, error) {
	var pol getPolicy
	if err := c.get("/v1/sys/policy/"+name, &pol); err != nil {
		return "", errors.Wrapf(err, "failed to get policy %q", name)
	}
	return pol.Rules, nil
}

func (c *client) SetPolicy(name, content string) error {
	bs, err := json.Marshal(getPolicy{
		Rules: content,
	})

	if err != nil {
		return errors.Wrapf(err, "failed to create json for setting policy %q", name)
	}

	if err := c.put("/v1/sys/policy/"+name, string(bs)); err != nil {
		return errors.Wrapf(err, "failed to set policy %q", name)
	}

	return nil
}

func (c *client) DeletePolicy(name string) error {
	if err := c.delete("/v1/sys/policy/" + name); err != nil {
		return errors.Wrapf(err, "failed to delete policy %q", name)
	}
	return nil
}

// A SealStatus is returned upon requesting the sealed status from
// vault.
//
// More information about the vault seal mechanism can be found here:
// https://www.vaultproject.io/docs/concepts/seal.html.
type SealStatus struct {
	Sealed      bool   `json:"sealed"`
	Threshold   int    `json:"t"`
	Shares      int    `json:"n"`
	Progress    int    `json:"progress"`
	Version     string `json:"version"`
	ClusterName string `json:"cluster_name"`
	ClusterID   string `json:"cluster_id"`
}

func (c *client) SealStatus() (SealStatus, error) {
	var ss SealStatus
	if err := c.get("/v1/sys/seal-status", &ss); err != nil {
		return SealStatus{}, errors.Wrap(err, "failed to get sealed status")
	}
	return ss, nil
}
