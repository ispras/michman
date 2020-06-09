package utils

import (
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
	UsePackageMirror string `yaml:"use_package_mirror,omitempty"`
	UsePipMirror     string `yaml:"use_pip_mirror,omitempty"`
	YumMirrorAddress string `yaml:"yum_mirror_address,omitempty"`
	AptMirrorAddress string `yaml:"apt_mirror_address,omitempty"`
	PipMirrorAddress string `yaml:"pip_mirror_address,omitempty"`
	PipTrustedHost   string `yaml:"pip_trusted_host,omitempty"`

	// Registry
	SelfignedRegistry bool `yaml:"docker_selfsigned_registry,omitempty"`
	InsecureRegistry  bool `yaml:"docker_insecure_registry,omitempty"`
	GitlabRegistry    bool `yaml:"gitlab_registry,omitempty"`

	SelfsignedRegistryIp string `yaml:"docker_selfsigned_registry_ip,omitempty"`
	InsecureRegistryIp   string `yaml:"docker_insecure_registry_ip,omitempty"`

	SelfignedRegistryUrl  string `yaml:"docker_selfsigned_registry_url,omitempty"`
	SelfignedRegistryCert string `yaml:"docker_cert_path,omitempty"`
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