package check

import (
	"github.com/ispras/michman/internal/database"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/utils"
)

// ServiceTypeClass checks that service type class belongs to one of the classes:
// Master-slave, StandAlone, Storage
func ServiceTypeClass(st *protobuf.ServiceType) error {
	if st.Class == utils.ClassMasterSlave || st.Class == utils.ClassStandAlone || st.Class == utils.ClassStorage {
		return nil
	}
	return ErrServiceTypeClass
}

// ServiceTypePort checks that service type port is valid
func ServiceTypePort(port int32) error {
	//TODO: add another checks for port?
	if port > 0 {
		return nil
	}
	return ErrServiceTypePort
}

// ServiceTypePorts checks all ports for correctness
func ServiceTypePorts(ports []*protobuf.ServicePort) error {
	for _, p := range ports {
		err := ServiceTypePort(p.Port)
		if err != nil {
			return err
		}
	}
	return nil
}

// ServiceTypeConfigDefaultValue checks that default value is set right
func ServiceTypeConfigDefaultValue(defaultValue string, possibleValues []string) error {
	for _, value := range possibleValues {
		if value == defaultValue {
			return nil
		}
	}
	return ErrServiceTypeVersionConfigDefaultValue
}

// ServiceTypeDefaultVersion checks service type default version for correctness
func ServiceTypeDefaultVersion(sTypeVersions []*protobuf.ServiceVersion, defaultVersion string) error {
	for _, curVersion := range sTypeVersions {
		if curVersion.Version == defaultVersion {
			return nil
		}
	}
	return ErrServiceTypeDefaultVersion
}

func ServiceTypeVersionConfigsUnique(sTypeVersionConfigs []*protobuf.ServiceConfig, newConfig protobuf.ServiceConfig) error {
	for _, curConfig := range sTypeVersionConfigs {
		if curConfig.ParameterName == newConfig.ParameterName {
			return ErrServiceTypeVersionConfigUnique(curConfig.ParameterName)
		}
	}
	return nil
}

func ServiceTypeVersionDependenciesUnique(sTypeVersionDependencies []*protobuf.ServiceDependency, newDependency protobuf.ServiceDependency) error {
	for _, curDependency := range sTypeVersionDependencies {
		if curDependency.ServiceType == newDependency.ServiceType {
			return ErrServiceTypeVersionDependencyUnique(curDependency.ServiceType)
		}
	}
	return nil
}

func ServiceTypeVersionUnique(sTypeVersions []*protobuf.ServiceVersion, newVersion protobuf.ServiceVersion) error {
	for _, curVersion := range sTypeVersions {
		if curVersion.Version == newVersion.Version {
			return ErrServiceTypeVersionUnique(curVersion.Version)
		}
	}
	return nil
}

// ServiceTypeVersionConfig checks that service type version config is unique and checks all fields for correctness
func ServiceTypeVersionConfig(config *protobuf.ServiceConfig, versionConfigs []*protobuf.ServiceConfig) error {
	// check param type
	err := SupportedType(config.Type)
	if err != nil {
		return err
	}

	// check param name is set
	if config.ParameterName == "" {
		return ErrServiceTypeVersionConfigParamEmpty("ParameterName")
	}

	// check config is unique by parameter name
	err = ServiceTypeVersionConfigsUnique(versionConfigs, *config)
	if err != nil {
		return err
	}

	//check param possible values
	if config.PossibleValues != nil {
		err = PossibleValues(config.PossibleValues, config.Type, config.IsList)
		if err != nil {
			return err
		}
		if config.DefaultValue == "" {
			return ErrServiceTypeVersionConfigDefaultValueEmpty
		} else {
			err = ServiceTypeConfigDefaultValue(config.DefaultValue, config.PossibleValues)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// ServiceTypeVersionConfigs checks all configs
func ServiceTypeVersionConfigs(versionConfigs []*protobuf.ServiceConfig) error {
	for i, curConfig := range versionConfigs {
		err := ServiceTypeVersionConfig(curConfig, versionConfigs[i+1:])
		if err != nil {
			return err
		}
	}
	return nil
}

// serviceTypeVersion checks that version is unique,
// checks its configs and dependencies
func serviceTypeVersion(db database.Database, sTypeVersion *protobuf.ServiceVersion, sTypeVersions []*protobuf.ServiceVersion) error {
	// check service version is unique

	err := ServiceTypeVersionUnique(sTypeVersions, *sTypeVersion)
	if err != nil {
		return err
	}

	// check service version config
	if sTypeVersion.Configs != nil {
		err = ServiceTypeVersionConfigs(sTypeVersion.Configs)
		if err != nil {
			return err
		}
	}

	//check service version dependencies
	err = ServiceTypeVersionDependencies(db, sTypeVersion.Dependencies)
	if err != nil {
		return err
	}
	return nil
}

// ServiceTypeVersions checks all versions
func ServiceTypeVersions(db database.Database, sTypeVersions []*protobuf.ServiceVersion) error {
	for i, serviceVersion := range sTypeVersions {
		err := serviceTypeVersion(db, serviceVersion, sTypeVersions[i+1:])
		if err != nil {
			return err
		}
	}
	return nil
}

// ServiceTypeVersionDependencyNotExists checks that service type version not present in all service types dependencies
func ServiceTypeVersionDependencyNotExists(serviceTypes []protobuf.ServiceType, serviceType *protobuf.ServiceType, serviceTypeVersion *protobuf.ServiceVersion) error {
	for _, curServiceType := range serviceTypes {
		for _, serviceVersion := range curServiceType.Versions {
			for _, serviceVersionDependency := range serviceVersion.Dependencies {
				if serviceVersionDependency.ServiceType == serviceType.Type {
					for _, serviceVersionDependencyVersion := range serviceVersionDependency.ServiceVersions {
						if serviceVersionDependencyVersion == serviceTypeVersion.Version {
							return ErrServiceTypeDependenceVersionExists(serviceVersionDependencyVersion, curServiceType.Type)
						}
					}
				}
			}
		}
	}
	return nil
}

// ServiceTypeDependencyNotExists checks that service type not present in all versions and their dependencies
func ServiceTypeDependencyNotExists(serviceType string, serviceTypes []protobuf.ServiceType) error {
	for _, curServiceType := range serviceTypes {
		for _, serviceVersion := range curServiceType.Versions {
			for _, serviceVersionDependency := range serviceVersion.Dependencies {
				if serviceVersionDependency.ServiceType == serviceType {
					return ErrConfigServiceTypeDependenceExists
				}
			}
		}
	}
	return nil
}

// ServiceTypeVersionDependencyPossibleVersions checks service type version dependency possible service versions
func ServiceTypeVersionDependencyPossibleVersions(serviceDependency *protobuf.ServiceDependency, sType *protobuf.ServiceType) error {
	for _, dependencyServiceVersion := range serviceDependency.ServiceVersions {
		flag := false

		for _, sv := range sType.Versions {
			if dependencyServiceVersion == sv.Version {
				flag = true
				break
			}
		}

		if !flag {
			return ErrConfigServiceDependencyVersionNotFound
		}
	}
	return nil
}

// ServiceTypeVersionDependencyDefaultServiceVersion checks service type version dependency default service version
func ServiceTypeVersionDependencyDefaultServiceVersion(serviceDependency *protobuf.ServiceDependency, defaultVersion string) error {
	flagDefaultVersion := false
	for _, dependencyServiceVersion := range serviceDependency.ServiceVersions {
		if dependencyServiceVersion == defaultVersion {
			flagDefaultVersion = true
			break
		}
	}
	if !flagDefaultVersion {
		return ErrConfigServiceDependencyDefaultVersionNotFound
	}
	return nil
}

// ServiceTypeVersionDependency checks service type version dependency for correctness
func ServiceTypeVersionDependency(db database.Database, serviceDependency *protobuf.ServiceDependency, versionDependencies []*protobuf.ServiceDependency) error {
	// read from database service type on which it depends
	sType, err := db.ReadServiceType(serviceDependency.ServiceType)
	if err != nil {
		return err
	}

	if sType.Type == "" {
		return ErrServiceDependenciesNotExists(serviceDependency.ServiceType)
	}

	// check dependency is unique by service type
	err = ServiceTypeVersionDependenciesUnique(versionDependencies, *serviceDependency)
	if err != nil {
		return err
	}

	if serviceDependency.ServiceVersions == nil {
		return ErrServiceTypeDependencyServiceEmptyField
	}

	if serviceDependency.DefaultServiceVersion == "" {
		return ErrServiceTypeDependencyServiceEmptyField
	}

	//check correctness of dependency versions list
	err = ServiceTypeVersionDependencyPossibleVersions(serviceDependency, sType)
	if err != nil {
		return err
	}

	//check correctness of default service version
	err = ServiceTypeVersionDependencyDefaultServiceVersion(serviceDependency, serviceDependency.DefaultServiceVersion)
	if err != nil {
		return err
	}
	return nil
}

// ServiceTypeVersionDependencies checks all dependencies
func ServiceTypeVersionDependencies(db database.Database, serviceDependencies []*protobuf.ServiceDependency) error {
	for i, serviceDependency := range serviceDependencies {
		err := ServiceTypeVersionDependency(db, serviceDependency, serviceDependencies[i+1:])
		if err != nil {
			return err
		}
	}
	return nil
}
