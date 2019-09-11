package vaultapi

import (
	"fmt"
	"sort"

	"github.com/pkg/errors"
)

// A KV represents the key-value store built into vault.
//
// Although vault supports arbitrary bytes as keys and values,
// this library assumes all keys and values are strings. This
// helps simplify code for clients, the 99% use case for which
// is writing secret passwords and other stringy information
// into vault for safe keeping.
type KV interface {
	// Get will return the value defined at path.
	Get(path string) (string, error)
	// Put will set value at path.
	Put(path, value string) error
	// Delete will remove the value at path.
	Delete(path string) error
	// Keys will list all of the subpaths under path in asciibetical
	// order. The returned paths may be terminal (ie, the value is
	// stored content) or they may traversable like a directory.
	Keys(path string) ([]string, error)
}

var (
	// ErrNoValue indicates that no value exists for a requested path.
	ErrNoValue = errors.New("no value defined for given path")
)

func (c *client) Get(path string) (string, error) {
	fullpath := fixup("/v1/secret", path, [2]string{"list", "false"})
	var data keyData
	err := c.get(fullpath, &data)
	if err != nil {
		return "", err
	}

	value, exists := data.Data["value"]
	if !exists {
		return "", ErrNoValue
	}

	return value, nil
}

func (c *client) Put(path, value string) error {
	fullpath := fixup("/v1/secret", path, [2]string{})
	body := fmt.Sprintf(`{%q:%q}`, "value", value)
	return c.post(fullpath, body, nil)
}

func (c *client) Delete(path string) error {
	fullpath := fixup("/v1/secret", path, [2]string{})
	return c.delete(fullpath)
}

func (c *client) Keys(path string) ([]string, error) {
	fullpath := fixup("/v1/secret", path, [2]string{"list", "true"})
	var data keysData
	err := c.get(fullpath, &data)
	if err != nil {
		return nil, err
	}
	keys := data.Data["keys"]
	sort.Strings(keys)
	return keys, nil
}

type keyData struct {
	Data map[string]string `json:"data"`
}

type keysData struct {
	Data map[string][]string `json:"data"`
}
