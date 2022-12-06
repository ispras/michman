package auth

import (
	"errors"
	"fmt"
	"github.com/ispras/michman/internal/rest"
	"github.com/ispras/michman/internal/utils"
)

const (
	errAuthCodeNil        = "authorization code required, can't be nil"
	errAuthTokenNil       = "X-Auth-Token from headers required, can't be nil"
	errSubjectTokenNil    = "X-Subject-Token from headers required, can't be nil"
	errAccessTokenEmpty   = "access token required, can't be empty"
	errAuthHeaderEmpty    = "authorization header required, can't be empty"
	errAuthHeaderBadToken = "bad token in authorization header"
	errUsrInfoGroupsEmpty = "user has no groups"
	errConnectVault       = "can't connect to vault secrets storage"
	errCreateAuth         = "can't create new authenticator"
)

var (
	ErrAuthCodeNil        = rest.MakeError(errAuthCodeNil, utils.AuthorizationError)
	ErrAuthTokenNil       = rest.MakeError(errAuthTokenNil, utils.AuthorizationError)
	ErrSubjectTokenNil    = rest.MakeError(errSubjectTokenNil, utils.AuthorizationError)
	ErrAccessTokenEmpty   = rest.MakeError(errAccessTokenEmpty, utils.AuthorizationError)
	ErrUsrInfoGroupsEmpty = rest.MakeError(errUsrInfoGroupsEmpty, utils.AuthorizationError)
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
	errMessage := fmt.Sprintf("failed %s request body parsing", param)
	return rest.MakeError(errMessage, utils.ParseError)
}

func ErrThirdParty(param string) error {
	return rest.MakeError(param, utils.UnexpectedError)
}
