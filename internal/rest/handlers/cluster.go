package handlers

import (
	"encoding/json"
	"github.com/google/uuid"
	proto "github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/utils"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (hS HttpServer) ClustersGet(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	projectIdOrName := params.ByName("projectIdOrName")
	hS.Logger.Print("Get /projects/", projectIdOrName, "/clusters GET")

	project, err := hS.Db.ReadProject(projectIdOrName)
	if err != nil {
		mess, _ := hS.RespHandler.Handle(w, DBerror, DBerrorMessage, err)
		hS.Logger.Print(mess)
		return
	}

	if project.Name == "" {
		hS.Logger.Printf("Project with name '%s' not found", projectIdOrName)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	//reading cluster info from database
	clusters, err := hS.Db.ReadProjectClusters(project.ID)
	if err != nil {
		mess, _ := hS.RespHandler.Handle(w, DBerror, DBerrorMessage, err)
		hS.Logger.Print(mess)
		return
	}

	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	err = enc.Encode(clusters)
	if err != nil {
		mess, _ := hS.RespHandler.Handle(w, LibErrorStructToJson, LibErrorStructToJsonMessage, err)
		hS.Logger.Print(mess)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

func (hS HttpServer) ClusterCreate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	projectIdOrName := params.ByName("projectIdOrName")
	hS.Logger.Print("Get /project/" + projectIdOrName + "/clusters POST")

	project, err := hS.Db.ReadProject(projectIdOrName)
	if err != nil {
		mess, _ := hS.RespHandler.Handle(w, DBerror, DBerrorMessage, err)
		hS.Logger.Print(mess)
		return
	}

	if project.Name == "" {
		hS.Logger.Printf("Project with name '%s' not found", projectIdOrName)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var clusterRes *proto.Cluster
	err = json.NewDecoder(r.Body).Decode(&clusterRes)
	if err != nil {
		mess, _ := hS.RespHandler.Handle(w, JSONerrorIncorrect, JSONerrorIncorrectMessage, err)
		hS.Logger.Print(mess)
		return
	}

	// set default project flavors if not specifie
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

	// validate struct
	validateRes, err := ValidateCluster(hS, clusterRes)
	if err != nil { //only db error returns
		mess, _ := hS.RespHandler.Handle(w, DBerror, DBerrorMessage, err)
		hS.Logger.Print(mess)
		return
	}
	if !validateRes {
		mess, _ := hS.RespHandler.Handle(w, JSONerrorIncorrectField, JSONerrorIncorrectFieldMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	//check, that cluster with such name doesn't exist
	searchedName := clusterRes.DisplayName + "-" + project.Name
	cluster, err := hS.Db.ReadCluster(project.ID, searchedName)
	if err != nil {
		mess, _ := hS.RespHandler.Handle(w, DBerror, DBerrorMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	clusterExists := false

	if cluster.Name != "" {
		clusterExists = true
		if cluster.EntityStatus != utils.StatusFailed {
			mess, _ := hS.RespHandler.Handle(w, UserErrorClusterExisted, UserErrorClusterExistedMessage, nil)
			hS.Logger.Print(mess)
			return
		}
	}

	// If cluster was failed
	if clusterExists {
		clusterRes = cluster
	} else {
		cUuid, err := uuid.NewRandom()
		if err != nil {
			mess, _ := hS.RespHandler.Handle(w, LibErrorUUID, LibErrorUUIDMessage, nil)
			hS.Logger.Print(mess)
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
						mess, _ := hS.RespHandler.Handle(w, DBerror, DBerrorMessage, nil)
						hS.Logger.Print(mess)
						return
					}

					if len(st.HealthCheck) == 0 {
						mess, _ := hS.RespHandler.Handle(w, DBerror, DBemptyHealthCheck, nil)
						hS.Logger.Print(mess)
						return
					}

					if s.Version == "" {
						s.Version = st.DefaultVersion
					}

					//add services from dependencies
					sToAdd, err := AddDependencies(hS, clusterRes, s)
					if err != nil {
						hS.Logger.Println(err)
						mess, _ := hS.RespHandler.Handle(w, UserErrorBadServiceVersion, UserErrorBadServiceVersionMessage, nil)
						hS.Logger.Print(mess)
						//w.WriteHeader(http.StatusBadRequest)
						return
					}
					hS.Logger.Println(sToAdd)

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
		validateRes, err := ValidateCluster(hS, clusterRes)
		if err != nil { //only db error returns
			mess, _ := hS.RespHandler.Handle(w, DBerror, DBerrorMessage, err)
			hS.Logger.Print(mess)
			return
		}
		if !validateRes {
			mess, _ := hS.RespHandler.Handle(w, JSONerrorIncorrectField, JSONerrorIncorrectFieldMessage, nil)
			hS.Logger.Print(mess)
			return
		}

		clusterRes.ProjectID = project.ID
		clusterRes.Name = clusterRes.DisplayName + "-" + project.Name
		//set uuids for all cluster services
		for _, s := range clusterRes.Services {
			sUuid, err := uuid.NewRandom()
			if err != nil {
				mess, _ := hS.RespHandler.Handle(w, LibErrorUUID, LibErrorUUIDMessage, nil)
				hS.Logger.Print(mess)
				return
			}
			s.ID = sUuid.String()
		}
	}
	// set default project Image if not specified
	if clusterRes.Image == "" {
		clusterRes.Image = project.DefaultImage
	}

	clusterRes.EntityStatus = utils.StatusInited
	if !clusterExists {
		err = hS.Db.WriteCluster(clusterRes)
		if err != nil {
			mess, _ := hS.RespHandler.Handle(w, DBerror, DBerrorMessage, nil)
			hS.Logger.Print(mess)
			return
		}
	}
	go hS.Gc.StartClusterCreation(clusterRes)

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(clusterRes)
}

func (hS HttpServer) ClustersGetByName(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	projectIdOrName := params.ByName("projectIdOrName")
	clusterIdOrName := params.ByName("clusterIdOrName")
	hS.Logger.Print("Get /projects/"+projectIdOrName+"/clusters/", clusterIdOrName, " GET")

	project, err := hS.Db.ReadProject(projectIdOrName)
	if err != nil {
		mess, _ := hS.RespHandler.Handle(w, DBerror, DBerrorMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	if project.Name == "" {
		hS.Logger.Printf("Project with name '%s' not found", projectIdOrName)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// reading cluster info from database
	cluster, err := hS.Db.ReadCluster(project.ID, clusterIdOrName)
	if err != nil {
		mess, _ := hS.RespHandler.Handle(w, DBerror, DBerrorMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	if cluster.Name == "" {
		hS.Logger.Printf("Cluster with name or ID '%s' not found", clusterIdOrName)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	err = enc.Encode(cluster)
	if err != nil {
		mess, _ := hS.RespHandler.Handle(w, LibErrorStructToJson, LibErrorStructToJsonMessage, nil)
		hS.Logger.Print(mess)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

func (hS HttpServer) ClustersStatusGetByName(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	projectIdOrName := params.ByName("projectIdOrName")
	clusterIdOrName := params.ByName("clusterIdOrName")
	hS.Logger.Print("Get /projects/"+projectIdOrName+"/clusters/", clusterIdOrName, "/status", " GET")

	project, err := hS.Db.ReadProject(projectIdOrName)
	if err != nil {
		mess, _ := hS.RespHandler.Handle(w, DBerror, DBerrorMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	if project.Name == "" {
		hS.Logger.Printf("Project with name '%s' not found", projectIdOrName)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// reading cluster info from database
	cluster, err := hS.Db.ReadCluster(project.ID, clusterIdOrName)
	if err != nil {
		mess, _ := hS.RespHandler.Handle(w, DBerror, DBerrorMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	if cluster.Name == "" {
		hS.Logger.Printf("Cluster with name or ID '%s' not found", clusterIdOrName)
		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(utils.StatusMissing))
		if err != nil {
			hS.Logger.Print(err)
			w.WriteHeader(http.StatusBadRequest)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(cluster.EntityStatus))
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (hS HttpServer) ClustersUpdate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	projectIdOrName := params.ByName("projectIdOrName")
	clusterIdOrName := params.ByName("clusterIdOrName")
	hS.Logger.Print("Get /projects/"+projectIdOrName+"/clusters/", clusterIdOrName, " PUT")

	project, err := hS.Db.ReadProject(projectIdOrName)
	if err != nil {
		mess, _ := hS.RespHandler.Handle(w, DBerror, DBerrorMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	if project.Name == "" {
		hS.Logger.Printf("Project with name '%s' not found", projectIdOrName)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	//check that cluster exists
	hS.Logger.Print("Sending request to db-service to check that cluster exists...")

	// reading cluster info from database
	cluster, err := hS.Db.ReadCluster(project.ID, clusterIdOrName)
	if err != nil {
		mess, _ := hS.RespHandler.Handle(w, DBerror, DBerrorMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	if cluster.Name == "" {
		hS.Logger.Printf("Cluster with name or ID '%s' not found", clusterIdOrName)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if cluster.EntityStatus != utils.StatusActive && cluster.EntityStatus != utils.StatusFailed {
		mess, _ := hS.RespHandler.Handle(w, UserErrorClusterStatus, UserErrorClusterStatusMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	// validate request struct
	var newC proto.Cluster
	err = json.NewDecoder(r.Body).Decode(&newC)
	if err != nil {
		mess, _ := hS.RespHandler.Handle(w, JSONerrorIncorrect, JSONerrorIncorrectMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	if newC.ID != "" || newC.Name != "" || newC.DisplayName != "" || newC.EntityStatus != "" ||
		newC.HostURL != "" || newC.MasterIP != "" || newC.ProjectID != "" || newC.Image != "" {
		mess, _ := hS.RespHandler.Handle(w, UserErrorProjectUnmodField, UserErrorProjectUnmodFieldMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	//check correctness of new services
	for _, s := range newC.Services {
		if res, _ := ValidateService(hS, s); !res {
			mess, _ := hS.RespHandler.Handle(w, UserErrorProjectUnmodField, UserErrorProjectUnmodFieldMessage, nil)
			hS.Logger.Print(mess)
			return
		}
	}

	sTypes, err := hS.Db.ReadServicesTypesList()
	if err != nil {
		mess, _ := hS.RespHandler.Handle(w, DBerror, DBerrorMessage, nil)
		hS.Logger.Print(mess)
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
				mess, _ := hS.RespHandler.Handle(w, LibErrorUUID, LibErrorUUIDMessage, nil)
				hS.Logger.Print(mess)
				return
			}
			s.ID = sUuid.String()
			cluster.Services = append(cluster.Services, s)
		}

		st, err := hS.Db.ReadServiceType(s.Type)
		if err != nil {
			hS.Logger.Print(err)
			w.WriteHeader(http.StatusBadRequest)
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
					mess, _ := hS.RespHandler.Handle(w, DBerror, DBerrorMessage, nil)
					hS.Logger.Print(mess)
					return
				}

				if s.Version == "" {
					s.Version = st.DefaultVersion
				}

				//add services from dependencies
				sToAdd, err := AddDependencies(hS, cluster, s)
				if err != nil {
					hS.Logger.Println(err)
					mess, _ := hS.RespHandler.Handle(w, UserErrorBadServiceVersion, UserErrorBadServiceVersionMessage, nil)
					hS.Logger.Print(mess)
					return
				}
				hS.Logger.Println(sToAdd)

				changesFlag := false
				if sToAdd != nil {
					for _, curS := range sToAdd {
						sUuid, err := uuid.NewRandom()
						if err != nil {
							mess, _ := hS.RespHandler.Handle(w, LibErrorUUID, LibErrorUUIDMessage, nil)
							hS.Logger.Print(mess)
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

	cluster.EntityStatus = utils.StatusInited
	if newC.NHosts != 0 || newHost {
		go hS.Gc.StartClusterCreation(cluster)
	} else {
		go hS.Gc.StartClusterModification(cluster)
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(cluster)
}

func (hS HttpServer) ClustersDelete(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	projectIdOrName := params.ByName("projectIdOrName")
	clusterIdOrName := params.ByName("clusterIdOrName")

	project, err := hS.Db.ReadProject(projectIdOrName)
	if err != nil {
		mess, _ := hS.RespHandler.Handle(w, DBerror, DBerrorMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	if project.Name == "" {
		hS.Logger.Printf("Project with name '%s' not found", projectIdOrName)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// reading cluster info from database
	cluster, err := hS.Db.ReadCluster(project.ID, clusterIdOrName)
	if err != nil {
		mess, _ := hS.RespHandler.Handle(w, DBerror, DBerrorMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	if cluster.Name == "" {
		hS.Logger.Printf("Cluster with name or ID '%s' not found", clusterIdOrName)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if cluster.EntityStatus != utils.StatusActive && cluster.EntityStatus != utils.StatusFailed {
		mess, _ := hS.RespHandler.Handle(w, UserErrorClusterStatus, UserErrorClusterStatusMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	cluster.EntityStatus = utils.StatusStopping

	go hS.Gc.StartClusterDestroying(cluster)
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(cluster)
}
