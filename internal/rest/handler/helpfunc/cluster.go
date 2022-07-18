package helpfunc

import (
	"github.com/google/uuid"
	"github.com/ispras/michman/internal/database"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/utils"
	"net/http"
)

type ServiceExists struct {
	Exists  bool
	Service *protobuf.Service
}

// GetServiceTypeIdx returns the ordinal number of the desired service in the list of all existing service types
func GetServiceTypeIdx(service *protobuf.Service, ServiceTypes []protobuf.ServiceType) (int, error) {
	for i, serviceType := range ServiceTypes {
		if serviceType.Type == service.Type {
			return i, nil
		}
	}
	return 0, ErrClusterServiceTypeNotSupported(service.Type)
}

// GetServiceVersionIdx returns the ordinal number of the desired service version in the list of all existing versions of the service
func GetServiceVersionIdx(service *protobuf.Service, ServiceTypes []protobuf.ServiceType, stIdx int) (int, error) {
	for i, sv := range ServiceTypes[stIdx].Versions {
		if sv.Version == service.Version {
			return i, nil
		}
	}
	return 0, ErrClusterServiceVersionNotSupported(service.Version, service.Type)
}

// SetClusterServicesUuids set uuids for all cluster services
func SetClusterServicesUuids(cluster *protobuf.Cluster) error {
	for _, service := range cluster.Services {
		sUuid, err := uuid.NewRandom()
		if err != nil {
			return err
		}
		service.ID = sUuid.String()
	}
	return nil
}

// SetDefaults set some cluster fields by default from project if not specified by user
func SetDefaults(cluster *protobuf.Cluster, project *protobuf.Project) {
	// set default project flavors if not specified
	if cluster.MasterFlavor == "" {
		cluster.MasterFlavor = project.DefaultMasterFlavor
	}
	if cluster.StorageFlavor == "" {
		cluster.StorageFlavor = project.DefaultStorageFlavor
	}
	if cluster.SlavesFlavor == "" {
		cluster.SlavesFlavor = project.DefaultSlavesFlavor
	}
	if cluster.MonitoringFlavor == "" {
		cluster.MonitoringFlavor = project.DefaultMonitoringFlavor
	}

	// set default project image if not specified
	if cluster.Image == "" {
		cluster.Image = project.DefaultImage
	}
}

// GetDependencies get and return serviceList of services and their services from dependencies
func GetDependencies(db database.Database, cluster *protobuf.Cluster, curService *protobuf.Service) ([]*protobuf.Service, error, int) {
	var serviceToAdd *protobuf.Service = nil
	var servicesList []*protobuf.Service = nil

	// a set of services specified in the user request or existed in cluster
	clusterServiceTypesInfo, _, err := SetServiceExistInfo(db, cluster)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}

	serviceVersion, err := db.ReadServiceTypeVersion(curService.Type, curService.Version)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}

	// check if version has dependencies
	if serviceVersion.Dependencies != nil {
		for _, serviceDependency := range serviceVersion.Dependencies {
			// checking that the service from the dependencies is explicitly specified by the user and there is no need to add a default
			if servInfo := clusterServiceTypesInfo[serviceDependency.ServiceType]; servInfo.Exists == true {
				if !utils.ItemExists(serviceDependency.ServiceVersions, servInfo.Service.Version) {
					return nil, ErrClusterDependenceServicesIncompatibleVersion(servInfo.Service.Type, curService.Type), http.StatusBadRequest
				}
				continue
			}

			// add service from dependencies with default configurations
			serviceToAdd = &protobuf.Service{
				Name:    curService.Name + "-dependent", //TODO: use better service name?
				Type:    serviceDependency.ServiceType,
				Version: serviceDependency.DefaultServiceVersion,
			}
			servicesList = append(servicesList, serviceToAdd)
		}
	}

	return servicesList, nil, http.StatusOK
}

// SetServiceExistInfo set information struct about services exist or not exist in old cluster
func SetServiceExistInfo(db database.Database, oldCluster *protobuf.Cluster) (map[string]ServiceExists, int, error) {
	sTypes, err := db.ReadServicesTypesList()
	if err != nil {
		return nil, 0, err
	}

	var serviceTypesOld = make(map[string]ServiceExists)

	// services not exist in old cluster struct
	for _, serviceType := range sTypes {
		serviceTypesOld[serviceType.Type] = ServiceExists{
			Exists:  false,
			Service: nil,
		}
	}

	// services exist in old cluster struct
	for _, service := range oldCluster.Services {
		serviceTypesOld[service.Type] = ServiceExists{
			Exists:  true,
			Service: service,
		}
	}

	// number of old services
	oldServiceNumber := len(oldCluster.Services)

	return serviceTypesOld, oldServiceNumber, nil
}

// AppendNewServices append new services to the cluster structure and return bool variable that shows if new slave node must be created
func AppendNewServices(db database.Database, serviceTypesOld map[string]ServiceExists, newCluster *protobuf.Cluster, resCluster *protobuf.Cluster) (bool, error) {
	// new nodes must be added for some special services types
	newHost := false

	for _, service := range newCluster.Services {
		if serviceTypesOld[service.Type].Exists == false {
			// generating UUID for new services
			sUuid, err := uuid.NewRandom()
			if err != nil {
				return false, ErrUuidLibError
			}
			service.ID = sUuid.String()
			resCluster.Services = append(resCluster.Services, service)
		}

		st, err := db.ReadServiceType(service.Type)
		if err != nil {
			return false, err
		}
		if st.Class == utils.ClassStorage {
			newHost = true
		}
	}

	return newHost, nil
}

// AppendDependentServices append services from dependencies to the cluster structure
func AppendDependentServices(servicesToAdd []*protobuf.Service, resCluster *protobuf.Cluster) (bool, error, int) {
	changesFlag := false
	if servicesToAdd != nil {
		for _, curService := range servicesToAdd {
			// generating UUID for added new services from dependencies
			sUuid, err := uuid.NewRandom()
			if err != nil {
				return false, ErrUuidLibError, http.StatusInternalServerError
			}
			curService.ID = sUuid.String()

			resCluster.Services = append(resCluster.Services, curService)
		}
		changesFlag = true
	}

	return changesFlag, nil, http.StatusOK
}

// UpdateRangeValuesAppendedServices used for updating range values of appended services and append services from their dependencies
func UpdateRangeValuesAppendedServices(db database.Database, oldServiceNumber int, resCluster *protobuf.Cluster, action string) (error, int) {
	retryFlag := true
	startIdx := oldServiceNumber

	// first for cycle is used for updating range values with appended services
	for retryFlag {
		for i, service := range resCluster.Services[startIdx:] {
			// read service type from database
			serviceType, err := db.ReadServiceType(service.Type)
			if err != nil {
				return err, http.StatusInternalServerError
			}

			if action == utils.ActionCreate && len(serviceType.HealthCheck) == 0 {
				return ErrClusterServiceHealthCheck(serviceType.Type), http.StatusInternalServerError
			}

			if service.Version == "" {
				service.Version = serviceType.DefaultVersion
			}

			// get services from dependencies
			servicesToAdd, err, status := GetDependencies(db, resCluster, service)
			if err != nil {
				return err, status
			}

			// append services from dependencies to resCluster struct
			changesFlag, err, status := AppendDependentServices(servicesToAdd, resCluster)
			if err != nil {
				return err, status
			}

			if !changesFlag {
				retryFlag = false
			} else {
				// update range values if new services has been added and start new iteration from the next value
				startIdx = i + 1
				break
			}
		}
	}

	return nil, http.StatusOK
}
