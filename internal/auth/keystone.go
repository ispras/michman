package auth

import (
	"encoding/json"
	"github.com/alexedwards/scs/v2"
	"github.com/ispras/michman/internal/utils"
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
	if err := config.MakeCfg(); err != nil {
		return nil, err
	}
	k.config = config
	k.keystoneUrl = k.config.KeystoneAddr
	return k, nil
}

func (keystone KeystoneAuthenticate) CheckAuth(token string) (bool, error) {
	return true, nil
}

func (keystone KeystoneAuthenticate) SetAuth(sm *scs.SessionManager, r *http.Request) (error, int) {
	//set session manager
	sessionManager = sm

	//get auth and subject tokens from headers
	authToken := r.Header.Get(authTokenKey)
	if authToken == "" {
		return ErrAuthTokenNil, http.StatusBadRequest
	}

	subToken := r.Header.Get(subTokenKey)
	if subToken == "" {
		return ErrSubjectTokenNil, http.StatusBadRequest
	}

	//prepare request
	tokenReq, err := http.NewRequest(http.MethodGet, keystone.keystoneUrl+checkTokenPath, nil)
	if err != nil {
		return err, http.StatusBadRequest
	}

	tokenReq.Header.Add(authTokenKey, authToken)
	tokenReq.Header.Add(subTokenKey, subToken)

	client := &http.Client{}
	//make token request for getting information about user roles
	resp, err := client.Do(tokenReq)
	if err != nil || resp.StatusCode != http.StatusOK {
		return err, http.StatusBadRequest
	}

	//parse request body
	var tokenBody *keystoneToken
	err = json.NewDecoder(resp.Body).Decode(&tokenBody)
	if err != nil {
		return ErrParseRequest, http.StatusBadRequest
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
		return err, http.StatusInternalServerError
	}

	//save in user session information about groups and tokens
	sessionManager.Put(r.Context(), authTokenKey, authToken)
	sessionManager.Put(r.Context(), subTokenKey, subToken)
	if userGroups.String() != "" {
		sessionManager.Put(r.Context(), utils.GroupKey, userGroups.String())
	}

	return nil, http.StatusOK
}

func (keystone KeystoneAuthenticate) RetrieveToken(r *http.Request) (string, error) {
	return "", nil
}
