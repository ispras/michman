package main

import (
	"flag"
	"fmt"
	"github.com/ispras/michman/database"
	"github.com/ispras/michman/utils"
	"io"
	"log"
	"net/http"
	"os"

	grpc_client "github.com/ispras/michman/rest/grpc"
	"github.com/ispras/michman/rest/handlers"
	"github.com/julienschmidt/httprouter"
)

const (
	addressAnsibleService = "localhost:5000"
	addressDBService      = "localhost:5001"
	restDefaultPort			  = "8081"
)

var VersionID string = "Default"

func main() {
	fmt.Printf("Build version: %v\n", VersionID)

	//set flags for config path and ansible service adress
	configPath := flag.String("config", utils.ConfigPath, "Path to the config.yaml file")
	launcherAddr := flag.String("launcher", addressAnsibleService, "Launcher service address")
	restPort := flag.String("port", restDefaultPort, "Rest service port")
	flag.Parse()

	//set config file path
	utils.SetConfigPath(*configPath)

	// creating grpc client for communicating with services
	grpcClientLogger := log.New(os.Stdout, "GRPC_CLIENT: ", log.Ldate|log.Ltime)
	vaultCommunicator := utils.VaultCommunicator{}
	vaultCommunicator.Init()
	db, err := database.NewCouchBase(&vaultCommunicator)
	if err != nil {
		fmt.Println("Can't create couchbase communicator")
		os.Exit(1)
	}
	gc := grpc_client.GrpcClient{Db: db}
	gc.SetLogger(grpcClientLogger)
	gc.SetConnection(*launcherAddr)

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

	// Routes for Configs module
	router.POST("/configs", hS.ConfigsCreateService)
	router.GET("/configs", hS.ConfigsGetServices)
	router.GET("/configs/:serviceType", hS.ConfigsGetService)
	router.PUT("/configs/:serviceType", hS.ConfigsUpdateService)
	router.DELETE("/configs/:serviceType", hS.ConfigsDeleteService)
	router.GET("/configs/:serviceType/versions", hS.ConfigsGetVersions)
	router.POST("/configs/:serviceType/versions", hS.ConfigsCreateVersion)
	router.GET("/configs/:serviceType/versions/:versionId", hS.ConfigsGetVersion)
	router.PUT("/configs/:serviceType/versions/:versionId", hS.ConfigsUpdateVersion)
	router.DELETE("/configs/:serviceType/versions/:versionId", hS.ConfigsDeleteVersion)
	router.POST("/configs/:serviceType/versions/:versionId/configs", hS.ConfigsCreateConfigParam)

	// swagger UI route
	router.ServeFiles("/api/*filepath", http.Dir("./rest/swaggerui"))

	// logs routes
	router.Handle("GET", "/logs/ansible_output", hS.ServeAnsibleOutput)
	router.Handle("GET", "/logs/launcher", hS.ServeAnsibleServiceLog)
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

	router.Handle("GET", "/images", hS.ImagesGetList)
	router.Handle("GET", "/images/:imageName", hS.ImageGet)
	router.Handle("POST", "/images", hS.ImagesPost)
	router.Handle("PUT", "/images/:imageName", hS.ImagePut)
	router.Handle("DELETE", "/images/:imageName", hS.ImageDelete)

	httpServerLogger.Print("Server starts to work")
	httpServerLogger.Fatal(http.ListenAndServe(*restPort, router))

}
