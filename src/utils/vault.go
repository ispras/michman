package utils

import (
	vaultapi "github.com/hashicorp/vault/api"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type OsCredentials struct {
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

type AuthConfig struct {
	Token     string `yaml:"token"`
	VaultAddr string `yaml:"vault_addr"`
	OsKey     string `yaml:"os_key"`
	SshKey    string `yaml:"ssh_key"`
}

func ConnectVault() (*vaultapi.Client, *AuthConfig) {
	path, err := os.Getwd() //file must be executed from spark-openstack directory
	if err != nil {
		log.Fatalln(err)
	}

	workingDir := filepath.Base(path)
	if workingDir != BasePath { //checking that current directory is correct
		log.Fatalln("Error: working directory must be spark-openstack")
		return nil, nil
	}


	//getting openstack credential info from vault
	vaultConfigPath := filepath.Join(path, VaultCfg)
	vaultBs, err := ioutil.ReadFile(vaultConfigPath)
	var vaultCfg *AuthConfig
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
	return client, vaultCfg
}