package helpfunc

import (
	"errors"
	"fmt"
	"github.com/ispras/michman/internal/utils"
)

const errUuidLibError = "uuid generating error"

var (
	HandlerHelpFuncsErrorMap = make(map[error]int)
	ErrUuidLibError          = errors.New(errUuidLibError)
)

func ErrClusterDependenceServicesIncompatibleVersion(service string, currentService string) error {
	ErrParamType := fmt.Errorf("service '%s' has incompatible version for service '%s'", service, currentService)
	HandlerHelpFuncsErrorMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func ErrClusterServiceTypeNotSupported(param string) error {
	ErrParamType := fmt.Errorf("service '%s' is not supported", param)
	HandlerHelpFuncsErrorMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func ErrClusterServiceVersionNotSupported(param string, service string) error {
	ErrParamType := fmt.Errorf("'%s' service version '%s' is not supported", service, param)
	HandlerHelpFuncsErrorMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func ErrClusterServiceHealthCheck(service string) error {
	ErrParamType := fmt.Errorf("'%s' HealthCheck field is empty", service)
	HandlerHelpFuncsErrorMap[ErrParamType] = utils.DatabaseError
	return ErrParamType
}

func ErrObjectUnique(param string) error {
	ErrParamType := fmt.Errorf("param %s is not unique", param)
	HandlerHelpFuncsErrorMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func init() {
	HandlerHelpFuncsErrorMap[ErrUuidLibError] = utils.LibError
}
