package helpfunc

import (
	"fmt"
	"github.com/ispras/michman/internal/rest"
	"github.com/ispras/michman/internal/utils"
)

const errUuidLibError = "uuid generation error"

var (
	ErrUuidLibError = rest.MakeError(errUuidLibError, utils.LibError)
)

func ErrClusterDependenceServicesIncompatibleVersion(service string, currentService string) error {
	errMessage := fmt.Sprintf("service '%s' has incompatible version for service '%s'", service, currentService)
	return rest.MakeError(errMessage, utils.ValidationError)
}

func ErrClusterServiceTypeNotSupported(param string) error {
	errMessage := fmt.Sprintf("service '%s' is not supported", param)
	return rest.MakeError(errMessage, utils.ValidationError)
}

func ErrClusterServiceVersionNotSupported(param string, service string) error {
	errMessage := fmt.Sprintf("'%s' service version '%s' is not supported", service, param)
	return rest.MakeError(errMessage, utils.ValidationError)
}

func ErrClusterServiceHealthCheck(service string) error {
	errMessage := fmt.Sprintf("'%s' HealthCheck field is empty", service)
	return rest.MakeError(errMessage, utils.DatabaseError)
}

func ErrObjectUnique(param string) error {
	errMessage := fmt.Sprintf("param %s is not unique", param)
	return rest.MakeError(errMessage, utils.ValidationError)
}
