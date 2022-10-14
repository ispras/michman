package validate

import (
	"github.com/ispras/michman/internal/database"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/rest/handler/check"
	"github.com/ispras/michman/internal/rest/response"
	"github.com/ispras/michman/internal/utils"
)

// FlavorCreate validates fields of the flavor structure for correct filling when creating
func FlavorCreate(flavor *protobuf.Flavor) error {
	if flavor.ID != "" {
		return ErrFlavorGeneratedField
	}
	if flavor.Name == "" || flavor.VCPUs == 0 || flavor.Disk == 0 || flavor.RAM == 0 {
		return ErrFlavorEmptyField
	}

	err := check.FlavorFields(flavor)
	if err != nil {
		return err
	}
	return nil
}

// FlavorUpdate validates fields of the flavor structure for correct filling when updating
func FlavorUpdate(db database.Database, oldFlavor *protobuf.Flavor, newFlavor *protobuf.Flavor) error {
	if newFlavor.ID != "" {
		return ErrFlavorUnmodField
	}

	if newFlavor.Name != "" && newFlavor.Name != oldFlavor.Name {
		dbFlavor, err := db.ReadFlavor(newFlavor.Name)
		if dbFlavor != nil {
			return ErrObjectExists("flavor", newFlavor.Name)
		}
		if err != nil && response.ErrorClass(err) != utils.ObjectNotFound {
			return err
		}
	}

	used, err := check.FlavorUsed(db, oldFlavor.Name)
	if err != nil {
		return err
	}
	if used {
		return ErrFlavorUsed
	}

	err = check.FlavorFields(newFlavor)
	if err != nil {
		return err
	}
	return nil
}

func FlavorDelete(db database.Database, flavor *protobuf.Flavor) error {
	used, err := check.FlavorUsed(db, flavor.Name)
	if err != nil {
		return err
	}
	if used {
		return ErrFlavorUsed
	}
	return nil
}
