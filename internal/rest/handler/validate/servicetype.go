package validate

import (
	"github.com/ispras/michman/internal/database"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/rest/handler/check"
	"net/http"
)

// ServiceTypeCreate validates fields of the service type structure for correct filling when creating
func ServiceTypeCreate(db database.Database, sType *protobuf.ServiceType) (error, int) {
	// check service class
	err := check.ServiceTypeClass(sType)
	if err != nil {
		return err, http.StatusBadRequest
	}

	// check service access port
	if sType.AccessPort != 0 { //0 if port not provided
		err = check.ServiceTypePort(sType.AccessPort)
		if err != nil {
			return err, http.StatusBadRequest
		}
	}

	// check all possible ports
	if sType.Ports != nil {
		err = check.ServiceTypePorts(sType.Ports)
		if err != nil {
			return err, http.StatusBadRequest
		}
	}

	// check service type versions
	err, status := check.ServiceTypeVersions(db, sType.Versions)
	if err != nil {
		return err, status
	}

	// check default version
	err = check.ServiceTypeDefaultVersion(sType.Versions, sType.DefaultVersion)
	if err != nil {
		return err, http.StatusBadRequest
	}

	return nil, 0
}

// ServiceTypeUpdate validates fields of the service type version structure for correct filling when updating
func ServiceTypeUpdate(oldServiceType *protobuf.ServiceType, newServiceType *protobuf.ServiceType) (error, int) {
	if newServiceType.ID != "" || newServiceType.Type != "" {
		return ErrServiceTypeUnmodFields, http.StatusBadRequest
	}
	if newServiceType.Versions != nil {
		return ErrServiceTypeUnmodVersionsField, http.StatusBadRequest
	}

	if newServiceType.DefaultVersion != "" {
		err := check.ServiceTypeDefaultVersion(oldServiceType.Versions, newServiceType.DefaultVersion)
		if err != nil {
			return err, http.StatusBadRequest
		}
	}

	if newServiceType.Class != "" {
		err := check.ServiceTypeClass(newServiceType)
		if err != nil {
			return err, http.StatusBadRequest
		}
	}

	if newServiceType.AccessPort != 0 { //0 if port not provided
		err := check.ServiceTypePort(newServiceType.AccessPort)
		if err != nil {
			return err, http.StatusBadRequest
		}
	}

	if newServiceType.Ports != nil {
		err := check.ServiceTypePorts(newServiceType.Ports)
		if err != nil {
			return err, http.StatusBadRequest
		}
	}
	return nil, 0
}

// ServiceTypeDelete validates fields of the service type structure dependencies for correct deletion
func ServiceTypeDelete(db database.Database, serviceType string) (error, int) {
	//check that service type doesn't exist in dependencies
	serviceTypes, err := db.ReadServicesTypesList()
	if err != nil {
		return err, http.StatusInternalServerError
	}
	err, status := check.ServiceTypeDependencyNotExists(serviceType, serviceTypes)
	if err != nil {
		return err, status
	}
	return nil, 0
}

// ServiceTypeVersionCreate validates fields of the service type version structure for correct filling when creating
func ServiceTypeVersionCreate(db database.Database, versions []*protobuf.ServiceVersion, newServiceTypeVersion protobuf.ServiceVersion) (error, int) {
	if newServiceTypeVersion.ID != "" {
		return ErrServiceTypeVersionUnmodFields, http.StatusBadRequest
	}

	if newServiceTypeVersion.Version == "" {
		return ErrServiceTypeVersionEmptyVersionField, http.StatusBadRequest
	}

	//check that version is unique
	if versions != nil {
		err := check.ServiceTypeVersionUnique(versions, newServiceTypeVersion)
		if err != nil {
			return err, http.StatusBadRequest
		}
	}

	//check service version config
	if newServiceTypeVersion.Configs != nil {
		err, status := check.ServiceTypeConfigs(newServiceTypeVersion.Configs)
		if err != nil {
			return err, status
		}
	}

	//check service version dependencies
	if newServiceTypeVersion.Dependencies != nil {
		err, status := check.ServiceTypeVersionDependencies(db, newServiceTypeVersion.Dependencies)
		if err != nil {
			return err, status
		}
	}
	return nil, 0
}

// ServiceTypeVersionUpdate validates fields of the service type version structure for correct filling when updating
func ServiceTypeVersionUpdate(newServiceTypeVersion protobuf.ServiceVersion) (error, int) {
	if newServiceTypeVersion.ID != "" || newServiceTypeVersion.Version != "" {
		return ErrServiceTypeVersionUnmodFields, http.StatusBadRequest
	}

	if newServiceTypeVersion.Configs != nil || newServiceTypeVersion.Dependencies != nil {
		return ErrServiceTypeUnmodVersionFields, http.StatusBadRequest
	}
	return nil, 0
}

// ServiceTypeVersionDelete validates fields of the service type version structure dependencies for correct deletion
func ServiceTypeVersionDelete(db database.Database, serviceType *protobuf.ServiceType, serviceTypeVersion *protobuf.ServiceVersion) (error, int) {
	//check that this service version doesn't present in dependencies
	serviceTypes, err := db.ReadServicesTypesList()
	if err != nil {
		return err, http.StatusInternalServerError
	}

	// check dependencies in other service types
	err, status := check.ServiceTypeVersionDependencyNotExists(serviceTypes, serviceType, serviceTypeVersion)
	if err != nil {
		return err, status
	}

	if serviceType.DefaultVersion == serviceTypeVersion.Version {
		return ErrServiceTypeDeleteVersionDefault, http.StatusBadRequest
	}
	return nil, 0
}

// ServiceTypeVersionConfigCreate validates fields of the service type version config structure for correct filling when creating
func ServiceTypeVersionConfigCreate(newServiceTypeConfig *protobuf.ServiceConfig, oldConfigs []*protobuf.ServiceConfig) (error, int) {
	err, status := check.ServiceTypeConfig(newServiceTypeConfig, oldConfigs)
	if err != nil {
		return err, status
	}
	return nil, 0
}

// ServiceTypeVersionConfigUpdate validates fields of the service type version config structure for correct filling when updating
func ServiceTypeVersionConfigUpdate(newServiceTypeConfig *protobuf.ServiceConfig) (error, int) {
	if newServiceTypeConfig.ParameterName != "" || newServiceTypeConfig.AnsibleVarName != "" {
		return ErrServiceTypeVersionConfigUnmodFields, http.StatusBadRequest
	}
	if newServiceTypeConfig.Type != "" {
		err := check.SupportedType(newServiceTypeConfig.Type)
		if err != nil {
			return err, http.StatusBadRequest
		}
	}
	if newServiceTypeConfig.PossibleValues != nil {
		err := check.PossibleValues(newServiceTypeConfig.PossibleValues, newServiceTypeConfig.Type, newServiceTypeConfig.IsList)
		if err != nil {
			return err, http.StatusBadRequest
		}
		if newServiceTypeConfig.DefaultValue == "" {
			return check.ErrServiceTypeVersionConfiqDefaultValueEmpty, http.StatusBadRequest
		} else {
			err = check.ServiceTypeConfigDefaultValue(newServiceTypeConfig.DefaultValue, newServiceTypeConfig.PossibleValues)
			if err != nil {
				return err, http.StatusBadRequest
			}
		}
	}
	return nil, 0
}
