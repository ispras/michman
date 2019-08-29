package handlers

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/database"
	protobuf "gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/protobuf"
	"gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/utils"
	"log"
	"net/http"
	"regexp"
)

const (
	NEW_CLUSTER = -1
)

type GrpcClient interface {
	GetID(c *protobuf.Cluster) (int32, error)
	StartClusterCreation(c *protobuf.Cluster)
	StartClusterDestroying(c *protobuf.Cluster)
	StartClusterModification(c *protobuf.Cluster)
}

type HttpServer struct {
	Gc     GrpcClient
	Logger *log.Logger
}

func ValidateCluster(cluster *protobuf.Cluster) bool {
	validName := regexp.MustCompile(`^[A-Za-z][A-Za-z0-9-]+$`).MatchString

	if !validName(cluster.Name) {
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

func (hS HttpServer) ClusterCreate(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	hS.Logger.Print("Get /clusters POST")
	var c protobuf.Cluster
	err := json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		hS.Logger.Print("ERROR:")
		hS.Logger.Print(err)
		hS.Logger.Print(r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// validate struct
	if !ValidateCluster(&c) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//check, that cluster with such name doesn't exist
	db := database.CouchDatabase{}
	dbRes, err := db.ReadCluster(c.Name)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if dbRes.Name != "" && dbRes.EntityStatus != utils.StatusFailed {
		hS.Logger.Print("Cluster with this name exists")
		w.WriteHeader(http.StatusBadRequest)
		enc := json.NewEncoder(w)
		err := enc.Encode("Cluster with this name exists")
		if err != nil {
			hS.Logger.Print(err)
			w.WriteHeader(http.StatusBadRequest)
		}
		return
	}

	if c.ID == NEW_CLUSTER {
		hS.Logger.Print("Sending request to db-service to get cluster ID")
		newID, err := hS.Gc.GetID(&c)
		if err != nil {
			hS.Logger.Print("DB server don't ...")
		}
		//newID = 1
		c.ID = newID
	}

	c.EntityStatus = utils.StatusInited
	go hS.Gc.StartClusterCreation(&c)

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(c)
}

func (hS HttpServer) ClustersGetList(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	hS.Logger.Print("Get /clusters GET")
	//reading cluster info from database
	hS.Logger.Print("Reading cluster information from db...")
	db := database.CouchDatabase{}
	clusters, err := db.ListClusters()
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	err = enc.Encode(clusters)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (hS HttpServer) ClustersGet(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	clusterName := params.ByName("clusterName")
	hS.Logger.Print("Get /clusters/", clusterName, " GET")

	//reading cluster info from database
	hS.Logger.Print("Reading cluster information from db...")
	db := database.CouchDatabase{}
	cluster, err := db.ReadCluster(clusterName)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if cluster.Name == "" {
		hS.Logger.Print("Cluster not found")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	err = enc.Encode(cluster)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (hS HttpServer) ClustersUpdate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	clusterName := params.ByName("clusterName")
	hS.Logger.Print("Get /clusters/", clusterName, " PUT")

	//check that cluster exists
	hS.Logger.Print("Sending request to db-service to check that cluster exists...")

	db := database.CouchDatabase{}
	oldC, err := db.ReadCluster(clusterName)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if oldC.Name == "" {
		hS.Logger.Print("Cluster not found")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if oldC.EntityStatus != utils.StatusCreated && oldC.EntityStatus != utils.StatusFailed {
		hS.Logger.Print("ERROR: status of cluster to update must be 'CREATED'")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// validate request struct
	var newC protobuf.Cluster
	err = json.NewDecoder(r.Body).Decode(&newC)
	if err != nil {
		hS.Logger.Print("ERROR:")
		hS.Logger.Print(err)
		hS.Logger.Print(r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !ValidateCluster(&newC) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	newC.EntityStatus = utils.StatusInited
	go hS.Gc.StartClusterModification(&newC)

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(newC)
}

func (hS HttpServer) ClustersDelete(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	clusterName := params.ByName("clusterName")
	hS.Logger.Print("Get /clusters/", clusterName, " DELETE")

	//check that cluster exists
	hS.Logger.Print("Sending request to db-service to check that cluster exists...")

	//cluster for testing
	db := database.CouchDatabase{}
	c, err := db.ReadCluster(clusterName)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if c.Name == "" {
		hS.Logger.Print("Cluster not found")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if c.EntityStatus != utils.StatusCreated && c.EntityStatus != utils.StatusFailed {
		hS.Logger.Print("ERROR: status of cluster to update must be 'CREATED'")
		hS.Logger.Print(r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	c.EntityStatus = utils.StatusStopping

	go hS.Gc.StartClusterDestroying(c)
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(c)
}