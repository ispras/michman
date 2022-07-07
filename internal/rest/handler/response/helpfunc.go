package response

import (
	"github.com/ispras/michman/internal/database"
	"github.com/ispras/michman/internal/rest/handler/check"
	"github.com/ispras/michman/internal/rest/handler/helpfunc"
	"github.com/ispras/michman/internal/rest/handler/validate"
	"github.com/ispras/michman/internal/utils"
)

func FindErrorType(err error) int {
	if HandlerErrorsMap[err] != 0 {
		return HandlerErrorsMap[err]
	}
	if database.DbErrorsMap[err] != 0 {
		return database.DbErrorsMap[err]
	}
	if check.HandlerCheckersErrorMap[err] != 0 {
		return check.HandlerCheckersErrorMap[err]
	}
	if helpfunc.HandlerHelpFuncsErrorMap[err] != 0 {
		return helpfunc.HandlerHelpFuncsErrorMap[err]
	}
	if HandlerResponseErrorMap[err] != 0 {
		return HandlerResponseErrorMap[err]
	}
	if validate.HandlerValidateErrorMap[err] != 0 {
		return validate.HandlerValidateErrorMap[err]
	}
	return utils.UnknownError
}
