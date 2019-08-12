package main

import (
	vaultapi "github.com/hashicorp/vault/api"
	protobuf "gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/protobuf"
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
)

type ansibleLaunche interface {
	Run(c *protobuf.Cluster, osCreds *osCredentials, osConfig *osConfig) error
}

// ansibleService implements ansible service
type ansibleService struct {
	logger        *log.Logger
	ansibleRunner ansibleLaunche
}

type authConfig struct {
	Token     string `yaml:"token"`
	VaultAddr string `yaml:"vault_addr"`
	OsKey     string `yaml:"os_key"`
}

type osConfig struct {
	Key     string `yaml:"os_key_name"`
	VirtualNetwork string `yaml:"virtual_network"`
	OsImage     string `yaml:"os_image"`
	FloatingIP     string `yaml:"floating_ip_pool"`
	Flavor     string `yaml:"flavor"`
}

type osCredentials struct {
	OsAuthUrl string
	OsPassword string
	OsProjectName string
	OsRegionName string
	OsTenantId string
	OsTenantName string
	OsUserName string
	OsSwiftUserName string
	OsSwiftPassword string
}

func (aS *ansibleService) RunAnsible(in *protobuf.Cluster, stream protobuf.AnsibleRunner_RunAnsibleServer) error {
	aS.logger.Print("Getting create cluster request...")
	aS.logger.Print("Cluster info:")
	in.PrintClusterData()

	aS.logger.Print("Getting vault secrets...")

	path, err := os.Getwd() //file must be executed from spark-openstack directory
	if err != nil {
		log.Fatalln(err)
	}

	workingDir := filepath.Base(path)
	if workingDir != "spark-openstack" { //checking that current directory is correct
		log.Fatalln("Error: working directory must be spark-openstack")
		return nil
	}


	//getting openstack credential info from vault
	vaultConfigPath := filepath.Join(path, "vault.yaml")
	vaultBs, err := ioutil.ReadFile(vaultConfigPath)
	var vaultCfg authConfig
	if err := yaml.Unmarshal(vaultBs, &vaultCfg); err != nil {
		log.Fatalln(err)
	}

	client, err := vaultapi.NewClient(&vaultapi.Config{
		Address: vaultCfg.VaultAddr,
	})
	if err != nil {
		log.Fatalln(err)
	}

	client.SetToken(vaultCfg.Token)

	keyName := vaultCfg.OsKey
	secretValues, err := client.Logical().Read(keyName)
	if err != nil {
		log.Fatalln(err)
	}

	var osCreds osCredentials
	osCreds.OsAuthUrl = secretValues.Data["OS_AUTH_URL"].(string)
	osCreds.OsPassword = secretValues.Data["OS_PASSWORD"].(string)
	osCreds.OsProjectName = secretValues.Data["OS_PROJECT_NAME"].(string)
	osCreds.OsRegionName = secretValues.Data["OS_REGION_NAME"].(string)
	osCreds.OsTenantId = secretValues.Data["OS_TENANT_ID"].(string)
	osCreds.OsTenantName = secretValues.Data["OS_TENANT_NAME"].(string)
	osCreds.OsUserName = secretValues.Data["OS_USERNAME"].(string)

	if uname, ok := secretValues.Data["OS_SWIFT_USERNAME"]; ok {
		osCreds.OsSwiftUserName = uname.(string)
	} else {
		osCreds.OsSwiftUserName = ""
	}
	if pass, ok := secretValues.Data["OS_SWIFT_PASSWORD"]; ok {
		osCreds.OsSwiftUserName = pass.(string)
	} else {
		osCreds.OsSwiftPassword = ""
	}
	//aS.logger.Print(osCreds)

	//getting openstack config info
	osConfigPath := filepath.Join(path, "openstack_config.yaml")
	osBs, err := ioutil.ReadFile(osConfigPath)
	var osCfg osConfig
	if err := yaml.Unmarshal(osBs, &osCfg); err != nil {
		log.Fatalln(err)
	}

	// here ansible will run
	aS.ansibleRunner.Run(in, &osCreds, &osCfg)

	if err := stream.Send(&protobuf.TaskStatus{Status: "OK"}); err != nil {
		return err
	}

	//	if err := stream.Send(io.EOF); err != nil {
	//		return err
	//	}

	return nil
}

func main() {
	ansibleServiceLogger := log.New(os.Stdout, "ANSIBLE_SERVICE: ", log.Ldate|log.Ltime)
	ansibleLaunche := AnsibleLauncher{}

	lis, err := net.Listen("tcp", inputPort)
	if err != nil {
		ansibleServiceLogger.Fatalf("failed to listen: %v", err)
	}

	gas := grpc.NewServer()
	protobuf.RegisterAnsibleRunnerServer(gas, &ansibleService{ansibleServiceLogger, ansibleLaunche})

	ansibleServiceLogger.Print("Ansible runner start work...\n")
	if err := gas.Serve(lis); err != nil {
		ansibleServiceLogger.Fatalf("failed to serve: %v", err)
	}
}
