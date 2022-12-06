package auth

import (
	"github.com/alexedwards/scs/v2"
	"github.com/ispras/michman/internal/utils"
	"net/http"
)

var (
	sessionManager *scs.SessionManager
)

type Authenticate interface {
	CheckAuth(token string) (bool, error)
	SetAuth(sm *scs.SessionManager, r *http.Request) error
	RetrieveToken(r *http.Request) (string, error)
}

func InitAuth(authMode string) (Authenticate, error) {
	switch authMode {
	case utils.OAuth2Mode:
		hydraAuth, err := NewHydraAuthenticate()
		if err != nil {
			return nil, ErrCreateAuthenticator(errCreateAuth, err.Error())
		}
		return hydraAuth, nil
	case utils.KeystoneMode:
		keystoneAuth, err := NewKeystoneAuthenticate()
		if err != nil {
			return nil, ErrCreateAuthenticator(errCreateAuth, err.Error())
		}
		return keystoneAuth, nil
	case utils.NoneAuthMode:
		noneAuth, err := NewNoneAuthenticate()
		if err != nil {
			return nil, ErrCreateAuthenticator(errCreateAuth, err.Error())
		}
		return noneAuth, nil
	}

	return nil, ErrCreateAuth
}
