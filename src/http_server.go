package main

import (
	"gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/database"
	"gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/utils"
	"log"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	grpc_client "gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/grpcclients"
	"gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/handlers"
)

const (
	addressAnsibleService = "localhost:5000"
	addressDBService      = "localhost:5001"
)

func main() {
	// creating grpc client for communicating with services
	grpcClientLogger := log.New(os.Stdout, "GRPC_CLIENT: ", log.Ldate|log.Ltime)
	vaultCommunicator := utils.VaultCommunicator{}
	gc := grpc_client.GrpcClient{Db: database.CouchDatabase{VaultCommunicator: &vaultCommunicator}}
	gc.SetLogger(grpcClientLogger)
	gc.SetConnection(addressAnsibleService, addressDBService)

	httpServerLogger := log.New(os.Stdout, "HTTP_SERVER: ", log.Ldate|log.Ltime)

	hS := handlers.HttpServer{Gc: gc, Logger: httpServerLogger, Db: database.CouchDatabase{VaultCommunicator: &vaultCommunicator}}

	router := httprouter.New()

	router.GET("/projects", hS.ProjectsGetList)
	router.POST("/projects", hS.ProjectCreate)
	router.GET("/projects/:projectName", hS.ProjectGetByName)
	router.PUT("/projects/:projectName", hS.ProjectUpdate)
	router.DELETE("/projects/:projectName", hS.ProjectDelete)

	router.GET("/projects/:projectName/clusters", hS.ClustersGet)
	router.POST("/projects/:projectName/clusters", hS.ClusterCreate)
	router.GET("/projects/:projectName/clusters/:clusterName", hS.ClustersGetByName)
	router.PUT("/projects/:projectName/clusters/:clusterName", hS.ClustersUpdate)
	router.DELETE("/projects/:projectName/clusters/:clusterName", hS.ClustersDelete)

	router.GET("/templates", hS.TemplatesGetList)
	router.POST("/templates", hS.TemplateCreate)
	router.GET("/templates/:templateID", hS.TemplateGet)
	router.PUT("/templates/:templateID", hS.TemplateUpdate)
	router.DELETE("/templates/:templateID", hS.TemplateDelete)

	router.GET("/projects/:projectID/templates", hS.TemplatesGetList)
	router.POST("/projects/:projectID/templates", hS.TemplateCreate)
	router.GET("/projects/:projectID/templates/:templateID", hS.TemplateGet)
	router.PUT("/projects/:projectID/templates/:templateID", hS.TemplateUpdate)
	router.DELETE("/projects/:projectID/templates/:templateID", hS.TemplateDelete)

	httpServerLogger.Print("Server starts to work")
	httpServerLogger.Fatal(http.ListenAndServe(":8080", router))
}
