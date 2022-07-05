package handlers

import (
	"errors"
	"fmt"
	"github.com/ispras/michman/internal/utils"
)

var HandlerErrorsMap = make(map[error]int)

const (
	warnFlavorVCPU = "VCPUs value is less than or equal to zero (Ignored)"
	warnFlavorRAM  = "RAM value is less than or equal to zero (Ignored)"
	warnFlavorDisk = "Disk value is less than or equal to zero (Ignored)"
)

const (
	errJsonIncorrect = "incorrect json"
	errJsonEncode    = "json encode error"
	errUuidLibError  = "uuid generating error"

	//flavor:
	errFlavorNotFound   = "flavor with this id or name not found"
	errFlavorValidation = "flavor validation error"
	errFlavorUsed       = "flavor already in use. it can't be modified or deleted"
	errFlavorUnmodField = "some flavor fields can't be modified (ID)"
	errFlavorExisted    = "flavor with this name already exists"
	errFlavorIdNotEmpty = "flavor ID is generated field. It can't be filled in by user"
	errFlavorEmptyName  = "flavor Name can't be empty"

	//image:
	errImageNotFound    = "image with this name not found"
	errImageUsed        = "image already in use. it can't be modified or deleted"
	errImageUnmodFields = "some image fields can't be modified (ID)"
	errImageValidation  = "image validation error"
	errImageExisted     = "image with this name already exists"
	errImageIdNotEmpty  = "image ID is generated field. It can't be filled in by user"

	//project:
	errProjectNotFound      = "project with this id or name not found"
	errProjectHasClusters   = "project has clusters. Delete them first"
	errProjectUnmodFields   = "some project fields can't be modified (ID, Name, GroupID)"
	errProjectValidation    = "project validation error. Bad name. You should use only alpha-numeric characters and '-' symbols and only alphabetic characters for leading symbol"
	errProjectExisted       = "project with this name already exists"
	errProjectImageNotFound = "specified DefaultImage not found"

	//cluster:
	errClusterNotFound                    = "cluster with this id or name not found"
	errClusterUnmodFields                 = "some cluster fields can't be modified"
	errClusterStatus                      = "cluster status must be 'ACTIVE' or 'FAILED' for UPDATE or DELETE"
	errClusterExisted                     = "cluster with this name already exists"
	errClusterBadName                     = "cluster validation error. Bad name. You should use only alpha-numeric characters and '-' symbols and only alphabetic characters for leading symbol"
	errClusterNhostsZero                  = "NHosts parameter must be number >= 0"
	errClustersNhostsMasterSlave          = "NHosts parameter must be number >= 1 if you want to install master-slave services"
	errClusterServiceTypeEmpty            = "service type field can't be empty"
	errClusterServicesIncompatibleVersion = "Incompatible versions between services"
	errClusterImageNotFound               = "specified Image not found"

	//logs
	errBadActionParam = "bad action param. Supported query variables for action parameter are 'create', 'update' and 'delete'. Action 'create' is default"

	//service type:
	errServiceTypeUnmodFields                        = "some service types fields can't be modified (ID, Type)"
	errServiceTypeUnmodVersionsField                 = "service types versions field can't be modified in this response. Use specified one"
	errServiceTypeNotFound                           = "service type with this id or name not found"
	errServiceTypeExisted                            = "service type with this name already exists"
	errServiceTypeClass                              = "class for service type is not supported"
	errServiceTypeAccessPort                         = "port is incorrect"
	errServiceTypeDefaultVersion                     = "default version not found in versions list"
	errConfigPossibleValueEmpty                      = "config possible value is empty"
	errConfigDependencyServiceVersionEmpty           = "service versions list in dependencies can't be empty"
	errConfigDependencyServiceDefaultVersionEmpty    = "service default version in dependency can't be empty"
	errConfigServiceDependencyVersionNotFound        = "service version in dependency doesn't exist"
	errConfigServiceDependencyDefaultVersionNotFound = "service default version in dependencies doesn't exist"
	errGetQueryParams                                = "bad view param. Supported query variables for view parameter are 'full' and 'summary', 'summary' is default"
	errConfigServiceTypeDependenceExists             = "service type presents in dependencies for another service"

	errServiceTypeVersionUnmodFields       = "some service type version fields can't be modified (ID, Version)"
	errServiceTypeVersionEmptyVersionField = "version field must be set"
	errServiceTypeVersionNotFound          = "service type version with this id or name not found"
	errServiceTypeVersionExisted           = "service type version with this name already exists"
	errServiceTypeUnmodVersionFields       = "service types version fields (config, dependencies) can't be modified in this response. Use specified one"
	errServiceTypeDeleteVersionDefault     = "service type version set in default version"

	errServiceTypeVersionConfigNotFound          = "service type version config with this id or name not found"
	errServiceTypeVersionConfigExists            = "service type version config with this id or name already exists"
	errServiceTypeVersionConfigDefaultValueEmpty = "service type version config default value must be set"
	errServiceTypeVersionConfigDefaultValue      = "service type version config default value not in possible values"

	errServiceTypeVersionConfigUnmodFields = "some service type version config fields can't be modified (ParameterName, AnsibleVarName)"
)

var (
	ErrJsonIncorrect    = errors.New(errJsonIncorrect)
	ErrJsonEncode       = errors.New(errJsonEncode)
	ErrFlavorValidation = errors.New(errFlavorValidation)
	ErrUuidLibError     = errors.New(errUuidLibError)

	ErrFlavorNotFound   = errors.New(errFlavorNotFound)
	ErrFlavorUsed       = errors.New(errFlavorUsed)
	ErrFlavorUnmodField = errors.New(errFlavorUnmodField)
	ErrFlavorExisted    = errors.New(errFlavorExisted)
	ErrFlavorIdNotEmpty = errors.New(errFlavorIdNotEmpty)
	ErrFlavorEmptyName  = errors.New(errFlavorEmptyName)

	ErrImageNotFound    = errors.New(errImageNotFound)
	ErrImageUsed        = errors.New(errImageUsed)
	ErrImageUnmodFields = errors.New(errImageUnmodFields)
	ErrImageValidation  = errors.New(errImageValidation)
	ErrImageExisted     = errors.New(errImageExisted)
	ErrImageIdNotEmpty  = errors.New(errImageIdNotEmpty)

	ErrProjectNotFound      = errors.New(errProjectNotFound)
	ErrProjectHasClusters   = errors.New(errProjectHasClusters)
	ErrProjectUnmodFields   = errors.New(errProjectUnmodFields)
	ErrProjectValidation    = errors.New(errProjectValidation)
	ErrProjectExisted       = errors.New(errProjectExisted)
	ErrProjectImageNotFound = errors.New(errProjectImageNotFound)

	ErrClusterNotFound                    = errors.New(errClusterNotFound)
	ErrClusterUnmodFields                 = errors.New(errClusterUnmodFields)
	ErrClusterStatus                      = errors.New(errClusterStatus)
	ErrClusterExisted                     = errors.New(errClusterExisted)
	ErrClusterBadName                     = errors.New(errClusterBadName)
	ErrClusterNhostsZero                  = errors.New(errClusterNhostsZero)
	ErrClustersNhostsMasterSlave          = errors.New(errClustersNhostsMasterSlave)
	ErrClusterServiceTypeEmpty            = errors.New(errClusterServiceTypeEmpty)
	ErrClusterServicesIncompatibleVersion = errors.New(errClusterServicesIncompatibleVersion)
	ErrClusterImageNotFound               = errors.New(errClusterImageNotFound)

	ErrLogsBadActionParam = errors.New(errBadActionParam)

	ErrServiceTypeUnmodFields                        = errors.New(errServiceTypeUnmodFields)
	ErrServiceTypeUnmodVersionsField                 = errors.New(errServiceTypeUnmodVersionsField)
	ErrServiceTypeNotFound                           = errors.New(errServiceTypeNotFound)
	ErrServiceTypeExisted                            = errors.New(errServiceTypeExisted)
	ErrServiceTypeClass                              = errors.New(errServiceTypeClass)
	ErrServiceTypePort                               = errors.New(errServiceTypeAccessPort)
	ErrServiceTypeDefaultVersion                     = errors.New(errServiceTypeDefaultVersion)
	ErrConfigPossibleValueEmpty                      = errors.New(errConfigPossibleValueEmpty)
	ErrConfigDependencyServiceVersionEmpty           = errors.New(errConfigDependencyServiceVersionEmpty)
	ErrConfigDependencyServiceDefaultVersionEmpty    = errors.New(errConfigDependencyServiceDefaultVersionEmpty)
	ErrConfigServiceDependencyVersionNotFound        = errors.New(errConfigServiceDependencyVersionNotFound)
	ErrConfigServiceDependencyDefaultVersionNotFound = errors.New(errConfigServiceDependencyDefaultVersionNotFound)
	ErrGetQueryParams                                = errors.New(errGetQueryParams)
	ErrConfigServiceTypeDependenceExists             = errors.New(errConfigServiceTypeDependenceExists)
	ErrServiceTypeVersionUnmodFields                 = errors.New(errServiceTypeVersionUnmodFields)

	ErrServiceTypeVersionExisted           = errors.New(errServiceTypeVersionExisted)
	ErrServiceTypeVersionNotFound          = errors.New(errServiceTypeVersionNotFound)
	ErrServiceTypeVersionEmptyVersionField = errors.New(errServiceTypeVersionEmptyVersionField)
	ErrServiceTypeUnmodVersionFields       = errors.New(errServiceTypeUnmodVersionFields)
	ErrServiceTypeDeleteVersionDefault     = errors.New(errServiceTypeDeleteVersionDefault)

	ErrServiceTypeVersionConfigNotFound          = errors.New(errServiceTypeVersionConfigNotFound)
	ErrServiceTypeVersionConfiqDefaultValutEmpty = errors.New(errServiceTypeVersionConfigDefaultValueEmpty)
	ErrServiceTypeVersionConfigDefaultValue      = errors.New(errServiceTypeVersionConfigDefaultValue)

	ErrServiceTypeVersionConfigExists      = errors.New(errServiceTypeVersionConfigExists)
	ErrServiceTypeVersionConfigUnmodFields = errors.New(errServiceTypeVersionConfigUnmodFields)
)

func ErrValidTypeParam(param string) error {
	ErrParamType := fmt.Errorf("parameter type must be int, float, bool, string. Got: %s", param)
	HandlerErrorsMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func ErrFlavorParamVal(param string) error {
	ErrParamVal := fmt.Errorf("flavor %s can't be less than or equal to zero", param)
	HandlerErrorsMap[ErrParamVal] = utils.ValidationError
	return ErrParamVal
}

func ErrFlavorParamType(param string) error {
	ErrParamType := fmt.Errorf("flavor %s must be int type", param)
	HandlerErrorsMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func ErrImageValidationParam(param string) error {
	ErrParamType := fmt.Errorf("image %s can't be empty", param)
	HandlerErrorsMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func ErrProjectFieldEmpty(param string) error {
	ErrParamType := fmt.Errorf("required project field '%s' is empty", param)
	HandlerErrorsMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func ErrProjectFieldIsGenerated(param string) error {
	ErrParamType := fmt.Errorf("project %s is generated field. It can't be filled in by user", param)
	HandlerErrorsMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func ErrFlavorFieldValueNotFound(param string) error {
	ErrParamType := fmt.Errorf("specified %s not found", param)
	HandlerErrorsMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func ErrServiceTypeVersionUnique(param string) error {
	ErrParamType := fmt.Errorf("version %s is not unique", param)
	HandlerErrorsMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func ErrServiceTypeVersionConfigUnique(param string) error {
	ErrParamType := fmt.Errorf("config with parameter name %s is not unique", param)
	HandlerErrorsMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func ErrServiceTypeVersionConfigPossibleValuesUnique(param string) error {
	ErrParamType := fmt.Errorf("config possible value %s is not unique", param)
	HandlerErrorsMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func ErrServiceTypeVersionConfigParamEmpty(param string) error {
	ErrParamType := fmt.Errorf("config parameter %s must be set", param)
	HandlerErrorsMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func ErrServiceTypeVersionConfigPossibleValues(param string) error {
	ErrParamType := fmt.Errorf("config possible value %s set incorrectly", param)
	HandlerErrorsMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func ErrServiceDependenciesNotExists(param string) error {
	ErrParamType := fmt.Errorf("service with type %s from dependencies doesn't exist", param)
	HandlerErrorsMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func ErrConfigServiceTypeDependenceVersionExists(param1 string, param2 string) error {
	ErrParamType := fmt.Errorf("service type version %s presents in dependencies versions in %s service", param1, param2)
	HandlerErrorsMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func ErrClusterServiceVersionsEmpty(param string) error {
	ErrParamType := fmt.Errorf("'%s' service version and default version are not specified", param)
	HandlerErrorsMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func ErrClusterServiceTypeNotSupported(param string) error {
	ErrParamType := fmt.Errorf("service '%s' is not supported", param)
	HandlerErrorsMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func ErrClusterServiceVersionNotSupported(param string, service string) error {
	ErrParamType := fmt.Errorf("'%s' service version '%s' is not supported", service, param)
	HandlerErrorsMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func ErrClusterServiceConfigNotSupported(param string, service string) error {
	ErrParamType := fmt.Errorf("'%s' service config param name '%s' is not supported", service, param)
	HandlerErrorsMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func ErrClusterServiceConfigIncorrectType(param string, service string) error {
	ErrParamType := fmt.Errorf("'%s' service config param '%s' has incorrect value type", service, param)
	HandlerErrorsMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func ErrClusterServiceConfigNotPossibleValue(param string, service string) error {
	ErrParamType := fmt.Errorf("'%s' service config param '%s' value is not supported", service, param)
	HandlerErrorsMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func ErrClusterServiceHealthCheck(service string) error {
	ErrParamType := fmt.Errorf("'%s' HealthCheck field is empty", service)
	HandlerErrorsMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func ErrClusterDependenceServicesIncompatibleVersion(service string, currentService string) error {
	ErrParamType := fmt.Errorf("service '%s' has incompatible version for service '%s'", service, currentService)
	HandlerErrorsMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func init() {
	HandlerErrorsMap[ErrJsonIncorrect] = utils.JsonError
	HandlerErrorsMap[ErrJsonEncode] = utils.JsonError
	HandlerErrorsMap[ErrFlavorValidation] = utils.ValidationError
	HandlerErrorsMap[ErrUuidLibError] = utils.LibError

	HandlerErrorsMap[ErrFlavorNotFound] = utils.DatabaseError
	HandlerErrorsMap[ErrFlavorUsed] = utils.ObjectUsed
	HandlerErrorsMap[ErrFlavorUnmodField] = utils.ObjectUnmodified
	HandlerErrorsMap[ErrFlavorExisted] = utils.ObjectExists
	HandlerErrorsMap[ErrFlavorIdNotEmpty] = utils.ValidationError
	HandlerErrorsMap[ErrFlavorEmptyName] = utils.ValidationError

	HandlerErrorsMap[ErrImageNotFound] = utils.DatabaseError
	HandlerErrorsMap[ErrImageUsed] = utils.ObjectUsed
	HandlerErrorsMap[ErrImageUnmodFields] = utils.ObjectUnmodified
	HandlerErrorsMap[ErrImageValidation] = utils.ValidationError
	HandlerErrorsMap[ErrImageExisted] = utils.ObjectExists
	HandlerErrorsMap[ErrImageIdNotEmpty] = utils.ValidationError

	HandlerErrorsMap[ErrProjectNotFound] = utils.DatabaseError
	HandlerErrorsMap[ErrProjectHasClusters] = utils.ObjectUsed
	HandlerErrorsMap[ErrProjectUnmodFields] = utils.ObjectUnmodified
	HandlerErrorsMap[ErrProjectValidation] = utils.ValidationError
	HandlerErrorsMap[ErrProjectExisted] = utils.ObjectExists
	HandlerErrorsMap[ErrProjectImageNotFound] = utils.ValidationError

	HandlerErrorsMap[ErrClusterNotFound] = utils.DatabaseError
	HandlerErrorsMap[ErrClusterUnmodFields] = utils.ObjectUnmodified
	HandlerErrorsMap[ErrClusterStatus] = utils.DatabaseError
	HandlerErrorsMap[ErrClusterExisted] = utils.ObjectExists
	HandlerErrorsMap[ErrClusterBadName] = utils.ValidationError
	HandlerErrorsMap[ErrClusterNhostsZero] = utils.ValidationError
	HandlerErrorsMap[ErrClustersNhostsMasterSlave] = utils.ValidationError
	HandlerErrorsMap[ErrClusterServiceTypeEmpty] = utils.ValidationError
	HandlerErrorsMap[ErrClusterServicesIncompatibleVersion] = utils.ValidationError
	HandlerErrorsMap[ErrClusterImageNotFound] = utils.ValidationError

	HandlerErrorsMap[ErrLogsBadActionParam] = utils.LogsError

	HandlerErrorsMap[ErrServiceTypeNotFound] = utils.DatabaseError
	HandlerErrorsMap[ErrServiceTypeExisted] = utils.ValidationError
	HandlerErrorsMap[ErrServiceTypeClass] = utils.ValidationError
	HandlerErrorsMap[ErrServiceTypePort] = utils.ValidationError
	HandlerErrorsMap[ErrServiceTypeDefaultVersion] = utils.ValidationError
	HandlerErrorsMap[ErrConfigPossibleValueEmpty] = utils.ValidationError
	HandlerErrorsMap[ErrConfigDependencyServiceVersionEmpty] = utils.ValidationError
	HandlerErrorsMap[ErrConfigDependencyServiceDefaultVersionEmpty] = utils.ValidationError
	HandlerErrorsMap[ErrConfigServiceDependencyVersionNotFound] = utils.ValidationError
	HandlerErrorsMap[ErrConfigServiceDependencyDefaultVersionNotFound] = utils.ValidationError
	HandlerErrorsMap[ErrServiceTypeUnmodFields] = utils.ValidationError
	HandlerErrorsMap[ErrGetQueryParams] = utils.ValidationError
	HandlerErrorsMap[ErrServiceTypeUnmodVersionsField] = utils.ValidationError
	HandlerErrorsMap[ErrConfigServiceTypeDependenceExists] = utils.ObjectUsed

	HandlerErrorsMap[ErrServiceTypeVersionUnmodFields] = utils.ValidationError
	HandlerErrorsMap[ErrServiceTypeVersionEmptyVersionField] = utils.ValidationError
	HandlerErrorsMap[ErrServiceTypeVersionNotFound] = utils.DatabaseError
	HandlerErrorsMap[ErrServiceTypeVersionExisted] = utils.DatabaseError
	HandlerErrorsMap[ErrServiceTypeUnmodVersionFields] = utils.DatabaseError

	HandlerErrorsMap[ErrServiceTypeDeleteVersionDefault] = utils.ValidationError

	HandlerErrorsMap[ErrServiceTypeVersionConfigNotFound] = utils.DatabaseError
	HandlerErrorsMap[ErrServiceTypeVersionConfiqDefaultValutEmpty] = utils.ValidationError
	HandlerErrorsMap[ErrServiceTypeVersionConfigDefaultValue] = utils.ValidationError

	HandlerErrorsMap[ErrServiceTypeVersionConfigExists] = utils.ValidationError

	HandlerErrorsMap[ErrServiceTypeVersionConfigUnmodFields] = utils.ValidationError
}
