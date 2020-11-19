package auth

import (
	"github.com/alexedwards/scs/v2"
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