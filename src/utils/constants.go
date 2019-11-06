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
	ServiceTypeFanlight   string = "fanlight"

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

	//base path
	BasePath = "spark-openstack"
	//vault config file path
	VaultCfg = "vault.yaml"
	//openstack config file path
	OpenstackCfg = "openstack_config.yaml"

	//ansible main role path
	AnsibleMainRole = "src/ansible/ansible/main.yml"
	//ansible get master ip role path
	AnsibleMasterIpRole = "src/ansible/ansible/get_master.yml"
	//ansible get ip of any instance with role
	AnsibleIpRole = "src/ansible/ansible/get_ip.yml"

	//openstack secrets keys
	OsAuthUrl       = "OS_AUTH_URL"
	OsPassword      = "OS_PASSWORD"
	OsProjectName   = "OS_PROJECT_NAME"
	OsRegionName    = "OS_REGION_NAME"
	OsTenantId      = "OS_TENANT_ID"
	OsTenantName    = "OS_TENANT_NAME"
	OsUsername      = "OS_USERNAME"
	OsSwiftUsername = "OS_SWIFT_USERNAME"
	OsSwiftPassword = "OS_SWIFT_PASSWORD"

	//ssh secrets keys
	VaultSshKey = "key_bgt"

	//Entity statuses
	StatusInited   = "INITED"
	StatusCreated  = "CREATED"
	StatusFailed   = "FAILED"
	StatusStopping = "STOPPING"

	//default IDs
	CommonProjectID string = "None"
)

var (
	SshKeyPath = os.Getenv("PWD") + "/src/ansible/ansible/files/ssh_key"
)
