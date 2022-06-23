package ansible

import (
	"encoding/json"
	"fmt"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/ispras/michman/internal/database"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/utils"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const (
	LauncherDefaultPort = "5000"
)

type InterfaceMap map[string]interface{}

type ServiceExists struct {
	exists  bool
	service *protobuf.Service
}

func (aL LauncherServer) GetElasticConnectorJar() string {
	elasticHadoopUrl := "http://download.elastic.co/hadoop/elasticsearch-hadoop-5.5.0.zip"
	elasticHadoopFilename := filepath.Join("/tmp", filepath.Base(elasticHadoopUrl))
	elasticDir := filepath.Join("/tmp", "elasticsearch-hadoop/")
	archivePath := "elasticsearch-hadoop-5.5.0/dist/elasticsearch-hadoop-5.5.0.jar"
	elasticPath := filepath.Join(elasticDir, archivePath)
	if _, err := os.Stat(elasticPath); err != nil {
		if os.IsNotExist(err) {
			// file does not exist
			aL.Logger.Info("Downloading ElasticSearch Hadoop integration")
			utils.DownloadFile(elasticHadoopUrl, elasticHadoopFilename)

			if _, err := utils.Unzip(elasticHadoopFilename, elasticDir); err != nil {
				aL.Logger.Warn(err)
			}
		}
	}
	return elasticPath
}

func (aL LauncherServer) GetCassandraConnectorJar(sparkVersion string) string {
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
			aL.Logger.Info("Downloading Spark Cassandra Connector for Spark version ", sparkVersion)
			utils.DownloadFile(sparkCassandraConnectorFile, sparkCassandraConnectorUrl)
		}
	}
	return sparkCassandraConnectorFile
}

func (aL LauncherServer) AddJar(path string) (map[string]string, error) {
	var absPath string
	if v, err := filepath.Abs(path); err != nil {
		return nil, err
	} else {
		absPath = v
	}
	var newElem = map[string]string{
		"name": filepath.Base(path), "path": absPath,
	}
	return newElem, nil
}

func SetDeployService(stype string) string {
	return "deploy_" + stype
}

func SetServiceVersion(stype string) string {
	return stype + "_version"
}

func (aL LauncherServer) ConvertParamValue(value string, vType string, flagLst bool) interface{} {
	if !flagLst {
		switch vType {
		case "int":
			if v, err := strconv.ParseInt(value, 10, 32); err != nil {
				aL.Logger.Warn(err)
				return nil
			} else {
				return v
			}
		case "float":
			if v, err := strconv.ParseFloat(value, 64); err != nil {
				aL.Logger.Warn(err)
				return nil
			} else {
				return v
			}
		case "bool":
			if v, err := strconv.ParseBool(value); err != nil {
				aL.Logger.Warn(err)
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
				aL.Logger.Warn(err)
				return err
			} else {
				return valList
			}
		case "float":
			var valList []float64
			if err := json.Unmarshal([]byte(value), &valList); err != nil {
				aL.Logger.Warn(err)
				return err
			} else {
				return valList
			}
		case "bool":
			var valList []bool
			if err := json.Unmarshal([]byte(value), &valList); err != nil {
				aL.Logger.Warn(err)
				return err
			} else {
				return valList
			}
		case "string":
			var valList []string
			if err := json.Unmarshal([]byte(value), &valList); err != nil {
				aL.Logger.Warn(err)
				return err
			} else {
				return valList
			}
		}
	}

	return nil
}

func (aL LauncherServer) MakeExtraVars(db database.Database, cluster *protobuf.Cluster, osCreds *utils.OsCredentials, osConfig *utils.Config, action string) (InterfaceMap, error) {
	sTypes, err := db.ListServicesTypes()
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
			extraVars[SetDeployService(curS.Type)] = true

			//set service version
			if curS.Version != "" {
				extraVars[SetServiceVersion(curS.Type)] = curS.Version
			} else {
				extraVars[SetServiceVersion(curS.Type)] = st.DefaultVersion
			}

			//set version config params
			var curSv *protobuf.ServiceVersion
			for _, sv := range st.Versions {
				if sv.Version == extraVars[SetServiceVersion(curS.Type)] {
					curSv = sv
					break
				}
			}

			for _, sc := range curSv.Configs {
				//check if in request presents current config param
				if value, ok := curS.Config[sc.ParameterName]; ok {
					extraVars[sc.AnsibleVarName] = aL.ConvertParamValue(value, sc.Type, sc.IsList)
				} else if sc.Required {
					//set default value if param is obligated
					extraVars[sc.AnsibleVarName] = aL.ConvertParamValue(sc.DefaultValue, sc.Type, sc.IsList)
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
						extraVars[hc.AnsibleVarName] = aL.ConvertParamValue(value, hc.Type, hc.IsList)
					} else if hc.Required {
						extraVars[hc.AnsibleVarName] = aL.ConvertParamValue(hc.DefaultValue, hc.Type, hc.IsList)
					}
				}
			}
		} else {
			extraVars[SetDeployService(st.Type)] = false
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
	extraVars["master_flavor"] = cluster.MasterFlavor
	extraVars["slaves_flavor"] = cluster.SlavesFlavor
	extraVars["storage_flavor"] = cluster.StorageFlavor
	extraVars["monitoring_flavor"] = cluster.MonitoringFlavor
	extraVars["boot_from_volume"] = false

	image, err := db.ReadImage(cluster.Image)
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
	if extraVars[SetDeployService("cassandra")] == true {
		cassandraJar := aL.GetCassandraConnectorJar(extraVars["spark_version"].(string))
		alAddJar, err := aL.AddJar(cassandraJar)
		if err != nil {
			return nil, err
		}
		extraJars = append(extraJars, alAddJar)
	}

	//TODO: change this
	if extraVars[SetDeployService("elastic")] == true {
		elasticJar := aL.GetElasticConnectorJar()
		alAddJar, err := aL.AddJar(elasticJar)
		if err != nil {
			return nil, err
		}
		extraJars = append(extraJars, alAddJar)
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

	if extraVars["create_monitoring"] == true {
		extraVars["deploy_consul"] = true
	}

	return extraVars, nil
}

func ValidateIP(input string) bool {
	pattern := "^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$"
	regEx := regexp.MustCompile(pattern)
	fmt.Println(input)
	return regEx.FindString(input) != ""
}

func FindIP(input string) string {
	numBlock := "(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])"
	regexPattern := numBlock + "\\." + numBlock + "\\." + numBlock + "\\." + numBlock

	regEx := regexp.MustCompile(regexPattern)
	return regEx.FindString(input)
}

func SetServiceUrl(ip string, port int32) string {
	return ip + ":" + fmt.Sprintf("%d", port)
}

func (aL LauncherServer) RunAnsible(cmd string, args []string, stdout io.Writer, stderr io.Writer) (bool, error) {
	prepCmd := exec.Command(cmd, args...)
	prepCmd.Stdout = stdout
	prepCmd.Stderr = stderr

	err := os.Setenv(utils.AnsibleConfigVar, utils.AnsibleConfigPath)
	if err != nil {
		aL.Logger.Warn(err)
		return false, err
	}

	if err := prepCmd.Start(); err != nil {
		aL.Logger.Warn(err)
		return false, err
	}

	if err := prepCmd.Wait(); err != nil {
		aL.Logger.Warn(err)
		return false, err
	}
	return true, nil
}

// SetOsVar TODO It's necessary to do correct error handling
func (aL LauncherServer) SetOsVar(utilVal string, secretValues *vaultapi.Secret) string {
	osCred := secretValues.Data[utilVal].(string)
	err := os.Setenv(utilVal, osCred)
	if err != nil {
		aL.Logger.Fatal(err)
	}
	return osCred
}

// main.go:
func (aL LauncherServer) MakeOsCreds(keyName string, vaultClient *vaultapi.Client, version string) (*utils.OsCredentials, error) {
	secretValues, err := vaultClient.Logical().Read(keyName)
	if err != nil {
		aL.Logger.Warn(err)
		return nil, err
	}
	var osCreds utils.OsCredentials
	switch version {
	case utils.OsUssuriVersion:
		osCreds.OsAuthUrl = aL.SetOsVar(utils.OsAuthUrl, secretValues)
		osCreds.OsProjectName = aL.SetOsVar(utils.OsProjectName, secretValues)
		osCreds.OsProjectID = aL.SetOsVar(utils.OsProjectID, secretValues)
		osCreds.OsInterface = aL.SetOsVar(utils.OsInterface, secretValues)
		osCreds.OsPassword = aL.SetOsVar(utils.OsPassword, secretValues)
		osCreds.OsRegionName = aL.SetOsVar(utils.OsRegionName, secretValues)
		osCreds.OsUserName = aL.SetOsVar(utils.OsUserName, secretValues)
		osCreds.OsUserDomainName = aL.SetOsVar(utils.OsUserDomainName, secretValues)
		osCreds.OsProjectDomainID = aL.SetOsVar(utils.OsProjectDomainID, secretValues)
		osCreds.OsIdentityApiVersion = aL.SetOsVar(utils.OsIdentityApiVersion, secretValues)
	case utils.OsSteinVersion:
		osCreds.OsAuthUrl = aL.SetOsVar(utils.OsAuthUrl, secretValues)
		osCreds.OsPassword = aL.SetOsVar(utils.OsPassword, secretValues)
		osCreds.OsProjectName = aL.SetOsVar(utils.OsProjectName, secretValues)
		osCreds.OsRegionName = aL.SetOsVar(utils.OsRegionName, secretValues)
		osCreds.OsUserName = aL.SetOsVar(utils.OsUserName, secretValues)
		osCreds.OsComputeApiVersion = aL.SetOsVar(utils.OsComputeApiVersion, secretValues)
		osCreds.OsNovaVersion = aL.SetOsVar(utils.OsNovaVersion, secretValues)
		osCreds.OsAuthType = aL.SetOsVar(utils.OsAuthType, secretValues)
		osCreds.OsCloudname = aL.SetOsVar(utils.OsCloudname, secretValues)
		osCreds.OsIdentityApiVersion = aL.SetOsVar(utils.OsIdentityApiVersion, secretValues)
		osCreds.OsImageApiVersion = aL.SetOsVar(utils.OsImageApiVersion, secretValues)
		osCreds.OsNoCache = aL.SetOsVar(utils.OsNoCache, secretValues)
		osCreds.OsProjectDomainName = aL.SetOsVar(utils.OsProjectDomainName, secretValues)
		osCreds.OsUserDomainName = aL.SetOsVar(utils.OsUserDomainName, secretValues)
		osCreds.OsVolumeApiVersion = aL.SetOsVar(utils.OsVolumeApiVersion, secretValues)
		osCreds.OsPythonwarnings = aL.SetOsVar(utils.OsPythonwarnings, secretValues)
		osCreds.OsNoProxy = aL.SetOsVar(utils.OsNoProxy, secretValues)
	case utils.OsLibertyVersion:
		osCreds.OsAuthUrl = aL.SetOsVar(utils.OsAuthUrl, secretValues)
		osCreds.OsPassword = aL.SetOsVar(utils.OsPassword, secretValues)
		osCreds.OsProjectName = aL.SetOsVar(utils.OsProjectName, secretValues)
		osCreds.OsRegionName = aL.SetOsVar(utils.OsRegionName, secretValues)
		osCreds.OsTenantId = aL.SetOsVar(utils.OsTenantId, secretValues)
		osCreds.OsTenantName = aL.SetOsVar(utils.OsTenantName, secretValues)
		osCreds.OsUserName = aL.SetOsVar(utils.OsUserName, secretValues)
		if _, ok := secretValues.Data[utils.OsSwiftUserName]; ok {
			osCreds.OsSwiftUserName = aL.SetOsVar(utils.OsSwiftUserName, secretValues)
		} else {
			osCreds.OsSwiftUserName = ""
		}
		if _, ok := secretValues.Data[utils.OsSwiftPassword]; ok {
			osCreds.OsSwiftUserName = aL.SetOsVar(utils.OsSwiftPassword, secretValues)
		} else {
			osCreds.OsSwiftPassword = ""
		}
	default: //liberty as default version
		osCreds.OsAuthUrl = aL.SetOsVar(utils.OsAuthUrl, secretValues)
		osCreds.OsPassword = aL.SetOsVar(utils.OsPassword, secretValues)
		osCreds.OsProjectName = aL.SetOsVar(utils.OsProjectName, secretValues)
		osCreds.OsRegionName = aL.SetOsVar(utils.OsRegionName, secretValues)
		osCreds.OsTenantId = aL.SetOsVar(utils.OsTenantId, secretValues)
		osCreds.OsTenantName = aL.SetOsVar(utils.OsTenantName, secretValues)
		osCreds.OsUserName = aL.SetOsVar(utils.OsUserName, secretValues)
		if _, ok := secretValues.Data[utils.OsSwiftUserName]; ok {
			osCreds.OsSwiftUserName = aL.SetOsVar(utils.OsSwiftUserName, secretValues)
		} else {
			osCreds.OsSwiftUserName = ""
		}
		if _, ok := secretValues.Data[utils.OsSwiftPassword]; ok {
			osCreds.OsSwiftPassword = aL.SetOsVar(utils.OsSwiftPassword, secretValues)
		} else {
			osCreds.OsSwiftPassword = ""
		}
	}
	return &osCreds, nil
}

func (aL LauncherServer) MakeDockerCreds(keyName string, vaultClient *vaultapi.Client) (*utils.DockerCredentials, error) {
	secrets, err := vaultClient.Logical().Read(keyName)
	if err != nil {
		return nil, err
	}
	var res utils.DockerCredentials
	res.Url = secrets.Data[utils.DockerLoginUlr].(string)
	res.User = secrets.Data[utils.DockerLoginUser].(string)
	res.Password = secrets.Data[utils.DockerLoginPassword].(string)
	return &res, nil
}

func (aL LauncherServer) CheckSshKey(keyName string, vaultClient *vaultapi.Client) error {
	sshPath := filepath.Join(utils.SshKeyPath)
	if _, err := os.Stat(sshPath); os.IsNotExist(err) {
		secretValues, err := vaultClient.Logical().Read(keyName)
		if err != nil {
			return err
		}
		sshKey := secretValues.Data[utils.VaultSshKey].(string)
		f, err := os.Create(sshPath)
		if err != nil {
			return err
		}
		err = os.Chmod(sshPath, 0777)
		if err != nil {
			return err
		}
		_, err = f.WriteString(sshKey)
		if err != nil {
			return err
		}
		err = f.Close()
		if err != nil {
			return err
		}
		err = os.Chmod(sshPath, 0400)
		if err != nil {
			return err
		}
	}
	return nil
}
