package auth

import (
	"errors"
	"github.com/alexedwards/scs/v2"
	"github.com/ispras/michman/internal/utils"
	"github.com/sirupsen/logrus"
	"net/http"
)

var (
	sessionManager *scs.SessionManager
)

type Authenticate interface {
	CheckAuth(token string) (bool, error)
	SetAuth(sm *scs.SessionManager, w http.ResponseWriter, r *http.Request) (http.ResponseWriter, error)
	RetrieveToken(r *http.Request) (string, error)
}

func InitAuth(httpLogger *logrus.Logger, authMode string) (Authenticate, error) {
	switch authMode {
	case utils.OAuth2Mode:
		hydraAuth, err := NewHydraAuthenticate()
		if err != nil {
			return nil, errors.New("Can't create new authenticator")
		}
		return hydraAuth, nil
	case utils.KeystoneMode:
		keystoneAuth, err := NewKeystoneAuthenticate()
		if err != nil {
			return nil, errors.New("Can't create new authenticator")
		}
		return keystoneAuth, nil
	case utils.NoneAuthMode:
		noneAuth, err := NewNoneAuthenticate()
		if err != nil {
			return nil, errors.New("Can't create new authenticator")
		}
		return noneAuth, nil
	}
	return nil, errors.New("Can't create new authenticator")
}
