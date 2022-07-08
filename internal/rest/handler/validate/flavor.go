package validate

import (
	"github.com/ispras/michman/internal/database"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/rest/handler/check"
	"net/http"
)

// FlavorCreate validates fields of the flavor structure for correct filling when creating
func FlavorCreate(flavor *protobuf.Flavor) error {
	if flavor.ID != "" {
		return ErrFlavorIdNotEmpty
	}
	if flavor.Name == "" {
		return ErrFlavorEmptyName
	}
	if flavor.VCPUs == 0 || flavor.Disk == 0 || flavor.RAM == 0 {
		return ErrFlavorZeroField
	}

	err := check.FlavorFields(flavor)
	if err != nil {
		return err
	}
	return nil
}

// FlavorUpdate validates fields of the flavor structure for correct filling when updating
func FlavorUpdate(db database.Database, oldFlavor *protobuf.Flavor, newFlavor *protobuf.Flavor) (error, int) {
	if newFlavor.Name != "" && newFlavor.Name != oldFlavor.Name {
		dbFlavor, err := db.ReadFlavor(newFlavor.Name)
		if err != nil {
			return err, http.StatusInternalServerError
		}
		if dbFlavor.ID != "" {
			return ErrFlavorExisted, http.StatusBadRequest
		}
	}

	err := check.FlavorFields(newFlavor)
	if err != nil {
		return err, http.StatusBadRequest
	}
	return nil, 0
}
