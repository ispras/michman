package utils

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type Config struct {
	// Openstack
	Key            		string `yaml:"os_key_name"`
	VirtualNetwork 		string `yaml:"virtual_network"`
	OsImage        		string `yaml:"os_image"`
	FloatingIP     		string `yaml:"floating_ip_pool"`
	MasterFlavor   		string `yaml:"master_flavor"`
	SlavesFlavor   		string `yaml:"slaves_flavor"`
	StorageFlavor  		string `yaml:"storage_flavor"`
	MonitoringFlavor  	string `yaml:"monitoring_flavor"`
	OsVersion      		string `yaml:"os_version"` //Now are supported only 'stein', 'ussuri' and 'liberty' versions

	// Vault
	Token       string `yaml:"token"`
	VaultAddr   string `yaml:"vault_addr"`
	OsKey       string `yaml:"os_key"`
	SshKey      string `yaml:"ssh_key"`
	CbKey       string `yaml:"cb_key"`
	RegistryKey string `yaml:"registry_key"`
	HydraKey    string `yaml:"hydra_key"`

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

	//Authentication
	UseAuth            bool   `yaml:"use_auth"`
	AuthorizationModel string `yaml:"authorization_model,omitempty"`   //Now are supported only 'oauth2', 'none' or 'keystone' values
	AdminGroup         string `yaml:"admin_group,omitempty"`           //name of the Admin user group
	SessionIdleTimeout int    `yaml:"session_idle_timeout,omitempty"'` //time in minutes, controls the maximum length of time a session can be inactive before it expires
	SessionLifetime    int    `yaml:"session_lifetime,omitempty"`      //time in minutes, controls the maximum length of time that a session is valid for before it expires
	HydraAdmin         string `yaml:"hydra_admin,omitempty"`           //hydra admin address
	HydraClient        string `yaml:"hydra_client,omitempty"`          //hydra client address
	KeystoneAddr       string `yaml:"keystone_addr,omitempty"`         //keystone service address

	//Cluster logs
	LogsOutput   string `yaml:"logs_output"`              //file or logstash
	LogsFilePath string `yaml:"logs_file_path,omitempty"` //path to directory with cluster logs if file output is used
	LogstashAddr string `yaml:"logstash_addr,omitempty"`  //logstash address if logstash output is used
	ElasticAddr  string `yaml:"elastic_addr,omitempty"`   //elastic address if logstash output is used
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
		return err
	}

	if Cfg.UseAuth && Cfg.AuthorizationModel != NoneAuthMode && Cfg.AuthorizationModel != OAuth2Mode &&
		Cfg.AuthorizationModel != KeystoneMode {
		log.Fatalln("For config parameter 'authorization_model' are supported only 'none', 'oauth2' or 'keystone' values")
	}

	//check hydra address for oauth2 mode
	if Cfg.UseAuth && Cfg.AuthorizationModel == OAuth2Mode {
		if Cfg.HydraAdmin == "" || Cfg.HydraClient == "" {
			log.Fatalln("For oauth2 authorization mode config parameters 'hydra_admin' and 'hydra_client' couldn't be empty")
		}
	}

	//check keystone address for keystone mode
	if Cfg.UseAuth && Cfg.AuthorizationModel == KeystoneMode && Cfg.KeystoneAddr == "" {
		log.Fatalln("For keystone authorization mode config parameters 'keystone_addr' couldn't be empty")
	}

	//check logs output values
	if Cfg.LogsOutput != LogsFileOutput && Cfg.LogsOutput != LogsLogstashOutput {
		log.Fatalln("For config parameter 'logs_output` are supported only 'file' or 'logstash' values")
	}

	//check file path not empty if logs output is 'file'
	if Cfg.LogsOutput == LogsFileOutput && Cfg.LogsFilePath == "" {
		log.Fatalln("'logs_file_path' couldn't be empty")
	}

	//check if directory for logs exists and create it if not
	if Cfg.LogsOutput == LogsFileOutput {
		if _, err := os.Stat(Cfg.LogsFilePath); os.IsNotExist(err) {
			err := os.Mkdir(Cfg.LogsFilePath, os.ModePerm)
			if err != nil {
				log.Fatalln(err)
			}
		}
	}

	//check logstash addr not empty
	if Cfg.LogsOutput == LogsLogstashOutput && (Cfg.LogstashAddr == "" || Cfg.ElasticAddr == "") {
		log.Fatalln("For logstash logs output config parameters 'logstash_addr' and 'elastic_addr' couldn't be empty")
	}
	return nil
}
