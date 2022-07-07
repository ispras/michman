package response

import (
	"errors"
	"github.com/ispras/michman/internal/utils"
)

// HandlerErrorsMap for all handlers to avoid cycle import
var HandlerErrorsMap = make(map[error]int)

var HandlerResponseErrorMap = make(map[error]int)

const (
	errJsonEncode = "json encode error"
)

var (
	ErrJsonEncode = errors.New(errJsonEncode)
)

func init() {
	HandlerResponseErrorMap[ErrJsonEncode] = utils.JsonError
}
