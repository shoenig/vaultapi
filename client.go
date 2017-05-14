// Author hoenig

package vaultapi

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/shoenig/toolkit"
)

const (
	headerVaultToken  = "X-Vault-Token"
	headerContentType = "Content-Type"
	mimeJSON          = "application/json"
	mimeText          = "text/plain"
)

type Client interface {
	KV
	Sys
}

type ClientOptions struct {
	Servers             []string
	HTTPTimeout         time.Duration
	SkipTLSVerification bool
	Logger              *log.Logger
}

func New(opts ClientOptions, tokener Tokener) (Client, error) {
	if opts.Logger == nil {
		opts.Logger = log.New(ioutil.Discard, "", 0)
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: opts.SkipTLSVerification,
		},
	}

	return &client{
		opts:    opts,
		tokener: tokener,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   opts.HTTPTimeout,
		},
	}, nil
}

type client struct {
	opts ClientOptions

	tokener    Tokener
	httpClient *http.Client
}

func (c *client) token() (string, error) {
	// the tokener is responsible for locking
	// its own token, whatever that means
	return c.tokener.Token()
}

func fixup(prefix, path string, params ...[2]string) string {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	values := make(url.Values)

	for _, param := range params {
		if param[1] != "" {
			values.Add(param[0], param[1])
		}
	}

	query := values.Encode()

	url := prefix + path
	if len(query) > 0 {
		url += "?" + query
	}
	return url
}

func (c *client) get(path string, i interface{}) error {
	for _, address := range c.opts.Servers {
		if err := c.singleGet(address, path, i); err != nil {
			c.opts.Logger.Printf("GET request failed: %v", err)
		} else {
			return nil
		}
	}
	return errors.Errorf("all attempts for GET request failed to: %v", c.opts.Servers)
}

func (c *client) singleGet(address, path string, i interface{}) error {
	url := address + path

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return errors.Wrapf(err, "failed to build GET request to %q", url)
	}

	token, err := c.token()
	if err != nil {
		return errors.Wrap(err, "failed to get token for request")
	}

	request.Header.Set(headerVaultToken, token)
	request.Header.Set(headerContentType, mimeText)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return errors.Wrapf(err, "failed to execute GET request to %q", url)
	}

	defer toolkit.Drain(response.Body)

	if response.StatusCode >= 400 {
		return errors.Errorf("bad status code: %d, url: %s", response.StatusCode, url)
	}

	if err := json.NewDecoder(response.Body).Decode(i); err != nil {
		return errors.Wrapf(err, "failed to read response from %q", url)
	}

	return nil
}

func (c *client) post(path, body string, i interface{}) error {
	for _, address := range c.opts.Servers {
		if err := c.singlePost(address, path, body, i); err != nil {
			c.opts.Logger.Printf("POST request failed: %v", err)
		} else {
			return nil
		}
	}
	return errors.Errorf("all attempts for POST request failed to: %v", c.opts.Servers)
}

func (c *client) singlePost(address, path, body string, i interface{}) error {
	url := address + path

	request, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		return errors.Wrapf(err, "failed to build POST request to %q", url)
	}

	token, err := c.token()
	if err != nil {
		return errors.Wrap(err, "failed to get token for request")
	}

	request.Header.Set(headerVaultToken, token)
	request.Header.Set(headerContentType, mimeJSON)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return errors.Wrapf(err, "failed to execute POST request to %q", url)
	}

	if response.StatusCode >= 400 {
		return errors.Errorf("bad status code: %d, url: %s", response.StatusCode, url)
	}

	if i != nil {
		// only read response if we passed a thing to read it into
		defer toolkit.Drain(response.Body)
		if err := json.NewDecoder(response.Body).Decode(i); err != nil {
			return errors.Wrapf(err, "failed to read response from %q", url)
		}
	}

	return nil
}

func (c *client) put(path, body string) error {
	for _, address := range c.opts.Servers {
		if err := c.singlePut(address, path, body); err != nil {
			c.opts.Logger.Printf("PUT request failed: %v", err)
		} else {
			return nil
		}
	}
	return errors.Errorf("all attempts for PUT request failed to: %v", c.opts.Servers)
}

func (c *client) singlePut(address, path, body string) error {
	url := address + path

	request, err := http.NewRequest(http.MethodPut, url, strings.NewReader(body))
	if err != nil {
		return errors.Wrapf(err, "failed to build PUT request to %q", url)
	}

	token, err := c.token()
	if err != nil {
		return errors.Wrap(err, "failed to get token for request")
	}

	request.Header.Set(headerVaultToken, token)
	request.Header.Set(headerContentType, mimeJSON)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return errors.Wrapf(err, "failed to execute PUT request to %q", url)
	}

	// do not read response

	if response.StatusCode >= 400 {
		return errors.Errorf("bad status code: %d, url: %s", response.StatusCode, url)
	}

	return nil
}

// we have to implement recursion ourselves - which will
// be the case for paths that end in a trailing slash
// see: https://github.com/hashicorp/vault/issues/885
func (c *client) delete(path string) error {
	c.opts.Logger.Printf("delete %q", path)
	noprefix := strings.TrimPrefix(path, "/v1/secret")

	// recursively descend if this path is a directory
	if strings.HasSuffix(path, "/") {
		keys, err := c.Keys(noprefix)
		if err != nil {
			c.opts.Logger.Printf("delete recursion error: %v", err)
			return err
		}
		c.opts.Logger.Print("recursive keys:", keys)
		// call delete on every key under this path
		for _, subpath := range keys {
			if err := c.delete(path + subpath); err != nil {
				return err
			}
		}

		return nil
	} else {
		// base case: actually delete this path, which is a concrete
		// key and not a directory
		c.opts.Logger.Printf("delete concrete path: %q", path)
		return c.deleteKey(path)
	}
}

func (c *client) deleteKey(path string) error {
	for _, address := range c.opts.Servers {
		if err := c.singleDelete(address, path); err != nil {
			c.opts.Logger.Printf("DELETE request failed: %v", err)
		} else {
			return nil
		}
	}
	return errors.Errorf("all attempts for DELETE request failed to: %v", c.opts.Servers)
}

func (c *client) singleDelete(address, path string) error {
	url := address + path
	c.opts.Logger.Printf("delete url: %q", url)

	request, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	token, err := c.token()
	if err != nil {
		return errors.Wrap(err, "failed to get token for request")
	}

	request.Header.Set(headerVaultToken, token)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	// do not read response

	if response.StatusCode >= 400 {
		return errors.Errorf("bad status code: %d", response.StatusCode)
	}
	c.opts.Logger.Printf("delete status code: %d", response.StatusCode)

	return nil
}
