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
	noneAuth := new(NoneAuthenticate)
	config := utils.Config{}
	if err := config.MakeCfg(); err != nil {
		return nil, err
	}
	noneAuth.config = config
	return noneAuth, nil
}

func (n NoneAuthenticate) CheckAuth(_ string) (bool, error) {
	return true, nil
}

func (n NoneAuthenticate) SetAuth(sm *scs.SessionManager, r *http.Request) error {
	// set session manager
	sessionManager = sm

	// init session for current user
	err := sessionManager.RenewToken(r.Context())
	if err != nil {
		return ErrThirdParty(err.Error())
	}

	sessionManager.Put(r.Context(), utils.GroupKey, n.config.AdminGroup)
	return nil
}

func (n NoneAuthenticate) RetrieveToken(_ *http.Request) (string, error) {
	return "", nil
}
