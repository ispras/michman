package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	grpc_client "gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/grpcclients"
	protobuf "gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/protobuf"
)

const (
	addressAnsibleService = "localhost:5000"
	addressDBService      = "localhost:5001"
	NEW_CLUSTER           = -1
)

type grpcClient interface {
	GetID(c *protobuf.Cluster) (int32, error)
	StartClusterCreation(c *protobuf.Cluster)
}

type httpServer struct {
	gc     grpcClient
	logger *log.Logger
}

func (hS httpServer) clustersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		hS.logger.Print("Get /clusters POST")
		w.WriteHeader(http.StatusOK)

	case "POST":
		hS.logger.Print("Get /clusters POST")
		var c protobuf.Cluster
		err := json.NewDecoder(r.Body).Decode(&c)
		if err != nil {
			hS.logger.Print("ERROR:")
			hS.logger.Print(err)
			hS.logger.Print(r.Body)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// validate struct?

		if c.ID == NEW_CLUSTER {
			hS.logger.Print("Sending request to db-service to get cluster ID")
			newID, err := hS.gc.GetID(&c)
			if err != nil {
				hS.logger.Print("DB server don't ...")
			}
			//newID = 1
			c.ID = newID
		}

		c.EntityStatus = protobuf.Cluster_INITED
		go hS.gc.StartClusterCreation(&c)

		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.Encode(c)

	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func main() {
	// creating grpc client for communicating with services
	grpcClientLogger := log.New(os.Stdout, "GRPC_CLIENT: ", log.Ldate|log.Ltime)
	gc := grpc_client.GrpcClient{}
	gc.SetLogger(grpcClientLogger)
	gc.SetConnection(addressAnsibleService, addressDBService)

	httpServerLogger := log.New(os.Stdout, "HTTP_SERVER: ", log.Ldate|log.Ltime)
	hS := httpServer{gc, httpServerLogger}

	http.HandleFunc("/clusters", hS.clustersHandler)
	httpServerLogger.Print("Server starts to work")
	httpServerLogger.Fatal(http.ListenAndServe(":8080", nil))
}
