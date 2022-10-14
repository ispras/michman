package handler

import (
	"fmt"
	"github.com/ispras/michman/internal/rest"
	"github.com/ispras/michman/internal/utils"
)

const (
	// common:
	errJsonIncorrect = "incorrect json format"
	errUuidLibError  = "uuid generation error"

	//flavor:

	//image:

	//project:

	//cluster:

	//log:
	errBadActionParam = "bad action param. Supported query variables for action parameter are 'create', 'update' and 'delete'. Action 'create' is default"

	//service type:
	errGetQueryParams = "bad view param. Supported query variables for view parameter are 'full' and 'summary', 'summary' is default"
)

var (
	// common:
	ErrJsonIncorrect = rest.MakeError(errJsonIncorrect, utils.JsonError)
	ErrUuidLibError  = rest.MakeError(errUuidLibError, utils.LibError)

	// flavor:

	// service type:
	ErrGetQueryParams = rest.MakeError(errGetQueryParams, utils.InputIncorrect)

	// image:

	// project:

	// log:
	ErrLogsBadActionParam = rest.MakeError(errBadActionParam, utils.LogsError)
)

func ErrObjectExists(object string, idOrName string) error {
	errMessage := fmt.Sprintf("%s with this name or id (%s) already exists", object, idOrName)
	return rest.MakeError(errMessage, utils.ObjectExists)
}
