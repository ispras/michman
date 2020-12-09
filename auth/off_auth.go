package auth

import (
	"github.com/alexedwards/scs/v2"
	"github.com/ispras/michman/utils"
	"net/http"
)

type NoneAuthenticate struct {
	config utils.Config
}

func NewNoneAuthenticate() (Authenticate, error) {
	n := new(NoneAuthenticate)
	config := utils.Config{}
	config.MakeCfg()
	n.config = config
	return n, nil
}

func (n NoneAuthenticate) CheckAuth(token string) (bool, error) {
	return true, nil
}

func (n NoneAuthenticate) SetAuth(sm *scs.SessionManager, w http.ResponseWriter, r *http.Request) (http.ResponseWriter, error) {
	//set session manager
	sessionManager = sm

	//init session for current user
	err := sessionManager.RenewToken(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return w, err
	}

	sessionManager.Put(r.Context(), utils.GroupKey, n.config.AdminGroup)
	return w, nil
}

func (n NoneAuthenticate) RetrieveToken(r *http.Request) (string, error) {
	return "", nil
}
