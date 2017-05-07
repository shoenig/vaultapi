// Author hoenig

package vaultapi

type Tokener interface {
	Token() string
}

// -- static token --

type staticToken struct {
	token string
}

var _ Tokener = (*staticToken)(nil)

func NewStaticToken(token string) Tokener {
	return &staticToken{token: token}
}

func (t *staticToken) Token() string {
	return t.token
}

// -- reloading token --

//type ReloadingTokenOptions struct {
//	ReloadFrequency time.Duration
//	Filepath        string
//}
//
//type reloadingToken struct {
//	opts ReloadingTokenOptions
//
//	lock         sync.RWMutex
//	currentToken string
//}
//
//var _ Tokener = (*reloadingToken)(nil)
//
//func (t *reloadingToken) Token() string {
//	t.lock.RLock()
//	defer t.lock.RUnlock()
//
//	return t.currentToken
//}
//
//func (t *reloadingToken) reload() {
//	// this needs server context
//}
