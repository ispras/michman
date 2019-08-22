package main

import (
	"errors"
	vaultapi "github.com/hashicorp/vault/api"
	protobuf "gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/protobuf"
	"gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/utils"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"
	"io/ioutil"
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
	Run(c *protobuf.Cluster, osCreds *utils.OsCredentials, osConfig *osConfig, action string) string
}

// ansibleService implements ansible service
type ansibleService struct {
	logger        *log.Logger
	ansibleRunner ansibleLaunch
}

type osConfig struct {
	Key     string `yaml:"os_key_name"`
	VirtualNetwork string `yaml:"virtual_network"`
	OsImage     string `yaml:"os_image"`
	FloatingIP     string `yaml:"floating_ip_pool"`
	Flavor     string `yaml:"flavor"`
}

func makeOsCreds(keyName string, vaultClient *vaultapi.Client) *utils.OsCredentials {
	secretValues, err := vaultClient.Logical().Read(keyName)
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	var osCreds utils.OsCredentials
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
	return &osCreds
}

func (osCfg *osConfig) makeOsCfg() error {
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

	osConfigPath := filepath.Join(path, utils.OpenstackCfg)
	osBs, err := ioutil.ReadFile(osConfigPath)
	if err := yaml.Unmarshal(osBs, &osCfg); err != nil {
		log.Fatalln(err)
	}
	return nil
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

	sshPath := filepath.Join(path, utils.SshKeyPath)
	if _, err := os.Stat(sshPath); os.IsNotExist(err) {
		// ssh-key does not exist, getting it from vault
		secretValues, err := vaultClient.Logical().Read(keyName)
		if err != nil {
			log.Fatalln(err)
			return err
		}

		sshKey := secretValues.Data[utils.VaultSshKey].(string)
		f, err := os.Create(utils.SshKeyPath)
		if err != nil {
			log.Fatalln(err)
			return err
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
	}
	return nil
}

func (aS *ansibleService) Delete(in *protobuf.Cluster, stream protobuf.AnsibleRunner_DeleteServer) error {
	aS.logger.Print("Getting delete cluster request...")
	aS.logger.Print("Cluster info:")
	in.PrintClusterData()

	aS.logger.Print("Getting vault secrets...")

	vaultClient, vaultCfg := utils.ConnectVault()
	if vaultClient == nil {
		log.Fatalln("Error: can't connect to vault secrets storage")
		return nil
	}

	keyName := vaultCfg.OsKey

	osCreds := makeOsCreds(keyName, vaultClient)
	if osCreds == nil {
		return nil
	}
	//aS.logger.Print(osCreds)

	//getting openstack config info
	var osCfg osConfig
	err := osCfg.makeOsCfg()
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	//check file with ssh-key and create it if it doesn't exist
	err = checkSshKey(vaultCfg.SshKey, vaultClient)
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	// here ansible will run
	ansibleStatus := aS.ansibleRunner.Run(in, osCreds, &osCfg, actionDelete)

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
	in.PrintClusterData()

	aS.logger.Print("Getting vault secrets...")

	vaultClient, vaultCfg := utils.ConnectVault()
	if vaultClient == nil {
		log.Fatalln("Error: can't connect to vault secrets storage")
		return nil
	}

	osCreds := makeOsCreds(vaultCfg.OsKey, vaultClient)
	if osCreds == nil {
		return nil
	}
	//aS.logger.Print(osCreds)

	//getting openstack config info
	var osCfg osConfig
	err := osCfg.makeOsCfg()
	if err != nil {
		log.Fatalln(err)
		return nil
	}
	//check file with ssh-key and create it if it doesn't exist
	err = checkSshKey(vaultCfg.SshKey, vaultClient)
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	// here ansible will run
	ansibleStatus := aS.ansibleRunner.Run(in, osCreds, &osCfg, actionUpdate)

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
	in.PrintClusterData()

	aS.logger.Print("Getting vault secrets...")

	vaultClient, vaultCfg := utils.ConnectVault()
	if vaultClient == nil {
		log.Fatalln("Error: can't connect to vault secrets storage")
		return nil
	}

	osCreds := makeOsCreds(vaultCfg.OsKey, vaultClient)
	if osCreds == nil {
		return nil
	}
	//aS.logger.Print(osCreds)

	//getting openstack config info
	var osCfg osConfig
	err := osCfg.makeOsCfg()
	if err != nil {
		log.Fatalln(err)
		return nil
	}
	//check file with ssh-key and create it if it doesn't exist
	err = checkSshKey(vaultCfg.SshKey, vaultClient)
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	// here ansible will run
	ansibleStatus := aS.ansibleRunner.Run(in, osCreds, &osCfg, actionCreate)

	if err := stream.Send(&protobuf.TaskStatus{Status: ansibleStatus}); err != nil {
		return err
	}

	//	if err := stream.Send(io.EOF); err != nil {
	//		return err
	//	}

	return nil
}

func main() {
	ansibleServiceLogger := log.New(os.Stdout, "ANSIBLE_SERVICE: ", log.Ldate|log.Ltime)
	ansibleLaunch := AnsibleLauncher{}

	lis, err := net.Listen("tcp", inputPort)
	if err != nil {
		ansibleServiceLogger.Fatalf("failed to listen: %v", err)
	}

	gas := grpc.NewServer()
	protobuf.RegisterAnsibleRunnerServer(gas, &ansibleService{ansibleServiceLogger, ansibleLaunch})

	ansibleServiceLogger.Print("Ansible runner start work...\n")
	if err := gas.Serve(lis); err != nil {
		ansibleServiceLogger.Fatalf("failed to serve: %v", err)
	}
}
