package validate

import (
	"github.com/ispras/michman/internal/database"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/rest/handler/helpfunc"
	"github.com/sirupsen/logrus"
	"net/http"
)

func Service(db database.Database, logger *logrus.Logger, service *protobuf.Service) (error, int) {
	logger.Info("Validating service type and config params...")

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

	//check service version
	if service.Version == "" && sTypes[stIdx].DefaultVersion != "" {
		service.Version = sTypes[stIdx].DefaultVersion
	} else if service.Version == "" && sTypes[stIdx].DefaultVersion == "" {
		return ErrClusterServiceVersionsEmpty(service.Type), http.StatusBadRequest
	}

	svIdx, err := helpfunc.GetServiceVersionIdx(service, sTypes, stIdx)
	if err != nil {
		return err, http.StatusBadRequest
	}

	if err, status := Configs(service, sTypes[stIdx].Versions[svIdx].Configs); err != nil {
		return err, status
	}

	return nil, 0
}
