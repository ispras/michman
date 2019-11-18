package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/database"
	protobuf "gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/protobuf"
	"gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/utils"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const (
	jupyterPort = "8888"
)

var sparkVersions = map[string]map[string][]string{
	"2.3.0": {"hadoop_versions": {"2.6", "2.7"}},
	"2.2.1": {"hadoop_versions": {"2.6", "2.7"}},
	"2.2.0": {"hadoop_versions": {"2.6", "2.7"}},
	"2.1.0": {"hadoop_versions": {"2.3", "2.4", "2.6", "2.7"}},
	"2.0.2": {"hadoop_versions": {"2.3", "2.4", "2.6", "2.7"}},
	"2.0.1": {"hadoop_versions": {"2.3", "2.4", "2.6", "2.7"}},
	"2.0.0": {"hadoop_versions": {"2.3", "2.4", "2.6", "2.7"}},
	"1.6.2": {"hadoop_versions": {"1", "cdh4", "2.3", "2.4", "2.6"}},
	"1.6.1": {"hadoop_versions": {"1", "cdh4", "2.3", "2.4", "2.6"}},
	"1.6.0": {"hadoop_versions": {"1", "cdh4", "2.3", "2.4", "2.6"}},
	"1.5.2": {"hadoop_versions": {"1", "cdh4", "2.3", "2.4", "2.6"}},
	"1.5.1": {"hadoop_versions": {"1", "cdh4", "2.3", "2.4", "2.6"}},
	"1.5.0": {"hadoop_versions": {"1", "cdh4", "2.3", "2.4", "2.6"}},
	"1.4.1": {"hadoop_versions": {"1", "cdh4", "2.3", "2.4", "2.6"}},
	"1.4.0": {"hadoop_versions": {"1", "cdh4", "2.3", "2.4", "2.6"}},
	"1.3.1": {"hadoop_versions": {"1", "cdh4", "2.3", "2.4", "2.6"}},
	"1.3.0": {"hadoop_versions": {"1", "cdh4", "2.3", "2.4"}},
	"1.2.2": {"hadoop_versions": {"1", "cdh4", "2.3", "2.4"}},
	"1.2.1": {"hadoop_versions": {"1", "cdh4", "2.3", "2.4"}},
	"1.2.0": {"hadoop_versions": {"1", "cdh4", "2.3", "2.4"}},
	"1.1.1": {"hadoop_versions": {"1", "cdh4", "2.3", "2.4"}},
	"1.1.0": {"hadoop_versions": {"1", "cdh4", "2.3", "2.4"}},
	"1.0.2": {"hadoop_versions": {"1", "cdh4"}},
	"1.0.1": {"hadoop_versions": {"1", "cdh4"}},
	"1.0.0": {"hadoop_versions": {"1", "cdh4"}},
}

var toreeVersions = map[string]string{
	"1": "https://www.apache.org/dist/incubator/toree/0.1.0-incubating/toree-pip/apache-toree-0.1.0.tar.gz",
	"2": "https://www.apache.org/dist/incubator/toree/0.2.0-incubating/toree-pip/toree-0.2.0.tar.gz",
	"3": "https://www.apache.org/dist/incubator/toree/0.3.0-incubating/toree-pip/toree-0.3.0.tar.gz",
}

type ServiceExists struct {
	exists  bool
	service *protobuf.Service
}

type AnsibleExtraVars struct {
	IgniteVersion            string              `json:"ignite_version"`
	EsHeapSize               string              `json:"es_heap_size"`
	CreateCluster            bool                `json:"create_cluster"`
	DeployCassandra          bool                `json:"deploy_cassandra"`
	Sync                     string              `json:"sync"`
	AnsibleUser              string              `json:"ansible_user"`
	DeploySpark              bool                `json:"deploy_spark"`
	DeployElastic            bool                `json:"deploy_elastic"`
	Mountnfs                 bool                `json:"mountnfs"`
	Flavor                   string              `json:"flavor"`
	BootFromVolume           bool                `json:"boot_from_volume"`
	DeployJupyter            bool                `json:"deploy_jupyter"`
	ToreeVersion             string              `json:"toree_version,omitempty"`
	HadoopUser               string              `json:"hadoop_user"`
	MasterFlavor             string              `json:"master_flavor"`
	NSlaves                  int32               `json:"n_slaves"`
	DeployIgnite             bool                `json:"deploy_ignite"`
	ClusterName              string              `json:"cluster_name"`
	SparkVersion             string              `json:"spark_version"`
	OsImage                  string              `json:"os_image"`
	SkipPackages             bool                `json:"skip_packages"`
	OsProjectName            string              `json:"os_project_name"`
	NfsShares                []string            `json:"nfs_shares"` //check if type is correct
	UseYarn                  bool                `json:"use_yarn"`
	FloatingIpPool           string              `json:"floating_ip_pool"`
	OsAuthUrl                string              `json:"os_auth_url"`
	UseOracleJava            bool                `json:"use_oracle_java"`
	AnsibleSshPrivateKeyFile string              `json:"ansible_ssh_private_key_file"`
	HadoopVersion            string              `json:"hadoop_version"`
	CassandraVersion         string              `json:"cassandra_version"`
	ExtraJars                []map[string]string `json:"extra_jars"`
	Act                      string              `json:"act"`
	VirtualNetwork           string              `json:"virtual_network"`
	OsKeyName                string              `json:"os_key_name"`
	DeployJupyterhub         bool                `json:"deploy_jupyterhub"`
	OsSwiftUserName          string              `json:"os_swift_user_name,omitempty"`
	OsSwiftPassword          string              `json:"os_swift_password,omitempty"`
	SparkWorkerMemMb         int                 `json:"spark_worker_mem_mb,omitempty"`
	IgniteMemory             int                 `json:"ignite_memory,omitempty"`
	YarnMasterMemMb          int                 `json:"yarn_master_mem_mb,omitempty"`
	DeployFanlight           bool                `json:"create_fanlight"`
	FanlightInstanceUrl      string              `json:"fanlight_instance_url"`
	DesktopAccessUrl         string              `json:"desktop_access_url"`
	DeployNFS                bool                `json:"create_storage"`
	UsersAdd                 string              `json:"users_add"`
	AppsAdd                  string              `json:"apps_add"`
	CustomOidcProvidersHost  string              `json:"custom_oidc_providers_host"`
	CustomOidcProvidersIP    string              `json:"custom_oidc_providers_ip"`
	UseMirror                string              `json:"use_mirror"`
	MirrorAddress            string              `json:"mirror_address"`
}

func GetElasticConnectorJar() string {
	elasticHadoopUrl := "http://download.elastic.co/hadoop/elasticsearch-hadoop-5.5.0.zip"
	elasticHadoopFilename := filepath.Join("/tmp", filepath.Base(elasticHadoopUrl))
	elasticDir := filepath.Join("/tmp", "elasticsearch-hadoop/")
	archivePath := "elasticsearch-hadoop-5.5.0/dist/elasticsearch-hadoop-5.5.0.jar"
	elasticPath := filepath.Join(elasticDir, archivePath)
	if _, err := os.Stat(elasticPath); err != nil {
		if os.IsNotExist(err) {
			// file does not exist
			log.Print("Downloading ElasticSearch Hadoop integration")
			utils.DownloadFile(elasticHadoopUrl, elasticHadoopFilename)

			if _, err := utils.Unzip(elasticHadoopFilename, elasticDir); err != nil {
				log.Print(err)
			}
		}
	}
	return elasticPath
}

func GetCassandraConnectorJar(sparkVersion string) string {
	var sparkCassandraConnectorUrl string
	if strings.HasPrefix(sparkVersion, "1.6") {
		sparkCassandraConnectorUrl = "http://dl.bintray.com/spark-packages/maven/datastax/spark-cassandra-connector/1.6.8-s_2.10/spark-cassandra-connector-1.6.8-s_2.10.jar"
	} else {
		sparkCassandraConnectorUrl = "http://dl.bintray.com/spark-packages/maven/datastax/spark-cassandra-connector/2.0.3-s_2.11/spark-cassandra-connector-2.0.3-s_2.11.jar"
	}
	sparkCassandraConnectorFile := filepath.Join("/tmp", filepath.Base(sparkCassandraConnectorUrl))

	//checking if file exists
	if _, err := os.Stat(sparkCassandraConnectorFile); err != nil {
		if os.IsNotExist(err) {
			// file does not exist
			log.Print("Downloading Spark Cassandra Connector for Spark version ", sparkVersion)
			utils.DownloadFile(sparkCassandraConnectorFile, sparkCassandraConnectorUrl)
		}
	}

	return sparkCassandraConnectorFile
}

func AddJar(path string) map[string]string {
	var absPath string
	if v, err := filepath.Abs(path); err != nil {
		log.Fatalln(err)
	} else {
		absPath = v
	}
	var newElem = map[string]string{
		"name": filepath.Base(path), "path": absPath,
	}
	return newElem
}

func MakeExtraVars(cluster *protobuf.Cluster, osCreds *utils.OsCredentials, osConfig *utils.OsConfig, action string) AnsibleExtraVars {
	//available services types
	var serviceTypes = map[string]ServiceExists{
		utils.ServiceTypeCassandra: {
			exists:  false,
			service: nil,
		},
		utils.ServiceTypeSpark: {
			exists:  false,
			service: nil,
		},
		utils.ServiceTypeElastic: {
			exists:  false,
			service: nil,
		},
		utils.ServiceTypeJupyter: {
			exists:  false,
			service: nil,
		},
		utils.ServiceTypeIgnite: {
			exists:  false,
			service: nil,
		},
		utils.ServiceTypeJupyterhub: {
			exists:  false,
			service: nil,
		},
		utils.ServiceTypeFanlight: {
			exists:  false,
			service: nil,
		},
		utils.ServiceTypeNFS: {
			exists:  false,
			service: nil,
		},
	}

	//iterating over services for looking, which services are presented
	for _, service := range cluster.Services {
		if _, ok := serviceTypes[service.Type]; ok {
			serviceTypes[service.Type] = ServiceExists{
				exists:  true,
				service: service,
			}
		}
	}

	var extraVars AnsibleExtraVars

	//must be True in method "/clusters" POST, else False
	extraVars.CreateCluster = false
	if action == actionCreate {
		extraVars.CreateCluster = true
	}

	//filling services
	extraVars.DeployCassandra = serviceTypes[utils.ServiceTypeCassandra].exists
	extraVars.DeploySpark = serviceTypes[utils.ServiceTypeSpark].exists
	extraVars.DeployElastic = serviceTypes[utils.ServiceTypeElastic].exists
	extraVars.DeployJupyter = serviceTypes[utils.ServiceTypeJupyter].exists
	extraVars.DeployIgnite = serviceTypes[utils.ServiceTypeIgnite].exists
	extraVars.DeployJupyterhub = serviceTypes[utils.ServiceTypeJupyterhub].exists
	extraVars.DeployFanlight = serviceTypes[utils.ServiceTypeFanlight].exists
	extraVars.DeployNFS = serviceTypes[utils.ServiceTypeNFS].exists

	//must be always async mode
	extraVars.Sync = "async"
	extraVars.AnsibleUser = "ubuntu"

	extraVars.IgniteVersion = "2.7.5"
	if serviceTypes[utils.ServiceTypeIgnite].exists && serviceTypes[utils.ServiceTypeIgnite].service.Version != "" {
		extraVars.IgniteVersion = serviceTypes[utils.ServiceTypeIgnite].service.Version
	}

	extraVars.EsHeapSize = "1g"
	if serviceTypes[utils.ServiceTypeElastic].exists && serviceTypes[utils.ServiceTypeElastic].service.Config != nil {
		if size, ok := serviceTypes[utils.ServiceTypeElastic].service.Config[utils.ElasticHeapSize]; ok {
			extraVars.EsHeapSize = size
		}
	}

	extraVars.Mountnfs = false
	extraVars.Flavor = osConfig.Flavor
	extraVars.BootFromVolume = false

	extraVars.HadoopUser = "ubuntu"
	extraVars.MasterFlavor = osConfig.Flavor
	extraVars.NSlaves = cluster.NHosts

	extraVars.ClusterName = cluster.Name

	extraVars.SparkVersion = "1.6.2"
	if serviceTypes[utils.ServiceTypeSpark].exists && serviceTypes[utils.ServiceTypeSpark].service.Version != "" {
		extraVars.SparkVersion = serviceTypes[utils.ServiceTypeSpark].service.Version
	}

	extraVars.OsImage = osConfig.OsImage
	extraVars.SkipPackages = false
	extraVars.OsProjectName = osCreds.OsProjectName
	extraVars.NfsShares = []string{}

	extraVars.UseYarn = false
	//getting latest hadoop version for selected spark version
	hadoopVersions := sparkVersions[extraVars.SparkVersion]["hadoop_versions"]
	extraVars.HadoopVersion = hadoopVersions[len(hadoopVersions)-1]
	//checking spark config params
	if serviceTypes[utils.ServiceTypeSpark].exists && serviceTypes[utils.ServiceTypeSpark].service.Config != nil {
		if yarn, ok := serviceTypes[utils.ServiceTypeSpark].service.Config[utils.SparkUseYarn]; ok {
			b, err := strconv.ParseBool(yarn)
			if err != nil {
				log.Fatalln(err)
			}
			extraVars.UseYarn = b
			extraVars.YarnMasterMemMb = 10240 //change it
		}
		if version, ok := serviceTypes[utils.ServiceTypeSpark].service.Config[utils.SparkHadoopVersion]; ok {
			hadoopVersions := sparkVersions[extraVars.SparkVersion]["hadoop_versions"]
			versionOk := false
			for _, v := range hadoopVersions {
				if v == version {
					extraVars.HadoopVersion = version
					versionOk = true
				}
			}
			if !versionOk {
				log.Print("Bad Hadoop version in Spark config")
				extraVars.HadoopVersion = hadoopVersions[len(hadoopVersions)-1]
			}
		}

		if mem, ok := serviceTypes[utils.ServiceTypeSpark].service.Config[utils.SparkWorkerMemMb]; ok {
			memInt, err := strconv.Atoi(mem)
			if err != nil {
				log.Fatalln(err)
			}
			extraVars.SparkWorkerMemMb = memInt
		}
		if mem, ok := serviceTypes[utils.ServiceTypeSpark].service.Config[utils.SparkYarnMasterMemMb]; ok {
			memInt, err := strconv.Atoi(mem)
			if err != nil {
				log.Fatalln(err)
			}
			extraVars.YarnMasterMemMb = memInt
		}
	}

	extraVars.FloatingIpPool = osConfig.FloatingIP
	extraVars.OsAuthUrl = osCreds.OsAuthUrl
	extraVars.UseOracleJava = false //must be always false
	extraVars.AnsibleSshPrivateKeyFile = utils.SshKeyPath

	extraVars.CassandraVersion = utils.CassandraDefaultVersion
	if serviceTypes[utils.ServiceTypeCassandra].exists && serviceTypes[utils.ServiceTypeCassandra].service.Version != "" {
		extraVars.CassandraVersion = serviceTypes[utils.ServiceTypeCassandra].service.Version
	}

	if serviceTypes[utils.ServiceTypeJupyter].exists && serviceTypes[utils.ServiceTypeJupyter].service.Config != nil {
		if version, ok := serviceTypes[utils.ServiceTypeJupyter].service.Config[utils.JupyterToreeVersion]; ok {
			if v, ok := toreeVersions[version]; ok {
				extraVars.ToreeVersion = v
			} else {
				log.Print("Bad Toree version in Jupyter config")
				extraVars.ToreeVersion = toreeVersions[string(extraVars.SparkVersion[0])]
			}
		}
	} else if serviceTypes[utils.ServiceTypeJupyter].exists {
		extraVars.ToreeVersion = toreeVersions[string(extraVars.SparkVersion[0])]
	}

	if serviceTypes[utils.ServiceTypeIgnite].exists && serviceTypes[utils.ServiceTypeIgnite].service.Config != nil {
		if mem, ok := serviceTypes[utils.ServiceTypeIgnite].service.Config[utils.IgniteMemory]; ok {
			memInt, err := strconv.Atoi(mem)
			if err != nil {
				log.Fatalln(err)
			}
			extraVars.IgniteMemory = memInt
		}
	}

	//action must be "launch" in method "/clusters" POST and /clusters/{clusterName} PUT
	//action must be "destroy" in method /clusters/{clusterName} DELETE
	if action == actionCreate || action == actionUpdate {
		extraVars.Act = utils.AnsibleLaunch
	} else if action == actionDelete {
		extraVars.Act = utils.AnsibleDestroy
	}

	extraVars.VirtualNetwork = osConfig.VirtualNetwork
	extraVars.OsKeyName = osConfig.Key

	extraVars.OsSwiftUserName = osCreds.OsSwiftUserName
	extraVars.OsSwiftPassword = osCreds.OsSwiftPassword

	//make extra jars
	var extraJars []map[string]string
	if extraVars.DeployCassandra {
		cassandraJar := GetCassandraConnectorJar(extraVars.SparkVersion)
		extraJars = append(extraJars, AddJar(cassandraJar))
	}

	if extraVars.DeployElastic {
		elasticJar := GetElasticConnectorJar()
		extraJars = append(extraJars, AddJar(elasticJar))
	}

	extraVars.ExtraJars = extraJars
	if extraVars.ExtraJars == nil {
		extraVars.ExtraJars = []map[string]string{}
	}

	//check fanlight config
	if serviceTypes[utils.ServiceTypeFanlight].exists && serviceTypes[utils.ServiceTypeFanlight].service.Config != nil {
		if fInstanceUrl, ok := serviceTypes[utils.ServiceTypeFanlight].service.Config[utils.FanlightInstanceUrl]; ok {
			extraVars.FanlightInstanceUrl = fInstanceUrl
		}
		if desktopUrl, ok := serviceTypes[utils.ServiceTypeFanlight].service.Config[utils.FanlightDesktopAccessUrl]; ok {
			extraVars.DesktopAccessUrl = desktopUrl
		}
		if usersAdd, ok := serviceTypes[utils.ServiceTypeFanlight].service.Config[utils.FanlightUsersAdd]; ok {
			extraVars.UsersAdd = usersAdd
		}
		if appssAdd, ok := serviceTypes[utils.ServiceTypeFanlight].service.Config[utils.FanlightAppsAdd]; ok {
			extraVars.AppsAdd = appssAdd
		}
	}

	//check nextcloud config
	if serviceTypes[utils.ServiceTypeNFS].exists && serviceTypes[utils.ServiceTypeNFS].service.Config != nil {
		if customOidcHost, ok := serviceTypes[utils.ServiceTypeNFS].service.Config[utils.NFSCustomOidcProvidersHost]; ok {
			extraVars.CustomOidcProvidersHost = customOidcHost
		}
		if customOidcIP, ok := serviceTypes[utils.ServiceTypeNFS].service.Config[utils.NFSCustomOidcProvidersIP]; ok {
			extraVars.CustomOidcProvidersIP = customOidcIP
		}
	}

	// load mirror config
	mirrorC, err := utils.GetMirrorConfig()
	if err != nil {
		log.Fatalln(err)
	}
	extraVars.UseMirror = mirrorC.Enable
	enable, err := strconv.ParseBool(mirrorC.Enable)
	if enable && !validateIP(mirrorC.Address) {
		log.Fatalln("ERROR: bad mirror's IP address")
	}

	extraVars.MirrorAddress = mirrorC.Address

	return extraVars
}

type AnsibleLauncher struct {
	couchbaseCommunicator database.Database
}

func validateIP(input string) bool {
	pattern := "^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$"
	regEx := regexp.MustCompile(pattern)
	fmt.Println(input)
	return regEx.FindString(input) != ""
}

func findIP(input string) string {
	numBlock := "(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])"
	regexPattern := numBlock + "\\." + numBlock + "\\." + numBlock + "\\." + numBlock

	regEx := regexp.MustCompile(regexPattern)
	return regEx.FindString(input)
}

func setOsVars(osCreds *utils.OsCredentials, version string) error {
	switch version {
	case utils.OsSteinVersion:
		err := os.Setenv(utils.OsAuthUrl, osCreds.OsAuthUrl)
		if err != nil {
			log.Fatalln(err)
		}

		err = os.Setenv(utils.OsProjectName, osCreds.OsProjectName)
		if err != nil {
			log.Fatalln(err)
		}
		err = os.Setenv(utils.OsUsername, osCreds.OsUserName)
		if err != nil {
			log.Fatalln(err)
		}
		err = os.Setenv(utils.OsPassword, osCreds.OsPassword)
		if err != nil {
			log.Fatalln(err)
		}
		err = os.Setenv(utils.OsRegionName, osCreds.OsRegionName)
		if err != nil {
			log.Fatalln(err)
		}

		err = os.Setenv(utils.OsIdentityApiVersion, osCreds.OsIdentityApiVersion)
		if err != nil {
			log.Fatalln(err)
		}

		err = os.Setenv(utils.OsImageApiVersion, osCreds.OsImageApiVersion)
		if err != nil {
			log.Fatalln(err)
		}

		err = os.Setenv(utils.OsNoCache, osCreds.OsNoCache)
		if err != nil {
			log.Fatalln(err)
		}

		err = os.Setenv(utils.OsProjectDomainName, osCreds.OsProjectDomainName)
		if err != nil {
			log.Fatalln(err)
		}

		err = os.Setenv(utils.OsUserDomainName, osCreds.OsUserDomainName)
		if err != nil {
			log.Fatalln(err)
		}

		err = os.Setenv(utils.OsAuthType, osCreds.OsAuthType)
		if err != nil {
			log.Fatalln(err)
		}

		err = os.Setenv(utils.OsCloudname, osCreds.OsCloudname)
		if err != nil {
			log.Fatalln(err)
		}

		err = os.Setenv(utils.OsNovaVersion, osCreds.OsNovaVersion)
		if err != nil {
			log.Fatalln(err)
		}
		err = os.Setenv(utils.OsComputeApiVersion, osCreds.OsComputeApiVersion)
		if err != nil {
			log.Fatalln(err)
		}

		err = os.Setenv(utils.OsNoProxy, osCreds.OsNoProxy)
		if err != nil {
			log.Fatalln(err)
		}

		err = os.Setenv(utils.OsVolumeApiVersion, osCreds.OsVolumeApiVersion)
		if err != nil {
			log.Fatalln(err)
		}

		err = os.Setenv(utils.OsPythonwarnings, osCreds.OsPythonwarnings)
		if err != nil {
			log.Fatalln(err)
		}
	case utils.OsLibertyVersion:
		err := os.Setenv(utils.OsAuthUrl, osCreds.OsAuthUrl)
		if err != nil {
			log.Fatalln(err)
		}

		err = os.Setenv(utils.OsProjectName, osCreds.OsProjectName)
		if err != nil {
			log.Fatalln(err)
		}
		err = os.Setenv(utils.OsUsername, osCreds.OsUserName)
		if err != nil {
			log.Fatalln(err)
		}
		err = os.Setenv(utils.OsPassword, osCreds.OsPassword)
		if err != nil {
			log.Fatalln(err)
		}
		err = os.Setenv(utils.OsRegionName, osCreds.OsRegionName)
		if err != nil {
			log.Fatalln(err)
		}

		err = os.Setenv(utils.OsTenantId, osCreds.OsTenantId)
		if err != nil {
			log.Fatalln(err)
		}
		err = os.Setenv(utils.OsTenantName, osCreds.OsTenantName)
		if err != nil {
			log.Fatalln(err)
		}

		if osCreds.OsSwiftUserName != "" {
			err = os.Setenv(utils.OsSwiftUsername, osCreds.OsSwiftUserName)
			if err != nil {
				log.Fatalln(err)
			}
		}

		if osCreds.OsSwiftPassword != "" {
			err = os.Setenv(utils.OsSwiftPassword, osCreds.OsSwiftPassword)
			if err != nil {
				log.Fatalln(err)
			}
		}
	default: //liberty as default version
		err := os.Setenv(utils.OsAuthUrl, osCreds.OsAuthUrl)
		if err != nil {
			log.Fatalln(err)
		}

		err = os.Setenv(utils.OsProjectName, osCreds.OsProjectName)
		if err != nil {
			log.Fatalln(err)
		}
		err = os.Setenv(utils.OsUsername, osCreds.OsUserName)
		if err != nil {
			log.Fatalln(err)
		}
		err = os.Setenv(utils.OsPassword, osCreds.OsPassword)
		if err != nil {
			log.Fatalln(err)
		}
		err = os.Setenv(utils.OsRegionName, osCreds.OsRegionName)
		if err != nil {
			log.Fatalln(err)
		}

		err = os.Setenv(utils.OsTenantId, osCreds.OsTenantId)
		if err != nil {
			log.Fatalln(err)
		}
		err = os.Setenv(utils.OsTenantName, osCreds.OsTenantName)
		if err != nil {
			log.Fatalln(err)
		}

		if osCreds.OsSwiftUserName != "" {
			err = os.Setenv(utils.OsSwiftUsername, osCreds.OsSwiftUserName)
			if err != nil {
				log.Fatalln(err)
			}
		}

		if osCreds.OsSwiftPassword != "" {
			err = os.Setenv(utils.OsSwiftPassword, osCreds.OsSwiftPassword)
			if err != nil {
				log.Fatalln(err)
			}
		}
	}

	return nil
}

func (aL AnsibleLauncher) Run(cluster *protobuf.Cluster, osCreds *utils.OsCredentials, osConfig *utils.OsConfig, action string) string {
	log.SetPrefix("ANSIBLE_LAUNCHER: ")

	// creating ansible-playbook commands according to cluster object

	//exporting ansible variables
	err := setOsVars(osCreds, osConfig.OsVersion)
	if err != nil {
		log.Fatalln(err)
	}

	//constructing ansible-playbook command
	extraVars := MakeExtraVars(cluster, osCreds, osConfig, action)
	ansibleArgs, err := json.Marshal(extraVars)
	if err != nil {
		log.Fatalln(err)
	}

	cmdName := utils.AnsiblePlaybookCmd
	cmdArgs := []string{"-vvv", utils.AnsibleMainRole, "--extra-vars", string(ansibleArgs)}

	//saving cluster to database
	log.Print("Writing new cluster to db...")
	err = aL.couchbaseCommunicator.WriteCluster(cluster)
	if err != nil {
		log.Fatalln(err)
	}

	log.Print("Running ansible...")

	// create output log
	f, err := os.Create("logs/ansible_output.log")

	defer f.Close()
	ansibleCmd := exec.Command(cmdName, cmdArgs...)
	stdout, err := ansibleCmd.StdoutPipe()
	if err != nil {
		log.Fatalln(err)
	}
	stderr, err := ansibleCmd.StderrPipe()
	if err != nil {
		log.Fatalln(err)
	}

	stdoutScanner := bufio.NewScanner(stdout)
	stderrScanner := bufio.NewScanner(stderr)
	go func() {
		for stdoutScanner.Scan() {
			_, err := f.WriteString(stdoutScanner.Text() + "\n")
			if err != nil {
				log.Fatalln(err)
			}
		}
	}()
	go func() {
		for stderrScanner.Scan() {
			_, err = f.WriteString(stderrScanner.Text() + "\n")
			if err != nil {
				log.Fatalln(err)
			}
		}
	}()

	ansibleOk := true
	if err := ansibleCmd.Start(); err != nil {
		ansibleOk = false
		log.Print("Error: ", err)
	}

	if err := ansibleCmd.Wait(); err != nil {
		ansibleOk = false
		log.Print("Error: ", err)
	}

	//Get Master IP for Cluster create or update action and save it
	if ansibleOk && (action == actionCreate || action == actionUpdate) {

		var v = map[string]string{
			"cluster_name": cluster.Name,
		}

		ipExtraVars, err := json.Marshal(v)
		if err != nil {
			log.Fatalln(err)
		}

		cmdName := utils.AnsiblePlaybookCmd
		args := []string{"-v", utils.AnsibleMasterIpRole, "--extra-vars", string(ipExtraVars)}

		log.Print("Running ansible for getting master IP...")
		cmd := exec.Command(cmdName, args...)
		var outb bytes.Buffer
		cmd.Stdout = &outb

		if err := cmd.Start(); err != nil {
			log.Print("Error: ", err)
		}

		if err := cmd.Wait(); err != nil {
			ansibleOk = false
			log.Print("Error: ", err)
		}

		masterIp := findIP(outb.String())
		fanlightIp := ""
		nfsIp := ""
		if extraVars.DeployFanlight {
			v = map[string]string{
				"cluster_name":  cluster.Name,
				"extended_role": "fanlight",
			}
			ipExtraVars, err = json.Marshal(v)
			if err != nil {
				log.Fatalln(err)
			}
			cmdName = utils.AnsiblePlaybookCmd
			args = []string{"-v", utils.AnsibleIpRole, "--extra-vars", string(ipExtraVars)}
			log.Print("Running ansible for getting fanlight IP...")
			cmd := exec.Command(cmdName, args...)
			var outb bytes.Buffer
			cmd.Stdout = &outb
			if err := cmd.Start(); err != nil {
				log.Print("Error: ", err)
			}

			if err := cmd.Wait(); err != nil {
				ansibleOk = false
				log.Print("Error: ", err)
			}
			fanlightIp = findIP(outb.String())
		}
		if extraVars.DeployNFS {
			v = map[string]string{
				"cluster_name":  cluster.Name,
				"extended_role": "storage",
			}
			ipExtraVars, err = json.Marshal(v)
			if err != nil {
				log.Fatalln(err)
			}
			cmdName = utils.AnsiblePlaybookCmd
			args = []string{"-v", utils.AnsibleIpRole, "--extra-vars", string(ipExtraVars)}
			log.Print("Running ansible for getting NFS server IP...")
			cmd := exec.Command(cmdName, args...)
			var outb bytes.Buffer
			cmd.Stdout = &outb
			if err := cmd.Start(); err != nil {
				log.Print("Error: ", err)
			}

			if err := cmd.Wait(); err != nil {
				ansibleOk = false
				log.Print("Error: ", err)
			}
			nfsIp = findIP(outb.String())
		}
		//filling services URLs:
		if masterIp != "" {
			log.Print("Master IP is: ", masterIp)
			cluster.MasterIP = masterIp

			for i, service := range cluster.Services {
				if service.Type == utils.ServiceTypeJupyter {
					cluster.Services[i].URL = masterIp + ":" + jupyterPort
				}
			}

			log.Print("Saving master IP...")
			err = aL.couchbaseCommunicator.WriteCluster(cluster)
			if err != nil {
				log.Fatalln(err)
			}
		} else {
			log.Print("There is no IP in Ansible output")
		}
		if fanlightIp != "" {
			log.Print("Fanlight IP is: ", fanlightIp)
			for i, service := range cluster.Services {
				if service.Type == utils.ServiceTypeFanlight {
					cluster.Services[i].ServiceURL = fanlightIp
				}
			}
			log.Print("Saving fanlight IP...")
			err = aL.couchbaseCommunicator.WriteCluster(cluster)
			if err != nil {
				log.Fatalln(err)
			}
		}
		if nfsIp != "" {
			log.Print("NFS server IP is: ", nfsIp)
			for i, service := range cluster.Services {
				if service.Type == utils.ServiceTypeNFS {
					cluster.Services[i].ServiceURL = nfsIp
				}
			}
			log.Print("Saving NFS server IP...")
			err = aL.couchbaseCommunicator.WriteCluster(cluster)
			if err != nil {
				log.Fatalln(err)
			}
		}

	}

	if ansibleOk {
		log.Print("Launch: OK")
		return utils.AnsibleOk
	} else {
		log.Print("Ansible has failed, check logs for mor information.")
		return utils.AnsibleFail
	}
}