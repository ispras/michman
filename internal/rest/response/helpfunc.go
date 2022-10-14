package response

import (
	"errors"
	"github.com/ispras/michman/internal/database"
	"github.com/ispras/michman/internal/rest"
	"github.com/ispras/michman/internal/utils"
	"net/http"
)

type fn func(w http.ResponseWriter, errMsg string, class int)

var ErrorMap = make(map[int]fn)

func ErrorClass(err error) int {
	var dbError *database.Error
	var restError *rest.Error
	if errors.As(err, &dbError) {
		return err.(*database.Error).Class
	}
	if errors.As(err, &restError) {
		return err.(*rest.Error).Class
	}
	return utils.UnexpectedError
}

func Error(w http.ResponseWriter, err error) {
	var dbError *database.Error
	var restError *rest.Error
	if errors.As(err, &dbError) {
		respErr := err.(*database.Error)
		ErrorMap[respErr.Class](w, respErr.Error(), respErr.Class)
	}
	if errors.As(err, &restError) {
		respErr := err.(*rest.Error)
		ErrorMap[respErr.Class](w, respErr.Error(), respErr.Class)
	}
}

func init() {
	ErrorMap[utils.DatabaseError] = InternalError
	ErrorMap[utils.LibError] = InternalError
	ErrorMap[utils.LogsError] = InternalError
	ErrorMap[utils.AuthorizationError] = InternalError

	ErrorMap[utils.ObjectNotFound] = NotFound

	ErrorMap[utils.JsonError] = BadRequest
	ErrorMap[utils.ValidationError] = BadRequest
	ErrorMap[utils.ObjectUsed] = BadRequest
	ErrorMap[utils.ObjectExists] = BadRequest
	ErrorMap[utils.ObjectUnmodified] = BadRequest
	ErrorMap[utils.InputIncorrect] = BadRequest
}
