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
	errFlavorZeroField                     = "flavor VCPUs | Disk | RAM can't be zero"
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
	errServiceTypeVersionUnmodFields       = "some service type version fields can't be modified (ID, Version)"
	errServiceTypeVersionEmptyVersionField = "version field must be set"
	errFlavorExisted                       = "flavor with this name already exists"
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
	ErrFlavorExisted    = errors.New(errFlavorExisted)
	ErrFlavorZeroField  = errors.New(errFlavorZeroField)

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
	ErrServiceTypeVersionConfigUnmodFields = errors.New(errServiceTypeVersionConfigUnmodFields)
)

func ErrFlavorFieldValueNotFound(param string) error {
	ErrParamType := fmt.Errorf("specified %s not found", param)
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
	HandlerValidateErrorMap[ErrServiceTypeVersionUnmodFields] = utils.ValidationError
	HandlerValidateErrorMap[ErrServiceTypeVersionEmptyVersionField] = utils.ValidationError
	HandlerValidateErrorMap[ErrServiceTypeUnmodVersionFields] = utils.DatabaseError
	HandlerValidateErrorMap[ErrFlavorExisted] = utils.ObjectExists
	HandlerValidateErrorMap[ErrServiceTypeDeleteVersionDefault] = utils.ValidationError
	HandlerValidateErrorMap[ErrServiceTypeVersionConfigUnmodFields] = utils.ValidationError
	HandlerValidateErrorMap[ErrFlavorZeroField] = utils.ValidationError
}
