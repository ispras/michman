package utils

import (
	vaultapi "github.com/hashicorp/vault/api"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type SecretStorage interface {
	ConnectVault() (*vaultapi.Client, *Config)
}

type VaultCommunicator struct {
	config Config
}

func (vc *VaultCommunicator) Init() error {
	path, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}

	var vaultConfigPath string
	if UseBasePath {
		vaultConfigPath = filepath.Join(path, ConfigPath)
	} else {
		vaultConfigPath = ConfigPath
	}

	vaultBs, err := ioutil.ReadFile(vaultConfigPath)
	if err := yaml.Unmarshal(vaultBs, &vc.config); err != nil {
		log.Fatalln(err)
	}
	return nil
}

func (vc *VaultCommunicator) ConnectVault() (*vaultapi.Client, *Config) {
	client, err := vaultapi.NewClient(&vaultapi.Config{
		Address: vc.config.VaultAddr,
	})
	if err != nil {
		log.Fatalln(err)
	}

	client.SetToken(vc.config.Token)
	return client, &vc.config
}
