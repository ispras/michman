package validate

import (
	"github.com/ispras/michman/internal/database"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/rest/handler/check"
	"net/http"
)

func ServiceTypeCreate(db database.Database, sType *protobuf.ServiceType) (error, int) {
	// check service class
	err := check.Class(sType)
	if err != nil {
		return err, http.StatusBadRequest
	}

	// check service access port
	if sType.AccessPort != 0 { //0 if port not provided
		err = check.Port(sType.AccessPort)
		if err != nil {
			return err, http.StatusBadRequest
		}
	}

	// check all ports
	if sType.Ports != nil {
		for _, p := range sType.Ports {
			err = check.Port(p.Port)
			if err != nil {
				return err, http.StatusBadRequest
			}
		}
	}

	err, status := ServiceTypeVersions(db, sType.Versions)
	if err != nil {
		return err, status
	}

	//check default version
	err = check.DefaultVersion(sType.Versions, sType.DefaultVersion)
	if err != nil {
		return err, http.StatusBadRequest
	}

	return nil, 0
}

func ServiceTypeVersions(db database.Database, sTypeVersions []*protobuf.ServiceVersion) (error, int) {
	for i, serviceVersion := range sTypeVersions {
		// check service version is unique
		err := check.VersionUnique(sTypeVersions[i+1:], *serviceVersion)
		if err != nil {
			return err, http.StatusBadRequest
		}

		//check service version config
		if serviceVersion.Configs != nil {
			err, status := check.Configs(serviceVersion.Configs)
			if err != nil {
				return err, status
			}
		}

		//check service version dependencies
		err, status := check.Dependencies(db, serviceVersion.Dependencies)
		if err != nil {
			return err, status
		}
	}
	return nil, 0
}

func ServiceTypeUpdate(oldServiceType *protobuf.ServiceType, newServiceType *protobuf.ServiceType) (error, int) {
	if newServiceType.ID != "" || newServiceType.Type != "" {
		return ErrServiceTypeUnmodFields, http.StatusBadRequest
	}
	if newServiceType.Versions != nil {
		return ErrServiceTypeUnmodVersionsField, http.StatusBadRequest
	}

	if newServiceType.DefaultVersion != "" {
		err := check.DefaultVersion(oldServiceType.Versions, newServiceType.DefaultVersion)
		if err != nil {
			return err, http.StatusBadRequest
		}
	}

	if newServiceType.Class != "" {
		err := check.Class(newServiceType)
		if err != nil {
			return err, http.StatusBadRequest
		}
	}

	if newServiceType.AccessPort != 0 { //0 if port not provided
		err := check.Port(newServiceType.AccessPort)
		if err != nil {
			return err, http.StatusBadRequest
		}
	}

	if newServiceType.Ports != nil {
		for _, port := range newServiceType.Ports {
			err := check.Port(port.Port)
			if err != nil {
				return err, http.StatusBadRequest
			}
		}
	}
	return nil, 0
}

func ServiceTypeVersionCreate(db database.Database, versions []*protobuf.ServiceVersion, newServiceTypeVersion protobuf.ServiceVersion) (error, int) {
	if newServiceTypeVersion.ID != "" {
		return ErrServiceTypeVersionUnmodFields, http.StatusBadRequest
	}

	if newServiceTypeVersion.Version == "" {
		return ErrServiceTypeVersionEmptyVersionField, http.StatusBadRequest
	}

	//check that version is unique
	if versions != nil {
		err := check.VersionUnique(versions, newServiceTypeVersion)
		if err != nil {
			return err, http.StatusBadRequest
		}
	}

	//check service version config
	if newServiceTypeVersion.Configs != nil {
		err, status := check.Configs(newServiceTypeVersion.Configs)
		if err != nil {
			return err, status
		}
	}

	//check service version dependencies
	if newServiceTypeVersion.Dependencies != nil {
		err, status := check.Dependencies(db, newServiceTypeVersion.Dependencies)
		if err != nil {
			return err, status
		}
	}
	return nil, 0
}

func ServiceTypeVersionUpdate(newServiceTypeVersion protobuf.ServiceVersion) (error, int) {
	if newServiceTypeVersion.ID != "" || newServiceTypeVersion.Version != "" {
		return ErrServiceTypeVersionUnmodFields, http.StatusBadRequest
	}

	if newServiceTypeVersion.Configs != nil || newServiceTypeVersion.Dependencies != nil {
		return ErrServiceTypeUnmodVersionFields, http.StatusBadRequest
	}
	return nil, 0
}

func ServiceTypeVersionDelete(db database.Database, serviceType *protobuf.ServiceType, serviceTypeVersion *protobuf.ServiceVersion) (error, int) {
	//check that this service version doesn't present in dependencies
	serviceTypes, err := db.ReadServicesTypesList()
	if err != nil {
		return err, http.StatusInternalServerError
	}
	for _, curServiceType := range serviceTypes {
		for _, serviceVersion := range curServiceType.Versions {
			for _, serviceVersionDependency := range serviceVersion.Dependencies {
				if serviceVersionDependency.ServiceType == serviceType.Type {
					for _, serviceVersionDependencyVersion := range serviceVersionDependency.ServiceVersions {
						if serviceVersionDependencyVersion == serviceTypeVersion.Version {
							return ErrConfigServiceTypeDependenceVersionExists(serviceVersionDependencyVersion, curServiceType.Type), http.StatusBadRequest
						}
					}
				}
			}
		}
	}
	if serviceType.DefaultVersion == serviceTypeVersion.Version {
		return ErrServiceTypeDeleteVersionDefault, http.StatusBadRequest
	}
	return nil, 0
}

func ServiceTypeDelete(db database.Database, serviceType string) (error, int) {
	//check that service type doesn't exist in dependencies
	serviceTypes, err := db.ReadServicesTypesList()
	if err != nil {
		return err, http.StatusInternalServerError
	}
	for _, curServiceType := range serviceTypes {
		for _, serviceVersion := range curServiceType.Versions {
			for _, serviceVersionDependency := range serviceVersion.Dependencies {
				if serviceVersionDependency.ServiceType == serviceType {
					return ErrConfigServiceTypeDependenceExists, http.StatusBadRequest
				}
			}
		}
	}
	return nil, 0
}

func ServiceTypeVersionConfigCreate(newServiceTypeConfig *protobuf.ServiceConfig, oldConfigs []*protobuf.ServiceConfig) (error, int) {
	err, status := check.Config(newServiceTypeConfig, oldConfigs)
	if err != nil {
		return err, status
	}
	return nil, 0
}

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
			err = check.ConfigDefaultValue(newServiceTypeConfig.DefaultValue, newServiceTypeConfig.PossibleValues)
			if err != nil {
				return err, http.StatusBadRequest
			}
		}
	}
	return nil, 0
}
