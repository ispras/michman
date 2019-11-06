package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	protobuf "gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/protobuf"
	"gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/utils"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/database"
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

	return extraVars
}

type AnsibleLauncher struct{
	couchbaseCommunicator database.Database
}

func findIP(input string) string {
	numBlock := "(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])"
	regexPattern := numBlock + "\\." + numBlock + "\\." + numBlock + "\\." + numBlock

	regEx := regexp.MustCompile(regexPattern)
	return regEx.FindString(input)
}

func (aL AnsibleLauncher) Run(cluster *protobuf.Cluster, osCreds *utils.OsCredentials, osConfig *utils.OsConfig, action string) string {
	log.SetPrefix("ANSIBLE_LAUNCHER: ")

	// creating ansible-playbook commands according to cluster object

	//exporting ansible variables
	err := os.Setenv("OS_AUTH_URL", osCreds.OsAuthUrl)
	if err != nil {
		log.Fatalln(err)
	}
	err = os.Setenv("OS_TENANT_ID", osCreds.OsTenantId)
	if err != nil {
		log.Fatalln(err)
	}
	err = os.Setenv("OS_TENANT_NAME", osCreds.OsTenantName)
	if err != nil {
		log.Fatalln(err)
	}
	err = os.Setenv("OS_PROJECT_NAME", osCreds.OsProjectName)
	if err != nil {
		log.Fatalln(err)
	}
	err = os.Setenv("OS_USERNAME", osCreds.OsUserName)
	if err != nil {
		log.Fatalln(err)
	}
	err = os.Setenv("OS_PASSWORD", osCreds.OsPassword)
	if err != nil {
		log.Fatalln(err)
	}
	err = os.Setenv("OS_REGION_NAME", osCreds.OsRegionName)
	if err != nil {
		log.Fatalln(err)
	}

	if osCreds.OsSwiftUserName != "" {
		err = os.Setenv("OS_SWIFT_USERNAME", osCreds.OsSwiftUserName)
		if err != nil {
			log.Fatalln(err)
		}
	}

	if osCreds.OsSwiftPassword != "" {
		err = os.Setenv("OS_SWIFT_PASSWORD", osCreds.OsSwiftPassword)
		if err != nil {
			log.Fatalln(err)
		}
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
	// retry on command fail
	retries := cluster.NHosts + 1
	success := false
	for retry := int32(1); !success && retry <= retries; retry++ {
		log.Printf("Try %v of %v", retry, retries)
		ansibleCmd := exec.Command(cmdName, cmdArgs...)
		stdout, err := ansibleCmd.StdoutPipe()
		if err != nil {
			log.Fatalln(err)
		}

		scanner := bufio.NewScanner(stdout)

		go func() {
			for scanner.Scan() {
				_, err := f.WriteString(scanner.Text() + "\n")
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
		if action == actionCreate || action == actionUpdate {
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

			ip := findIP(outb.String())
			if ip != "" {
				log.Print("Master IP is: ", ip)
				cluster.MasterIP = ip

				//filling services URLs:

				for i, service := range cluster.Services {
					if service.Type == utils.ServiceTypeJupyter {
						cluster.Services[i].URL = ip + ":" + jupyterPort
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

		}

		if ansibleOk {
			log.Print("Launch: OK")
			success = true
		} else {
			log.Print("Ansible has failed, check logs for mor information.")
		}
	}
	if success {
		return utils.AnsibleOk
	} else {
		return utils.AnsibleFail
	}
}