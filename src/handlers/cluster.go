package handlers

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	protobuf "gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/protobuf"
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
	//check that name is unique, request to couchbase

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

	if c.ID == NEW_CLUSTER {
		hS.Logger.Print("Sending request to db-service to get cluster ID")
		newID, err := hS.Gc.GetID(&c)
		if err != nil {
			hS.Logger.Print("DB server don't ...")
		}
		//newID = 1
		c.ID = newID
	}

	c.EntityStatus = protobuf.Cluster_INITED
	go hS.Gc.StartClusterCreation(&c)

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(c)
}

func (hS HttpServer) ClustersGetList(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	hS.Logger.Print("Get /clusters GET")
	w.WriteHeader(http.StatusOK)
}

func (hS HttpServer) ClustersGet(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	clusterName := params.ByName("clusterName")
	hS.Logger.Print("Get /clusters/", clusterName, " GET")
	w.WriteHeader(http.StatusOK)
}

func (hS HttpServer) ClustersUpdate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	clusterName := params.ByName("clusterName")
	hS.Logger.Print("Get /clusters/", clusterName, " PUT")

	//check that cluster exists
	hS.Logger.Print("Sending request to db-service to check that cluster exists...")

	var c protobuf.Cluster
	err := json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		hS.Logger.Print("ERROR:")
		hS.Logger.Print(err)
		hS.Logger.Print(r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if c.EntityStatus != protobuf.Cluster_CREATED {
		hS.Logger.Print("ERROR: status of cluster to update must be 'CREATED'")
		hS.Logger.Print(r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// validate struct
	if !ValidateCluster(&c) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	c.EntityStatus = protobuf.Cluster_INITED
	go hS.Gc.StartClusterModification(&c)

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(c)
}

func (hS HttpServer) ClustersDelete(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	clusterName := params.ByName("clusterName")
	hS.Logger.Print("Get /clusters/", clusterName, " DELETE")

	//check that cluster exists
	hS.Logger.Print("Sending request to db-service to check that cluster exists...")

	//cluster for testing
	c := protobuf.Cluster{
		ID: 1,
		Name: clusterName,
		NHosts: 1,
		EntityStatus: protobuf.Cluster_CREATED,
	}

	if c.EntityStatus == protobuf.Cluster_CREATED {
		c.EntityStatus = protobuf.Cluster_STOPPING
		//send changes in status to database

		go hS.Gc.StartClusterDestroying(&c)
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.Encode(c)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		hS.Logger.Print("Error: entity status of destroying cluster must be 'CREATED")
		enc := json.NewEncoder(w)
		enc.Encode("Error: entity status of destroying cluster must be 'CREATED")
		return
	}

}