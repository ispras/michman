package check

import (
	"fmt"
	"github.com/ispras/michman/internal/rest"
	"github.com/ispras/michman/internal/utils"
)

const (
	// cluster:
	errClusterBadName = "cluster validation error. Bad name. You should use only alpha-numeric characters and '-' symbols and only alphabetic characters for leading symbol"

	// flavor:

	// service type:
	errServiceTypeClass                              = "class for service type is not supported"
	errServiceTypePort                               = "port is incorrect"
	errServiceTypeVersionConfigDefaultValueEmpty     = "service type version config default value must be set"
	errServiceTypeVersionConfigDefaultValue          = "service type version config default value not in possible values"
	errServiceTypeDependencyServiceEmptyField        = "service default version | versions list in dependency can't be empty"
	errConfigServiceDependencyVersionNotFound        = "service version in versions list doesn't exist"
	errConfigServiceDependencyDefaultVersionNotFound = "service default version in dependencies doesn't exist"
	errServiceTypeDefaultVersion                     = "default version not found in versions list"

	// project:

	// log:
	errOsStat = "error occurred while reading file info describing the named file"

	errConfigPossibleValueEmpty          = "config possible value is empty"
	errConfigServiceTypeDependenceExists = "service type presents in dependencies for another service"
)

var (
	ErrClusterBadName = rest.MakeError(errClusterBadName, utils.ValidationError)
	// flavor:

	// service type:
	ErrServiceTypeClass                              = rest.MakeError(errServiceTypeClass, utils.ValidationError)
	ErrServiceTypePort                               = rest.MakeError(errServiceTypePort, utils.ValidationError)
	ErrServiceTypeVersionConfigDefaultValueEmpty     = rest.MakeError(errServiceTypeVersionConfigDefaultValueEmpty, utils.ValidationError)
	ErrServiceTypeVersionConfigDefaultValue          = rest.MakeError(errServiceTypeVersionConfigDefaultValue, utils.ValidationError)
	ErrServiceTypeDependencyServiceEmptyField        = rest.MakeError(errServiceTypeDependencyServiceEmptyField, utils.ValidationError)
	ErrConfigServiceDependencyVersionNotFound        = rest.MakeError(errConfigServiceDependencyVersionNotFound, utils.ObjectNotFound)
	ErrConfigServiceDependencyDefaultVersionNotFound = rest.MakeError(errConfigServiceDependencyDefaultVersionNotFound, utils.ObjectNotFound)
	ErrServiceTypeDefaultVersion                     = rest.MakeError(errServiceTypeDefaultVersion, utils.ObjectNotFound)
	ErrConfigServiceTypeDependenceExists             = rest.MakeError(errConfigServiceTypeDependenceExists, utils.ObjectExists)
	ErrConfigPossibleValueEmpty                      = rest.MakeError(errConfigPossibleValueEmpty, utils.ValidationError)

	// project:

	// log:
	ErrOsStat = rest.MakeError(errOsStat, utils.ValidationError)
)

// common:
func ErrObjectExists(param1 string, param2 string) error {
	errMessage := fmt.Sprintf("%s with this name or id (%s) already exists", param1, param2)
	return rest.MakeError(errMessage, utils.ObjectExists)
}

func ErrValidTypeParam(param string) error {
	errMessage := fmt.Sprintf("parameter type must be int, float, bool, string. Got: %s", param)
	return rest.MakeError(errMessage, utils.ValidationError)
}

func ErrPossibleValues(param string) error {
	errMessage := fmt.Sprintf("config possible value %s set incorrectly", param)
	return rest.MakeError(errMessage, utils.ValidationError)
}

// flavor:

func ErrFlavorParamVal(param string) error {
	errMessage := fmt.Sprintf("flavor %s can't be less than or equal to zero", param)
	return rest.MakeError(errMessage, utils.ValidationError)
}

func ErrFlavorParamType(param string) error {
	errMessage := fmt.Sprintf("flavor %s must be int type", param)
	return rest.MakeError(errMessage, utils.ValidationError)
}

// service type:

func ErrServiceTypeVersionUnique(param string) error {
	errMessage := fmt.Sprintf("version %s is not unique", param)
	return rest.MakeError(errMessage, utils.ValidationError)
}

func ErrServiceTypeVersionConfigParamEmpty(param string) error {
	errMessage := fmt.Sprintf("config parameter %s must be set", param)
	return rest.MakeError(errMessage, utils.ValidationError)
}

func ErrServiceTypeVersionConfigUnique(param string) error {
	errMessage := fmt.Sprintf("config with parameter name %s is not unique", param)
	return rest.MakeError(errMessage, utils.ValidationError)
}

func ErrServiceDependenciesNotExists(param string) error {
	errMessage := fmt.Sprintf("service with type %s from dependencies doesn't exist", param)
	return rest.MakeError(errMessage, utils.ValidationError)
}

func ErrServiceTypeVersionDependencyUnique(param string) error {
	errMessage := fmt.Sprintf("dependency with service type %s is not unique", param)
	return rest.MakeError(errMessage, utils.ValidationError)
}

func ErrServiceTypeDependenceVersionExists(param1 string, param2 string) error {
	errMessage := fmt.Sprintf("service type version %s presents in dependencies versions in %s service", param1, param2)
	return rest.MakeError(errMessage, utils.ValidationError)
}

func ErrDependencyServiceTypeNotExists(used_service, dep_version string) error {
	errMessage := fmt.Sprintf("service type %s not found in dependencies of version %s", used_service, dep_version)
	return rest.MakeError(errMessage, utils.ObjectNotFound)
}

func ErrServiceTypeVersionConfigPossibleValuesUnique(param string) error {
	errMessage := fmt.Sprintf("config possible value %s is not unique", param)
	return rest.MakeError(errMessage, utils.ValidationError)
}

func ErrClusterServiceConfigIncorrectType(param string, service string) error {
	errMessage := fmt.Sprintf("'%s' service config param '%s' has incorrect value type", service, param)
	return rest.MakeError(errMessage, utils.ValidationError)
}

func ErrClusterServiceConfigNotPossibleValue(param string, service string) error {
	errMessage := fmt.Sprintf("'%s' service config param '%s' value is not supported", service, param)
	return rest.MakeError(errMessage, utils.ValidationError)
}

func ErrClusterServiceConfigNotSupported(param string, service string) error {
	errMessage := fmt.Sprintf("'%s' service config param name '%s' is not supported", service, param)
	return rest.MakeError(errMessage, utils.ValidationError)
}
