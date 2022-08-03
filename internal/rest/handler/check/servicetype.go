package check

import (
	"github.com/ispras/michman/internal/database"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/utils"
	"net/http"
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
func ServiceTypeVersionConfig(config *protobuf.ServiceConfig, versionConfigs []*protobuf.ServiceConfig) (error, int) {
	// check param type
	err := SupportedType(config.Type)
	if err != nil {
		return err, http.StatusBadRequest
	}

	// check param name is set
	if config.ParameterName == "" {
		return ErrServiceTypeVersionConfigParamEmpty("ParameterName"), http.StatusBadRequest
	}

	// check config is unique by parameter name
	err = ServiceTypeVersionConfigsUnique(versionConfigs, *config)
	if err != nil {
		return err, http.StatusBadRequest
	}

	//check param possible values
	if config.PossibleValues != nil {
		err = PossibleValues(config.PossibleValues, config.Type, config.IsList)
		if err != nil {
			return err, http.StatusBadRequest
		}
		if config.DefaultValue == "" {
			return ErrServiceTypeVersionConfigDefaultValueEmpty, http.StatusBadRequest
		} else {
			err = ServiceTypeConfigDefaultValue(config.DefaultValue, config.PossibleValues)
			if err != nil {
				return err, http.StatusBadRequest
			}
		}
	}
	return nil, 0
}

// ServiceTypeVersionConfigs checks all configs
func ServiceTypeVersionConfigs(versionConfigs []*protobuf.ServiceConfig) (error, int) {
	for i, curConfig := range versionConfigs {
		err, status := ServiceTypeVersionConfig(curConfig, versionConfigs[i+1:])
		if err != nil {
			return err, status
		}
	}
	return nil, 0
}

// serviceTypeVersion checks that version is unique,
// checks its configs and dependencies
func serviceTypeVersion(db database.Database, sTypeVersion *protobuf.ServiceVersion, sTypeVersions []*protobuf.ServiceVersion) (error, int) {
	// check service version is unique

	err := ServiceTypeVersionUnique(sTypeVersions, *sTypeVersion)
	if err != nil {
		return err, http.StatusBadRequest
	}

	// check service version config
	if sTypeVersion.Configs != nil {
		err, status := ServiceTypeVersionConfigs(sTypeVersion.Configs)
		if err != nil {
			return err, status
		}
	}

	//check service version dependencies
	err, status := ServiceTypeVersionDependencies(db, sTypeVersion.Dependencies)
	if err != nil {
		return err, status
	}
	return nil, 0
}

// ServiceTypeVersions checks all versions
func ServiceTypeVersions(db database.Database, sTypeVersions []*protobuf.ServiceVersion) (error, int) {
	for i, serviceVersion := range sTypeVersions {
		err, status := serviceTypeVersion(db, serviceVersion, sTypeVersions[i+1:])
		if err != nil {
			return err, status
		}
	}
	return nil, 0
}

// ServiceTypeVersionDependencyNotExists checks that service type version not present in all service types dependencies
func ServiceTypeVersionDependencyNotExists(serviceTypes []protobuf.ServiceType, serviceType *protobuf.ServiceType, serviceTypeVersion *protobuf.ServiceVersion) (error, int) {
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
	return nil, 0
}

// ServiceTypeDependencyNotExists checks that service type not present in all versions and their dependencies
func ServiceTypeDependencyNotExists(serviceType string, serviceTypes []protobuf.ServiceType) (error, int) {
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

// ServiceTypeVersionDependencyPossibleVersions checks service type version dependency possible service versions
func ServiceTypeVersionDependencyPossibleVersions(serviceDependency *protobuf.ServiceDependency, sType *protobuf.ServiceType) (error, int) {
	for _, dependencyServiceVersion := range serviceDependency.ServiceVersions {
		flag := false

		for _, sv := range sType.Versions {
			if dependencyServiceVersion == sv.Version {
				flag = true
				break
			}
		}

		if !flag {
			return ErrConfigServiceDependencyVersionNotFound, http.StatusBadRequest
		}
	}
	return nil, 0
}

// ServiceTypeVersionDependencyDefaultServiceVersion checks service type version dependency default service version
func ServiceTypeVersionDependencyDefaultServiceVersion(serviceDependency *protobuf.ServiceDependency, defaultVersion string) (error, int) {
	flagDefaultVersion := false
	for _, dependencyServiceVersion := range serviceDependency.ServiceVersions {
		if dependencyServiceVersion == defaultVersion {
			flagDefaultVersion = true
			break
		}
	}
	if !flagDefaultVersion {
		return ErrConfigServiceDependencyDefaultVersionNotFound, http.StatusBadRequest
	}
	return nil, 0
}

// ServiceTypeVersionDependency checks service type version dependency for correctness
func ServiceTypeVersionDependency(db database.Database, serviceDependency *protobuf.ServiceDependency, versionDependencies []*protobuf.ServiceDependency) (error, int) {
	// read from database service type on which it depends
	sType, err := db.ReadServiceType(serviceDependency.ServiceType)
	if err != nil {
		return err, http.StatusInternalServerError
	}

	if sType.Type == "" {
		return ErrServiceDependenciesNotExists(serviceDependency.ServiceType), http.StatusBadRequest
	}

	// check dependency is unique by service type
	err = ServiceTypeVersionDependenciesUnique(versionDependencies, *serviceDependency)
	if err != nil {
		return err, http.StatusBadRequest
	}

	if serviceDependency.ServiceVersions == nil {
		return ErrConfigDependencyServiceVersionEmpty, http.StatusBadRequest
	}

	if serviceDependency.DefaultServiceVersion == "" {
		return ErrConfigDependencyServiceDefaultVersionEmpty, http.StatusBadRequest
	}

	//check correctness of dependency versions list
	err, status := ServiceTypeVersionDependencyPossibleVersions(serviceDependency, sType)
	if err != nil {
		return err, status
	}

	//check correctness of default service version
	err, status = ServiceTypeVersionDependencyDefaultServiceVersion(serviceDependency, serviceDependency.DefaultServiceVersion)
	if err != nil {
		return err, status
	}
	return nil, 0
}

// ServiceTypeVersionDependencies checks all dependencies
func ServiceTypeVersionDependencies(db database.Database, serviceDependencies []*protobuf.ServiceDependency) (error, int) {
	for i, serviceDependency := range serviceDependencies {
		err, status := ServiceTypeVersionDependency(db, serviceDependency, serviceDependencies[i+1:])
		if err != nil {
			return err, status
		}
	}
	return nil, 0
}
