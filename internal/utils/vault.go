package utils

import (
	vaultapi "github.com/hashicorp/vault/api"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

type OsCredentials map[string]string

type DockerCredentials struct {
	Url      string
	User     string
	Password string
}

type CbCredentials struct {
	Address  string `yaml:"cb_address"`
	Username string `yaml:"cb_username"`
	Password string `yaml:"cb_password"`
}

type HydraCredentials struct {
	RedirectUri  string
	ClientId     string
	ClientSecret string
}

type SecretStorage interface {
	ConnectVault() (*vaultapi.Client, *Config, error)
}

type VaultCommunicator struct {
	config Config
}

func (vc *VaultCommunicator) Init() error {
	path, err := os.Getwd()
	if err != nil {
		return ErrGetwd
	}

	var vaultConfigPath string
	if UseBasePath {
		vaultConfigPath = filepath.Join(path, ConfigPath)
	} else {
		vaultConfigPath = ConfigPath
	}

	vaultBs, err := ioutil.ReadFile(vaultConfigPath)
	if err != nil {
		return ErrVaultReadFile
	}

	err = yaml.Unmarshal(vaultBs, &vc.config)
	if err != nil {
		return ErrUnmarshal
	}
	return nil
}

func (vc *VaultCommunicator) ConnectVault() (*vaultapi.Client, *Config, error) {
	client, err := vaultapi.NewClient(&vaultapi.Config{
		Address: vc.config.VaultAddr,
	})
	if err != nil {
		return nil, nil, ErrVaultNewClient
	}

	client.SetToken(vc.config.Token)
	return client, &vc.config, nil
}
