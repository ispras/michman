package main

import (
	"fmt"
	"gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/database"
	"gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/utils"
	"io"
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

var VersionID string = "Default"

func main() {
	fmt.Printf("Build version: %v\n", VersionID)
	// creating grpc client for communicating with services
	grpcClientLogger := log.New(os.Stdout, "GRPC_CLIENT: ", log.Ldate|log.Ltime)
	vaultCommunicator := utils.VaultCommunicator{}
	db, err := database.NewCouchBase(&vaultCommunicator)
	if err != nil {
		fmt.Printf("Can't create couchbase communicator")
		os.Exit(1)
	}
	gc := grpc_client.GrpcClient{Db: db}
	gc.SetLogger(grpcClientLogger)
	gc.SetConnection(addressAnsibleService)

	// create a multiwriter which writes to stout and a file simultaneously
	logFile, err := os.OpenFile("logs/http_server.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("Can't create a log file. Exit...")
		os.Exit(1)
	}
	mw := io.MultiWriter(os.Stdout, logFile)

	httpServerLogger := log.New(mw, "HTTP_SERVER: ", log.Ldate|log.Ltime)

	errHandler := handlers.HttpErrorHandler{}


	hS := handlers.HttpServer{Gc: gc, Logger: httpServerLogger, Db: db,
		ErrHandler: errHandler}

	router := httprouter.New()

	router.GET("/projects", hS.ProjectsGetList)
	router.POST("/projects", hS.ProjectCreate)
	router.GET("/projects/:projectIdOrName", hS.ProjectGetByName)
	router.PUT("/projects/:projectIdOrName", hS.ProjectUpdate)
	router.DELETE("/projects/:projectIdOrName", hS.ProjectDelete)

	router.GET("/projects/:projectIdOrName/clusters", hS.ClustersGet)
	router.POST("/projects/:projectIdOrName/clusters", hS.ClusterCreate)
	router.GET("/projects/:projectIdOrName/clusters/:clusterIdOrName", hS.ClustersGetByName)
	router.GET("/projects/:projectIdOrName/clusters/:clusterIdOrName/status", hS.ClustersStatusGetByName)
	router.PUT("/projects/:projectIdOrName/clusters/:clusterIdOrName", hS.ClustersUpdate)
	router.DELETE("/projects/:projectIdOrName/clusters/:clusterIdOrName", hS.ClustersDelete)

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

	// logs routes
	router.Handle("GET", "/logs/ansible_output", hS.ServeAnsibleOutput)
	router.Handle("GET", "/logs/ansible_service", hS.ServeAnsibleServiceLog)
	router.Handle("GET", "/logs/http_server", hS.ServeHttpServerLog)

	// version of service
	router.Handle("GET", "/version", func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(VersionID))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			hS.Logger.Println(err)
		}
	})

	httpServerLogger.Print("Server starts to work")
	httpServerLogger.Fatal(http.ListenAndServe(":8080", router))

}
