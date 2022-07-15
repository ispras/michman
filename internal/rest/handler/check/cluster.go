package check

import (
	"github.com/ispras/michman/internal/database"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/rest/handler/helpfunc"
	"github.com/ispras/michman/internal/utils"
	"net/http"
)

// ValuesAllowed checks whether the passed value is in the list of possible values
func ValuesAllowed(value string, possibleValues []string, IsList bool) bool {
	if IsList {
		value = helpfunc.DeleteSpaces(value)
	}
	for _, pv := range possibleValues {
		if value == pv {
			return true
		}
	}
	return false
}

// ClusterExist checks if a cluster with the same name exists and return old cluster structure if it's status failed
func ClusterExist(db database.Database, clusterRes *protobuf.Cluster, project *protobuf.Project) (bool, *protobuf.Cluster, error, int) {
	clusterExists := false
	searchedName := clusterRes.DisplayName + "-" + project.Name
	cluster, err := db.ReadCluster(project.ID, searchedName)
	if err != nil {
		return false, nil, err, http.StatusInternalServerError
	}
	if cluster.Name != "" {
		clusterExists = true
		if cluster.EntityStatus != utils.StatusFailed {
			return clusterExists, cluster, ErrClusterExisted, http.StatusBadRequest
		}
	}
	return clusterExists, cluster, nil, http.StatusOK
}

// ServiceConfigCorrectValue checks all cluster configs for correct type, possible value and supporting
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
					if !ValuesAllowed(configValue, serviceConfig.PossibleValues, serviceConfig.IsList) {
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
