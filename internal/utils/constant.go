package utils

import "os"

const (
	//statuses for ansible runner
	AnsibleOk   string = "OK"
	AnsibleFail string = "FAIL"
	RunFail     string = "RUN_FAIL"

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
	UserIdKey      = "user_id"

	//cluster logs outputs
	LogsFileOutput     = "file"
	LogsLogstashOutput = "logstash"

	//ansible actions
	ActionCreate = "create"
	ActionUpdate = "update"
	ActionDelete = "delete"

	//log file names
	HttpLogFileName     = "http_server.log"
	LauncherLogFileName = "launcher.log"

	//Pattern strings
	ClusterNamePattern = `^[A-Za-z][A-Za-z0-9-]+$`
	ProjectNamePattern = `^[A-Za-z][A-Za-z0-9-]+$`
	ProjectPathPattern = `^/projects/`
	RegexPattern       = "Bearer " + "[A-Za-z0-9\\-\\._~\\+\\/]+=*"

	//openstack secrets keys value names
	OsAuthUrl            = "OsAuthUrl"
	OsPassword           = "OsPassword"
	OsProjectName        = "OsProjectName"
	OsProjectID          = "OsProjectID"
	OsProjectDomainID    = "OsProjectDomainID"
	OsInterface          = "OsInterface"
	OsRegionName         = "OsRegionName"
	OsTenantId           = "OsTenantId"
	OsTenantName         = "OsTenantName"
	OsUserName           = "OsUserName"
	OsSwiftUserName      = "OsSwiftUserName"
	OsSwiftPassword      = "OsSwiftPassword"
	OsComputeApiVersion  = "OsComputeApiVersion"
	OsNovaVersion        = "OsNovaVersion"
	OsAuthType           = "OsAuthType"
	OsCloudName          = "OsCloudName"
	OsIdentityApiVersion = "OsIdentityApiVersion"
	OsImageApiVersion    = "OsImageApiVersion"
	OsNoCache            = "OsNoCache"
	OsProjectDomainName  = "OsProjectDomainName"
	OsUserDomainName     = "OsUserDomainName"
	OsVolumeApiVersion   = "OsVolumeApiVersion"
	OsPythonWarnings     = "OsPythonWarnings"
	OsNoProxy            = "OsNoProxy"
)

var (
	SshKeyPath  = os.Getenv("PWD") + "/ansible/files/ssh_key"
	ConfigPath  = "configs/config.yaml"
	UseBasePath = true
)

var OpenstackSecretsKeys = map[string]map[string]string{
	// ussuri:
	OsUssuriVersion: {
		OsAuthUrl:            "OS_AUTH_URL",
		OsProjectName:        "OS_PROJECT_NAME",
		OsProjectID:          "OS_PROJECT_ID",
		OsInterface:          "OS_INTERFACE",
		OsPassword:           "OS_PASSWORD",
		OsRegionName:         "OS_REGION_NAME",
		OsUserName:           "OS_USERNAME",
		OsUserDomainName:     "OS_USER_DOMAIN_NAME",
		OsProjectDomainID:    "OS_PROJECT_DOMAIN_ID",
		OsIdentityApiVersion: "OS_IDENTITY_API_VERSION",
	},

	// stein:
	OsSteinVersion: {
		OsAuthUrl:            "OS_AUTH_URL",
		OsPassword:           "OS_PASSWORD",
		OsProjectName:        "OS_PROJECT_NAME",
		OsRegionName:         "OS_REGION_NAME",
		OsUserName:           "OS_USERNAME",
		OsComputeApiVersion:  "COMPUTE_API_VERSION",
		OsNovaVersion:        "NOVA_VERSION",
		OsAuthType:           "OS_AUTH_TYPE",
		OsCloudName:          "OS_CLOUDNAME",
		OsIdentityApiVersion: "OS_IDENTITY_API_VERSION",
		OsImageApiVersion:    "OS_IMAGE_API_VERSION",
		OsNoCache:            "OS_NO_CACHE",
		OsProjectDomainName:  "OS_PROJECT_DOMAIN_NAME",
		OsUserDomainName:     "OS_USER_DOMAIN_NAME",
		OsVolumeApiVersion:   "OS_VOLUME_API_VERSION",
		OsPythonWarnings:     "PYTHONWARNINGS",
		OsNoProxy:            "no_proxy",
	},

	// liberty:
	OsLibertyVersion: {
		OsAuthUrl:       "OS_AUTH_URL",
		OsPassword:      "OS_PASSWORD",
		OsProjectName:   "OS_PROJECT_NAME",
		OsRegionName:    "OS_REGION_NAME",
		OsTenantId:      "OS_TENANT_ID",
		OsTenantName:    "OS_TENANT_NAME",
		OsSwiftUserName: "OS_SWIFT_USERNAME",
		OsSwiftPassword: "OS_SWIFT_PASSWORD",
	},
}
