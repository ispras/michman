package validate

import (
	"github.com/ispras/michman/internal/protobuf"
	"github.com/sirupsen/logrus"
)

func Flavor(logger *logrus.Logger, flavor *protobuf.Flavor) error {
	logger.Info("Validating flavor...")
	if flavor.ID != "" {
		return ErrFlavorIdNotEmpty
	}
	if flavor.Name == "" {
		return ErrFlavorEmptyName
	}

	switch interface{}(flavor.VCPUs).(type) {
	case int32:
		if flavor.VCPUs <= 0 {
			return ErrFlavorParamVal("VCPUs")
		}
	default:
		return ErrFlavorParamType("VCPUs")
	}

	switch interface{}(flavor.RAM).(type) {
	case int32:
		if flavor.RAM <= 0 {
			return ErrFlavorParamVal("RAM")
		}
	default:
		return ErrFlavorParamType("RAM")
	}

	switch interface{}(flavor.Disk).(type) {
	case int32:
		if flavor.Disk <= 0 {
			return ErrFlavorParamVal("Disk")
		}
	default:
		return ErrFlavorParamType("RAM")
	}
	return nil
}
