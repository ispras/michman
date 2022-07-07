package validate

import (
	"github.com/ispras/michman/internal/database"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/rest/handler/check"
	"github.com/ispras/michman/internal/utils"
	"github.com/sirupsen/logrus"
	"net/http"
)

func Cluster(db database.Database, logger *logrus.Logger, cluster *protobuf.Cluster) (error, int) {
	logger.Info("Validating cluster...")
	if err, status := check.ValidName(cluster.DisplayName, utils.ClusterNamePattern, ErrClusterBadName); err != nil {
		return err, status
	}

	for _, service := range cluster.Services {
		if err, status := Service(db, logger, service); err != nil {
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
	return nil, 0
}
