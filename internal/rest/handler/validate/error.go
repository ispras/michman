package validate

import (
	"errors"
	"fmt"
	"github.com/ispras/michman/internal/utils"
)

var HandlerValidateErrorMap = make(map[error]int)

const (
	errClusterBadName                      = "cluster validation error. Bad name. You should use only alpha-numeric characters and '-' symbols and only alphabetic characters for leading symbol"
	errClusterNhostsZero                   = "NHosts parameter must be number >= 0"
	errClustersNhostsMasterSlave           = "NHosts parameter must be number >= 1 if you want to install master-slave services"
	errClusterImageNotFound                = "specified Image not found"
	errFlavorIdNotEmpty                    = "flavor ID is generated field. It can't be filled in by user"
	errFlavorEmptyName                     = "flavor Name can't be empty"
	errImageIdNotEmpty                     = "image ID is generated field. It can't be filled in by user"
	errProjectValidation                   = "project validation error. Bad name. You should use only alpha-numeric characters and '-' symbols and only alphabetic characters for leading symbol"
	errProjectExisted                      = "project with this name already exists"
	errProjectUnmodFields                  = "some project fields can't be modified (ID, Name, GroupID)"
	errProjectImageNotFound                = "specified DefaultImage not found"
	errClusterServiceTypeEmpty             = "service type field can't be empty"
	errServiceTypeUnmodFields              = "some service types fields can't be modified (ID, Type)"
	errServiceTypeUnmodVersionsField       = "service types versions field can't be modified in this response. Use specified one"
	errServiceTypeVersionConfigUnmodFields = "some service type version config fields can't be modified (ParameterName, AnsibleVarName)"
	errServiceTypeUnmodVersionFields       = "service types version fields (config, dependencies) can't be modified in this response. Use specified one"
	errServiceTypeDeleteVersionDefault     = "service type version set in default version"
	errConfigServiceTypeDependenceExists   = "service type presents in dependencies for another service"
	errServiceTypeVersionUnmodFields       = "some service type version fields can't be modified (ID, Version)"
	errServiceTypeVersionEmptyVersionField = "version field must be set"
)

var (
	// cluster:
	ErrClusterBadName            = errors.New(errClusterBadName)
	ErrClusterNhostsZero         = errors.New(errClusterNhostsZero)
	ErrClustersNhostsMasterSlave = errors.New(errClustersNhostsMasterSlave)
	ErrClusterImageNotFound      = errors.New(errClusterImageNotFound)

	// flavor:
	ErrFlavorIdNotEmpty = errors.New(errFlavorIdNotEmpty)
	ErrFlavorEmptyName  = errors.New(errFlavorEmptyName)

	// image:
	ErrImageIdNotEmpty = errors.New(errImageIdNotEmpty)

	// project:
	ErrProjectValidation    = errors.New(errProjectValidation)
	ErrProjectExisted       = errors.New(errProjectExisted)
	ErrProjectUnmodFields   = errors.New(errProjectUnmodFields)
	ErrProjectImageNotFound = errors.New(errProjectImageNotFound)

	// service:
	ErrClusterServiceTypeEmpty = errors.New(errClusterServiceTypeEmpty)

	// service type:
	ErrServiceTypeUnmodFields              = errors.New(errServiceTypeUnmodFields)
	ErrServiceTypeUnmodVersionsField       = errors.New(errServiceTypeUnmodVersionsField)
	ErrServiceTypeVersionUnmodFields       = errors.New(errServiceTypeVersionUnmodFields)
	ErrServiceTypeVersionEmptyVersionField = errors.New(errServiceTypeVersionEmptyVersionField)
	ErrServiceTypeUnmodVersionFields       = errors.New(errServiceTypeUnmodVersionFields)
	ErrServiceTypeDeleteVersionDefault     = errors.New(errServiceTypeDeleteVersionDefault)
	ErrConfigServiceTypeDependenceExists   = errors.New(errConfigServiceTypeDependenceExists)
	ErrServiceTypeVersionConfigUnmodFields = errors.New(errServiceTypeVersionConfigUnmodFields)
)

func ErrFlavorFieldValueNotFound(param string) error {
	ErrParamType := fmt.Errorf("specified %s not found", param)
	HandlerValidateErrorMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func ErrFlavorParamVal(param string) error {
	ErrParamVal := fmt.Errorf("flavor %s can't be less than or equal to zero", param)
	HandlerValidateErrorMap[ErrParamVal] = utils.ValidationError
	return ErrParamVal
}

func ErrFlavorParamType(param string) error {
	ErrParamType := fmt.Errorf("flavor %s must be int type", param)
	HandlerValidateErrorMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func ErrImageValidationParam(param string) error {
	ErrParamType := fmt.Errorf("image %s can't be empty", param)
	HandlerValidateErrorMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func ErrProjectFieldEmpty(param string) error {
	ErrParamType := fmt.Errorf("required project field '%s' is empty", param)
	HandlerValidateErrorMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func ErrProjectFieldIsGenerated(param string) error {
	ErrParamType := fmt.Errorf("project %s is generated field. It can't be filled in by user", param)
	HandlerValidateErrorMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func ErrClusterServiceVersionsEmpty(param string) error {
	ErrParamType := fmt.Errorf("'%s' service version and default version are not specified", param)
	HandlerValidateErrorMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func ErrConfigServiceTypeDependenceVersionExists(param1 string, param2 string) error {
	ErrParamType := fmt.Errorf("service type version %s presents in dependencies versions in %s service", param1, param2)
	HandlerValidateErrorMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func ErrClusterServiceConfigIncorrectType(param string, service string) error {
	ErrParamType := fmt.Errorf("'%s' service config param '%s' has incorrect value type", service, param)
	HandlerValidateErrorMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func ErrClusterServiceConfigNotPossibleValue(param string, service string) error {
	ErrParamType := fmt.Errorf("'%s' service config param '%s' value is not supported", service, param)
	HandlerValidateErrorMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func ErrClusterServiceConfigNotSupported(param string, service string) error {
	ErrParamType := fmt.Errorf("'%s' service config param name '%s' is not supported", service, param)
	HandlerValidateErrorMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func init() {
	HandlerValidateErrorMap[ErrClusterBadName] = utils.ValidationError
	HandlerValidateErrorMap[ErrClusterNhostsZero] = utils.ValidationError
	HandlerValidateErrorMap[ErrClustersNhostsMasterSlave] = utils.ValidationError
	HandlerValidateErrorMap[ErrClusterImageNotFound] = utils.ValidationError
	HandlerValidateErrorMap[ErrFlavorIdNotEmpty] = utils.ValidationError
	HandlerValidateErrorMap[ErrFlavorEmptyName] = utils.ValidationError
	HandlerValidateErrorMap[ErrImageIdNotEmpty] = utils.ValidationError
	HandlerValidateErrorMap[ErrProjectUnmodFields] = utils.ObjectUnmodified
	HandlerValidateErrorMap[ErrProjectValidation] = utils.ValidationError
	HandlerValidateErrorMap[ErrProjectExisted] = utils.ObjectExists
	HandlerValidateErrorMap[ErrProjectImageNotFound] = utils.ValidationError
	HandlerValidateErrorMap[ErrClusterServiceTypeEmpty] = utils.ValidationError
	HandlerValidateErrorMap[ErrServiceTypeUnmodFields] = utils.ValidationError
	HandlerValidateErrorMap[ErrServiceTypeUnmodVersionsField] = utils.ValidationError
	HandlerValidateErrorMap[ErrConfigServiceTypeDependenceExists] = utils.ObjectUsed
	HandlerValidateErrorMap[ErrServiceTypeVersionUnmodFields] = utils.ValidationError
	HandlerValidateErrorMap[ErrServiceTypeVersionEmptyVersionField] = utils.ValidationError
	HandlerValidateErrorMap[ErrServiceTypeUnmodVersionFields] = utils.DatabaseError

	HandlerValidateErrorMap[ErrServiceTypeDeleteVersionDefault] = utils.ValidationError
	HandlerValidateErrorMap[ErrServiceTypeVersionConfigUnmodFields] = utils.ValidationError

}
