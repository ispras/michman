package response

import (
	"github.com/ispras/michman/internal/rest"
	"github.com/ispras/michman/internal/utils"
)

const (
	errJsonEncode = "json encode error"
)

var (
	ErrJsonEncode = rest.MakeError(errJsonEncode, utils.JsonError)
)
