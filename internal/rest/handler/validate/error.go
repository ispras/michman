package validate

import (
	"fmt"
	"github.com/ispras/michman/internal/rest"
	"github.com/ispras/michman/internal/utils"
)

const (
	// flavor:
	errFlavorGeneratedField = "flavor ID is generated field. It can't be filled in by user"
	errFlavorEmptyField     = "flavor Name | VCPUs | Disk | RAM can't be empty or equal to zero"
	errFlavorUnmodField     = "some flavor fields can't be modified (ID)"
	errFlavorUsed           = "flavor already in use. it can't be modified or deleted"

	// cluster:
	errClusterNSlavesZero         = "NSlaves parameter must be number >= 0"
	errClustersNSlavesMasterSlave = "NSlaves parameter must be number >= 1 because master-slave services will be installed"
	errClusterStatus              = "cluster status must be 'ACTIVE' or 'FAILED' for UPDATE or DELETE"

	// image:
	errImageGeneratedField = "image ID is generated field. It can't be filled in by user"
	errImageUsed           = "image already in use. it can't be modified or deleted"
	errImageUnmodFields    = "some image fields can't be modified (ID)"

	// project:
	errProjectUnmodFields      = "some project fields can't be modified (ID, Name)"
	errProjectHasClusters      = "project has clusters. Delete them first"
	errClusterServiceTypeEmpty = "service type field can't be empty"

	// service type:
	errServiceTypeUnmodFields                  = "some service types fields can't be modified (ID, Type)"
	errServiceTypeUnmodVersionsField           = "service types versions field can't be modified in this response. Use specified one"
	errServiceTypeVersionUnmodFields           = "some service type version fields can't be modified (ID, Version)"
	errServiceTypeDeleteVersionDefault         = "service type version set in default version"
	errServiceTypeVersionConfigUnmodFields     = "some service type version config fields can't be modified (ID, ParameterName, AnsibleVarName)"
	errServiceTypeVersionDependencyUnmodFields = "some service types version dependency fields can't be modified (Service Type)"
	errServiceTypeUnmodVersionFields           = "service types version fields (config, dependencies) can't be modified in this response. Use specified one"
	errServiceTypeVersionEmptyVersionField     = "version field must be set"
)

var (
	// cluster:
	ErrClusterNSlavesZero         = rest.MakeError(errClusterNSlavesZero, utils.ValidationError)
	ErrClustersNSlavesMasterSlave = rest.MakeError(errClustersNSlavesMasterSlave, utils.ValidationError)
	ErrClusterStatus              = rest.MakeError(errClusterStatus, utils.ValidationError)

	// flavor:
	ErrFlavorGeneratedField = rest.MakeError(errFlavorGeneratedField, utils.ValidationError)
	ErrFlavorEmptyField     = rest.MakeError(errFlavorEmptyField, utils.ValidationError)
	ErrFlavorUnmodField     = rest.MakeError(errFlavorUnmodField, utils.ValidationError)
	ErrFlavorUsed           = rest.MakeError(errFlavorUsed, utils.ObjectUsed)

	// image:
	ErrImageGeneratedField = rest.MakeError(errImageGeneratedField, utils.ValidationError)
	ErrImageUsed           = rest.MakeError(errImageUsed, utils.ValidationError)
	ErrImageUnmodFields    = rest.MakeError(errImageUnmodFields, utils.ValidationError)

	// project:
	ErrProjectUnmodFields = rest.MakeError(errProjectUnmodFields, utils.ValidationError)
	ErrProjectHasClusters = rest.MakeError(errProjectHasClusters, utils.ValidationError)

	// service:
	ErrClusterServiceTypeEmpty = rest.MakeError(errClusterServiceTypeEmpty, utils.ValidationError)

	// service type:
	ErrServiceTypeUnmodFields                  = rest.MakeError(errServiceTypeUnmodFields, utils.ValidationError)
	ErrServiceTypeUnmodVersionsField           = rest.MakeError(errServiceTypeUnmodVersionsField, utils.ValidationError)
	ErrServiceTypeVersionUnmodFields           = rest.MakeError(errServiceTypeVersionUnmodFields, utils.ObjectUnmodified)
	ErrServiceTypeUnmodVersionFields           = rest.MakeError(errServiceTypeUnmodVersionFields, utils.ValidationError)
	ErrServiceTypeDeleteVersionDefault         = rest.MakeError(errServiceTypeDeleteVersionDefault, utils.ValidationError)
	ErrServiceTypeVersionConfigUnmodFields     = rest.MakeError(errServiceTypeVersionConfigUnmodFields, utils.ObjectUnmodified)
	ErrServiceTypeVersionDependencyUnmodFields = rest.MakeError(errServiceTypeVersionDependencyUnmodFields, utils.ValidationError)
	ErrServiceTypeVersionEmptyVersionField     = rest.MakeError(errServiceTypeVersionEmptyVersionField, utils.ValidationError)
)

// common:
func ErrObjectExists(object string, idOrName string) error {
	errMessage := fmt.Sprintf("%s with this name or id (%s) already exists", object, idOrName)
	return rest.MakeError(errMessage, utils.ObjectExists)
}

func ErrGeneratedField(object, field string) error {
	errMessage := fmt.Sprintf("%s '%s' field is generated and can't be filled by user", object, field)
	return rest.MakeError(errMessage, utils.ValidationError)
}

func ErrEmptyField(object, field string) error {
	errMessage := fmt.Sprintf("required %s field '%s' is empty", object, field)
	return rest.MakeError(errMessage, utils.ValidationError)
}

// image:
func ErrImageValidationParam(param string) error {
	errMessage := fmt.Sprintf("image %s can't be empty", param)
	return rest.MakeError(errMessage, utils.ValidationError)
}

// cluster:
func ErrClusterServiceVersionsEmpty(param string) error {
	errMessage := fmt.Sprintf("'%s' service version and default version are not specified", param)
	return rest.MakeError(errMessage, utils.ValidationError)
}

func ErrClusterUnmodFields(field string) error {
	errMessage := fmt.Sprintf("cluster field '%s' can't be modified", field)
	return rest.MakeError(errMessage, utils.ObjectUnmodified)
}

// project:

//func ErrProjectFieldIsGenerated(param string) error {
//	errMessage := fmt.Sprintf("project %s is generated field. It can't be filled in by user", param)
//	return rest.MakeError(errMessage, utils.ValidationError)
//}
