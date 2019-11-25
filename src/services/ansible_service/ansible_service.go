package main

import (
	"errors"
	"fmt"
	vaultapi "github.com/hashicorp/vault/api"
	"gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/database"
	protobuf "gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/protobuf"
	"gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/utils"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
)

const (
	inputPort = ":5000"

	actionCreate = "create"
	actionUpdate = "update"
	actionDelete = "delete"
)

type ansibleLaunch interface {
	Run(c *protobuf.Cluster, osCreds *utils.OsCredentials, osConfig *utils.Config, action string) string
}

// ansibleService implements ansible service
type ansibleService struct {
	logger            *log.Logger
	ansibleRunner     ansibleLaunch
	vaultCommunicator utils.SecretStorage
	config utils.Config
}

func (aS *ansibleService) Init( logger *log.Logger, ansibleLaunch AnsibleLauncher,
	vaultCommunicator utils.SecretStorage) {
	config := utils.Config{}
	config.MakeCfg()
	aS.logger = logger
	aS.ansibleRunner = ansibleLaunch
	aS.vaultCommunicator = vaultCommunicator
	aS.config = config
}

func makeOsCreds(keyName string, vaultClient *vaultapi.Client, version string) *utils.OsCredentials {
	secretValues, err := vaultClient.Logical().Read(keyName)
	if err != nil {
		log.Fatalln(err)
		return nil
	}
	var osCreds utils.OsCredentials
	switch version {
	case utils.OsSteinVersion:
		osCreds.OsAuthUrl = secretValues.Data[utils.OsAuthUrl].(string)
		osCreds.OsPassword = secretValues.Data[utils.OsPassword].(string)
		osCreds.OsProjectName = secretValues.Data[utils.OsProjectName].(string)
		osCreds.OsRegionName = secretValues.Data[utils.OsRegionName].(string)
		osCreds.OsUserName = secretValues.Data[utils.OsUsername].(string)
		osCreds.OsComputeApiVersion = secretValues.Data[utils.OsComputeApiVersion].(string)
		osCreds.OsNovaVersion = secretValues.Data[utils.OsNovaVersion].(string)
		osCreds.OsAuthType = secretValues.Data[utils.OsAuthType].(string)
		osCreds.OsCloudname = secretValues.Data[utils.OsCloudname].(string)
		osCreds.OsIdentityApiVersion = secretValues.Data[utils.OsIdentityApiVersion].(string)
		osCreds.OsImageApiVersion = secretValues.Data[utils.OsImageApiVersion].(string)
		osCreds.OsNoCache = secretValues.Data[utils.OsNoCache].(string)
		osCreds.OsProjectDomainName = secretValues.Data[utils.OsProjectDomainName].(string)
		osCreds.OsUserDomainName = secretValues.Data[utils.OsUserDomainName].(string)
		osCreds.OsVolumeApiVersion = secretValues.Data[utils.OsVolumeApiVersion].(string)
		osCreds.OsPythonwarnings = secretValues.Data[utils.OsPythonwarnings].(string)
		osCreds.OsNoProxy = secretValues.Data[utils.OsNoProxy].(string)
	case utils.OsLibertyVersion:
		osCreds.OsAuthUrl = secretValues.Data[utils.OsAuthUrl].(string)
		osCreds.OsPassword = secretValues.Data[utils.OsPassword].(string)
		osCreds.OsProjectName = secretValues.Data[utils.OsProjectName].(string)
		osCreds.OsRegionName = secretValues.Data[utils.OsRegionName].(string)
		osCreds.OsTenantId = secretValues.Data[utils.OsTenantId].(string)
		osCreds.OsTenantName = secretValues.Data[utils.OsTenantName].(string)
		osCreds.OsUserName = secretValues.Data[utils.OsUsername].(string)
		if uname, ok := secretValues.Data[utils.OsSwiftUsername]; ok {
			osCreds.OsSwiftUserName = uname.(string)
		} else {
			osCreds.OsSwiftUserName = ""
		}
		if pass, ok := secretValues.Data[utils.OsSwiftPassword]; ok {
			osCreds.OsSwiftUserName = pass.(string)
		} else {
			osCreds.OsSwiftPassword = ""
		}
	default: //liberty as default version
		osCreds.OsAuthUrl = secretValues.Data[utils.OsAuthUrl].(string)
		osCreds.OsPassword = secretValues.Data[utils.OsPassword].(string)
		osCreds.OsProjectName = secretValues.Data[utils.OsProjectName].(string)
		osCreds.OsRegionName = secretValues.Data[utils.OsRegionName].(string)
		osCreds.OsTenantId = secretValues.Data[utils.OsTenantId].(string)
		osCreds.OsTenantName = secretValues.Data[utils.OsTenantName].(string)
		osCreds.OsUserName = secretValues.Data[utils.OsUsername].(string)
		if uname, ok := secretValues.Data[utils.OsSwiftUsername]; ok {
			osCreds.OsSwiftUserName = uname.(string)
		} else {
			osCreds.OsSwiftUserName = ""
		}
		if pass, ok := secretValues.Data[utils.OsSwiftPassword]; ok {
			osCreds.OsSwiftUserName = pass.(string)
		} else {
			osCreds.OsSwiftPassword = ""
		}
	}

	return &osCreds
}

func checkSshKey(keyName string, vaultClient *vaultapi.Client) error {
	path, err := os.Getwd() //file must be executed from spark-openstack directory
	if err != nil {
		log.Fatalln(err)
		return err
	}

	workingDir := filepath.Base(path)
	if workingDir != utils.BasePath { //checking that current directory is correct
		log.Fatalln("Error: working directory must be spark-openstack")
		return errors.New("Error: working directory must be spark-openstack")
	}

	sshPath := filepath.Join(utils.SshKeyPath)
	if _, err := os.Stat(sshPath); os.IsNotExist(err) {
		// ssh-key does not exist, getting it from vault
		secretValues, err := vaultClient.Logical().Read(keyName)
		if err != nil {
			log.Fatalln(err)
			return err
		}

		sshKey := secretValues.Data[utils.VaultSshKey].(string)
		f, err := os.Create(sshPath)
		if err != nil {
			log.Fatalln(err)
			return err
		}
		err = os.Chmod(sshPath, 0777)
		if err != nil {
			log.Fatalln(err)
		}
		_, err = f.WriteString(sshKey)
		if err != nil {
			log.Fatalln(err)
			return err
		}
		err = f.Close()
		if err != nil {
			log.Fatalln(err)
			return err
		}
		err = os.Chmod(sshPath, 0400)
		if err != nil {
			log.Fatalln(err)
		}
	}
	return nil
}

func (aS *ansibleService) Delete(in *protobuf.Cluster, stream protobuf.AnsibleRunner_DeleteServer) error {
	aS.logger.Print("Getting delete cluster request...")
	aS.logger.Print("Cluster info:")
	in.PrintClusterData(aS.logger)

	aS.logger.Print("Getting vault secrets...")

	vaultClient, vaultCfg := aS.vaultCommunicator.ConnectVault()
	if vaultClient == nil {
		log.Fatalln("Error: can't connect to vault secrets storage")
		return nil
	}

	keyName := vaultCfg.OsKey

	osCreds := makeOsCreds(keyName, vaultClient, aS.config.OsVersion)
	if osCreds == nil {
		return nil
	}

	//check file with ssh-key and create it if it doesn't exist
	err := checkSshKey(vaultCfg.SshKey, vaultClient)
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	// here ansible will run
	ansibleStatus := aS.ansibleRunner.Run(in, osCreds, &aS.config, actionDelete)

	if err := stream.Send(&protobuf.TaskStatus{Status: ansibleStatus}); err != nil {
		return err
	}

	//	if err := stream.Send(io.EOF); err != nil {
	//		return err
	//	}
	return nil
}

func (aS *ansibleService) Update(in *protobuf.Cluster, stream protobuf.AnsibleRunner_UpdateServer) error {
	aS.logger.Print("Getting update cluster request...")
	aS.logger.Print("Cluster info:")
	in.PrintClusterData(aS.logger)

	aS.logger.Print("Getting vault secrets...")

	vaultClient, vaultCfg := aS.vaultCommunicator.ConnectVault()
	if vaultClient == nil {
		log.Fatalln("Error: can't connect to vault secrets storage")
		return nil
	}

	osCreds := makeOsCreds(vaultCfg.OsKey, vaultClient, aS.config.OsVersion)
	if osCreds == nil {
		return nil
	}

	//check file with ssh-key and create it if it doesn't exist
	err := checkSshKey(vaultCfg.SshKey, vaultClient)
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	// here ansible will run
	ansibleStatus := aS.ansibleRunner.Run(in, osCreds, &aS.config, actionUpdate)

	if err := stream.Send(&protobuf.TaskStatus{Status: ansibleStatus}); err != nil {
		return err
	}

	//	if err := stream.Send(io.EOF); err != nil {
	//		return err
	//	}

	return nil
}

func (aS *ansibleService) Create(in *protobuf.Cluster, stream protobuf.AnsibleRunner_CreateServer) error {
	aS.logger.Print("Getting create cluster request...")
	aS.logger.Print("Cluster info:")
	in.PrintClusterData(aS.logger)

	aS.logger.Print("Getting vault secrets...")

	vaultClient, vaultCfg := aS.vaultCommunicator.ConnectVault()
	if vaultClient == nil {
		log.Fatalln("Error: can't connect to vault secrets storage")
		return nil
	}

	osCreds := makeOsCreds(vaultCfg.OsKey, vaultClient, aS.config.OsVersion)
	if osCreds == nil {
		return nil
	}
	//aS.logger.Print(osCreds)

	//check file with ssh-key and create it if it doesn't exist
	err := checkSshKey(vaultCfg.SshKey, vaultClient)
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	// here ansible will run
	ansibleStatus := aS.ansibleRunner.Run(in, osCreds, &aS.config, actionCreate)

	if err := stream.Send(&protobuf.TaskStatus{Status: ansibleStatus}); err != nil {
		return err
	}

	//	if err := stream.Send(io.EOF); err != nil {
	//		return err
	//	}

	return nil
}

func (aS *ansibleService) GetMasterIP(in *protobuf.Cluster, stream protobuf.AnsibleRunner_GetMasterIPServer) error {
	return nil
}

func main() {
	// check if we get path to config
	args := os.Args[1:]
	if len(args) > 0 {
		fmt.Printf("Config path is %v\n", args[0])
		utils.SetConfigPath(args[0])
	}

	// create a multiwriter which writes to stout and a file simultaneously
	logFile, err := os.OpenFile("logs/ansible_service.log", os.O_CREATE | os.O_APPEND | os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("Can't create a log file. Exit...")
		os.Exit(1)
	}
	mw := io.MultiWriter(os.Stdout, logFile)

	ansibleServiceLogger := log.New(mw, "ANSIBLE_SERVICE: ", log.Ldate|log.Ltime)
	vaultCommunicator := utils.VaultCommunicator{}
	vaultCommunicator.Init()
	db, err := database.NewCouchBase(&vaultCommunicator)
	if err != nil {
		fmt.Println("Can't create database connection. Exit...")
		os.Exit(1)
	}
	ansibleLaunch := AnsibleLauncher{couchbaseCommunicator: db}

	lis, err := net.Listen("tcp", inputPort)
	if err != nil {
		ansibleServiceLogger.Fatalf("failed to listen: %v", err)
	}

	gas := grpc.NewServer()
	aService := ansibleService{}
	aService.Init(ansibleServiceLogger, ansibleLaunch, &vaultCommunicator)
	protobuf.RegisterAnsibleRunnerServer(gas, &aService)

	ansibleServiceLogger.Print("Ansible runner start work...\n")
	if err := gas.Serve(lis); err != nil {
		ansibleServiceLogger.Fatalf("failed to serve: %v", err)
	}
}
