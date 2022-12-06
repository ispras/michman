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
	userIdKey      = "user_id"
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

type keystoneUser struct {
	Domain          interface{} `json:"domain"`
	Id              string      `json:"id"`
	Name            string      `json:"name"`
	PasswordExpired interface{} `json:"password_expires_at"`
}

type tokenInfo struct {
	Methods   []string       `json:"methods"`
	Links     interface{}    `json:"links"`
	User      keystoneUser   `json:"user"`
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
	keystoneAuth := new(KeystoneAuthenticate)

	config := utils.Config{}
	if err := config.MakeCfg(); err != nil {
		return nil, err
	}
	keystoneAuth.config = config
	keystoneAuth.keystoneUrl = keystoneAuth.config.KeystoneAddr
	return keystoneAuth, nil
}

func (keystone KeystoneAuthenticate) CheckAuth(_ string) (bool, error) {
	return true, nil
}

func (keystone KeystoneAuthenticate) SetAuth(sm *scs.SessionManager, r *http.Request) error {
	// set session manager
	sessionManager = sm

	// get X-Auth-Token from headers
	authToken := r.Header.Get(authTokenKey)
	if authToken == "" {
		return ErrAuthTokenNil
	}

	// get X-Subject-Token from headers
	subToken := r.Header.Get(subTokenKey)
	if subToken == "" {
		return ErrSubjectTokenNil
	}

	// prepare request for confirmation of user tokens
	tokenRequest, err := http.NewRequest(http.MethodGet, keystone.keystoneUrl+checkTokenPath, nil)
	if err != nil {
		return ErrThirdParty(err.Error())
	}

	// set headers for token confirmation request
	tokenRequest.Header.Add(authTokenKey, authToken)
	tokenRequest.Header.Add(subTokenKey, subToken)

	client := &http.Client{}

	// send user token confirmation request for getting information about user roles and userID
	resp, err := client.Do(tokenRequest)
	if err != nil || resp.StatusCode != http.StatusOK {
		return ErrThirdParty(err.Error())
	}

	// parse request response body
	var tokenBody *keystoneToken
	err = json.NewDecoder(resp.Body).Decode(&tokenBody)
	if err != nil {
		return ErrParseRequest("token")
	}

	// get user ID from token request
	userID := tokenBody.Token.User.Id

	// generate user groups by roles names
	var userGroups strings.Builder
	for i, role := range tokenBody.Token.Roles {
		if i == 0 {
			// join role without comma for the first time
			userGroups.WriteString(role.Name)
		} else {
			userGroups.WriteString("," + role.Name)
		}
	}

	// init session for current user
	err = sessionManager.RenewToken(r.Context())
	if err != nil {
		return ErrThirdParty(err.Error())
	}

	// save in user session information about groups, tokens and user ID
	sessionManager.Put(r.Context(), userIdKey, userID)
	sessionManager.Put(r.Context(), authTokenKey, authToken)
	sessionManager.Put(r.Context(), subTokenKey, subToken)
	if userGroups.String() != "" {
		sessionManager.Put(r.Context(), utils.GroupKey, userGroups.String())
	}

	return nil
}

func (keystone KeystoneAuthenticate) RetrieveToken(_ *http.Request) (string, error) {
	return "", nil
}
