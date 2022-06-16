package main

import (
	"flag"
	"github.com/ispras/michman/internal/ansible"
	"github.com/ispras/michman/internal/database"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/utils"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
	"os"
)

func main() {
	//set flags for config path and ansible service adress
	configPath := flag.String("config", utils.ConfigPath, "Path to the config.yaml file")
	launcherPort := flag.String("port", ansible.LauncherDefaultPort, "Launcher service default port")
	flag.Parse()

	//set config file path
	utils.SetConfigPath(*configPath)
	config := utils.Config{}
	if err := config.MakeCfg(); err != nil {
		panic(err)
	}
	logFile, err := os.OpenFile(config.LogsFilePath+"/launcher.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	mw := io.MultiWriter(os.Stdout, logFile)

	LauncherLogger := log.New(mw, "LAUNCHER: ", log.Ldate|log.Ltime)

	vaultCommunicator := utils.VaultCommunicator{}
	err = vaultCommunicator.Init()
	if err != nil {
		panic(err)
	}
	db, err := database.NewCouchBase(&vaultCommunicator)
	if err != nil {
		panic("Can't create database connection. Exit...")
	}
	lis, err := net.Listen("tcp", ":"+*launcherPort)
	if err != nil {
		LauncherLogger.Fatalf("failed to listen: %v", err)
	}

	gas := grpc.NewServer()
	aService := ansible.LauncherServer{Logger: LauncherLogger, Db: db,
		VaultCommunicator: &vaultCommunicator, Config: config}

	protobuf.RegisterAnsibleRunnerServer(gas, &aService)

	aService.Logger.Print("Ansible runner start work...\n")
	if err := gas.Serve(lis); err != nil {
		aService.Logger.Fatalf("failed to serve: %v", err)
	}
}
