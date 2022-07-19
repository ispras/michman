package authorization

import "errors"

const (
	errNoProjectInURL             = "no project ID or name in URL path"
	errUnauthorized               = "unauthorized to access the resource"
	errAuthenticationUnsuccessful = "Authentication unsuccessful! You are not a member of any group (Bad token)"
)

var (
	ErrNoProjectInURL             = errors.New(errNoProjectInURL)
	ErrUnauthorized               = errors.New(errUnauthorized)
	ErrAuthenticationUnsuccessful = errors.New(errAuthenticationUnsuccessful)
)
