// Author hoenig

package vaultapi

type KV interface {
	Get(path string) (string, error)
	Put(path, value string) error
	Delete(path string) error
	Keys(path string) ([]string, error)
	Recurse(path string) ([][2]string, error)
}


