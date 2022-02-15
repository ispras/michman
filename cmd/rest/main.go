package main

import (
	"encoding/json"
	"flag"
	"github.com/alexedwards/scs/v2"
	"github.com/casbin/casbin"
	auth "github.com/ispras/michman/internal/auth"
	"github.com/ispras/michman/internal/database"
	"github.com/ispras/michman/internal/rest/authorization"
	grpc_client "github.com/ispras/michman/internal/rest/grpc"
	"github.com/ispras/michman/internal/rest/handlers"
	"github.com/ispras/michman/internal/utils"
	"github.com/julienschmidt/httprouter"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	addressAnsibleService = "localhost:5000"
	restDefaultPort       = "8081"
)

var (
	VersionID        string = "Default"
	sessionManager   *scs.SessionManager
	httpServerLogger *log.Logger
)

func initAuth(authMode string) auth.Authenticate {
	switch authMode {
	case utils.OAuth2Mode:
		hydraAuth, err := auth.NewHydraAuthenticate()
		if err != nil {
			httpServerLogger.Println("Can't create new authenticator")
			os.Exit(1)
		}
		return hydraAuth
	case utils.KeystoneMode:
		keystoneAuth, err := auth.NewKeystoneAuthenticate()
		if err != nil {
			httpServerLogger.Println("Can't create new authenticator")
			os.Exit(1)
		}
		return keystoneAuth
	case utils.NoneAuthMode:
		noneAuth, err := auth.NewNoneAuthenticate()
		if err != nil {
			httpServerLogger.Println("Can't create new authenticator")
			os.Exit(1)
		}
		return noneAuth
	}
	return nil
}

func main() {
	//set flags for config path and ansible service adress
	configPath := flag.String("config", utils.ConfigPath, "Path to the config.yaml file")
	launcherAddr := flag.String("launcher", addressAnsibleService, "Launcher service address")
	restPort := flag.String("port", restDefaultPort, "Rest service port")
	flag.Parse()

	//set config file path
	utils.SetConfigPath(*configPath)
	// create a multiwriter which writes to stdout and a file simultaneously
	config := utils.Config{}
	err := config.MakeCfg()
	if err != nil {
		panic(err)
	}
	logFile, err := os.OpenFile(config.LogsFilePath+"/http_server.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	mw := io.MultiWriter(os.Stdout, logFile)

	httpServerLogger = log.New(mw, "HTTP_SERVER: ", log.Ldate|log.Ltime)

	httpServerLogger.Printf("Build version: %v\n", VersionID)

	//check rest port correctness
	iRestPort, err := strconv.Atoi(*restPort)
	if err != nil {
		httpServerLogger.Fatal(err)
	}
	if iRestPort <= 0 {
		*restPort = restDefaultPort
	}

	// setup casbin auth rules
	authEnforcer, err := casbin.NewEnforcerSafe("./configs/auth_model.conf", "./configs/policy.csv")
	if err != nil {
		httpServerLogger.Fatal(err)
	}

	// creating grpc client for communicating with services
	grpcClientLogger := log.New(os.Stdout, "GRPC_CLIENT: ", log.Ldate|log.Ltime)

	// creating vault communicator
	vaultCommunicator := utils.VaultCommunicator{}
	vaultCommunicator.Init()

	//initialize db connection
	db, err := database.NewCouchBase(&vaultCommunicator)
	if err != nil {
		httpServerLogger.Println("Can't create couchbase communicator")
		os.Exit(1)
	}
	gc := grpc_client.GrpcClient{Db: db}
	gc.SetLogger(grpcClientLogger)
	gc.SetConnection(*launcherAddr)

	//setup session manager
	sessionManager = scs.New()
	//set session configurations
	if config.SessionIdleTimeout > 0 {
		sessionManager.IdleTimeout = time.Duration(config.SessionIdleTimeout) * time.Minute
	}
	if config.SessionLifetime > 0 {
		sessionManager.Lifetime = time.Duration(config.SessionLifetime) * time.Minute
	}

	//setup authorize client
	authorizeClientLogger := log.New(os.Stdout, "AUTHORIZE_CLIENT: ", log.Ldate|log.Ltime)
	authorizeClient := authorization.AuthorizeClient{Logger: authorizeClientLogger, Db: db,
		Config: config, SessionManager: sessionManager}

	var usedAuth auth.Authenticate
	usedAuth = initAuth(config.AuthorizationModel)

	errHandler := handlers.HttpErrorHandler{}

	hS := handlers.HttpServer{Gc: gc, Logger: httpServerLogger, Db: db,
		ErrHandler: errHandler, Auth: usedAuth, Config: config}

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
	router.ServeFiles("/api/*filepath", http.Dir("./api/rest"))

	// logs routes
	router.Handle("GET", "/logs/launcher", hS.ServeAnsibleServiceLog)
	router.Handle("GET", "/logs/http_server", hS.ServeHttpServerLog)
	router.Handle("GET", "/logs/projects/:projectIdOrName/clusters/:clusterID", hS.ServeHttpServerLogstash)

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

	//auth route
	router.Handle("GET", "/auth", func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		//set auth facts
		w, err := hS.Auth.SetAuth(sessionManager, w, r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			hS.Logger.Println(err)
			return
		}

		g := sessionManager.GetString(r.Context(), utils.GroupKey)

		hS.Logger.Println("Authentication success!")
		hS.Logger.Println("User groups are: " + g)

		var userGroups string
		if g == "" {
			userGroups = "You are not a member of any group."
		} else {
			userGroups = "You are a member of the following groups: " + g
		}

		enc := json.NewEncoder(w)
		err = enc.Encode("Authentication success! " + userGroups)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			hS.Logger.Print(err)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	httpServerLogger.Print("Server starts to work")
	//serve with session and authorization if authentication is used
	if config.UseAuth {
		httpServerLogger.Fatal(http.ListenAndServe(":"+*restPort,
			sessionManager.LoadAndSave(authorizeClient.Authorizer(authEnforcer)(router))))
	} else {
		httpServerLogger.Fatal(http.ListenAndServe(":"+*restPort, router))
	}
}
