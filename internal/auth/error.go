package auth

import (
	"errors"
	"fmt"
)

const (
	errAuthCodeNil        = "authorization code required, can't be nil"
	errAuthTokenNil       = "X-Auth-Token from headers required, can't be nil"
	errSubjectTokenNil    = "X-Subject-Token from headers required, can't be nil"
	errAccessTokenEmpty   = "access token required, can't be empty"
	errAuthHeaderEmpty    = "authorization header required, can't be empty"
	errAuthHeaderBadToken = "bad token in authorization header"
	errConnectVault       = "can't connect to vault secrets storage"
	errCreateAuth         = "can't create new authenticator"
)

var (
	ErrAuthCodeNil        = errors.New(errAuthCodeNil)
	ErrAuthTokenNil       = errors.New(errAuthTokenNil)
	ErrSubjectTokenNil    = errors.New(errSubjectTokenNil)
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

func ErrParseRequest(param string) error {
	ErrParamType := fmt.Errorf("failed %s request body parsing", param)
	return ErrParamType
}
