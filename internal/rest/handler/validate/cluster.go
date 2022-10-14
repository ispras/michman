package validate

import (
	"github.com/ispras/michman/internal/database"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/rest/handler/check"
	"github.com/ispras/michman/internal/rest/handler/helpfunc"
	"github.com/ispras/michman/internal/utils"
)

// ClusterCreateGeneral function validates common cluster` information without services
func ClusterCreateGeneral(db database.Database, cluster *protobuf.Cluster) error {
	if cluster.DisplayName == "" {
		ErrEmptyField("cluster", "DisplayName")
	}
	err := check.ClusterValidName(cluster.DisplayName)
	if err != nil {
		return err
	}
	if cluster.ID != "" {
		return ErrGeneratedField("cluster", "ID")
	}
	if cluster.Name != "" {
		return ErrGeneratedField("cluster", "Name")
	}
	if cluster.OwnerID != "" {
		return ErrGeneratedField("cluster", "OwnerID")
	}
	if cluster.HostURL != "" {
		return ErrGeneratedField("cluster", "HostURL")
	}
	if cluster.EntityStatus != "" {
		return ErrGeneratedField("cluster", "EntityStatus")
	}
	if cluster.ClusterType != "" {
		return ErrGeneratedField("cluster", "ClusterType")
	}
	if cluster.MasterIP != "" {
		return ErrGeneratedField("cluster", "MasterIP")
	}
	if cluster.ProjectID != "" {
		return ErrGeneratedField("cluster", "ProjectID")
	}

	if cluster.NHosts < 0 {
		return ErrClusterNhostsZero
	}

	_, err = db.ReadImage(cluster.Image)
	if err != nil {
		return err
	}
	_, err = db.ReadFlavor(cluster.MasterFlavor)
	if err != nil {
		return err
	}
	_, err = db.ReadFlavor(cluster.SlavesFlavor)
	if err != nil {
		return err
	}
	_, err = db.ReadFlavor(cluster.StorageFlavor)
	if err != nil {
		return err
	}
	_, err = db.ReadFlavor(cluster.MonitoringFlavor)
	if err != nil {
		return err
	}
	return nil
}

// ClusterCreateServices validates fields of the cluster structure for correct filling when creating
func ClusterCreateServices(db database.Database, cluster *protobuf.Cluster) error {
	// check correctness of services
	for _, service := range cluster.Services {
		err := ClusterService(db, service)
		if err != nil {
			return err
		}
	}

	if cluster.NHosts == 0 {
		res, err := check.MSServices(db, cluster)
		if err != nil {
			return err
		}
		if res {
			return ErrClustersNhostsMasterSlave
		}
	}

	return nil
}

// ClusterUpdate validates fields of the cluster structure for correct filling when updating
func ClusterUpdate(db database.Database, oldCluster *protobuf.Cluster, newCluster *protobuf.Cluster) error {
	if oldCluster.EntityStatus != utils.StatusActive && oldCluster.EntityStatus != utils.StatusFailed {
		return ErrClusterStatus
	}
	if newCluster.ID != "" {
		return ErrClusterUnmodFields("ID")
	}
	if newCluster.Name != "" {
		return ErrClusterUnmodFields("Name")
	}
	if newCluster.EntityStatus != "" {
		return ErrClusterUnmodFields("EntityStatus")
	}
	if newCluster.NHosts != 0 {
		return ErrClusterUnmodFields("NHosts")
	}
	if newCluster.HostURL != "" {
		return ErrClusterUnmodFields("HostURL")
	}
	if newCluster.MasterIP != "" {
		return ErrClusterUnmodFields("MasterIP")
	}
	if newCluster.ProjectID != "" {
		return ErrClusterUnmodFields("ProjectID")
	}
	if newCluster.Image != "" {
		return ErrClusterUnmodFields("Image")
	}
	if newCluster.OwnerID != "" {
		return ErrClusterUnmodFields("OwnerID")
	}
	if newCluster.MasterFlavor != "" {
		return ErrClusterUnmodFields("MasterFlavor")
	}
	if newCluster.SlavesFlavor != "" {
		return ErrClusterUnmodFields("SlavesFlavor")
	}
	if newCluster.StorageFlavor != "" {
		return ErrClusterUnmodFields("StorageFlavor")
	}
	if newCluster.MonitoringFlavor != "" {
		return ErrClusterUnmodFields("MonitoringFlavor")
	}

	// check correctness of new services
	for _, services := range newCluster.Services {
		err := ClusterService(db, services)
		if err != nil {
			return err
		}
	}

	return nil
}

// ClusterDelete validates the cluster structure for the correct status when deleting
func ClusterDelete(cluster *protobuf.Cluster) error {
	if cluster.EntityStatus != utils.StatusActive && cluster.EntityStatus != utils.StatusFailed {
		return ErrClusterStatus
	}

	return nil
}

// ClusterServices validates service fields of the cluster structure after addition services from dependencies
func ClusterServices(db database.Database, cluster *protobuf.Cluster) error {
	// check correctness of services
	for _, service := range cluster.Services {
		if err := ClusterService(db, service); err != nil {
			return err
		}
	}
	return nil
}

// ClusterService validates service fields of the cluster structure for correct filling when updating or creating
func ClusterService(db database.Database, service *protobuf.Service) error {
	if service.Type == "" {
		return ErrClusterServiceTypeEmpty
	}

	sTypes, err := db.ReadServicesTypesList()
	if err != nil {
		return err
	}

	stIdx, err := helpfunc.GetServiceTypeIdx(service, sTypes)
	if err != nil {
		return err
	}

	// check service version
	if service.Version == "" && sTypes[stIdx].DefaultVersion != "" {
		service.Version = sTypes[stIdx].DefaultVersion
	} else if service.Version == "" && sTypes[stIdx].DefaultVersion == "" {
		return ErrClusterServiceVersionsEmpty(service.Type)
	}

	svIdx, err := helpfunc.GetServiceVersionIdx(service, sTypes, stIdx)
	if err != nil {
		return err
	}

	if err = check.ServiceConfigCorrectValue(service, sTypes[stIdx].Versions[svIdx].Configs); err != nil {
		return err
	}

	return nil
}
