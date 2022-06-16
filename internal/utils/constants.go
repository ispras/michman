package utils

import "os"

const (
	//list of supported services types
	ServiceTypeSpark      string = "spark"
	ServiceTypeIgnite     string = "ignite"
	ServiceTypeCassandra  string = "cassandra"
	ServiceTypeJupyter    string = "jupyter"
	ServiceTypeJupyterhub string = "jupyterhub"
	ServiceTypeElastic    string = "elastic"
	ServiceTypeNFS        string = "nfs-server"
	ServiceTypeNextCloud  string = "nextcloud"

	//supported spark configurations
	SparkUseYarn         string = "use-yarn"
	SparkHadoopVersion   string = "hadoop-version"
	SparkWorkerMemMb     string = "worker-mem-mb"
	SparkYarnMasterMemMb string = "yarn-master-mem-mb"

	//supported jupyter configurations
	JupyterToreeVersion string = "toree-version"

	//supported ignite configurations
	IgniteMemory string = "ignite-memory"

	//supported elastic configurations
	ElasticHeapSize string = "es-heap-size"

	//supported cassandra configurations
	CassandraDefaultVersion string = "3.11.4"

	//statuses for ansible runner
	AnsibleOk   string = "OK"
	AnsibleFail string = "FAIL"

	//supported actions for ansible
	AnsibleLaunch  = "launch"
	AnsibleDestroy = "destroy"

	//ansible-playbook command
	AnsiblePlaybookCmd = "ansible-playbook"

	//path for ansible.cfg
	AnsibleConfigVar  = "ANSIBLE_CONFIG"
	AnsibleConfigPath = "ansible/ansible.cfg"

	//base path
	BasePath = "michman"
	//ansible main role path
	AnsibleMainRole = "ansible/main.yml"
	//ansible get master ip role path
	AnsibleMasterIpRole = "ansible/get_master.yml"
	//ansible get ip of any instance with role
	AnsibleIpRole = "ansible/get_ip.yml"

	//openstack secrets keys
	OsAuthUrl            = "OS_AUTH_URL"
	OsPassword           = "OS_PASSWORD"
	OsProjectName        = "OS_PROJECT_NAME"
	OsProjectID          = "OS_PROJECT_ID"
	OsProjectDomainID    = "OS_PROJECT_DOMAIN_ID"
	OsInterface          = "OS_INTERFACE"
	OsRegionName         = "OS_REGION_NAME"
	OsTenantId           = "OS_TENANT_ID"
	OsTenantName         = "OS_TENANT_NAME"
	OsUserName           = "OS_USERNAME"
	OsSwiftUserName      = "OS_SWIFT_USERNAME"
	OsSwiftPassword      = "OS_SWIFT_PASSWORD"
	OsComputeApiVersion  = "COMPUTE_API_VERSION"
	OsNovaVersion        = "NOVA_VERSION"
	OsAuthType           = "OS_AUTH_TYPE"
	OsCloudname          = "OS_CLOUDNAME"
	OsIdentityApiVersion = "OS_IDENTITY_API_VERSION"
	OsImageApiVersion    = "OS_IMAGE_API_VERSION"
	OsNoCache            = "OS_NO_CACHE"
	OsProjectDomainName  = "OS_PROJECT_DOMAIN_NAME"
	OsUserDomainName     = "OS_USER_DOMAIN_NAME"
	OsVolumeApiVersion   = "OS_VOLUME_API_VERSION"
	OsPythonwarnings     = "PYTHONWARNINGS"
	OsNoProxy            = "no_proxy"

	// Docker login secrets keys
	DockerLoginUlr      = "url"
	DockerLoginUser     = "user"
	DockerLoginPassword = "password"

	//Couchbase secret keys
	CouchbasePath     = "path"
	CouchbaseUsername = "username"
	CouchbasePassword = "password"

	//Hydra secret keys
	HydraRedirectUri  = "redirect_uri"
	HydraClientId     = "client_id"
	HydraClientSecret = "client_secret"

	//ssh secrets keys
	VaultSshKey = "key_bgt"

	//Entity statuses
	StatusInited   = "INITED"
	StatusActive   = "ACTIVE"
	StatusFailed   = "FAILED"
	StatusStopping = "STOPPING"
	StatusMissing  = "MISSING"

	//default IDs
	CommonProjectID string = "None"

	//Openstack stein version
	OsSteinVersion   string = "stein"
	OsLibertyVersion string = "liberty"
	OsUssuriVersion  string = "ussuri"

	//Supported classes for service types
	ClassStorage     string = "storage"
	ClassMasterSlave string = "master-slave"
	ClassStandAlone  string = "stand-alone"

	//Authorization models
	OAuth2Mode   = "oauth2"
	NoneAuthMode = "none"
	KeystoneMode = "keystone"

	//sessions keys
	GroupKey       = "groups"
	AccessTokenKey = "AccessToken"

	//cluster logs outputs
	LogsFileOutput     = "file"
	LogsLogstashOutput = "logstash"

	//ansible actions
	ActionCreate = "create"
	ActionUpdate = "update"
	ActionDelete = "delete"
)

var (
	SshKeyPath  = os.Getenv("PWD") + "/ansible/files/ssh_key"
	ConfigPath  = "configs/config.yaml"
	UseBasePath = true
)
