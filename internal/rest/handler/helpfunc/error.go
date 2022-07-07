package helpfunc

import (
	"fmt"
	"github.com/ispras/michman/internal/utils"
)

var HandlerHelpFuncsErrorMap = make(map[error]int)

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
