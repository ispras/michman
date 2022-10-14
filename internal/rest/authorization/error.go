package authorization

import (
	"github.com/ispras/michman/internal/rest"
	"github.com/ispras/michman/internal/utils"
)

const (
	errNoProjectInURL             = "no project ID or name in URL path"
	errUnauthorized               = "unauthorized to access the resource"
	errAuthenticationUnsuccessful = "Authentication unsuccessful! You are not a member of any group (Bad token)"
)

var (
	ErrAuthenticationUnsuccessful = rest.MakeError(errAuthenticationUnsuccessful, utils.AuthorizationError)
	ErrNoProjectInURL             = rest.MakeError(errNoProjectInURL, utils.AuthorizationError)
	ErrUnauthorized               = rest.MakeError(errUnauthorized, utils.AuthorizationError)
)
