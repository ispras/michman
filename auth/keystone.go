package auth

import (
	"encoding/json"
	"errors"
	"github.com/alexedwards/scs/v2"
	"github.com/ispras/michman/utils"
	"net/http"
	"strings"
)

const (
	checkTokenPath = "/v3/auth/tokens"
	authTokenKey   = "X-Auth-Token"
	subTokenKey    = "X-Subject-Token"
)

type KeystoneAuthenticate struct {
	keystoneUrl string
	config      utils.Config
}

type keystoneRole struct {
	Id          string            `json:"id"`
	Links       map[string]string `json:"links,omitempty"`
	Description string            `json:"description,omitempty"`
	Name        string            `json:"name"`
}

type tokenInfo struct {
	Methods   []string       `json:"methods"`
	Links     interface{}    `json:"links"`
	User      interface{}    `json:"user"`
	Token     interface{}    `json:"token"`
	ExpiresAt string         `json:"expires_at"`
	Catalog   []interface{}  `json:"catalog,omitempty"`
	System    interface{}    `json:"system,omitempty"`
	Domain    interface{}    `json:"domain,omitempty"`
	Project   interface{}    `json:"project,omitempty"`
	Roles     []keystoneRole `json:"roles"`
	AuditIds  []string       `json:"audit_ids"`
	IssuedAt  string         `json:"issued_at"`
	Id        string         `json:"id,omitempty"`
	Name      string         `json:"name,omitempty"`
}

type keystoneToken struct {
	Token tokenInfo `json:"token"`
}

func NewKeystoneAuthenticate() (Authenticate, error) {
	k := new(KeystoneAuthenticate)

	config := utils.Config{}
	config.MakeCfg()
	k.config = config
	k.keystoneUrl = k.config.KeystoneAddr
	return k, nil
}

func (keystone KeystoneAuthenticate) CheckAuth(token string) (bool, error) {
	return true, nil
}

func (keystone KeystoneAuthenticate) SetAuth(sm *scs.SessionManager, w http.ResponseWriter, r *http.Request) (http.ResponseWriter, error) {
	//set session manager
	sessionManager = sm

	//get auth and subject tokens from headers
	authToken := r.Header.Get(authTokenKey)
	if authToken == "" {
		w.WriteHeader(http.StatusBadRequest)
		return w, errors.New("X-Auth-Token from headers is nil")
	}

	subToken := r.Header.Get(subTokenKey)
	if subToken == "" {
		w.WriteHeader(http.StatusBadRequest)
		return w, errors.New("X-Subject-Token from headers is nil")
	}

	//prepare request
	tokenReq, err := http.NewRequest(http.MethodGet, keystone.keystoneUrl+checkTokenPath, nil)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return w, err
	}
	tokenReq.Header.Add(authTokenKey, authToken)
	tokenReq.Header.Add(subTokenKey, subToken)

	client := &http.Client{}
	//make token request for getting information about user roles
	resp, err := client.Do(tokenReq)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return w, err
	}

	if resp.StatusCode != http.StatusOK {
		w.WriteHeader(resp.StatusCode)
		return w, err
	}

	//parse request body
	var tokenBody *keystoneToken
	err = json.NewDecoder(resp.Body).Decode(&tokenBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return w, err
	}

	//generate user groups by roles names
	var userGroups strings.Builder
	for i, r := range tokenBody.Token.Roles {
		if i == 0 {
			//join role without comma for the first time
			userGroups.WriteString(r.Name)
		} else {
			userGroups.WriteString("," + r.Name)
		}
	}

	//init session for current user
	err = sessionManager.RenewToken(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return w, err
	}

	//save in user session information about groups and tokens
	sessionManager.Put(r.Context(), authTokenKey, authToken)
	sessionManager.Put(r.Context(), subTokenKey, subToken)
	if userGroups.String() != "" {
		sessionManager.Put(r.Context(), utils.GroupKey, userGroups.String())
	}

	return w, nil
}

func (keystone KeystoneAuthenticate) RetrieveToken(r *http.Request) (string, error) {
	return "", nil
}
