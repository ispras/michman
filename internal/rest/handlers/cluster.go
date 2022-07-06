package handlers

import (
	"encoding/json"
	"github.com/google/uuid"
	proto "github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/utils"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (hS HttpServer) ClustersGetList(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	projectIdOrName := params.ByName("projectIdOrName")
	request := "/projects/" + projectIdOrName + "/clusters GET"
	hS.Logger.Info("Get " + request)

	project, err := hS.Db.ReadProject(projectIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}
	if project.Name == "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusNotFound, ": ", ErrProjectNotFound.Error())
		ResponseNotFound(w, ErrProjectNotFound)
		return
	}

	//reading cluster info from database
	clusters, err := hS.Db.ReadProjectClusters(project.ID)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	ResponseOK(w, clusters, request)
}

func (hS HttpServer) ClusterCreate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	projectIdOrName := params.ByName("projectIdOrName")
	request := "/project/" + projectIdOrName + "/clusters POST"
	hS.Logger.Info("Get " + request)

	project, err := hS.Db.ReadProject(projectIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}
	if project.Name == "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusNotFound, ": ", ErrProjectNotFound.Error())
		ResponseNotFound(w, ErrProjectNotFound)
		return
	}

	var clusterRes *proto.Cluster
	err = json.NewDecoder(r.Body).Decode(&clusterRes)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrJsonIncorrect.Error())
		ResponseBadRequest(w, ErrJsonIncorrect)
		return
	}

	// set default project flavors if not specified
	if clusterRes.MasterFlavor == "" {
		clusterRes.MasterFlavor = project.DefaultMasterFlavor
	}
	if clusterRes.StorageFlavor == "" {
		clusterRes.StorageFlavor = project.DefaultStorageFlavor
	}
	if clusterRes.SlavesFlavor == "" {
		clusterRes.SlavesFlavor = project.DefaultSlavesFlavor
	}
	if clusterRes.MonitoringFlavor == "" {
		clusterRes.MonitoringFlavor = project.DefaultMonitoringFlavor
	}

	// set default project image if not specified
	if clusterRes.Image == "" {
		clusterRes.Image = project.DefaultImage
	}

	// validate struct
	err, status := ValidateCluster(hS, clusterRes)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", status, ": ", err.Error())
		switch status {
		case http.StatusBadRequest:
			ResponseBadRequest(w, err)
			return
		case http.StatusInternalServerError:
			ResponseInternalError(w, err)
			return
		}
	}

	//check, that cluster with such name doesn't exist
	searchedName := clusterRes.DisplayName + "-" + project.Name
	cluster, err := hS.Db.ReadCluster(project.ID, searchedName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}

	clusterExists := false

	if cluster.Name != "" {
		clusterExists = true
		if cluster.EntityStatus != utils.StatusFailed {
			hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrClusterExisted.Error())
			ResponseBadRequest(w, ErrClusterExisted)
			return
		}
	}

	// If cluster was failed
	if clusterExists {
		clusterRes = cluster
	} else {
		cUuid, err := uuid.NewRandom()
		if err != nil {
			hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", ErrUuidLibError.Error())
			ResponseInternalError(w, ErrUuidLibError)
			return
		}
		clusterRes.ID = cUuid.String()

		//add services from user request and from dependencies
		if clusterRes.Services != nil {
			retryFlag := true
			startIdx := 0

			//first for cycle is used for updating range values with appended services
			for retryFlag {
				for i, s := range clusterRes.Services[startIdx:] {
					st, err := hS.Db.ReadServiceType(s.Type)
					if err != nil {
						hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
						ResponseInternalError(w, err)
						return
					}

					if len(st.HealthCheck) == 0 {
						hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", ErrClusterServiceHealthCheck(st.Type).Error())
						ResponseInternalError(w, ErrClusterServiceHealthCheck(st.Type))
						return
					}

					if s.Version == "" {
						s.Version = st.DefaultVersion
					}

					//add services from dependencies
					sToAdd, err, status := AddDependencies(hS, clusterRes, s)
					if err != nil {
						hS.Logger.Warn("Request ", request, " failed with status ", status, ": ", err.Error())
						switch status {
						case http.StatusBadRequest:
							ResponseBadRequest(w, err)
							return
						case http.StatusInternalServerError:
							ResponseInternalError(w, err)
							return
						}
						return
					}

					changesFlag := false
					if sToAdd != nil {
						for _, curS := range sToAdd {
							clusterRes.Services = append(clusterRes.Services, curS)
						}
						changesFlag = true
					}
					if !changesFlag {
						retryFlag = false
					} else {
						//update range values if new services has been added and start new iteration from the next value
						startIdx = i + 1
						break
					}
				}
			}
		}

		//cluster should be validated after addition services from dependencies
		err, status := ValidateCluster(hS, clusterRes)
		if err != nil {
			hS.Logger.Warn("Request ", request, " failed with status ", status, ": ", err.Error())
			switch status {
			case http.StatusBadRequest:
				ResponseBadRequest(w, err)
				return
			case http.StatusInternalServerError:
				ResponseInternalError(w, err)
				return
			}
		}

		clusterRes.ProjectID = project.ID
		clusterRes.Name = clusterRes.DisplayName + "-" + project.Name

		//set uuids for all cluster services
		for _, s := range clusterRes.Services {
			sUuid, err := uuid.NewRandom()
			if err != nil {
				hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", ErrUuidLibError.Error())
				ResponseInternalError(w, ErrUuidLibError)
				return
			}
			s.ID = sUuid.String()
		}
	}

	clusterRes.EntityStatus = utils.StatusInited
	if !clusterExists {
		err = hS.Db.WriteCluster(clusterRes)
		if err != nil {
			hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
			ResponseInternalError(w, err)
			return
		}
	}
	go hS.Gc.StartClusterCreation(clusterRes)

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusCreated)
	ResponseCreated(w, clusterRes, request)
}

func (hS HttpServer) ClusterGet(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	projectIdOrName := params.ByName("projectIdOrName")
	clusterIdOrName := params.ByName("clusterIdOrName")
	request := "/projects/" + projectIdOrName + "/clusters/" + clusterIdOrName + " GET"
	hS.Logger.Info("Get " + request)

	project, err := hS.Db.ReadProject(projectIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}
	if project.Name == "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusNotFound, ": ", ErrProjectNotFound.Error())
		ResponseNotFound(w, ErrProjectNotFound)
		return
	}

	// reading cluster info from database
	cluster, err := hS.Db.ReadCluster(project.ID, clusterIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}
	if cluster.Name == "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusNotFound, ": ", ErrClusterNotFound.Error())
		ResponseNotFound(w, ErrClusterNotFound)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	ResponseOK(w, cluster, request)
}

func (hS HttpServer) ClusterStatusGet(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	projectIdOrName := params.ByName("projectIdOrName")
	clusterIdOrName := params.ByName("clusterIdOrName")
	request := "/projects/" + projectIdOrName + "/clusters/" + clusterIdOrName + "/status GET"
	hS.Logger.Info("Get " + request)

	project, err := hS.Db.ReadProject(projectIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}
	if project.Name == "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusNotFound, ": ", ErrProjectNotFound.Error())
		ResponseNotFound(w, ErrProjectNotFound)
		return
	}

	// reading cluster info from database
	cluster, err := hS.Db.ReadCluster(project.ID, clusterIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}
	if cluster.Name == "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusNotFound, ": ", ErrClusterNotFound.Error())
		ResponseNotFound(w, ErrClusterNotFound)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	ResponseOK(w, cluster.EntityStatus, request)
}

func (hS HttpServer) ClustersUpdate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	projectIdOrName := params.ByName("projectIdOrName")
	clusterIdOrName := params.ByName("clusterIdOrName")
	request := "/projects/" + projectIdOrName + "/clusters/" + clusterIdOrName + " PUT"
	hS.Logger.Info("Get " + request)

	project, err := hS.Db.ReadProject(projectIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}
	if project.Name == "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusNotFound, ": ", ErrProjectNotFound.Error())
		ResponseNotFound(w, ErrProjectNotFound)
		return
	}

	//check that cluster exists
	hS.Logger.Info("Sending request to db-service to check that cluster exists...")

	// reading cluster info from database
	cluster, err := hS.Db.ReadCluster(project.ID, clusterIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}
	if cluster.Name == "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusNotFound, ": ", ErrClusterNotFound.Error())
		ResponseNotFound(w, ErrClusterNotFound)
		return
	}

	if cluster.EntityStatus != utils.StatusActive && cluster.EntityStatus != utils.StatusFailed {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", ErrClusterStatus.Error())
		ResponseInternalError(w, ErrClusterStatus)
		return
	}

	// validate request struct
	var newC proto.Cluster
	err = json.NewDecoder(r.Body).Decode(&newC)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrJsonIncorrect.Error())
		ResponseBadRequest(w, ErrJsonIncorrect)
		return
	}

	if newC.ID != "" || newC.Name != "" || newC.EntityStatus != "" || newC.NHosts != 0 ||
		newC.HostURL != "" || newC.MasterIP != "" || newC.ProjectID != "" || newC.Image != "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrClusterUnmodFields.Error())
		ResponseBadRequest(w, ErrClusterUnmodFields)
		return
	}

	//check correctness of new services
	for _, s := range newC.Services {
		err, status := ValidateService(hS, s)
		if err != nil {
			hS.Logger.Warn("Request ", request, " failed with status ", status, ": ", err.Error())
			switch status {
			case http.StatusBadRequest:
				ResponseBadRequest(w, err)
				return
			case http.StatusInternalServerError:
				ResponseInternalError(w, err)
				return
			}
		}
	}

	sTypes, err := hS.Db.ReadServicesTypesList()
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}
	//appending old services which does not exist in new cluster configuration
	var serviceTypesOld = make(map[string]serviceExists)

	for _, st := range sTypes {
		serviceTypesOld[st.Type] = serviceExists{
			exists:  false,
			service: nil,
		}
	}

	for _, s := range cluster.Services {
		serviceTypesOld[s.Type] = serviceExists{
			exists:  true,
			service: s,
		}
	}
	//new nodes must be added for some special services types
	newHost := false

	//number of old services
	oldSN := len(cluster.Services)

	for _, s := range newC.Services {
		if serviceTypesOld[s.Type].exists == false {
			sUuid, err := uuid.NewRandom()
			if err != nil {
				hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", ErrUuidLibError.Error())
				ResponseInternalError(w, ErrUuidLibError)
				return
			}
			s.ID = sUuid.String()
			cluster.Services = append(cluster.Services, s)
		}

		st, err := hS.Db.ReadServiceType(s.Type)
		if err != nil {
			hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
			ResponseInternalError(w, err)
			return
		}
		if st.Class == utils.ClassStorage {
			newHost = true
		}
	}

	//check if new services are added
	if oldSN != len(cluster.Services) {
		retryFlag := true
		startIdx := oldSN

		//first for cycle is used for updating range values with appended services
		for retryFlag {
			for i, s := range cluster.Services[startIdx:] {
				st, err := hS.Db.ReadServiceType(s.Type)
				if err != nil {
					hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
					ResponseInternalError(w, err)
					return
				}

				if s.Version == "" {
					s.Version = st.DefaultVersion
				}

				//add services from dependencies
				sToAdd, err, status := AddDependencies(hS, cluster, s)
				if err != nil {
					hS.Logger.Warn("Request ", request, " failed with status ", status, ": ", err.Error())
					switch status {
					case http.StatusBadRequest:
						ResponseBadRequest(w, err)
						return
					case http.StatusInternalServerError:
						ResponseInternalError(w, err)
						return
					}
					return
				}

				changesFlag := false
				if sToAdd != nil {
					for _, curS := range sToAdd {
						sUuid, err := uuid.NewRandom()
						if err != nil {
							hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", ErrUuidLibError.Error())
							ResponseInternalError(w, ErrUuidLibError)
							return
						}
						curS.ID = sUuid.String()

						cluster.Services = append(cluster.Services, curS)
					}
					changesFlag = true
				}

				if !changesFlag {
					retryFlag = false
				} else {
					//update range values if new services has been added and start new iteration from the next value
					startIdx = i + 1
					break
				}
			}
		}
	}

	if newC.Description != "" {
		cluster.Description = newC.Description
	}

	if newC.DisplayName != "" {
		cluster.DisplayName = newC.DisplayName
	}

	cluster.EntityStatus = utils.StatusInited
	if newC.NHosts != 0 || newHost {
		go hS.Gc.StartClusterCreation(cluster)
	} else {
		go hS.Gc.StartClusterModification(cluster)
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	ResponseOK(w, cluster, request)
}

func (hS HttpServer) ClustersDelete(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	projectIdOrName := params.ByName("projectIdOrName")
	clusterIdOrName := params.ByName("clusterIdOrName")
	request := "/projects/" + projectIdOrName + "/clusters/" + clusterIdOrName + " DELETE"
	hS.Logger.Info("Get " + request)

	project, err := hS.Db.ReadProject(projectIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}
	if project.Name == "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusNotFound, ": ", ErrProjectNotFound.Error())
		ResponseNotFound(w, ErrProjectNotFound)
		return
	}

	// reading cluster info from database
	cluster, err := hS.Db.ReadCluster(project.ID, clusterIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}
	if cluster.Name == "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusNotFound, ": ", ErrClusterNotFound.Error())
		ResponseNotFound(w, ErrClusterNotFound)
		return
	}

	if cluster.EntityStatus != utils.StatusActive && cluster.EntityStatus != utils.StatusFailed {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", ErrClusterStatus.Error())
		ResponseInternalError(w, ErrClusterStatus)
		return
	}

	cluster.EntityStatus = utils.StatusStopping

	go hS.Gc.StartClusterDestroying(cluster)

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	ResponseOK(w, cluster, request)
}
