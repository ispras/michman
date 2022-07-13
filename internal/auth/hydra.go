package auth

import (
	"bytes"
	"encoding/json"
	"github.com/alexedwards/scs/v2"
	"github.com/ispras/michman/internal/utils"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const (
	authHeader = "Authorization"

	tokenReqPath      = "/oauth2/token"
	uInfoReqPath      = "/userinfo"
	introspectReqPath = "/oauth2/introspect"
)

type HydraAuthenticate struct {
	hydraAdminUrl     string
	hydraClientUrl    string
	config            utils.Config
	vaultCommunicator utils.SecretStorage
	hydraAuth         *utils.HydraCredentials
}

type hydraIntrospect struct {
	Active            bool        `json:"active"`
	Aud               []string    `json:"aud,omitempty"`
	ClientId          string      `json:"client_id,omitempty"`
	Exp               int         `json:"exp,omitempty"`
	Ext               interface{} `json:"ext,omitempty"`
	Iat               int         `json:"iat,omitempty"`
	Iss               string      `json:"iss,omitempty"`
	Nbf               int         `json:"nbf,omitempty"`
	ObfuscatedSubject string      `json:"obfuscated_subject,omitempty"`
	Scope             string      `json:"scope,omitempty"`
	Sub               string      `json:"sub,omitempty"`
	TokenType         string      `json:"token_type,omitempty"`
	Username          string      `json:"username,omitempty"`
}

type hydraToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	IdToken     string `json:"id_token"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

type hydraUserInfo struct {
	Email      string `json:"email"`
	FamilyName string `json:"family_name"`
	Groups     string `json:"groups"`
	Sid        string `json:"sid"`
	Sub        string `json:"sub"`
}

func NewHydraAuthenticate() (Authenticate, error) {
	hydra := new(HydraAuthenticate)

	config := utils.Config{}
	if err := config.MakeCfg(); err != nil {
		return nil, err
	}
	hydra.config = config
	hydra.hydraAdminUrl = hydra.config.HydraAdmin
	hydra.hydraClientUrl = hydra.config.HydraClient

	client, vaultCfg, err := hydra.vaultCommunicator.ConnectVault()
	if client == nil || err != nil {
		return nil, ErrConnectVault
	}

	hydraSecrets, err := client.Logical().Read(vaultCfg.HydraKey)
	if err != nil {
		return nil, err
	}

	hydra.hydraAuth = &utils.HydraCredentials{
		RedirectUri:  hydraSecrets.Data[utils.HydraRedirectUri].(string),
		ClientId:     hydraSecrets.Data[utils.HydraClientId].(string),
		ClientSecret: hydraSecrets.Data[utils.HydraClientSecret].(string),
	}

	return hydra, nil
}

func (hydra HydraAuthenticate) CheckAuth(token string) (bool, error) {
	var body []byte
	body = []byte("token=" + token)

	req, err := http.NewRequest(http.MethodPost, hydra.hydraAdminUrl+introspectReqPath, bytes.NewBuffer(body))
	if err != nil {
		return false, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}

	var intrBody *hydraIntrospect
	err = json.NewDecoder(resp.Body).Decode(&intrBody)
	if err != nil {
		return false, ErrParseRequest
	}

	jBody, err := json.Marshal(intrBody)
	if err != nil {
		return false, err
	}

	if !intrBody.Active {
		// Token is not active/valid
		return false, ErrTokenActive(string(jBody))
	} else if intrBody.TokenType != "access_token" {
		// Token is not an access token (probably a refresh token)
		return false, ErrNotAccessToken(string(jBody))
	}

	return true, nil
}

func (hydra HydraAuthenticate) SetAuth(sm *scs.SessionManager, r *http.Request) (error, int) {
	//set session manager
	sessionManager = sm

	urlKeys := r.URL.Query()

	//get authorization code from url params
	code := urlKeys.Get("code")

	if code == "" {
		return ErrAuthCodeNil, http.StatusBadRequest
	}

	//set body params for token request
	body := url.Values{}
	body.Set("grant_type", "authorization_code")
	body.Set("code", code)
	body.Set("redirect_uri", hydra.hydraAuth.RedirectUri)
	body.Set("client_id", hydra.hydraAuth.ClientId)
	body.Set("client_secret", hydra.hydraAuth.ClientSecret)

	tokenReq, err := http.NewRequest(http.MethodPost, hydra.hydraClientUrl+tokenReqPath, strings.NewReader(body.Encode()))
	if err != nil {
		return err, http.StatusBadRequest
	}

	tokenReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	tokenReq.Header.Add("Accept", "application/json")

	client := &http.Client{}
	//make token request for getting information about access token
	resp, err := client.Do(tokenReq)
	if err != nil || resp.StatusCode != http.StatusOK {
		return err, http.StatusBadRequest
	}

	var tokenBody *hydraToken
	err = json.NewDecoder(resp.Body).Decode(&tokenBody)
	if err != nil {
		return ErrParseRequest, http.StatusBadRequest
	}

	//access token have to be not nil
	if tokenBody.AccessToken == "" {
		return ErrAccessTokenEmpty, http.StatusBadRequest
	}

	//set params for userinfo request
	uInfoReq, err := http.NewRequest(http.MethodGet, hydra.hydraClientUrl+uInfoReqPath, nil)
	if err != nil {
		return err, http.StatusBadRequest
	}

	uInfoReq.Header.Add("Authorization", "Bearer "+tokenBody.AccessToken)
	uInfoReq.Header.Add("Accept", "application/json")

	//make userinfo request for getting information about user group
	uInfoResp, err := client.Do(uInfoReq)
	if err != nil {
		return err, http.StatusBadRequest
	}

	var uInfoBody *hydraUserInfo
	err = json.NewDecoder(uInfoResp.Body).Decode(&uInfoBody)
	if err != nil {
		return ErrParseRequest, http.StatusBadRequest
	}

	if uInfoBody.Groups == "" {
		return ErrAccessTokenEmpty, http.StatusBadRequest
	}

	//init session for current user
	err = sessionManager.RenewToken(r.Context())
	if err != nil {
		return err, http.StatusInternalServerError
	}

	//save in user session information about group and access token
	sessionManager.Put(r.Context(), utils.GroupKey, uInfoBody.Groups)
	sessionManager.Put(r.Context(), utils.AccessTokenKey, tokenBody.AccessToken)

	return nil, http.StatusOK
}

func (hydra HydraAuthenticate) RetrieveToken(r *http.Request) (string, error) {
	bToken := r.Header.Get(authHeader)
	if bToken == "" {
		return bToken, ErrAuthHeaderEmpty
	}

	regEx := regexp.MustCompile(utils.RegexPattern)

	if regEx.FindString(bToken) == "" {
		return bToken, ErrAuthHeaderBadToken
	}

	token := strings.Fields(bToken)[1]
	return token, nil
}
