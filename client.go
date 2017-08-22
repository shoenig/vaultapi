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
	methodLIST        = "LIST" // ffs
)

// mocks generated with github.com/vektra/mockery
//go:generate mockery -name Client -case=underscore -outpkg vaultapitest -output vaultapitest

// A Client is used to communicate with vault. The interface is composed of
// other interfaces, which reflect the different categories of API supported
// by the vault server.
type Client interface {
	Auth
	KV
	Sys
}

var (
	// ErrNoServers indicates that a Client was created with
	// no URIs of vault servers to communicate with.
	ErrNoServers = errors.New("no servers were provided")

	// ErrInvalidHTTPTimeout indicates that a negative time.Duration
	// was provided as a value for client HTTP timeouts.
	ErrInvalidHTTPTimeout = errors.New("invalid HTTP timeout")

	// ErrPathNotFound indicates the requested path did not exist.
	ErrPathNotFound = errors.New("requested path not found")
)

// ClientOptions are used to configure options of a Client
// upon creation.
type ClientOptions struct {
	// Servers should be populated with complete URI including transport
	// and port number of each of the vault servers that are running.
	// An example URI: https://127.0.0.1:8200.
	Servers []string

	// HTTPTimeout configures how long underlying HTTP requests should
	// wait before giving up and returning a timeout error. By default,
	// this value is 10 seconds.
	HTTPTimeout time.Duration

	// SkipTLSVerification configures the underlying HTTP client
	// to ignore any TLS certificate validation errors. This is a
	// hacky option that can be used to work around environments that
	// are using self-signed certificates. For best security practices
	// do not use this option in production environments.
	SkipTLSVerification bool

	// Logger may be optionally configured as an output for trace
	// level logging produced by the Client. This can be helpful
	// for debugging logic errors in client code.
	Logger *log.Logger
}

// New creates a new Client that will connect to one or more vault
// servers as specified by opts.Servers. The tokener is used to
// aquire the token to be used to authenticate with vault. If
// opts.Logger is not nil, trace output will be emitted to it which
// can be helpful for debugging an application using the Client.
func New(opts ClientOptions, tokener Tokener) (Client, error) {
	if len(opts.Servers) == 0 {
		return nil, ErrNoServers
	}

	if opts.HTTPTimeout < 0 {
		return nil, ErrInvalidHTTPTimeout
	}

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

	// special case 404, because we need to be able to explicitly identify
	// cases where the requested path was not available.
	if response.StatusCode == http.StatusNotFound {
		return ErrPathNotFound
	}

	if response.StatusCode >= 400 {
		return errors.Errorf("bad status code: %d, url: %s", response.StatusCode, url)
	}

	if err := json.NewDecoder(response.Body).Decode(i); err != nil {
		return errors.Wrapf(err, "failed to read response from %q", url)
	}

	return nil
}

func (c *client) list(path string, i interface{}) error {
	for _, address := range c.opts.Servers {
		if err := c.singleList(address, path, i); err != nil {
			c.opts.Logger.Printf("LIST request failed: %v", err)
		} else {
			return nil
		}
	}
	return errors.Errorf("all attempts for LIST request failed to: %v", c.opts.Servers)
}

func (c *client) singleList(address, path string, i interface{}) error {
	url := address + path

	request, err := http.NewRequest(methodLIST, url, nil)
	if err != nil {
		return errors.Wrapf(err, "failed to build LIST request to: %q", url)
	}

	token, err := c.token()
	if err != nil {
		return errors.Wrap(err, "failed to get token for request")
	}

	request.Header.Set(headerVaultToken, token)
	request.Header.Set(headerContentType, mimeJSON)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return errors.Wrapf(err, "failed to execute LIST request to %q", url)
	}

	// special case 404, because we need to be able to explicitly identify
	// cases where the requested path was not available.
	if response.StatusCode == http.StatusNotFound {
		return ErrPathNotFound
	}

	if response.StatusCode >= 400 {
		return errors.Errorf("bad status code: %d, url: %s", response.StatusCode, url)
	}

	if i != nil {
		// read the response iff we have something to unmarshal it into
		defer toolkit.Drain(response.Body)
		if err := json.NewDecoder(response.Body).Decode(i); err != nil {
			return errors.Wrapf(err, "failed to read response from %q", url)
		}
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

	// special case 404, because we need to be able to explicitly identify
	// cases where the requested path was not available.
	if response.StatusCode == http.StatusNotFound {
		return ErrPathNotFound
	}

	if response.StatusCode >= 400 {
		return errors.Errorf("bad status code: %d, url: %s", response.StatusCode, url)
	}

	if i != nil {
		// read the response iff we have something to unmarshal it into
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

	// special case 404, because we need to be able to explicitly identify
	// cases where the requested path was not available.
	if response.StatusCode == http.StatusNotFound {
		return ErrPathNotFound
	}

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
	}
	// base case: actually delete this path, which is a concrete
	// key and not a directory
	c.opts.Logger.Printf("delete concrete path: %q", path)
	return c.deleteKey(path)
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

	// special case 404, because we need to be able to explicitly identify
	// cases where the requested path was not available.
	if response.StatusCode == http.StatusNotFound {
		return ErrPathNotFound
	}

	if response.StatusCode >= 400 {
		return errors.Errorf("bad status code: %d", response.StatusCode)
	}
	c.opts.Logger.Printf("delete status code: %d", response.StatusCode)

	return nil
}
