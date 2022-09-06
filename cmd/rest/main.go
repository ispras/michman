package main

import (
	"flag"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/casbin/casbin"
	"github.com/ispras/michman/cmd"
	"github.com/ispras/michman/internal/auth"
	"github.com/ispras/michman/internal/database"
	"github.com/ispras/michman/internal/logger"
	"github.com/ispras/michman/internal/rest/authorization"
	grpc_client "github.com/ispras/michman/internal/rest/grpc"
	"github.com/ispras/michman/internal/rest/handler"
	"github.com/ispras/michman/internal/utils"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

const (
	addressAnsibleService = "localhost:5000"
	restDefaultPort       = "8081"
)

func main() {
	//set flags for config path and ansible service address
	configPath := flag.String("config", utils.ConfigPath, "Path to the config.yaml file")
	launcherAddr := flag.String("launcher", addressAnsibleService, "Launcher service address")
	restPort := flag.String("port", restDefaultPort, "Rest service port")
	flag.Parse()

	//set config file path
	utils.SetConfigPath(*configPath)

	// create a multi-writer which writes to stdout and a file simultaneously
	config := utils.Config{}
	err := config.MakeCfg()
	if err != nil {
		panic(err)
	}

	logFile, err := os.OpenFile(config.LogsFilePath+"/http_server.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(cmd.ErrOpenFile)
	}
	mw := io.MultiWriter(os.Stdout, logFile)

	httpLogger := &logrus.Logger{
		Out:   mw,
		Level: logrus.InfoLevel,
		Formatter: &logger.Formatter{
			TimestampFormat: time.Stamp,
			NoColors:        true,
			NoFieldsColors:  true,
			ShowFullLevel:   false,
			LoggerName:      "HTTP_SERVER",
		},
	}

	grpcLogger := &logrus.Logger{
		Out:   mw,
		Level: logrus.InfoLevel,
		Formatter: &logger.Formatter{
			TimestampFormat: time.Stamp,
			NoColors:        true,
			NoFieldsColors:  true,
			ShowFullLevel:   false,
			LoggerName:      "GRPC_CLIENT",
		},
	}

	authorizeLogger := &logrus.Logger{
		Out:   mw,
		Level: logrus.InfoLevel,
		Formatter: &logger.Formatter{
			TimestampFormat: time.Stamp,
			NoColors:        true,
			NoFieldsColors:  true,
			ShowFullLevel:   false,
			LoggerName:      "AUTH_CLIENT",
		},
	}
	httpLogger.Infof("Build version: %v", handler.VersionID)

	//check rest port correctness
	iRestPort, err := strconv.Atoi(*restPort)
	if err != nil {
		httpLogger.SetOutput(os.Stderr)
		httpLogger.Fatal(cmd.ErrAtoi)
	}
	if iRestPort <= 0 {
		*restPort = restDefaultPort
	}

	var authEnforcer *casbin.Enforcer
	if config.UseAuth {
		exists, err := utils.FileExists(config.AuthConfigPath)
		if !exists {
			httpLogger.SetOutput(os.Stderr)
			httpLogger.Fatal(cmd.ErrAuthModelNotExists)
		}

		exists, err = utils.FileExists(config.PolicyPath)
		if !exists {
			httpLogger.SetOutput(os.Stderr)
			httpLogger.Fatal(cmd.ErrPolicyNotExists)
		}

		// setup casbin auth rules
		authEnforcer, err = casbin.NewEnforcerSafe(config.AuthConfigPath, config.PolicyPath)
		if err != nil {
			httpLogger.SetOutput(os.Stderr)
			httpLogger.Fatal(cmd.ErrNewEnforcerSafe)
		}
	}

	// creating vault communicator
	vaultCommunicator := utils.VaultCommunicator{}
	err = vaultCommunicator.Init()
	if err != nil {
		httpLogger.SetOutput(os.Stderr)
		httpLogger.Fatal(err)
	}

	//initialize db connection
	var db database.Database
	if config.Storage == utils.StorageMySQL {
		db, err = database.NewMySQL(&vaultCommunicator)
		if err != nil {
			httpLogger.Fatal(err)
		}
	} else {
		db, err = database.NewCouchBase(&vaultCommunicator)
		if err != nil {
			httpLogger.Fatal(err)
		}
	}

	gc := grpc_client.GrpcClient{Db: db}
	gc.SetLogger(grpcLogger)
	err = gc.SetConnection(*launcherAddr)
	if err != nil {
		httpLogger.SetOutput(os.Stderr)
		grpcLogger.Fatal(err)
	}

	//setup session manager
	sessionManager := scs.New()
	//set session configurations
	if config.SessionIdleTimeout > 0 {
		sessionManager.IdleTimeout = time.Duration(config.SessionIdleTimeout) * time.Minute
	}
	if config.SessionLifetime > 0 {
		sessionManager.Lifetime = time.Duration(config.SessionLifetime) * time.Minute
	}

	router := httprouter.New()

	httpLogger.Info("Server starts to work")

	//serve with session and authorization if authentication is used
	if config.UseAuth {
		var usedAuth auth.Authenticate
		usedAuth, err = auth.InitAuth(config.AuthorizationModel)
		if err != nil {
			httpLogger.SetOutput(os.Stderr)
			httpLogger.Fatal(err)
		}
		authorizeClient := authorization.AuthorizeClient{Logger: authorizeLogger, Db: db,
			Config: config, SessionManager: sessionManager, Auth: usedAuth, Router: router}

		authorizeClient.CreateRoutes()
		err = http.ListenAndServe(":"+*restPort, sessionManager.LoadAndSave(authorizeClient.Authorizer(authEnforcer)(router)))

		httpLogger.SetOutput(os.Stderr)
		httpLogger.Fatal(err)
	} else {
		hS := handler.HttpServer{Gc: gc, Logger: httpLogger, Db: db, Router: router, Config: config}
		hS.CreateRoutes()
		err = http.ListenAndServe(":"+*restPort, router)
		httpLogger.SetOutput(os.Stderr)
		httpLogger.Fatal(err)
	}
}
