package auth

import (
	"github.com/alexedwards/scs/v2"
	"github.com/ispras/michman/internal/utils"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
)

var (
	sessionManager *scs.SessionManager
)

type Authenticate interface {
	CheckAuth(token string) (bool, error)
	SetAuth(sm *scs.SessionManager, w http.ResponseWriter, r *http.Request) (http.ResponseWriter, error)
	RetrieveToken(r *http.Request) (string, error)
}

func InitAuth(httpLogger *logrus.Logger, authMode string) Authenticate {
	switch authMode {
	case utils.OAuth2Mode:
		hydraAuth, err := NewHydraAuthenticate()
		if err != nil {
			httpLogger.Println("Can't create new authenticator")
			os.Exit(1)
		}
		return hydraAuth
	case utils.KeystoneMode:
		keystoneAuth, err := NewKeystoneAuthenticate()
		if err != nil {
			httpLogger.Println("Can't create new authenticator")
			os.Exit(1)
		}
		return keystoneAuth
	case utils.NoneAuthMode:
		noneAuth, err := NewNoneAuthenticate()
		if err != nil {
			httpLogger.Println("Can't create new authenticator")
			os.Exit(1)
		}
		return noneAuth
	}
	return nil
}
