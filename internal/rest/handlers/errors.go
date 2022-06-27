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
	errProjectValidation    = "project validation error"
	errProjectExisted       = "project with this name already exists"
	errProjectImageNotFound = "specified DefaultImage not found"
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
)

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

func ErrProjectFlavorNotFound(param string) error {
	ErrParamType := fmt.Errorf("specified %s not found", param)
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
}

func main() {}
