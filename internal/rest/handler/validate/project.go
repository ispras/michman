package validate

import (
	"github.com/ispras/michman/internal/database"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/rest/handler/check"
	"github.com/ispras/michman/internal/utils"
	"net/http"
)

//ProjectCreate validates fields of the project structure for correct filling when creating
func ProjectCreate(db database.Database, project *protobuf.Project) (error, int) {
	if project.DisplayName == "" {
		return ErrProjectFieldEmpty("DisplayName"), http.StatusBadRequest
	}
	if project.ID != "" {
		return ErrProjectFieldIsGenerated("ID"), http.StatusBadRequest
	}
	if project.Name != "" {
		return ErrProjectFieldIsGenerated("Name"), http.StatusBadRequest
	}
	if err, status := check.ValidName(project.DisplayName, utils.ProjectNamePattern, ErrProjectValidation); err != nil {
		return err, status
	}
	if project.DefaultImage == "" {
		return ErrProjectFieldEmpty("DefaultImage"), http.StatusBadRequest
	}
	if project.DefaultMasterFlavor == "" {
		return ErrProjectFieldEmpty("DefaultMasterFlavor"), http.StatusBadRequest
	}
	if project.DefaultSlavesFlavor == "" {
		return ErrProjectFieldEmpty("DefaultSlavesFlavor"), http.StatusBadRequest
	}
	if project.DefaultStorageFlavor == "" {
		return ErrProjectFieldEmpty("DefaultStorageFlavor"), http.StatusBadRequest
	}
	if project.DefaultMonitoringFlavor == "" {
		return ErrProjectFieldEmpty("DefaultMonitoringFlavor"), http.StatusBadRequest
	}

	dbRes, err := db.ReadProject(project.Name)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if dbRes.Name != "" {
		return ErrProjectExisted, http.StatusBadRequest
	}

	err, status := ProjectFieldsDb(db, project)
	if err != nil {
		return err, status
	}

	return nil, http.StatusOK
}

// ProjectUpdate validates fields of the project structure for correct filling when updating
func ProjectUpdate(db database.Database, project *protobuf.Project) (error, int) {
	if project.ID != "" || project.Name != "" {
		return ErrProjectUnmodFields, http.StatusBadRequest
	}
	err, status := ProjectFieldsDb(db, project)
	if err != nil {
		return err, status
	}

	return nil, http.StatusOK
}

// ProjectFieldsDb used in validation and checks whether the specified values of project fields exist in the database
func ProjectFieldsDb(db database.Database, project *protobuf.Project) (error, int) {
	if project.DefaultImage != "" {
		dbImg, err := db.ReadImage(project.DefaultImage)
		if err != nil {
			return err, http.StatusInternalServerError
		}
		if dbImg.Name == "" {
			return ErrProjectImageNotFound, http.StatusBadRequest
		}
	}
	if project.DefaultMasterFlavor != "" {
		flavor, err := db.ReadFlavor(project.DefaultMasterFlavor)
		if err != nil {
			return err, http.StatusInternalServerError
		}
		if flavor.Name == "" {
			return ErrFlavorFieldValueNotFound("DefaultMasterFlavor"), http.StatusBadRequest
		}
	}
	if project.DefaultSlavesFlavor != "" {
		flavor, err := db.ReadFlavor(project.DefaultSlavesFlavor)
		if err != nil {
			return err, http.StatusInternalServerError
		}
		if flavor.Name == "" {
			return ErrFlavorFieldValueNotFound("DefaultSlavesFlavor"), http.StatusBadRequest
		}
	}
	if project.DefaultStorageFlavor != "" {
		flavor, err := db.ReadFlavor(project.DefaultStorageFlavor)
		if err != nil {
			return err, http.StatusInternalServerError
		}
		if flavor.Name == "" {
			return ErrFlavorFieldValueNotFound("DefaultStorageFlavor"), http.StatusBadRequest
		}
	}
	if project.DefaultMonitoringFlavor != "" {
		flavor, err := db.ReadFlavor(project.DefaultMonitoringFlavor)
		if err != nil {
			return err, http.StatusInternalServerError
		}
		if flavor.Name == "" {
			return ErrFlavorFieldValueNotFound("DefaultMonitoringFlavor"), http.StatusBadRequest
		}
	}
	return nil, http.StatusOK
}

// ProjectDelete checks the project structure for the presence of used clusters when deleting
func ProjectDelete(db database.Database, project *protobuf.Project) (error, int) {
	clusters, err := db.ReadProjectClusters(project.ID)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if len(clusters) > 0 {
		return ErrProjectHasClusters, http.StatusBadRequest
	}
	return nil, http.StatusOK
}
