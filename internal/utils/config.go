package utils

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Config struct {
	// Openstack
	Key            string `yaml:"os_key_name"`
	VirtualNetwork string `yaml:"virtual_network"`
	FloatingIP     string `yaml:"floating_ip_pool"`
	OsVersion      string `yaml:"os_version"` //Now are supported only 'stein', 'ussuri' and 'liberty' versions

	// Vault
	VaultAddr   string `yaml:"vault_addr"`
	Token       string `yaml:"token"`
	OsKey       string `yaml:"os_key"`
	SshKey      string `yaml:"ssh_key"`
	Storage     string `yaml:"storage"` //couchbase or mysql
	CbKey       string `yaml:"cb_key"`
	MySqlKey    string `yaml:"mysql_key"`
	RegistryKey string `yaml:"registry_key"`
	HydraKey    string `yaml:"hydra_key"`

	//Cluster logs
	LogsOutput   string `yaml:"logs_output"`              //file or logstash
	LogsFilePath string `yaml:"logs_file_path,omitempty"` //path to directory with cluster logs if file output is used
	LogstashAddr string `yaml:"logstash_addr,omitempty"`  //logstash address if logstash output is used
	ElasticAddr  string `yaml:"elastic_addr,omitempty"`   //elastic address if logstash output is used

	// Mirror
	UsePackageMirror bool   `yaml:"use_package_mirror,omitempty"`
	UsePipMirror     bool   `yaml:"use_pip_mirror,omitempty"`
	YumMirrorAddress string `yaml:"yum_mirror_address,omitempty"`
	AptMirrorAddress string `yaml:"apt_mirror_address,omitempty"`
	PipMirrorAddress string `yaml:"pip_mirror_address,omitempty"`
	PipTrustedHost   string `yaml:"pip_trusted_host,omitempty"`

	// Registry
	InsecureRegistry       bool   `yaml:"docker_insecure_registry,omitempty"`
	SelfignedRegistry      bool   `yaml:"docker_selfsigned_registry,omitempty"`
	GitlabRegistry         bool   `yaml:"docker_gitlab_registry,omitempty"`
	InsecureRegistryIp     string `yaml:"docker_insecure_registry_ip,omitempty"`
	SelfsignedRegistryIp   string `yaml:"docker_selfsigned_registry_ip,omitempty"`
	SelfsignedRegistryPort string `yaml:"docker_selfsigned_registry_port,omitempty"`
	SelfsignedRegistryUrl  string `yaml:"docker_selfsigned_registry_url,omitempty"`
	SelfsignedRegistryCert string `yaml:"docker_cert_path,omitempty"`

	//Authentication
	UseAuth            bool   `yaml:"use_auth"`
	AuthorizationModel string `yaml:"authorization_model,omitempty"`  //Now are supported only 'oauth2', 'none' or 'keystone' values
	AdminGroup         string `yaml:"admin_group,omitempty"`          //name of the Admin user group
	SessionIdleTimeout int    `yaml:"session_idle_timeout,omitempty"` //time in minutes, controls the maximum length of time a session can be inactive before it expires
	SessionLifetime    int    `yaml:"session_lifetime,omitempty"`     //time in minutes, controls the maximum length of time that a session is valid for before it expires
	HydraAdmin         string `yaml:"hydra_admin,omitempty"`          //hydra admin address
	HydraClient        string `yaml:"hydra_client,omitempty"`         //hydra client address
	KeystoneAddr       string `yaml:"keystone_addr,omitempty"`        //keystone service address
	AuthConfigPath     string `yaml:"auth_config_path,omitempty"`     //path to auth_model.conf
	PolicyPath         string `yaml:"policy_path,omitempty"`          //path to policy.csv
}

func SetConfigPath(configPath string) {
	ConfigPath = configPath
	UseBasePath = false
}

func (Cfg *Config) MakeCfg() error {
	path, err := os.Getwd()
	if err != nil {
		return ErrGetwd
	}

	var osConfigPath string
	if UseBasePath {
		osConfigPath = filepath.Join(path, ConfigPath)
	} else {
		osConfigPath = ConfigPath
	}

	osBs, err := ioutil.ReadFile(osConfigPath)
	if err != nil {
		return ErrVaultReadFile
	}
	err = yaml.Unmarshal(osBs, &Cfg)
	if err != nil {
		return ErrUnmarshal
	}

	if Cfg.UseAuth && Cfg.AuthorizationModel != NoneAuthMode && Cfg.AuthorizationModel != OAuth2Mode &&
		Cfg.AuthorizationModel != KeystoneMode {
		return ErrAuthorizationModel
	}

	//check hydra address for oauth2 mode
	if Cfg.UseAuth && Cfg.AuthorizationModel == OAuth2Mode {
		if Cfg.HydraAdmin == "" || Cfg.HydraClient == "" {
			return ErrOAuth2ModeAuthorization
		}
	}

	//check keystone address for keystone mode
	if Cfg.UseAuth && Cfg.AuthorizationModel == KeystoneMode && Cfg.KeystoneAddr == "" {
		return ErrKeystoneAuthorization
	}

	//check logs output values
	if Cfg.LogsOutput != LogsFileOutput && Cfg.LogsOutput != LogsLogstashOutput {
		return ErrLogsOutputParams
	}

	//check file path not empty if logs output is 'file'
	if Cfg.LogsOutput == LogsFileOutput && Cfg.LogsFilePath == "" {
		return ErrLogsFilePathEmpty
	}

	//check if directory for logs exists and create it if not
	if Cfg.LogsOutput == LogsFileOutput {
		if _, err := os.Stat(Cfg.LogsFilePath); os.IsNotExist(err) {
			err := os.Mkdir(Cfg.LogsFilePath, os.ModePerm)
			if err != nil {
				return ErrMkdir
			}
		}
	}

	//check logstash addr not empty
	if Cfg.LogsOutput == LogsLogstashOutput && (Cfg.LogstashAddr == "" || Cfg.ElasticAddr == "") {
		return ErrLogstashOutputParams
	}

	if Cfg.Storage != StorageCouchbase && Cfg.Storage != StorageMySQL {
		return ErrStorage
	}
	return nil
}
