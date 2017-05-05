// Author hoenig

package vaultapi

import (
	"crypto/tls"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/shoenig/toolkit"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client interface {
}

type ClientOptions struct {
	Servers             []string
	HTTPTimeout         time.Duration
	SkipTLSVerification bool
	Logger              *log.Logger
}

type discard struct{}

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
	opts       ClientOptions
	tokener    Tokener
	httpClient *http.Client
}

// params are url param kv pairs
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
			c.opts.Logger.Println("[get] request failed: %v", err)
		} else {
			return nil
		}
	}   
	return errors.Errorf("[get] all get requests failed to: %v", c.opts.Servers)
}

func (c *client) singleGet(address, path string, i interface{}) error {
	url := address + path
	response, err := c.httpClient.Get(url)
	if err != nil {
		return err
	}
	defer toolkit.Drain(response.Body)

	if response.StatusCode >= 400 {
		return errors.Errorf("bad status code: %d", response.StatusCode)
	}

	return json.NewDecoder(response.Body).Decode(i)
}
