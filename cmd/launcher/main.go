package main

import (
	"flag"
	"github.com/ispras/michman/cmd"
	"github.com/ispras/michman/internal/ansible"
	"github.com/ispras/michman/internal/database"
	"github.com/ispras/michman/internal/logger"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/utils"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"io"
	"net"
	"os"
	"time"
)

func main() {
	//set flags for config path and ansible service adress
	configPath := flag.String("config", utils.ConfigPath, "Path to the config.yaml file")
	launcherPort := flag.String("port", ansible.LauncherDefaultPort, "Launcher service default port")
	flag.Parse()

	//set config file path
	utils.SetConfigPath(*configPath)
	config := utils.Config{}
	err := config.MakeCfg()
	if err != nil {
		panic(err)
	}

	logFile, err := os.OpenFile(config.LogsFilePath+"/launcher.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(cmd.ErrOpenFile)
	}
	mw := io.MultiWriter(os.Stdout, logFile)

	LauncherLogger := &logrus.Logger{
		Out:   mw,
		Level: logrus.InfoLevel,
		Formatter: &logger.Formatter{
			TimestampFormat: time.Stamp,
			NoColors:        true,
			NoFieldsColors:  true,
			ShowFullLevel:   false,
			LoggerName:      "LAUNCHER",
		},
	}

	vaultCommunicator := utils.VaultCommunicator{}
	err = vaultCommunicator.Init()
	if err != nil {
		LauncherLogger.SetOutput(os.Stderr)
		LauncherLogger.Fatal(err)
	}
	db, err := database.NewCouchBase(&vaultCommunicator)
	if err != nil {
		LauncherLogger.SetOutput(os.Stderr)
		LauncherLogger.Fatal(err)
	}
	lis, err := net.Listen("tcp", ":"+*launcherPort)
	if err != nil {
		LauncherLogger.SetOutput(os.Stderr)
		LauncherLogger.Fatal(cmd.ErrTcpListen(*launcherPort))
	}
	gas := grpc.NewServer()

	vaultClient, vaultCfg, err := vaultCommunicator.ConnectVault()
	if vaultClient == nil || err != nil {
		LauncherLogger.SetOutput(os.Stderr)
		LauncherLogger.Fatal(err)
	}

	osCreds, err := ansible.MakeOsCreds(vaultCfg.OsKey, vaultClient, config.OsVersion)

	aService := ansible.LauncherServer{Logger: LauncherLogger, Db: db,
		VaultCommunicator: &vaultCommunicator, Config: config, OsCreds: osCreds}

	protobuf.RegisterAnsibleRunnerServer(gas, &aService)

	aService.Logger.Info("Ansible runner start work...")
	err = gas.Serve(lis)
	if err != nil {
		LauncherLogger.SetOutput(os.Stderr)
		aService.Logger.Fatal(cmd.ErrServe)
	}
}
