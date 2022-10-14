package validate

import (
	"github.com/ispras/michman/internal/database"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/rest/handler/check"
	"github.com/ispras/michman/internal/rest/response"
	"github.com/ispras/michman/internal/utils"
)

// ProjectCreate validates fields of the project structure for correct filling when creating
func ProjectCreate(db database.Database, project *protobuf.Project) error {
	if project.ID != "" {
		return ErrGeneratedField("project", "ID")
	}
	if project.Name != "" {
		return ErrGeneratedField("project", "Name")
	}
	if project.DisplayName == "" {
		return ErrEmptyField("project", "DisplayName")
	}
	if err := check.ProjectValidName(project.DisplayName); err != nil {
		return err
	}
	if project.DefaultImage == "" {
		return ErrEmptyField("project", "DefaultImage")
	}
	if project.DefaultMasterFlavor == "" {
		return ErrEmptyField("project", "DefaultMasterFlavor")
	}
	if project.DefaultSlavesFlavor == "" {
		return ErrEmptyField("project", "DefaultSlavesFlavor")
	}
	if project.DefaultStorageFlavor == "" {
		return ErrEmptyField("project", "DefaultStorageFlavor")
	}
	if project.DefaultMonitoringFlavor == "" {
		return ErrEmptyField("project", "DefaultMonitoringFlavor")
	}
	// Try to find project by DisplayName (because there's still no Name now)
	dbProject, err := db.ReadProject(project.DisplayName)
	if dbProject != nil {
		return ErrObjectExists("project", project.DisplayName)
	}
	if err != nil && response.ErrorClass(err) != utils.ObjectNotFound {
		return err
	}

	err = ProjectFieldsDb(db, project)
	if err != nil {
		return err
	}

	return nil
}

// ProjectUpdate validates fields of the project structure for correct filling when updating
func ProjectUpdate(db database.Database, project *protobuf.Project) error {
	if project.ID != "" || project.Name != "" {
		return ErrProjectUnmodFields
	}
	err := ProjectFieldsDb(db, project)
	if err != nil {
		return err
	}

	return nil
}

// ProjectFieldsDb used in validation and checks whether the specified values of project fields exist in the database
func ProjectFieldsDb(db database.Database, project *protobuf.Project) error {
	if project.DefaultImage != "" {
		_, err := db.ReadImage(project.DefaultImage)
		if err != nil {
			return err
		}
	}
	if project.DefaultMasterFlavor != "" {
		_, err := db.ReadFlavor(project.DefaultMasterFlavor)
		if err != nil {
			return err
		}
	}
	if project.DefaultSlavesFlavor != "" {
		_, err := db.ReadFlavor(project.DefaultSlavesFlavor)
		if err != nil {
			return err
		}
	}
	if project.DefaultStorageFlavor != "" {
		_, err := db.ReadFlavor(project.DefaultStorageFlavor)
		if err != nil {
			return err
		}
	}
	if project.DefaultMonitoringFlavor != "" {
		_, err := db.ReadFlavor(project.DefaultMonitoringFlavor)
		if err != nil {
			return err
		}
	}
	return nil
}

// ProjectDelete checks the project structure for the presence of used clusters when deleting
func ProjectDelete(db database.Database, project *protobuf.Project) error {
	clusters, err := db.ReadProjectClusters(project.ID)
	if err != nil {
		return err
	}
	if len(clusters) > 0 {
		return ErrProjectHasClusters
	}
	return nil
}
