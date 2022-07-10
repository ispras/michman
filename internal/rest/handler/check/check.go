package check

import (
	"encoding/json"
	"github.com/ispras/michman/internal/database"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/rest/handler/helpfunc"
	"github.com/ispras/michman/internal/utils"
	"net/http"
	"os"
	"regexp"
	"strconv"
)

// SupportedType list of supported types
func SupportedType(_type string) error {
	switch _type {
	case
		"int",
		"float",
		"bool",
		"string":
		return nil
	}
	return ErrValidTypeParam(_type)
}

// MSServices returns true if master-slave service exists
func MSServices(db database.Database, cluster *protobuf.Cluster) (bool, error) {
	for _, service := range cluster.Services {
		st, err := db.ReadServiceType(service.Type)
		if err != nil {
			return false, err
		}
		if st.Class == utils.ClassMasterSlave {
			return true, nil
		}
	}
	return false, nil
}

func ValuesAllowed(val string, posVal []string) bool {
	val = helpfunc.DeleteSpaces(val)
	for _, pv := range posVal {
		if val == pv {
			return true
		}
	}
	return false
}

func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	return false, err
}

func ValidName(name string, pattern string, errorType error) (error, int) {
	validName := regexp.MustCompile(pattern).MatchString
	if !validName(name) {
		return errorType, http.StatusBadRequest
	}
	return nil, 0
}

func Class(st *protobuf.ServiceType) error {
	if st.Class == utils.ClassMasterSlave || st.Class == utils.ClassStandAlone || st.Class == utils.ClassStorage {
		return nil
	}
	return ErrServiceTypeClass
}

func Port(port int32) error {
	//TODO: add another checks for port?
	if port > 0 {
		return nil
	}
	return ErrServiceTypePort
}

func DefaultVersion(sTypeVersions []*protobuf.ServiceVersion, defaultVersion string) error {
	for _, curVersion := range sTypeVersions {
		if curVersion.Version == defaultVersion {
			return nil
		}
	}
	return ErrServiceTypeDefaultVersion
}

func VersionUnique(sTypeVersions []*protobuf.ServiceVersion, newVersion protobuf.ServiceVersion) error {
	for _, curVersion := range sTypeVersions {
		if curVersion.Version == newVersion.Version {
			return ErrServiceTypeVersionUnique(curVersion.Version)
		}
	}
	return nil
}

func ConfigsUnique(sTypeVersionConfigs []*protobuf.ServiceConfig, newConfig protobuf.ServiceConfig) error {
	for _, curConfig := range sTypeVersionConfigs {
		if curConfig.ParameterName == newConfig.ParameterName {
			return ErrServiceTypeVersionConfigUnique(curConfig.ParameterName)
		}
	}
	return nil
}

func PossibleValuesUnique(possibleValues []string) error {
	for i, curVal := range possibleValues[:len(possibleValues)-1] {
		if curVal == "" {
			return ErrConfigPossibleValueEmpty
		}
		for _, otherVal := range possibleValues[i+1:] {
			if curVal == otherVal {
				return ErrServiceTypeVersionConfigPossibleValuesUnique(curVal)
			}
		}
	}
	return nil
}

func PossibleValues(possibleValues []string, vType string, IsList bool) error {
	//check PossibleValues type correct
	for _, value := range possibleValues {
		if err := CorrectType(value, vType, IsList); err != nil {
			return ErrServiceTypeVersionConfigPossibleValues(value)
		}
	}

	//format PossibleValue strings
	if IsList {
		for i, pV := range possibleValues {
			possibleValues[i] = helpfunc.DeleteSpaces(pV)
		}
	}

	//check PossibleValues are unique
	err := PossibleValuesUnique(possibleValues)
	if err != nil {
		return err
	}
	return nil
}

func ConfigDefaultValue(defaultValue string, possibleValues []string) error {
	for _, value := range possibleValues {
		if value == defaultValue {
			return nil
		}
	}
	return ErrServiceTypeVersionConfigDefaultValue
}

func Config(config *protobuf.ServiceConfig, versionConfigs []*protobuf.ServiceConfig) (error, int) {
	// check param type
	err := SupportedType(config.Type)
	if err != nil {
		return err, http.StatusBadRequest
	}

	// check param name is unique
	if config.ParameterName == "" {
		return ErrServiceTypeVersionConfigParamEmpty("ParameterName"), http.StatusBadRequest
	}

	// check config is unique by parameter name
	err = ConfigsUnique(versionConfigs, *config)
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
			return ErrServiceTypeVersionConfiqDefaultValueEmpty, http.StatusBadRequest
		} else {
			err = ConfigDefaultValue(config.DefaultValue, config.PossibleValues)
			if err != nil {
				return err, http.StatusBadRequest
			}
		}
	}
	return nil, 0
}

func Configs(versionConfigs []*protobuf.ServiceConfig) (error, int) {
	for i, curConfig := range versionConfigs {
		err, status := Config(curConfig, versionConfigs[i+1:])
		if err != nil {
			return err, status
		}
	}
	return nil, 0
}

func Dependencies(db database.Database, serviceDependencies []*protobuf.ServiceDependency) (error, int) {
	for _, serviceDependency := range serviceDependencies {
		sType, err := db.ReadServiceType(serviceDependency.ServiceType)
		if err != nil {
			return err, http.StatusInternalServerError
		}

		if sType.Type == "" {
			return ErrServiceDependenciesNotExists(serviceDependency.ServiceType), http.StatusBadRequest
		}

		if serviceDependency.ServiceVersions == nil {
			return ErrConfigDependencyServiceVersionEmpty, http.StatusBadRequest
		}

		if serviceDependency.DefaultServiceVersion == "" {
			return ErrConfigDependencyServiceDefaultVersionEmpty, http.StatusBadRequest
		}

		//check correctness of versions list
		flagDefaultVersion := false
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
			if dependencyServiceVersion == serviceDependency.DefaultServiceVersion {
				flagDefaultVersion = true
			}
		}

		if !flagDefaultVersion {
			return ErrConfigServiceDependencyDefaultVersionNotFound, http.StatusBadRequest
		}
	}
	return nil, 0
}

func ServiceConfigCorrectValue(service *protobuf.Service, Configs []*protobuf.ServiceConfig) error {
	for configName, configValue := range service.Config {
		flagPN := false
		for _, serviceConfig := range Configs {
			if configName == serviceConfig.ParameterName {
				flagPN = true

				if err := CorrectType(configValue, serviceConfig.Type, serviceConfig.IsList); err != nil {
					return ErrClusterServiceConfigIncorrectType(configName, service.Type)
				}

				//check for possible values
				if serviceConfig.PossibleValues != nil {
					if !ValuesAllowed(configValue, serviceConfig.PossibleValues) {
						return ErrClusterServiceConfigNotPossibleValue(configName, service.Type)
					}
				}

				break
			}
		}
		if !flagPN {
			return ErrClusterServiceConfigNotSupported(configName, service.Type)
		}
	}
	return nil
}

func CorrectType(value string, Type string, IsList bool) error {
	//check value type correct
	if !IsList {
		switch Type {
		case "int":
			if _, err := strconv.ParseInt(value, 10, 32); err != nil {
				return err
			}
		case "float":
			if _, err := strconv.ParseFloat(value, 64); err != nil {
				return err
			}
		case "bool":
			if _, err := strconv.ParseBool(value); err != nil {
				return err
			}
		}
	} else {
		switch Type {
		case "int":
			var valList []int64
			if err := json.Unmarshal([]byte(value), &valList); err != nil {
				return err
			}
		case "float":
			var valList []float64
			if err := json.Unmarshal([]byte(value), &valList); err != nil {
				return err
			}
		case "bool":
			var valList []bool
			if err := json.Unmarshal([]byte(value), &valList); err != nil {
				return err
			}
		case "string":
			var valList []string
			if err := json.Unmarshal([]byte(value), &valList); err != nil {
				return err
			}
		}
	}

	return nil
}
