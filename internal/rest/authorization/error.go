package authorization

import (
	"github.com/ispras/michman/internal/rest"
	"github.com/ispras/michman/internal/utils"
)

const (
	errEnforcerSafe   = "failed to call Enforce in a safe way"
	errNoProjectInURL = "no project ID or name in URL path"
	errUnauthorized   = "unauthorized to access the resource"
)

var (
	ErrEnforcerSafe   = rest.MakeError(errEnforcerSafe, utils.EnforcerError)
	ErrNoProjectInURL = rest.MakeError(errNoProjectInURL, utils.InputIncorrect)
	ErrUnauthorized   = rest.MakeError(errUnauthorized, utils.AuthorizationError)
)
