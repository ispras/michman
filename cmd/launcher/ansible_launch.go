package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ispras/michman/internal/database"
	cluster_logger "github.com/ispras/michman/internal/logger"
	protobuf "github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/utils"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type InterfaceMap map[string]interface{}

type ServiceExists struct {
	exists  bool
	service *protobuf.Service
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

func addJar(path string) map[string]string {
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

func setDeployService(stype string) string {
	return "deploy_" + stype
}

func setServiceVersion(stype string) string {
	return stype + "_version"
}

func convertParamValue(value string, vType string, flagLst bool) interface{} {
	if !flagLst {
		switch vType {
		case "int":
			if v, err := strconv.ParseInt(value, 10, 32); err != nil {
				log.Print(err)
				return nil
			} else {
				return v
			}
		case "float":
			if v, err := strconv.ParseFloat(value, 64); err != nil {
				log.Print(err)
				return nil
			} else {
				return v
			}
		case "bool":
			if v, err := strconv.ParseBool(value); err != nil {
				log.Print(err)
				return nil
			} else {
				return v
			}
		case "string":
			return value
		}
	} else {
		switch vType {
		case "int":
			var valList []int64
			if err := json.Unmarshal([]byte(value), &valList); err != nil {
				log.Print(err)
				return err
			} else {
				return valList
			}
		case "float":
			var valList []float64
			if err := json.Unmarshal([]byte(value), &valList); err != nil {
				log.Print(err)
				return err
			} else {
				return valList
			}
		case "bool":
			var valList []bool
			if err := json.Unmarshal([]byte(value), &valList); err != nil {
				log.Print(err)
				return err
			} else {
				return valList
			}
		case "string":
			var valList []string
			if err := json.Unmarshal([]byte(value), &valList); err != nil {
				log.Print(err)
				return err
			} else {
				return valList
			}
		}
	}

	return nil
}

func makeExtraVars(aL AnsibleLauncher, cluster *protobuf.Cluster, osCreds *utils.OsCredentials, osConfig *utils.Config, action string) (InterfaceMap, error) {
	sTypes, err := aL.couchbaseCommunicator.ListServicesTypes()
	if err != nil {
		return nil, err
	}
	//appending old services which does not exist in new cluster configuration
	var curServices = make(map[string]ServiceExists)

	for _, service := range cluster.Services {
		curServices[service.Type] = ServiceExists{
			exists:  true,
			service: service,
		}
	}

	var extraVars = make(InterfaceMap)

	extraVars["create_storage"] = false

	for _, st := range sTypes {
		if curServices[st.Type].exists {
			//set deploy_stype to True
			curS := curServices[st.Type].service
			extraVars[setDeployService(curS.Type)] = true

			//set service version
			if curS.Version != "" {
				extraVars[setServiceVersion(curS.Type)] = curS.Version
			} else {
				extraVars[setServiceVersion(curS.Type)] = st.DefaultVersion
			}

			//set version config params
			var curSv *protobuf.ServiceVersion
			for _, sv := range st.Versions {
				if sv.Version == extraVars[setServiceVersion(curS.Type)] {
					curSv = sv
					break
				}
			}

			for _, sc := range curSv.Configs {
				//check if in request presents current config param
				if value, ok := curS.Config[sc.ParameterName]; ok {
					extraVars[sc.AnsibleVarName] = convertParamValue(value, sc.Type, sc.IsList)
				} else if sc.Required {
					//set default value if param is obligated
					extraVars[sc.AnsibleVarName] = convertParamValue(sc.DefaultValue, sc.Type, sc.IsList)
				}
			}

			//set node-types deployment flags
			if st.Class == utils.ClassStorage {
				extraVars["create_storage"] = true
			} else if st.Class == utils.ClassStandAlone {
				extraVars["create_master"] = true
			} else {
				extraVars["create_master_slave"] = true
			}

			//for consul health checks
			if st.HealthCheck[0].CheckType != "NotSupported" {
				for _, hc := range st.HealthCheck[0].Configs {
					if value, ok := curS.Config[hc.ParameterName]; ok {
						extraVars[hc.AnsibleVarName] = convertParamValue(value, hc.Type, hc.IsList)
					} else if hc.Required {
						extraVars[hc.AnsibleVarName] = convertParamValue(hc.DefaultValue, hc.Type, hc.IsList)
					}
				}
			}
		} else {
			extraVars[setDeployService(st.Type)] = false
		}
	}

	//filling obligated params
	extraVars["sync"] = "async" //must be always async mode

	extraVars["create_cluster"] = false
	if action == utils.ActionCreate {
		extraVars["create_cluster"] = true
	}

	extraVars["n_slaves"] = cluster.NHosts
	extraVars["cluster_name"] = cluster.Name
	extraVars["create_monitoring"] = cluster.Monitoring

	extraVars["mountnfs"] = false
	extraVars["master_flavor"] = osConfig.MasterFlavor
	extraVars["slaves_flavor"] = osConfig.SlavesFlavor
	extraVars["storage_flavor"] = osConfig.StorageFlavor
	extraVars["monitoring_flavor"] = osConfig.MonitoringFlavor
	extraVars["boot_from_volume"] = false

	image, err := aL.couchbaseCommunicator.ReadImage(cluster.Image)
	if err != nil {
		return nil, err
	}
	extraVars["ansible_user"] = image.AnsibleUser
	extraVars["hadoop_user"] = image.AnsibleUser
	extraVars["os_image"] = image.CloudImageID
	extraVars["skip_packages"] = false
	extraVars["os_project_name"] = osCreds.OsProjectName

	extraVars["floating_ip_pool"] = osConfig.FloatingIP
	extraVars["os_auth_url"] = osCreds.OsAuthUrl
	extraVars["use_oracle_java"] = false //must be always false
	extraVars["ansible_ssh_private_key_file"] = utils.SshKeyPath

	//action must be "launch" in method "/clusters" POST and /clusters/{clusterName} PUT
	//action must be "destroy" in method /clusters/{clusterName} DELETE
	if action == utils.ActionCreate || action == utils.ActionUpdate {
		extraVars["act"] = utils.AnsibleLaunch
	} else if action == utils.ActionDelete {
		extraVars["act"] = utils.AnsibleDestroy
	}

	extraVars["virtual_network"] = osConfig.VirtualNetwork
	extraVars["os_key_name"] = osConfig.Key

	extraVars["os_swift_user_name"] = osCreds.OsSwiftUserName
	extraVars["os_swift_password"] = osCreds.OsSwiftPassword

	//make extra jars
	//TODO: change this
	var extraJars []map[string]string
	extraVars["spark_extra_jars"] = []map[string]string{}
	if extraVars[setDeployService("cassandra")] == true {
		cassandraJar := GetCassandraConnectorJar(extraVars["spark_version"].(string))
		extraJars = append(extraJars, addJar(cassandraJar))
	}

	//TODO: change this
	if extraVars[setDeployService("elastic")] == true {
		elasticJar := GetElasticConnectorJar()
		extraJars = append(extraJars, addJar(elasticJar))
	}

	if extraJars != nil {
		extraVars["spark_extra_jars"] = extraJars
	}

	extraVars["use_package_mirror"] = osConfig.UsePackageMirror
	extraVars["use_pip_mirror"] = osConfig.UsePipMirror
	extraVars["apt_mirror_address"] = osConfig.AptMirrorAddress
	extraVars["pip_mirror_address"] = osConfig.PipMirrorAddress
	extraVars["pip_trusted_host"] = osConfig.PipTrustedHost
	extraVars["yum_mirror_address"] = osConfig.YumMirrorAddress

	//if no services in cluster -- create master-slave nodes
	if len(cluster.Services) == 0 && cluster.NHosts > 0 {
		extraVars["create_master_slave"] = true
	} else if len(cluster.Services) == 0 && cluster.NHosts == 0 {
		extraVars["create_master"] = true
		extraVars["create_master_slave"] = false
	}

	//create slaves if NHosts > 0 and master is created
	if extraVars["create_master"] == true && cluster.NHosts > 0 {
		extraVars["create_master_slave"] = true
	}

	if cluster.Keys != nil && len(cluster.Keys) > 0 {
		extraVars["public_keys"] = cluster.Keys
	}

	if extraVars["create_monitoring"] == true{
		extraVars["deploy_consul"] = true
	}

	return extraVars, nil
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
	case utils.OsUssuriVersion:
		err := os.Setenv(utils.OsAuthUrl, osCreds.OsAuthUrl)
		if err != nil {
			log.Fatalln(err)
		}
		err = os.Setenv(utils.OsProjectID, osCreds.OsProjectID)
		if err != nil {
			log.Fatalln(err)
		}
		err = os.Setenv(utils.OsProjectDomainID, osCreds.OsProjectDomainID)
		if err != nil {
			log.Fatalln(err)
		}
		err = os.Setenv(utils.OsInterface, osCreds.OsInterface)
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
		err = os.Setenv(utils.OsUserDomainName, osCreds.OsUserDomainName)
		if err != nil {
			log.Fatalln(err)
		}
		err = os.Setenv(utils.OsIdentityApiVersion, osCreds.OsIdentityApiVersion)
		if err != nil {
			log.Fatalln(err)
		}
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

func setServiceUrl(ip string, port int32) string {
	return ip + ":" + fmt.Sprintf("%d", port)
}

func runAnsible(cmd string, args []string, stdout io.Writer, stderr io.Writer) (bool, error) {
	prepCmd := exec.Command(cmd, args...)
	prepCmd.Stdout = stdout
	prepCmd.Stderr = stderr

	err := os.Setenv(utils.AnsibleConfigVar, utils.AnsibleConfigPath)
	if err != nil {
		log.Fatalln(err)
	}

	if err := prepCmd.Start(); err != nil {
		return false, err
	}

	if err := prepCmd.Wait(); err != nil {
		log.Print("Error: ", err)
		return false, err
	}
	return true, nil
}

func (aL AnsibleLauncher) Run(cluster *protobuf.Cluster, osCreds *utils.OsCredentials, dockRegCreds *utils.DockerCredentials, osConfig *utils.Config, action string) string {
	log.SetPrefix("LAUNCHER: ")

	// creating ansible-playbook commands according to cluster object

	//exporting ansible variables
	err := setOsVars(osCreds, osConfig.OsVersion)
	if err != nil {
		log.Println(err)
	}
	//constructing ansible-playbook command
	newExtraVars, err := makeExtraVars(aL, cluster, osCreds, osConfig, action)
	if err != nil {
		log.Println(err)
	}

	newAnsibleArgs, err := json.Marshal(newExtraVars)
	if err != nil {
		log.Println(err)
	}

	cmdArgs := []string{"-vvv", utils.AnsibleMainRole, "--extra-vars", string(newAnsibleArgs)}
	//saving cluster to database
	log.Print("Writing new cluster to db...")
	err = aL.couchbaseCommunicator.WriteCluster(cluster)
	if err != nil {
		log.Println(err)
	}
	// initialize output log
	cLogger, err := cluster_logger.MakeNewClusterLogger(*osConfig, cluster.ID, action)
	if err != nil {
		log.Fatalln(err)
	}
	buf, err := cLogger.PrepClusterLogsWriter()
	log.Print("Running ansible...")
	res, err := runAnsible(utils.AnsiblePlaybookCmd, cmdArgs, buf, buf)
	//write cluster logs
	err = cLogger.FinClusterLogsWriter()
	if err != nil {
		log.Fatalln(err)
	}

	//post-deploy actions: get ip for master and storage nodes for Cluster create or update action
	if res && (action == utils.ActionCreate || action == utils.ActionUpdate) {
		masterIp := ""
		if newExtraVars["create_master"] == true || newExtraVars["create_master_slave"] == true {
			v := map[string]string{
				"cluster_name": cluster.Name,
			}

			ipExtraVars, err := json.Marshal(v)
			if err != nil {
				log.Fatalln(err)
			}

			args := []string{"-v", utils.AnsibleMasterIpRole, "--extra-vars", string(ipExtraVars)}

			log.Print("Running ansible for getting master IP...")
			var outb bytes.Buffer
			runAnsible(utils.AnsiblePlaybookCmd, args, &outb, nil)
			masterIp = findIP(outb.String())
		}

		storageIp := ""
		//check if cluster has storage
		if newExtraVars["create_storage"] == true {
			v := map[string]string{
				"cluster_name":  cluster.Name,
				"extended_role": "storage",
			}
			ipExtraVars, err := json.Marshal(v)
			if err != nil {
				log.Fatalln(err)
			}
			args := []string{"-v", utils.AnsibleIpRole, "--extra-vars", string(ipExtraVars)}
			log.Print("Running ansible for getting storage IP...")
			var outb bytes.Buffer
			runAnsible(utils.AnsiblePlaybookCmd, args, &outb, nil)
			storageIp = findIP(outb.String())
		}
		monitoringIp := ""
		//check if cluster has monitoring
		if newExtraVars["create_monitoring"] == true {
			v := map[string]string{
				"cluster_name":  cluster.Name,
				"extended_role": "monitoring",
			}
			ipExtraVars, err := json.Marshal(v)
			if err != nil {
				log.Fatalln(err)
			}
			args := []string{"-v", utils.AnsibleIpRole, "--extra-vars", string(ipExtraVars)}
			log.Print("Running ansible for getting monitoring IP...")
			var outb bytes.Buffer
			runAnsible(utils.AnsiblePlaybookCmd, args, &outb, nil)
			monitoringIp = findIP(outb.String())
		}

		//filling services URLs:
		sTypes, err := aL.couchbaseCommunicator.ListServicesTypes()
		if err != nil {
			log.Fatalln(err)
		}

		for i, service := range cluster.Services {
			for _, st := range sTypes {
				if st.Type == service.Type {
					//set service url for services on master host
					if (st.Class == utils.ClassStandAlone || st.Class == utils.ClassMasterSlave) && masterIp != "" {
						if st.AccessPort > 0 { //service has access port
							cluster.Services[i].URL = setServiceUrl(masterIp, st.AccessPort)
						} else {
							cluster.Services[i].URL = masterIp
						}
						//set service url for services on storage host
					} else if st.Class == utils.ClassStorage && storageIp != "" {
						if st.AccessPort > 0 { //service has access port
							cluster.Services[i].URL = setServiceUrl(storageIp, st.AccessPort)
						} else {
							cluster.Services[i].URL = storageIp
						}
					}
					break
				}
			}
		}

		if monitoringIp != "" {
			log.Print("Monitoring IP is: ", monitoringIp)
		}

		if masterIp != "" {
			log.Print("Master IP is: ", masterIp)
			cluster.MasterIP = masterIp
		}

		if storageIp != "" {
			log.Print("Storage IP is: ", storageIp)
		}

		log.Print("Saving IPs and URLs for services...")
		err = aL.couchbaseCommunicator.WriteCluster(cluster)
		if err != nil {
			log.Fatalln(err)
		}

	}

	if res {
		log.Print("Launch: OK")
		return utils.AnsibleOk
	} else {
		log.Print("Ansible has failed, check logs for more information.")
		return utils.AnsibleFail
	}
}
