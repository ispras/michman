package main

import (
	"flag"
	"fmt"
	"github.com/alexedwards/scs/v2"
	"github.com/casbin/casbin"
	"github.com/ispras/michman/internal/auth"
	"github.com/ispras/michman/internal/database"
	"github.com/ispras/michman/internal/logger"
	"github.com/ispras/michman/internal/rest/authorization"
	grpc_client "github.com/ispras/michman/internal/rest/grpc"
	"github.com/ispras/michman/internal/rest/handlers"
	"github.com/ispras/michman/internal/utils"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	addressAnsibleService = "localhost:5000"
	restDefaultPort       = "8081"
)

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

	httpLogger := &logrus.Logger{
		Out:   mw,
		Level: logrus.InfoLevel,
		Formatter: &logger.Formatter{
			TimestampFormat: time.Stamp,
			NoFieldsColors:  true,
			ShowFullLevel:   true,
			LoggerName:      "HTTP_SERVER",
		},
	}

	grpcLogger := &logrus.Logger{
		Out:   mw,
		Level: logrus.InfoLevel,
		Formatter: &logger.Formatter{
			TimestampFormat: time.Stamp,
			NoFieldsColors:  true,
			ShowFullLevel:   true,
			LoggerName:      "GRPC_CLIENT",
		},
	}

	authorizeLogger := &logrus.Logger{
		Out:   mw,
		Level: logrus.InfoLevel,
		Formatter: &logger.Formatter{
			TimestampFormat: time.Stamp,
			NoFieldsColors:  true,
			ShowFullLevel:   true,
			LoggerName:      "AUTHORIZE_CLIENT",
		},
	}

	httpLogger.Println(fmt.Sprintf("Build version: %v", handlers.VersionID))

	//check rest port correctness
	iRestPort, err := strconv.Atoi(*restPort)
	if err != nil {
		httpLogger.Fatal(err)
	}
	if iRestPort <= 0 {
		*restPort = restDefaultPort
	}

	// setup casbin auth rules
	authEnforcer, err := casbin.NewEnforcerSafe("./configs/auth_model.conf", "./configs/policy.csv")
	if err != nil {
		httpLogger.Fatal(err)
	}

	// creating vault communicator
	vaultCommunicator := utils.VaultCommunicator{}
	vaultCommunicator.Init()

	//initialize db connection
	db, err := database.NewCouchBase(&vaultCommunicator)
	if err != nil {
		httpLogger.Println("Can't create couchbase communicator")
		os.Exit(1)
	}

	gc := grpc_client.GrpcClient{Db: db}
	gc.SetLogger(grpcLogger)
	gc.SetConnection(*launcherAddr)

	//setup session manager
	sessionManager := scs.New()
	//set session configurations
	if config.SessionIdleTimeout > 0 {
		sessionManager.IdleTimeout = time.Duration(config.SessionIdleTimeout) * time.Minute
	}
	if config.SessionLifetime > 0 {
		sessionManager.Lifetime = time.Duration(config.SessionLifetime) * time.Minute
	}

	var usedAuth auth.Authenticate
	usedAuth = auth.InitAuth(httpLogger, config.AuthorizationModel)

	router := httprouter.New()
	RespHandler := handlers.HttpResponseHandler{}

	authorizeClient := authorization.AuthorizeClient{Logger: authorizeLogger, Db: db,
		Config: config, SessionManager: sessionManager, Auth: usedAuth, Router: router}

	hS := handlers.HttpServer{Gc: gc, Logger: httpLogger, Db: db,
		RespHandler: RespHandler, Router: router, Auth: usedAuth, Config: config}

	authorizeClient.CreateRoutes()
	hS.CreateRoutes()

	httpLogger.Println("Server starts to work")
	//serve with session and authorization if authentication is used
	if config.UseAuth {
		httpLogger.Fatal(http.ListenAndServe(":"+*restPort,
			sessionManager.LoadAndSave(authorizeClient.Authorizer(authEnforcer)(router))))
	} else {
		httpLogger.Fatal(http.ListenAndServe(":"+*restPort, router))
	}
}
