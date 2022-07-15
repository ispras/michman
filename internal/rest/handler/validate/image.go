package validate

import (
	"github.com/ispras/michman/internal/database"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/rest/handler/check"
	"net/http"
)

// ImageCreate validates fields of the image structure for correct filling when creating
func ImageCreate(db database.Database, image *protobuf.Image) (error, int) {
	if image.ID != "" {
		return ErrImageIdNotEmpty, http.StatusBadRequest
	}
	if image.Name == "" {
		return ErrImageValidationParam("Name"), http.StatusBadRequest
	}
	if image.AnsibleUser == "" {
		return ErrImageValidationParam("AnsibleUser"), http.StatusBadRequest
	}
	if image.CloudImageID == "" {
		return ErrImageValidationParam("ImageID"), http.StatusBadRequest
	}

	dbImg, err := db.ReadImage(image.Name)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if dbImg.ID != "" {
		return ErrImageExisted, http.StatusBadRequest
	}

	return nil, http.StatusOK
}

// ImageUpdate validates fields of the image structure for correct filling when updating
func ImageUpdate(db database.Database, oldImage *protobuf.Image, newImage *protobuf.Image) (error, int) {
	used, err := check.ImageUsed(db, oldImage.Name)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if used {
		return ErrImageUsed, http.StatusBadRequest
	}

	if newImage.ID != "" {
		return ErrImageUnmodFields, http.StatusBadRequest
	}

	if newImage.Name != "" && oldImage.Name != newImage.Name {
		dbImg, err := db.ReadImage(newImage.Name)
		if err != nil {
			return err, http.StatusInternalServerError
		}
		if dbImg.ID != "" {
			return ErrImageExisted, http.StatusBadRequest
		}
	}

	return nil, http.StatusOK
}

// ImageDelete checks the structure of the image for use when deleting
func ImageDelete(db database.Database, image *protobuf.Image) (error, int) {
	used, err := check.ImageUsed(db, image.Name)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if used {
		return ErrImageUsed, http.StatusBadRequest
	}

	return nil, http.StatusOK
}
