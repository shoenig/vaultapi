// Author hoenig

package vaultapi

import (
	"io/ioutil"
	"strings"
)

// A Tokener provides a token that can be used to
// authenticate with vault.
type Tokener interface {
	Token() (string, error)
}

type staticToken struct {
	// todo: maybe mlock this
	token string
}

var _ Tokener = (*staticToken)(nil)

// NewStaticToken creates a Tokener that will only
// ever return the one provided value for token.
func NewStaticToken(token string) Tokener {
	return &staticToken{token: token}
}

func (t *staticToken) Token() (string, error) {
	return t.token, nil
}

type fileToken struct {
	filename string
}

var _ Tokener = (*fileToken)(nil)

// NewFileToken will create a Tokener that will always
// reload the token value from the specified file.
func NewFileToken(filename string) Tokener {
	return &fileToken{filename: filename}
}

func (t *fileToken) Token() (string, error) {
	bs, err := ioutil.ReadFile(t.filename)
	return strings.TrimSpace(string(bs)), err
}
