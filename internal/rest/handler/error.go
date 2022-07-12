package handler

import (
	"errors"
	"fmt"
	"github.com/ispras/michman/internal/rest/handler/response"
	"github.com/ispras/michman/internal/utils"
)

const (
	WarnFlavorVCPU = "VCPUs value is less than or equal to zero (Ignored)"
	WarnFlavorRAM  = "RAM value is less than or equal to zero (Ignored)"
	WarnFlavorDisk = "Disk value is less than or equal to zero (Ignored)"
)

const (
	errJsonIncorrect = "incorrect json"
	errUuidLibError  = "uuid generating error"

	//flavor:
	errFlavorNotFound   = "flavor with this id or name not found"
	errFlavorValidation = "flavor validation error"
	errFlavorUsed       = "flavor already in use. it can't be modified or deleted"
	errFlavorUnmodField = "some flavor fields can't be modified (ID)"
	errFlavorExisted    = "flavor with this name already exists"

	//image:
	errImageNotFound    = "image with this name not found"
	errImageUsed        = "image already in use. it can't be modified or deleted"
	errImageUnmodFields = "some image fields can't be modified (ID)"
	errImageValidation  = "image validation error"
	errImageExisted     = "image with this name already exists"

	//project:
	errProjectNotFound    = "project with this id or name not found"
	errProjectHasClusters = "project has clusters. Delete them first"

	//cluster:
	errClusterNotFound                    = "cluster with this id or name not found"
	errClusterUnmodFields                 = "some cluster fields can't be modified"
	errClusterStatus                      = "cluster status must be 'ACTIVE' or 'FAILED' for UPDATE or DELETE"
	errClusterExisted                     = "cluster with this name already exists"
	errClusterServicesIncompatibleVersion = "incompatible versions between services"

	//logs
	errBadActionParam = "bad action param. Supported query variables for action parameter are 'create', 'update' and 'delete'. Action 'create' is default"

	//service type:
	errServiceTypeNotFound                = "service type with this id or name not found"
	errServiceTypeExisted                 = "service type with this name already exists"
	errGetQueryParams                     = "bad view param. Supported query variables for view parameter are 'full' and 'summary', 'summary' is default"
	errServiceTypeVersionNotFound         = "service type version with this id or name not found"
	errServiceTypeVersionExisted          = "service type version with this name already exists"
	errServiceTypeVersionConfigNotFound   = "service type version config with this id or name not found"
	errServiceTypeVersionConfigExists     = "service type version config with this id or name already exists"
	errServiceTypeVersionDependencyExists = "service type version dependency with this service type already exists"
)

var (
	ErrJsonIncorrect                      = errors.New(errJsonIncorrect)
	ErrFlavorValidation                   = errors.New(errFlavorValidation)
	ErrUuidLibError                       = errors.New(errUuidLibError)
	ErrFlavorNotFound                     = errors.New(errFlavorNotFound)
	ErrFlavorUsed                         = errors.New(errFlavorUsed)
	ErrFlavorUnmodField                   = errors.New(errFlavorUnmodField)
	ErrFlavorExisted                      = errors.New(errFlavorExisted)
	ErrImageNotFound                      = errors.New(errImageNotFound)
	ErrImageUsed                          = errors.New(errImageUsed)
	ErrImageUnmodFields                   = errors.New(errImageUnmodFields)
	ErrImageValidation                    = errors.New(errImageValidation)
	ErrImageExisted                       = errors.New(errImageExisted)
	ErrProjectNotFound                    = errors.New(errProjectNotFound)
	ErrProjectHasClusters                 = errors.New(errProjectHasClusters)
	ErrClusterNotFound                    = errors.New(errClusterNotFound)
	ErrClusterUnmodFields                 = errors.New(errClusterUnmodFields)
	ErrClusterStatus                      = errors.New(errClusterStatus)
	ErrClusterExisted                     = errors.New(errClusterExisted)
	ErrClusterServicesIncompatibleVersion = errors.New(errClusterServicesIncompatibleVersion)
	ErrLogsBadActionParam                 = errors.New(errBadActionParam)
	ErrServiceTypeNotFound                = errors.New(errServiceTypeNotFound)
	ErrServiceTypeExisted                 = errors.New(errServiceTypeExisted)
	ErrGetQueryParams                     = errors.New(errGetQueryParams)
	ErrServiceTypeVersionExisted          = errors.New(errServiceTypeVersionExisted)
	ErrServiceTypeVersionNotFound         = errors.New(errServiceTypeVersionNotFound)
	ErrServiceTypeVersionConfigNotFound   = errors.New(errServiceTypeVersionConfigNotFound)
	ErrServiceTypeVersionConfigExists     = errors.New(errServiceTypeVersionConfigExists)
	ErrServiceTypeVersionDependencyExists = errors.New(errServiceTypeVersionDependencyExists)
)

func ErrClusterServiceHealthCheck(service string) error {
	ErrParamType := fmt.Errorf("'%s' HealthCheck field is empty", service)
	response.HandlerErrorsMap[ErrParamType] = utils.ValidationError
	return ErrParamType
}

func init() {
	response.HandlerErrorsMap[ErrJsonIncorrect] = utils.JsonError
	response.HandlerErrorsMap[ErrFlavorValidation] = utils.ValidationError
	response.HandlerErrorsMap[ErrUuidLibError] = utils.LibError
	response.HandlerErrorsMap[ErrFlavorNotFound] = utils.DatabaseError
	response.HandlerErrorsMap[ErrFlavorUsed] = utils.ObjectUsed
	response.HandlerErrorsMap[ErrFlavorUnmodField] = utils.ObjectUnmodified
	response.HandlerErrorsMap[ErrFlavorExisted] = utils.ObjectExists
	response.HandlerErrorsMap[ErrImageNotFound] = utils.DatabaseError
	response.HandlerErrorsMap[ErrImageUsed] = utils.ObjectUsed
	response.HandlerErrorsMap[ErrImageUnmodFields] = utils.ObjectUnmodified
	response.HandlerErrorsMap[ErrImageValidation] = utils.ValidationError
	response.HandlerErrorsMap[ErrImageExisted] = utils.ObjectExists
	response.HandlerErrorsMap[ErrProjectNotFound] = utils.DatabaseError
	response.HandlerErrorsMap[ErrProjectHasClusters] = utils.ObjectUsed
	response.HandlerErrorsMap[ErrClusterNotFound] = utils.DatabaseError
	response.HandlerErrorsMap[ErrClusterUnmodFields] = utils.ObjectUnmodified
	response.HandlerErrorsMap[ErrClusterStatus] = utils.DatabaseError
	response.HandlerErrorsMap[ErrClusterExisted] = utils.ObjectExists
	response.HandlerErrorsMap[ErrClusterServicesIncompatibleVersion] = utils.ValidationError
	response.HandlerErrorsMap[ErrLogsBadActionParam] = utils.LogsError
	response.HandlerErrorsMap[ErrServiceTypeNotFound] = utils.DatabaseError
	response.HandlerErrorsMap[ErrServiceTypeExisted] = utils.ValidationError
	response.HandlerErrorsMap[ErrGetQueryParams] = utils.ValidationError
	response.HandlerErrorsMap[ErrServiceTypeVersionNotFound] = utils.DatabaseError
	response.HandlerErrorsMap[ErrServiceTypeVersionExisted] = utils.DatabaseError
	response.HandlerErrorsMap[ErrServiceTypeVersionConfigNotFound] = utils.DatabaseError
	response.HandlerErrorsMap[ErrServiceTypeVersionConfigExists] = utils.ValidationError
	response.HandlerErrorsMap[ErrServiceTypeVersionDependencyExists] = utils.ValidationError
}
