package validate

import (
	"github.com/ispras/michman/internal/protobuf"
	"github.com/sirupsen/logrus"
)

func Image(logger *logrus.Logger, image *protobuf.Image) error {
	logger.Info("Validating image...")
	if image.ID != "" {
		return ErrImageIdNotEmpty
	}
	if image.Name == "" {
		return ErrImageValidationParam("Name")
	}
	if image.AnsibleUser == "" {
		return ErrImageValidationParam("AnsibleUser")
	}
	if image.CloudImageID == "" {
		return ErrImageValidationParam("ImageID")
	}
	return nil
}
