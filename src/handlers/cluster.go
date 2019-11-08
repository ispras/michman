package handlers

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/database"
	proto "gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/protobuf"
	"gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/utils"
	"log"
	"net/http"
	"regexp"
)

const (
	NEW_CLUSTER         = -1
	CLUSTER_DIDNT_EXIST = -2
)

type GrpcClient interface {
	GetID(c *proto.Cluster) (int32, error)
	StartClusterCreation(c *proto.Cluster)
	StartClusterDestroying(c *proto.Cluster)
	StartClusterModification(c *proto.Cluster)
}

type HttpServer struct {
	Gc         GrpcClient
	Logger     *log.Logger
	Db         database.Database
	ErrHandler ErrorHandler
}

type serviceExists struct {
	exists  bool
	service *proto.Service
}

func ValidateCluster(cluster *proto.Cluster) bool {
	validName := regexp.MustCompile(`^[A-Za-z][A-Za-z0-9-]+$`).MatchString

	if !validName(cluster.DisplayName) {
		log.Print("ERROR: bad name for cluster. You should use only alpha-numeric characters and '-' symbols and only alphabetic characters for leading symbol.")
		return false
	}

	if cluster.NHosts < 1 {
		log.Print(cluster.NHosts)
		log.Print("ERROR: NHosts parameter must be number >= 1.")
		return false
	}

	for _, service := range cluster.Services {
		if !ValidateService(service) {
			return false
		}
	}

	return true
}

func (hS HttpServer) getCluster(projectID, idORname string) (*proto.Cluster, error) {
	is_uuid := true
	_, err := uuid.Parse(idORname)
	if err != nil {
		is_uuid = false
	}

	var cluster *proto.Cluster

	if is_uuid {
		cluster, err = hS.Db.ReadCluster(idORname)
	} else {
		cluster, err = hS.Db.ReadClusterByName(projectID, idORname)
	}

	return cluster, err
}

func (hS HttpServer) ClustersGet(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	projectIdOrName := params.ByName("projectIdOrName")
	hS.Logger.Print("Get /projects/", projectIdOrName, "/clusters GET")

	project, err := hS.getProject(projectIdOrName)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, DBerror, DBerrorMessage, err)
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
		mess, _ := hS.ErrHandler.Handle(w, DBerror, DBerrorMessage, err)
		hS.Logger.Print(mess)
		return
	}

	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	err = enc.Encode(clusters)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, LibErrorStructToJson, LibErrorStructToJsonMessage, err)
		hS.Logger.Print(mess)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

func (hS HttpServer) ClusterCreate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	projectIdOrName := params.ByName("projectIdOrName")
	hS.Logger.Print("Get /project/" + projectIdOrName + "/clusters POST")

	project, err := hS.getProject(projectIdOrName)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, DBerror, DBerrorMessage, err)
		hS.Logger.Print(mess)
		return
	}

	if project.Name == "" {
		hS.Logger.Printf("Project with name '%s' not found", projectIdOrName)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var c *proto.Cluster
	err = json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, JSONerrorIncorrect, JSONerrorIncorrectMessage, err)
		hS.Logger.Print(mess)
		return
	}

	// validate struct
	if !ValidateCluster(c) {
		mess, _ := hS.ErrHandler.Handle(w, JSONerrorIncorrectField, JSONerrorIncorrectFieldMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	//check, that cluster with such name doesn't exist
	searchedName := c.DisplayName + "-" + project.Name
	cluster, err := hS.Db.ReadClusterByName(project.ID, searchedName)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, DBerror, DBerrorMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	clusterExists := false

	if cluster.Name != "" {
		clusterExists = true
		if cluster.EntityStatus != utils.StatusFailed {
			mess, _ := hS.ErrHandler.Handle(w, UserErrorClusterExisted, UserErrorClusterExistedMessage, nil)
			hS.Logger.Print(mess)
			return
		}
	}

	// If cluster was failed
	if clusterExists {
		c = cluster
	} else {
		cUuid, err := uuid.NewRandom()
		if err != nil {
			mess, _ := hS.ErrHandler.Handle(w, LibErrorUUID, LibErrorUUIDMessage, nil)
			hS.Logger.Print(mess)
			return
		}
		c.ID = cUuid.String()
		for _, s := range c.Services {
			sUuid, err := uuid.NewRandom()
			if err != nil {
				mess, _ := hS.ErrHandler.Handle(w, LibErrorUUID, LibErrorUUIDMessage, nil)
				hS.Logger.Print(mess)
				return
			}
			s.ID = sUuid.String()
		}

		c.ProjectID = project.ID
		c.Name = c.DisplayName + "-" + project.Name
	}

	c.EntityStatus = utils.StatusInited
	if !clusterExists {
		err = hS.Db.WriteCluster(c)
		if err != nil {
			mess, _ := hS.ErrHandler.Handle(w, DBerror, DBerrorMessage, nil)
			hS.Logger.Print(mess)
			return
		}
	}
	go hS.Gc.StartClusterCreation(c)

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(c)
}

func (hS HttpServer) ClustersGetByName(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	projectIdOrName := params.ByName("projectIdOrName")
	clusterIdOrName := params.ByName("clusterIdOrName")
	hS.Logger.Print("Get /projects/"+projectIdOrName+"/clusters/", clusterIdOrName, " GET")

	project, err := hS.getProject(projectIdOrName)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, DBerror, DBerrorMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	if project.Name == "" {
		hS.Logger.Printf("Project with name '%s' not found", projectIdOrName)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// reading cluster info from database
	cluster, err := hS.getCluster(project.ID, clusterIdOrName)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, DBerror, DBerrorMessage, nil)
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
		mess, _ := hS.ErrHandler.Handle(w, LibErrorStructToJson, LibErrorStructToJsonMessage, nil)
		hS.Logger.Print(mess)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

func (hS HttpServer) ClustersStatusGetByName(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	projectIdOrName := params.ByName("projectIdOrName")
	clusterIdOrName := params.ByName("clusterIdOrName")
	hS.Logger.Print("Get /projects/"+projectIdOrName+"/clusters/", clusterIdOrName, "/status", " GET")

	project, err := hS.getProject(projectIdOrName)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, DBerror, DBerrorMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	if project.Name == "" {
		hS.Logger.Printf("Project with name '%s' not found", projectIdOrName)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// reading cluster info from database
	cluster, err := hS.getCluster(project.ID, clusterIdOrName)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, DBerror, DBerrorMessage, nil)
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

	project, err := hS.getProject(projectIdOrName)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, DBerror, DBerrorMessage, nil)
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
	cluster, err := hS.getCluster(project.ID, clusterIdOrName)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, DBerror, DBerrorMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	if cluster.Name == "" {
		hS.Logger.Printf("Cluster with name or ID '%s' not found", clusterIdOrName)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if cluster.EntityStatus != utils.StatusActive && cluster.EntityStatus != utils.StatusFailed {
		mess, _ := hS.ErrHandler.Handle(w, UserErrorClusterStatus, UserErrorClusterStatusMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	// validate request struct
	var newC proto.Cluster
	err = json.NewDecoder(r.Body).Decode(&newC)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, JSONerrorIncorrect, JSONerrorIncorrectMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	if newC.ID != "" || newC.Name != "" || newC.DisplayName != "" || newC.EntityStatus != "" ||
		newC.HostURL != "" || newC.MasterIP != "" || newC.ProjectID != "" {
		mess, _ := hS.ErrHandler.Handle(w, UserErrorProjectUnmodField, UserErrorProjectUnmodFieldMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	//appending old services which does not exist in new cluster configuration
	var serviceTypesOld = map[string]serviceExists{
		utils.ServiceTypeCassandra: {
			exists:  false,
			service: nil,
		},
		utils.ServiceTypeSpark: {
			exists:  false,
			service: nil,
		},
		utils.ServiceTypeElastic: {
			exists:  false,
			service: nil,
		},
		utils.ServiceTypeJupyter: {
			exists:  false,
			service: nil,
		},
		utils.ServiceTypeIgnite: {
			exists:  false,
			service: nil,
		},
		utils.ServiceTypeJupyterhub: {
			exists:  false,
			service: nil,
		},
		utils.ServiceTypeFanlight: {
			exists:  false,
			service: nil,
		},
	}

	for _, s := range cluster.Services {
		sUuid, err := uuid.NewRandom()
		if err != nil {
			mess, _ := hS.ErrHandler.Handle(w, LibErrorUUID, LibErrorUUIDMessage, nil)
			hS.Logger.Print(mess)
			return
		}
		s.ID = sUuid.String()
		serviceTypesOld[s.Type] = serviceExists{
			exists:  true,
			service: s,
		}
	}

	for _, s := range newC.Services {
		if serviceTypesOld[s.Type].exists == false {
			cluster.Services = append(cluster.Services, s)
		}
	}

	if newC.Description != "" {
		cluster.Description = newC.Description
	}

	newC.EntityStatus = utils.StatusInited
	go hS.Gc.StartClusterModification(cluster)

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(newC)
}

func (hS HttpServer) ClustersDelete(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	projectIdOrName := params.ByName("projectIdOrName")
	clusterIdOrName := params.ByName("clusterIdOrName")
	hS.Logger.Print("Get /projects/"+projectIdOrName+"/clusters/", clusterIdOrName, " DELETE")

	project, err := hS.getProject(projectIdOrName)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, DBerror, DBerrorMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	if project.Name == "" {
		hS.Logger.Printf("Project with name '%s' not found", projectIdOrName)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// reading cluster info from database
	cluster, err := hS.getCluster(project.ID, clusterIdOrName)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, DBerror, DBerrorMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	if cluster.Name == "" {
		hS.Logger.Printf("Cluster with name or ID '%s' not found", clusterIdOrName)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if cluster.EntityStatus != utils.StatusActive && cluster.EntityStatus != utils.StatusFailed {
		mess, _ := hS.ErrHandler.Handle(w, UserErrorClusterStatus, UserErrorClusterStatusMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	cluster.EntityStatus = utils.StatusStopping

	go hS.Gc.StartClusterDestroying(cluster)
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(cluster)
}
