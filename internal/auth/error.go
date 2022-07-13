package auth

import (
	"errors"
	"fmt"
)

const (
	errAuthCodeNil        = "Authorization code is nil"
	errAuthTokenNil       = "X-Auth-Token from headers is nil"
	errSubjectTokenNil    = "X-Subject-Token from headers is nil"
	errParseRequest       = "failed request body parsing"
	errAccessTokenEmpty   = "access token can't be empty"
	errAuthHeaderEmpty    = "authorization header is empty"
	errAuthHeaderBadToken = "bad token in authorization header"
	errConnectVault       = "can't connect to vault secrets storage"
	errCreateAuth         = "can't create new authenticator"
)

var (
	ErrAuthCodeNil        = errors.New(errAuthCodeNil)
	ErrAuthTokenNil       = errors.New(errAuthTokenNil)
	ErrSubjectTokenNil    = errors.New(errSubjectTokenNil)
	ErrParseRequest       = errors.New(errParseRequest)
	ErrAccessTokenEmpty   = errors.New(errAccessTokenEmpty)
	ErrAuthHeaderEmpty    = errors.New(errAuthHeaderEmpty)
	ErrAuthHeaderBadToken = errors.New(errAuthHeaderBadToken)
	ErrConnectVault       = errors.New(errConnectVault)
	ErrCreateAuth         = errors.New(errCreateAuth)
)

func ErrTokenActive(param string) error {
	ErrParamType := fmt.Errorf("token is not active %s", param)
	return ErrParamType
}

func ErrNotAccessToken(param string) error {
	ErrParamType := fmt.Errorf("token is not access token %s", param)
	return ErrParamType
}

func ErrCreateAuthenticator(error string, details string) error {
	ErrParamType := fmt.Errorf("%s - %s", error, details)
	return ErrParamType
}
