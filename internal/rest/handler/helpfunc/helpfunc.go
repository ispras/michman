package helpfunc

import (
	"github.com/ispras/michman/internal/database"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/utils"
	"github.com/sirupsen/logrus"
	"net/http"
)

func AddDependencies(db database.Database, cluster *protobuf.Cluster, curS *protobuf.Service) ([]*protobuf.Service, error, int) {
	var serviceToAdd *protobuf.Service = nil
	var servicesList []*protobuf.Service = nil

	sv, err := db.ReadServiceTypeVersion(curS.Type, curS.Version)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}

	//check if version has dependencies
	if sv.Dependencies != nil {
		for _, sd := range sv.Dependencies {
			//check if the service from dependencies has already listed in cluster and version is ok
			for _, clusterS := range cluster.Services {
				if clusterS.Type == sd.ServiceType {
					if !utils.ItemExists(sd.ServiceVersions, clusterS.Version) {
						return nil, ErrClusterDependenceServicesIncompatibleVersion(clusterS.Type, curS.Type), http.StatusBadRequest
					}
				}
			}

			//add service from dependencies with default configurations
			serviceToAdd = &protobuf.Service{
				Name:    curS.Name + "-dependent", //TODO: use better service name?
				Type:    sd.ServiceType,
				Version: sd.DefaultServiceVersion,
			}
			servicesList = append(servicesList, serviceToAdd)
		}
	}

	return servicesList, nil, http.StatusOK
}

func IsImageUsed(db database.Database, logger *logrus.Logger, name string) (bool, error) {
	logger.Info("Checking is image used...")
	clusters, err := db.ReadClustersList()
	if err != nil {
		return false, err
	}
	for _, c := range clusters {
		if c.Image == name {
			return true, nil
		}
	}
	projects, err := db.ReadProjectsList()
	if err != nil {
		return false, err
	}
	for _, p := range projects {
		if p.DefaultImage == name {
			return true, nil
		}
	}
	return false, nil
}

func DeleteSpaces(valStr string) string {
	resStr := ""
	for _, ch := range valStr {
		if ch != ' ' {
			resStr += string(ch)
		}
	}
	return resStr
}

func MakeLogFilePath(filename string, LogsFilePath string) string {
	if LogsFilePath[0] == '/' {
		return LogsFilePath + "/" + filename
	}
	return "./" + LogsFilePath + "/" + filename
}

func GetServiceTypeIdx(service *protobuf.Service, ServiceTypes []protobuf.ServiceType) (int, error) {
	for i, serviceType := range ServiceTypes {
		if serviceType.Type == service.Type {
			return i, nil
		}
	}
	return 0, ErrClusterServiceTypeNotSupported(service.Type)
}

func GetServiceVersionIdx(service *protobuf.Service, ServiceTypes []protobuf.ServiceType, stIdx int) (int, error) {
	for i, sv := range ServiceTypes[stIdx].Versions {
		if sv.Version == service.Version {
			return i, nil
		}
	}
	return 0, ErrClusterServiceVersionNotSupported(service.Version, service.Type)
}
