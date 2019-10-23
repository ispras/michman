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
	router.GET("/projects/:projectIdOrName", hS.ProjectGetByName)
	router.PUT("/projects/:projectIdOrName", hS.ProjectUpdate)
	router.DELETE("/projects/:projectIdOrName", hS.ProjectDelete)

	router.GET("/projects/:projectIdOrName/clusters", hS.ClustersGet)
	router.POST("/projects/:projectIdOrName/clusters", hS.ClusterCreate)
	router.GET("/projects/:projectIdOrName/clusters/:clusterName", hS.ClustersGetByName)
	router.PUT("/projects/:projectIdOrName/clusters/:clusterName", hS.ClustersUpdate)
	router.DELETE("/projects/:projectIdOrName/clusters/:clusterName", hS.ClustersDelete)

	router.GET("/templates", hS.TemplatesGetList)
	router.POST("/templates", hS.TemplateCreate)
	router.GET("/templates/:templateID", hS.TemplateGet)
	router.PUT("/templates/:templateID", hS.TemplateUpdate)
	router.DELETE("/templates/:templateID", hS.TemplateDelete)

	router.GET("/projects/:projectIdOrName/templates", hS.TemplatesGetList)
	router.POST("/projects/:projectIdOrName/templates", hS.TemplateCreate)
	router.GET("/projects/:projectIdOrName/templates/:templateID", hS.TemplateGet)
	router.PUT("/projects/:projectIdOrName/templates/:templateID", hS.TemplateUpdate)
	router.DELETE("/projects/:projectIdOrName/templates/:templateID", hS.TemplateDelete)

	// swagger UI route
	router.ServeFiles("/api/*filepath", http.Dir("./swaggerui"))

	httpServerLogger.Print("Server starts to work")
	httpServerLogger.Fatal(http.ListenAndServe(":8080", router))


}
