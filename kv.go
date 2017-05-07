// Author hoenig

package vaultapi

import (
	"fmt"
	"sort"

	"github.com/pkg/errors"
)

type KV interface {
	Get(path string) (string, error)
	Put(path, value string) error
	Delete(path string) error
	Keys(path string) ([]string, error)
}

type keysData struct {
	Data map[string][]string `json:"data"`
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

func (c *client) Get(path string) (string, error) {
	fullpath := fixup("/v1/secret", path, [2]string{"list", "false"})
	var data keyData
	err := c.get(fullpath, &data)
	if err != nil {
		return "", err
	}

	value, exists := data.Data["value"]
	if !exists {
		return "", errors.Errorf("no value at path %q", fullpath)
	}

	return value, nil
}

func (c *client) Put(path, value string) error {
	fullpath := fixup("/v1/secret", path, [2]string{})
	body := fmt.Sprintf(`{%q:%q}`, "value", value)
	return c.post(fullpath, body)
}

func (c *client) Delete(path string) error {
	fullpath := fixup("/v1/secret", path, [2]string{})
	return c.delete(fullpath)
}
