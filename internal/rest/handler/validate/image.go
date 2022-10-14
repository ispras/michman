package validate

import (
	"github.com/ispras/michman/internal/database"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/rest/handler/check"
	"github.com/ispras/michman/internal/rest/response"
	"github.com/ispras/michman/internal/utils"
)

// ImageCreate validates fields of the image structure for correct filling when creating
func ImageCreate(db database.Database, image *protobuf.Image) error {
	if image.ID != "" {
		return ErrImageGeneratedField
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

// ImageUpdate validates fields of the image structure for correct filling when updating
func ImageUpdate(db database.Database, oldImage *protobuf.Image, newImage *protobuf.Image) error {
	used, err := check.ImageUsed(db, oldImage.Name)
	if err != nil {
		return err
	}
	if used {
		return ErrImageUsed
	}

	if newImage.ID != "" {
		return ErrImageUnmodFields
	}

	if newImage.Name != "" && oldImage.Name != newImage.Name {
		_, err := db.ReadImage(newImage.Name)
		if err == nil {
			return ErrObjectExists("image", newImage.Name)
		}
		if err != nil && response.ErrorClass(err) != utils.ObjectNotFound {
			return err
		}
	}

	return nil
}

// ImageDelete checks the structure of the image for use when deleting
func ImageDelete(db database.Database, image *protobuf.Image) error {
	used, err := check.ImageUsed(db, image.Name)
	if err != nil {
		return err
	}
	if used {
		return ErrImageUsed
	}

	return nil
}
