package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"bufio"
	protobuf "gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/protobuf"
	"path/filepath"
	"strconv"
	"strings"
)

var sparkVersions = map[string]map[string][]string {
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

var toreeVersions = map[string]string {
	"1" : "https://www.apache.org/dist/incubator/toree/0.1.0-incubating/toree-pip/apache-toree-0.1.0.tar.gz",
	"2" : "https://www.apache.org/dist/incubator/toree/0.2.0-incubating/toree-pip/toree-0.2.0.tar.gz",
	"3" : "https://www.apache.org/dist/incubator/toree/0.3.0-incubating/toree-pip/toree-0.3.0.tar.gz",
}

type ServiceExists struct {
	exists bool
	service *protobuf.Service
}


type AnsibleExtraVars struct {
	IgniteVersion string `json:"ignite_version"`
	EsHeapSize string `json:"es_heap_size"`
	CreateCluster bool `json:"create_cluster"`
	DeployCassandra bool `json:"deploy_cassandra"`
	Sync string `json:"sync"`
	AnsibleUser string `json:"ansible_user"`
	DeploySpark bool `json:"deploy_spark"`
	DeployElastic bool `json:"deploy_elastic"`
	Mountnfs bool `json:"mountnfs"`
	Flavor string `json:"flavor"`
	BootFromVolume bool `json:"boot_from_volume"`
	DeployJupyter bool `json:"deploy_jupyter"`
	ToreeVersion string `json:"toree_version,omitempty"`
	HadoopUser string `json:"hadoop_user"`
	MasterFlavor string `json:"master_flavor"`
	NSlaves int32 `json:"n_slaves"`
	DeployIgnite bool `json:"deploy_ignite"`
	ClusterName string `json:"cluster_name"`
	SparkVersion string `json:"spark_version"`
	OsImage string `json:"os_image"`
	SkipPackages bool `json:"skip_packages"`
	OsProjectName string `json:"os_project_name"`
	NfsShares []string `json:"nfs_shares"` //check if type is correct
	UseYarn bool `json:"use_yarn"`
	FloatingIpPool string `json:"floating_ip_pool"`
	OsAuthUrl string `json:"os_auth_url"`
	UseOracleJava bool `json:"use_oracle_java"`
	AnsibleSshPrivateKeyFile string `json:"ansible_ssh_private_key_file"`
	HadoopVersion string `json:"hadoop_version"`
	CassandraVersion string `json:"cassandra_version"`
	ExtraJars []map[string]string `json:"extra_jars"`
	Act string `json:"act"`
	VirtualNetwork string `json:"virtual_network"`
	OsKeyName string `json:"os_key_name"`
	DeployJupyterhub bool `json:"deploy_jupyterhub"`
	OsSwiftUserName string `json:"os_swift_user_name,omitempty"`
	OsSwiftPassword string `json:"os_swift_password,omitempty"`
	SparkWorkerMemMb int `json:"spark_worker_mem_mb,omitempty"`
	IgniteMemory int `json:"ignite_memory,omitempty"`
	YarnMasterMemMb int `json:"yarn_master_mem_mb,omitempty"`
}

func downloadFile(filepath string, url string) (err error) {

	// Create the file
	out, err := os.Create(filepath)
	if err != nil  {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil  {
		return err
	}

	return nil
}

func Unzip(src string, dest string) ([]string, error) {

	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
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
			downloadFile(elasticHadoopUrl, elasticHadoopFilename)

			if _, err := Unzip(elasticHadoopFilename, elasticDir); err != nil {
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
			downloadFile(sparkCassandraConnectorFile, sparkCassandraConnectorUrl)
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
	var newElem = map[string]string {
		"name": filepath.Base(path), "path": absPath,
	}
	return newElem
}

func MakeExtraVars(cluster *protobuf.Cluster, osCreds *osCredentials, osConfig *osConfig) AnsibleExtraVars {
	//available services types
	var serviceTypes = map[string]ServiceExists {
		"cassandra": {
			exists:  false,
			service: nil,
		},
		"spark": {
			exists: false,
			service: nil,
		},
		"elastic": {
			exists: false,
			service: nil,
		},
		"jupyter": {
			exists: false,
			service: nil,
		},
		"ignite": {
			exists: false,
			service: nil,
		},
		"jupyterhub": {
			exists: false,
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
	if cluster.EntityStatus == protobuf.Cluster_INITED {
		extraVars.CreateCluster = true
	}

	//filling services
	extraVars.DeployCassandra = serviceTypes["cassandra"].exists
	extraVars.DeploySpark = serviceTypes["spark"].exists
	extraVars.DeployElastic = serviceTypes["elastic"].exists
	extraVars.DeployJupyter = serviceTypes["jupyter"].exists
	extraVars.DeployIgnite = serviceTypes["ignite"].exists
	extraVars.DeployJupyterhub = serviceTypes["jupyterhub"].exists

	//must be always async mode
	extraVars.Sync = "async"
	extraVars.AnsibleUser = "ubuntu"

	extraVars.IgniteVersion = "2.7.5"
	if serviceTypes["ignite"].exists && serviceTypes["ignite"].service.Version != "" {
		extraVars.IgniteVersion = serviceTypes["ignite"].service.Version
	}

	extraVars.EsHeapSize = "1g"
	if serviceTypes["elastic"].exists && serviceTypes["elastic"].service.Config != nil {
		if size, ok := serviceTypes["elastic"].service.Config["es-heap-size"]; ok {
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
	if serviceTypes["spark"].exists && serviceTypes["spark"].service.Version != "" {
		extraVars.SparkVersion = serviceTypes["spark"].service.Version
	}

	extraVars.OsImage = osConfig.OsImage
	extraVars.SkipPackages = false
	extraVars.OsProjectName = osCreds.OsProjectName
	extraVars.NfsShares = []string{}

	extraVars.UseYarn = false
	//getting latest hadoop version for selected spark version
	hadoopVersions := sparkVersions[extraVars.SparkVersion]["hadoop_versions"]
	extraVars.HadoopVersion = hadoopVersions[len(hadoopVersions) - 1]
	//checking spark config params
	if serviceTypes["spark"].exists && serviceTypes["spark"].service.Config != nil {
		if yarn, ok := serviceTypes["spark"].service.Config["use-yarn"]; ok {
			b, err := strconv.ParseBool(yarn)
			if err != nil {
				log.Fatalln(err)
			}
			extraVars.UseYarn = b
			extraVars.YarnMasterMemMb = 10240 //change it
		}
		if version, ok := serviceTypes["spark"].service.Config["hadoop-version"]; ok {
			extraVars.HadoopVersion = version
		}
		if mem, ok := serviceTypes["spark"].service.Config["spark-worker-mem-mb"]; ok {
			memInt, err := strconv.Atoi(mem)
			if err != nil {
				log.Fatalln(err)
			}
			extraVars.SparkWorkerMemMb = memInt
		}
		if mem, ok := serviceTypes["spark"].service.Config["yarn-master-mem-mb"]; ok {
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
	extraVars.AnsibleSshPrivateKeyFile = "/home/lenaaxenova/.ssh/id_rsa.pub" //TODO: get ssh key from vault

	extraVars.CassandraVersion = "3.11.4"
	if serviceTypes["cassandra"].exists && serviceTypes["cassandra"].service.Version != "" {
		extraVars.CassandraVersion = serviceTypes["cassandra"].service.Version
	}

	if serviceTypes["jupyter"].exists && serviceTypes["jupyter"].service.Config != nil {
		if version, ok := serviceTypes["jupyter"].service.Config["toree-version"]; ok {
			extraVars.ToreeVersion = version
		}
	} else if serviceTypes["jupyter"].exists {
		extraVars.ToreeVersion = toreeVersions[string(extraVars.SparkVersion[0])]
	}

	if serviceTypes["ignite"].exists && serviceTypes["ignite"].service.Config != nil {
		if mem, ok := serviceTypes["ignite"].service.Config["ignite-memory"]; ok {
			memInt, err := strconv.Atoi(mem)
			if err != nil {
				log.Fatalln(err)
			}
			extraVars.IgniteMemory = memInt
		}
	}

	extraVars.Act = "launch" ///must be always "launch" in method "/clusters" POST
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

type AnsibleLauncher struct{}

func (aL AnsibleLauncher) Run(cluster *protobuf.Cluster, osCreds *osCredentials, osConfig *osConfig) error {
	log.SetPrefix("ANSIBLE_LAUNCHER: ")

	// creating ansible-playbook commands according to cluster object

	log.Print("Running ansible...")
	log.Print(cluster)

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

	ansibleArgs, err := json.Marshal(MakeExtraVars(cluster, osCreds, osConfig))
	if err != nil {
		log.Fatalln(err)
	}
	log.Print(string(ansibleArgs))

	cmdName := "ansible-playbook"
	cmdArgs := []string{"-v", "src/ansible/ansible/main.yml", "--extra-vars", string(ansibleArgs)}
	
	ansibleCmd := exec.Command(cmdName, cmdArgs...)
	stdout, err := ansibleCmd.StdoutPipe()
	if err != nil {
		log.Fatalln(err)
	}

	scanner := bufio.NewScanner(stdout)

	f, err := os.Create("ansible_output.txt")
	go func() {
		for scanner.Scan() {
			_, err := f.WriteString(scanner.Text() + "\n")
			if err != nil {
				log.Fatalln(err)
			}
		}
	}()

	if err := ansibleCmd.Start(); err != nil {
		log.Fatalln(err)
	}

	if err := ansibleCmd.Wait(); err != nil {
		log.Fatal(err)
	}

	err = f.Close()
	if err != nil {
		log.Fatalln(err)
	}

	log.Print("Launch: OK")
	return nil
}
