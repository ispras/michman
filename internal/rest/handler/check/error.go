package check

import (
	"errors"
	"fmt"
	"github.com/ispras/michman/internal/utils"
)

var HandlerCheckersErrorMap = make(map[error]int)

const (
	errServiceTypeClass                              = "class for service type is not supported"
	errServiceTypeAccessPort                         = "port is incorrect"
	errServiceTypeDefaultVersion                     = "default version not found in versions list"
	errConfigPossibleValueEmpty                      = "config possible value is empty"
	errServiceTypeVersionConfigDefaultValue          = "service type version config default value not in possible values"
	errServiceTypeVersionConfigDefaultValueEmpty     = "service type version config default value must be set"
	errConfigDependencyServiceDefaultVersionEmpty    = "service default version in dependency can't be empty"
	errConfigServiceDependencyVersionNotFound        = "service version in dependency doesn't exist"
	errConfigServiceDependencyDefaultVersionNotFound = "service default version in dependencies doesn't exist"
	errConfigDependencyServiceVersionEmpty           = "service versions list in dependencies can't be empty"
)

var (
	ErrServiceTypeClass                              = errors.New(errServiceTypeClass)
	ErrServiceTypePort                               = errors.New(errServiceTypeAccessPort)
	ErrServiceTypeDefaultVersion                     = errors.New(errServiceTypeDefaultVersion)
	ErrConfigPossibleValueEmpty                      = errors.New(errConfigPossibleValueEmpty)
	ErrServiceTypeVersionConfigDefaultValue          = errors.New(errServiceTypeVersionConfigDefaultValue)
	ErrServiceTypeVersionConfiqDefaultValueEmpty     = errors.New(errServiceTypeVersionConfigDefaultValueEmpty)
	ErrConfigDependencyServiceDefaultVersionEmpty    = errors.New(errConfigDependencyServiceDefaultVersionEmpty)
	ErrConfigServiceDependencyVersionNotFound        = errors.New(errConfigServiceDependencyVersionNotFound)
	ErrConfigServiceDependencyDefaultVersionNotFound = errors.New(errConfigServiceDependencyDefaultVersionNotFound)
	ErrConfigDependencyServiceVersionEmpty           = errors.New(errConfigDependencyServiceVersionEmpty)
)

func ErrValidTypeParam(param string) error {
	ErrParamType := fmt.Errorf("parameter type must be int, float, bool, string. Got: %s", param)
	HandlerCheckersErrorMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func ErrServiceTypeVersionUnique(param string) error {
	ErrParamType := fmt.Errorf("version %s is not unique", param)
	HandlerCheckersErrorMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func ErrServiceTypeVersionConfigUnique(param string) error {
	ErrParamType := fmt.Errorf("config with parameter name %s is not unique", param)
	HandlerCheckersErrorMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func ErrServiceTypeVersionConfigPossibleValuesUnique(param string) error {
	ErrParamType := fmt.Errorf("config possible value %s is not unique", param)
	HandlerCheckersErrorMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func ErrServiceTypeVersionConfigPossibleValues(param string) error {
	ErrParamType := fmt.Errorf("config possible value %s set incorrectly", param)
	HandlerCheckersErrorMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func ErrServiceTypeVersionConfigParamEmpty(param string) error {
	ErrParamType := fmt.Errorf("config parameter %s must be set", param)
	HandlerCheckersErrorMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func ErrServiceDependenciesNotExists(param string) error {
	ErrParamType := fmt.Errorf("service with type %s from dependencies doesn't exist", param)
	HandlerCheckersErrorMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func ErrFlavorParamVal(param string) error {
	ErrParamVal := fmt.Errorf("flavor %s can't be less than or equal to zero", param)
	HandlerCheckersErrorMap[ErrParamVal] = utils.ValidationError
	return ErrParamVal
}

func ErrFlavorParamType(param string) error {
	ErrParamType := fmt.Errorf("flavor %s must be int type", param)
	HandlerCheckersErrorMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func init() {
	HandlerCheckersErrorMap[ErrServiceTypeClass] = utils.ValidationError
	HandlerCheckersErrorMap[ErrServiceTypePort] = utils.ValidationError
	HandlerCheckersErrorMap[ErrServiceTypeDefaultVersion] = utils.ValidationError
	HandlerCheckersErrorMap[ErrConfigPossibleValueEmpty] = utils.ValidationError
	HandlerCheckersErrorMap[ErrConfigDependencyServiceDefaultVersionEmpty] = utils.ValidationError
	HandlerCheckersErrorMap[ErrConfigServiceDependencyVersionNotFound] = utils.ValidationError
	HandlerCheckersErrorMap[ErrConfigServiceDependencyDefaultVersionNotFound] = utils.ValidationError
	HandlerCheckersErrorMap[ErrServiceTypeVersionConfiqDefaultValueEmpty] = utils.ValidationError
	HandlerCheckersErrorMap[ErrServiceTypeVersionConfigDefaultValue] = utils.ValidationError
	HandlerCheckersErrorMap[ErrConfigDependencyServiceVersionEmpty] = utils.ValidationError

}
