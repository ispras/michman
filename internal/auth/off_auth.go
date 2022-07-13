package auth

import (
	"github.com/alexedwards/scs/v2"
	"github.com/ispras/michman/internal/utils"
	"net/http"
)

type NoneAuthenticate struct {
	config utils.Config
}

func NewNoneAuthenticate() (Authenticate, error) {
	n := new(NoneAuthenticate)
	config := utils.Config{}
	if err := config.MakeCfg(); err != nil {
		return nil, err
	}
	n.config = config
	return n, nil
}

func (n NoneAuthenticate) CheckAuth(token string) (bool, error) {
	return true, nil
}

func (n NoneAuthenticate) SetAuth(sm *scs.SessionManager, r *http.Request) (error, int) {
	//set session manager
	sessionManager = sm

	//init session for current user
	err := sessionManager.RenewToken(r.Context())
	if err != nil {
		return err, http.StatusInternalServerError
	}

	sessionManager.Put(r.Context(), utils.GroupKey, n.config.AdminGroup)
	return nil, http.StatusOK
}

func (n NoneAuthenticate) RetrieveToken(r *http.Request) (string, error) {
	return "", nil
}
