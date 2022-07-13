package authorization

import "errors"

const (
	errNoProjectInURL = "no project ID or name in URL path"
	errUnauthorized   = "unauthorized to access the resource"
)

var (
	ErrNoProjectInURL = errors.New(errNoProjectInURL)
	ErrUnauthorized   = errors.New(errUnauthorized)
)
