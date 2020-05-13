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
	OsAuthUrl            string
	OsPassword           string
	OsProjectName        string
	OsRegionName         string
	OsTenantId           string
	OsTenantName         string
	OsUserName           string
	OsSwiftUserName      string
	OsSwiftPassword      string
	OsComputeApiVersion  string
	OsNovaVersion        string
	OsAuthType           string
	OsCloudname          string
	OsIdentityApiVersion string
	OsImageApiVersion    string
	OsNoCache            string
	OsProjectDomainName  string
	OsUserDomainName     string
	OsVolumeApiVersion   string
	OsPythonwarnings     string
	OsNoProxy            string
}

type DockerCredentials struct {
	Url      string
	User     string
	Password string
}

type Config struct {
	// Openstack
	Key            string `yaml:"os_key_name"`
	VirtualNetwork string `yaml:"virtual_network"`
	OsImage        string `yaml:"os_image"`
	FloatingIP     string `yaml:"floating_ip_pool"`
	MasterFlavor   string `yaml:"master_flavor"`
	SlavesFlavor   string `yaml:"slaves_flavor"`
	StorageFlavor  string `yaml:"storage_flavor"`
	OsVersion      string `yaml:"os_version"` //Now are supported only 'stein' and 'liberty' versions

	// Vault
	Token       string `yaml:"token"`
	VaultAddr   string `yaml:"vault_addr"`
	OsKey       string `yaml:"os_key"`
	SshKey      string `yaml:"ssh_key"`
	CbKey       string `yaml:"cb_key"`
	RegistryKey string `yaml:"registry_key"`

	// Mirror
	UsePackageMirror string `yaml:"use_package_mirror"`
	UsePipMirror     string `yaml:"use_pip_mirror"`
	YumMirrorAddress string `yaml:"yum_mirror_address"`
	AptMirrorAddress string `yaml:"apt_mirror_address"`
	PipMirrorAddress string `yaml:"pip_mirror_address"`
	PipTrustedHost   string `yaml:"pip_trusted_host"`

	// Registry
	SelfignedRegistry bool `yaml:"docker_selfsigned_registry"`
	InsecureRegistry  bool `yaml:"docker_insecure_registry"`
	GitlabRegistry    bool `yaml:"gitlab_registry"`

	SelfsignedRegistryIp string `yaml:"docker_selfsigned_registry_ip"`
	InsecureRegistryIp   string `yaml:"docker_insecure_registry_ip"`

	SelfignedRegistryUrl  string `yaml:"docker_selfsigned_registry_url"`
	SelfignedRegistryCert string `yaml:"docker_cert_path"`

	// User to connect
	AnsibleUser string `yaml:ansible_user`
}

func SetConfigPath(configPath string) {
	ConfigPath = configPath
	UseBasePath = false
}

func (Cfg *Config) MakeCfg() error {
	path, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
		return err
	}

	var osConfigPath string
	if UseBasePath {
		osConfigPath = filepath.Join(path, ConfigPath)
	} else {
		osConfigPath = ConfigPath
	}

	osBs, err := ioutil.ReadFile(osConfigPath)
	if err := yaml.Unmarshal(osBs, &Cfg); err != nil {
		log.Fatalln(err)
	}
	return nil
}

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
