package validate

import (
	"github.com/ispras/michman/internal/database"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/rest/handler/check"
	"github.com/ispras/michman/internal/rest/handler/helpfunc"
	"github.com/ispras/michman/internal/utils"
	"net/http"
)

// ClusterCreate validates fields of the cluster structure for correct filling when creating
func ClusterCreate(db database.Database, cluster *protobuf.Cluster) (error, int) {
	if err, status := check.ValidName(cluster.DisplayName, utils.ClusterNamePattern, ErrClusterBadName); err != nil {
		return err, status
	}

	if cluster.OwnerID != "" {
		return ErrClusterOwnerNotEmpty, http.StatusBadRequest
	}

	// check correctness of services
	for _, service := range cluster.Services {
		if err, status := ClusterService(db, service); err != nil {
			return err, status
		}
	}

	if cluster.NHosts < 0 {
		return ErrClusterNhostsZero, http.StatusBadRequest
	}

	if cluster.NHosts == 0 {
		res, err := check.MSServices(db, cluster)
		if err != nil {
			return err, http.StatusInternalServerError
		}
		if res {
			return ErrClustersNhostsMasterSlave, http.StatusBadRequest
		}
	}

	dbImg, err := db.ReadImage(cluster.Image)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if dbImg.Name == "" {
		return ErrClusterImageNotFound, http.StatusBadRequest
	}
	dbFlavor, err := db.ReadFlavor(cluster.MasterFlavor)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if dbFlavor.ID == "" {
		return ErrFlavorFieldValueNotFound("MasterFlavor"), http.StatusBadRequest
	}
	dbFlavor, err = db.ReadFlavor(cluster.SlavesFlavor)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if dbFlavor.ID == "" {
		return ErrFlavorFieldValueNotFound("SlavesFlavor"), http.StatusBadRequest
	}
	dbFlavor, err = db.ReadFlavor(cluster.StorageFlavor)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if dbFlavor.ID == "" {
		return ErrFlavorFieldValueNotFound("StorageFlavor"), http.StatusBadRequest
	}
	dbFlavor, err = db.ReadFlavor(cluster.MonitoringFlavor)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if dbFlavor.ID == "" {
		return ErrFlavorFieldValueNotFound("MonitoringFlavor"), http.StatusBadRequest
	}
	return nil, http.StatusOK
}

// ClusterUpdate validates fields of the cluster structure for correct filling when updating
func ClusterUpdate(db database.Database, oldCluster *protobuf.Cluster, newCluster *protobuf.Cluster) (error, int) {
	if oldCluster.EntityStatus != utils.StatusActive && oldCluster.EntityStatus != utils.StatusFailed {
		return ErrClusterStatus, http.StatusInternalServerError
	}
	if newCluster.ID != "" {
		return ErrClusterUnmodFields("ID"), http.StatusBadRequest
	}
	if newCluster.Name != "" {
		return ErrClusterUnmodFields("Name"), http.StatusBadRequest
	}
	if newCluster.EntityStatus != "" {
		return ErrClusterUnmodFields("EntityStatus"), http.StatusBadRequest
	}
	if newCluster.NHosts != 0 {
		return ErrClusterUnmodFields("NHosts"), http.StatusBadRequest
	}
	if newCluster.HostURL != "" {
		return ErrClusterUnmodFields("HostURL"), http.StatusBadRequest
	}
	if newCluster.MasterIP != "" {
		return ErrClusterUnmodFields("MasterIP"), http.StatusBadRequest
	}
	if newCluster.ProjectID != "" {
		return ErrClusterUnmodFields("ProjectID"), http.StatusBadRequest
	}
	if newCluster.Image != "" {
		return ErrClusterUnmodFields("Image"), http.StatusBadRequest
	}
	if newCluster.OwnerID != "" {
		return ErrClusterUnmodFields("OwnerID"), http.StatusBadRequest
	}
	if newCluster.MasterFlavor != "" {
		return ErrClusterUnmodFields("MasterFlavor"), http.StatusBadRequest
	}
	if newCluster.SlavesFlavor != "" {
		return ErrClusterUnmodFields("SlavesFlavor"), http.StatusBadRequest
	}
	if newCluster.StorageFlavor != "" {
		return ErrClusterUnmodFields("StorageFlavor"), http.StatusBadRequest
	}
	if newCluster.MonitoringFlavor != "" {
		return ErrClusterUnmodFields("MonitoringFlavor"), http.StatusBadRequest
	}

	// check correctness of new services
	for _, services := range newCluster.Services {
		err, status := ClusterService(db, services)
		if err != nil {
			return err, status
		}
	}

	return nil, http.StatusOK
}

// ClusterDelete validates the cluster structure for the correct status when deleting
func ClusterDelete(cluster *protobuf.Cluster) (error, int) {
	if cluster.EntityStatus != utils.StatusActive && cluster.EntityStatus != utils.StatusFailed {
		return ErrClusterStatus, http.StatusInternalServerError
	}

	return nil, http.StatusOK
}

// ClusterAddedServices validates service fields of the cluster structure after addition services from dependencies
func ClusterAddedServices(db database.Database, cluster *protobuf.Cluster) (error, int) {
	// check correctness of services
	for _, service := range cluster.Services {
		if err, status := ClusterService(db, service); err != nil {
			return err, status
		}
	}
	return nil, http.StatusOK
}

// ClusterService validates service fields of the cluster structure for correct filling when updating or creating
func ClusterService(db database.Database, service *protobuf.Service) (error, int) {
	if service.Type == "" {
		return ErrClusterServiceTypeEmpty, http.StatusBadRequest
	}

	sTypes, err := db.ReadServicesTypesList()
	if err != nil {
		return err, http.StatusInternalServerError
	}

	stIdx, err := helpfunc.GetServiceTypeIdx(service, sTypes)
	if err != nil {
		return err, http.StatusBadRequest
	}

	// check service version
	if service.Version == "" && sTypes[stIdx].DefaultVersion != "" {
		service.Version = sTypes[stIdx].DefaultVersion
	} else if service.Version == "" && sTypes[stIdx].DefaultVersion == "" {
		return ErrClusterServiceVersionsEmpty(service.Type), http.StatusBadRequest
	}

	svIdx, err := helpfunc.GetServiceVersionIdx(service, sTypes, stIdx)
	if err != nil {
		return err, http.StatusBadRequest
	}

	if err = check.ServiceConfigCorrectValue(service, sTypes[stIdx].Versions[svIdx].Configs); err != nil {
		return err, http.StatusBadRequest
	}

	return nil, 0
}
