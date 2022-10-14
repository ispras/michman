package validate

import (
	"github.com/ispras/michman/internal/database"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/rest/handler/check"
)

// ServiceTypeCreate validates fields of the service type structure for correct filling when creating
func ServiceTypeCreate(db database.Database, sType *protobuf.ServiceType) error {
	// check service class
	err := check.ServiceTypeClass(sType)
	if err != nil {
		return err
	}

	// check service access port
	if sType.AccessPort != 0 { //0 if port not provided
		err = check.ServiceTypePort(sType.AccessPort)
		if err != nil {
			return err
		}
	}

	// check all possible ports
	if sType.Ports != nil {
		err = check.ServiceTypePorts(sType.Ports)
		if err != nil {
			return err
		}
	}

	// check service type versions
	err = check.ServiceTypeVersions(db, sType.Versions)
	if err != nil {
		return err
	}

	// check default version
	err = check.ServiceTypeDefaultVersion(sType.Versions, sType.DefaultVersion)
	if err != nil {
		return err
	}

	return nil
}

// ServiceTypeUpdate validates fields of the service type version structure for correct filling when updating
func ServiceTypeUpdate(oldServiceType *protobuf.ServiceType, newServiceType *protobuf.ServiceType) error {
	if newServiceType.ID != "" || newServiceType.Type != "" {
		return ErrServiceTypeUnmodFields
	}
	if newServiceType.Versions != nil {
		return ErrServiceTypeUnmodVersionsField
	}

	if newServiceType.DefaultVersion != "" {
		err := check.ServiceTypeDefaultVersion(oldServiceType.Versions, newServiceType.DefaultVersion)
		if err != nil {
			return err
		}
	}

	if newServiceType.Class != "" {
		err := check.ServiceTypeClass(newServiceType)
		if err != nil {
			return err
		}
	}

	if newServiceType.AccessPort != 0 { //0 if port not provided
		err := check.ServiceTypePort(newServiceType.AccessPort)
		if err != nil {
			return err
		}
	}

	if newServiceType.Ports != nil {
		err := check.ServiceTypePorts(newServiceType.Ports)
		if err != nil {
			return err
		}
	}
	return nil
}

// ServiceTypeDelete validates fields of the service type structure dependencies for correct deletion
func ServiceTypeDelete(db database.Database, serviceType string) error {
	//check that service type doesn't exist in dependencies
	serviceTypes, err := db.ReadServicesTypesList()
	if err != nil {
		return err
	}
	err = check.ServiceTypeDependencyNotExists(serviceType, serviceTypes)
	if err != nil {
		return err
	}
	return nil
}

// ServiceTypeVersionCreate validates fields of the service type version structure for correct filling when creating
func ServiceTypeVersionCreate(db database.Database, versions []*protobuf.ServiceVersion, newServiceTypeVersion protobuf.ServiceVersion) error {
	if newServiceTypeVersion.ID != "" {
		return ErrServiceTypeVersionUnmodFields
	}

	if newServiceTypeVersion.Version == "" {
		return ErrServiceTypeVersionEmptyVersionField
	}

	//check that version is unique
	if versions != nil {
		err := check.ServiceTypeVersionUnique(versions, newServiceTypeVersion)
		if err != nil {
			return err
		}
	}

	//check service version config
	if newServiceTypeVersion.Configs != nil {
		err := check.ServiceTypeVersionConfigs(newServiceTypeVersion.Configs)
		if err != nil {
			return err
		}
	}

	//check service version dependencies
	if newServiceTypeVersion.Dependencies != nil {
		err := check.ServiceTypeVersionDependencies(db, newServiceTypeVersion.Dependencies)
		if err != nil {
			return err
		}
	}
	return nil
}

// ServiceTypeVersionUpdate validates fields of the service type version structure for correct filling when updating
func ServiceTypeVersionUpdate(newServiceTypeVersion protobuf.ServiceVersion) error {
	if newServiceTypeVersion.ID != "" || newServiceTypeVersion.Version != "" {
		return ErrServiceTypeVersionUnmodFields
	}

	if newServiceTypeVersion.Configs != nil || newServiceTypeVersion.Dependencies != nil {
		return ErrServiceTypeUnmodVersionFields
	}
	return nil
}

// ServiceTypeVersionDelete validates fields of the service type version structure dependencies for correct deletion
func ServiceTypeVersionDelete(db database.Database, serviceType *protobuf.ServiceType, serviceTypeVersion *protobuf.ServiceVersion) error {
	//check that this service version doesn't present in dependencies
	serviceTypes, err := db.ReadServicesTypesList()
	if err != nil {
		return err
	}

	// check dependencies in other service types
	err = check.ServiceTypeVersionDependencyNotExists(serviceTypes, serviceType, serviceTypeVersion)
	if err != nil {
		return err
	}

	if serviceType.DefaultVersion == serviceTypeVersion.Version {
		return ErrServiceTypeDeleteVersionDefault
	}
	return nil
}

// ServiceTypeVersionConfigCreate validates fields of the service type version config structure for correct filling when creating
func ServiceTypeVersionConfigCreate(newServiceTypeConfig *protobuf.ServiceConfig, oldConfigs []*protobuf.ServiceConfig) error {
	err := check.ServiceTypeVersionConfig(newServiceTypeConfig, oldConfigs)
	if err != nil {
		return err
	}
	return nil
}

// ServiceTypeVersionConfigUpdate validates fields of the service type version config structure for correct filling when updating
func ServiceTypeVersionConfigUpdate(newServiceTypeConfig, oldConfig *protobuf.ServiceConfig) error {
	if newServiceTypeConfig.ParameterName != "" ||
		newServiceTypeConfig.AnsibleVarName != "" ||
		newServiceTypeConfig.ID != "" {
		return ErrServiceTypeVersionConfigUnmodFields
	}
	if newServiceTypeConfig.Type != "" {
		err := check.SupportedType(newServiceTypeConfig.Type)
		if err != nil {
			return err
		}
	}
	if newServiceTypeConfig.PossibleValues != nil {
		err := check.PossibleValues(newServiceTypeConfig.PossibleValues, newServiceTypeConfig.Type, newServiceTypeConfig.IsList)
		if err != nil {
			return err
		}
		if newServiceTypeConfig.DefaultValue != "" {
			if err := check.ServiceTypeConfigDefaultValue(newServiceTypeConfig.DefaultValue, newServiceTypeConfig.PossibleValues); err != nil {
				return err
			}
		} else {
			if err := check.ServiceTypeConfigDefaultValue(oldConfig.DefaultValue, newServiceTypeConfig.PossibleValues); err != nil {
				return err
			}
		}
	}
	if newServiceTypeConfig.DefaultValue != "" && newServiceTypeConfig.PossibleValues == nil && oldConfig.PossibleValues != nil {
		if err := check.ServiceTypeConfigDefaultValue(newServiceTypeConfig.DefaultValue, oldConfig.PossibleValues); err != nil {
			return err
		}
	}
	return nil
}

// ServiceTypeVersionDependencyCreate validates fields of the service type version config structure for correct filling when creating
func ServiceTypeVersionDependencyCreate(db database.Database, newServiceTypeDependency *protobuf.ServiceDependency, oldDependencies []*protobuf.ServiceDependency) error {
	err := check.ServiceTypeVersionDependency(db, newServiceTypeDependency, oldDependencies)
	if err != nil {
		return err
	}
	return nil
}

// ServiceTypeVersionDependencyUpdate validates fields of the service type version dependency structure for correct filling when updating
func ServiceTypeVersionDependencyUpdate(db database.Database, oldServiceTypeDependency *protobuf.ServiceDependency, newServiceTypeDependency *protobuf.ServiceDependency) error {
	if newServiceTypeDependency.ServiceType != "" {
		return ErrServiceTypeVersionDependencyUnmodFields
	}

	if newServiceTypeDependency.ServiceVersions != nil {
		sType, err := db.ReadServiceType(oldServiceTypeDependency.ServiceType)
		if err != nil {
			return err
		}

		//check correctness of new dependency versions list
		err = check.ServiceTypeVersionDependencyPossibleVersions(newServiceTypeDependency, sType)
		if err != nil {
			return err
		}
	}

	if newServiceTypeDependency.DefaultServiceVersion != "" {
		//check correctness of default service version
		errNew := check.ServiceTypeVersionDependencyDefaultServiceVersion(newServiceTypeDependency, newServiceTypeDependency.DefaultServiceVersion)
		errOld := check.ServiceTypeVersionDependencyDefaultServiceVersion(oldServiceTypeDependency, newServiceTypeDependency.DefaultServiceVersion)
		if errNew != nil && errOld != nil {
			return errNew
		}
	}
	return nil
}
