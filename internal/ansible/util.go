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

func (aL LauncherServer) GetElasticConnectorJar() (string, error) {
	elasticHadoopUrl := "http://download.elastic.co/hadoop/elasticsearch-hadoop-5.5.0.zip"
	elasticHadoopFilename := filepath.Join("/tmp", filepath.Base(elasticHadoopUrl))
	elasticDir := filepath.Join("/tmp", "elasticsearch-hadoop/")
	archivePath := "elasticsearch-hadoop-5.5.0/dist/elasticsearch-hadoop-5.5.0.jar"
	elasticPath := filepath.Join(elasticDir, archivePath)
	if _, err := os.Stat(elasticPath); err != nil {
		if os.IsNotExist(err) {
			// file does not exist
			aL.Logger.Info("Downloading ElasticSearch Hadoop integration")
			err = utils.DownloadFile(elasticHadoopUrl, elasticHadoopFilename)
			if err != nil {
				return "", ErrDownload
			}

			_, err = utils.Unzip(elasticHadoopFilename, elasticDir)
			if err != nil {
				return "", err
			}
		}
	}
	return elasticPath, nil
}

func (aL LauncherServer) GetCassandraConnectorJar(sparkVersion string) (string, error) {
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
			err = utils.DownloadFile(sparkCassandraConnectorFile, sparkCassandraConnectorUrl)
			if err != nil {
				return "", ErrDownload
			}
		}
	}
	return sparkCassandraConnectorFile, nil
}

func (aL LauncherServer) AddJar(path string) (map[string]string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, ErrAbs
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

func (aL LauncherServer) ConvertParamValue(value string, vType string, flagLst bool) (interface{}, error) {
	if !flagLst {
		switch vType {
		case "int":
			parsedValue, err := strconv.ParseInt(value, 10, 32)
			if err != nil {
				return nil, ErrParseValue(value)
			}
			return parsedValue, nil
		case "float":
			parsedValue, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, ErrParseValue(value)
			}
			return parsedValue, nil
		case "bool":
			parsedValue, err := strconv.ParseBool(value)
			if err != nil {
				return nil, ErrParseValue(value)
			}
			return parsedValue, nil
		case "string":
			return value, nil
		}
	}

	switch vType {
	case "int":
		var valList []int64
		err := json.Unmarshal([]byte(value), &valList)
		if err != nil {
			return nil, ErrUnMarshal
		}
		return valList, nil
	case "float":
		var valList []float64
		err := json.Unmarshal([]byte(value), &valList)
		if err != nil {
			return nil, ErrUnMarshal
		}
		return valList, nil
	case "bool":
		var valList []bool
		err := json.Unmarshal([]byte(value), &valList)
		if err != nil {
			return nil, ErrUnMarshal
		}
		return valList, nil
	case "string":
		var valList []string
		err := json.Unmarshal([]byte(value), &valList)
		if err != nil {
			return nil, ErrUnMarshal
		}
		return valList, nil
	}
	return nil, ErrConvertParam
}

func (aL LauncherServer) MakeExtraVars(db database.Database, cluster *protobuf.Cluster, osConfig *utils.Config, action string) (InterfaceMap, error) {
	sTypes, err := db.ReadServicesTypesList()
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
					extraVars[sc.AnsibleVarName], err = aL.ConvertParamValue(value, sc.Type, sc.IsList)
					if err != nil {
						return nil, err
					}
				} else if sc.Required {
					//set default value if param is obligated
					extraVars[sc.AnsibleVarName], err = aL.ConvertParamValue(sc.DefaultValue, sc.Type, sc.IsList)
					if err != nil {
						return nil, err
					}
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
						extraVars[hc.AnsibleVarName], err = aL.ConvertParamValue(value, hc.Type, hc.IsList)
						if err != nil {
							return nil, err
						}
					} else if hc.Required {
						extraVars[hc.AnsibleVarName], err = aL.ConvertParamValue(hc.DefaultValue, hc.Type, hc.IsList)
						if err != nil {
							return nil, err
						}
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
	extraVars["os_project_name"] = aL.OsCreds[utils.OsProjectName]

	extraVars["floating_ip_pool"] = osConfig.FloatingIP
	extraVars["os_auth_url"] = aL.OsCreds[utils.OsAuthUrl]
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

	extraVars["os_swift_user_name"] = aL.OsCreds[utils.OsSwiftUserName]
	extraVars["os_swift_password"] = aL.OsCreds[utils.OsSwiftPassword]

	//make extra jars
	//TODO: change this
	var extraJars []map[string]string
	extraVars["spark_extra_jars"] = []map[string]string{}
	if extraVars[SetDeployService("cassandra")] == true {
		cassandraJar, err := aL.GetCassandraConnectorJar(extraVars["spark_version"].(string))
		if err != nil {
			return nil, err
		}
		alAddJar, err := aL.AddJar(cassandraJar)
		if err != nil {
			return nil, err
		}
		extraJars = append(extraJars, alAddJar)
	}

	//TODO: change this
	if extraVars[SetDeployService("elastic")] == true {
		elasticJar, err := aL.GetElasticConnectorJar()
		if err != nil {
			return nil, err
		}
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

	extraVars["cluster_owner"] = cluster.OwnerID

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
		return false, ErrSetEnv
	}

	err = prepCmd.Start()
	if err != nil {
		return false, ErrCmdStart
	}

	err = prepCmd.Wait()
	if err != nil {
		return false, ErrCmdWait
	}
	return true, nil
}

func MakeOsCreds(keyName string, vaultClient *vaultapi.Client, version string) (utils.OsCredentials, error) {
	secretValues, err := vaultClient.Logical().Read(keyName)
	if err != nil {
		return nil, ErrCouchSecretsRead
	}
	osCreds := make(utils.OsCredentials)
	var resVersion string
	_, exists := utils.OpenstackSecretsKeys[version]
	if exists {
		resVersion = version
	} else {
		resVersion = utils.OsUssuriVersion
	}
	for key, value := range utils.OpenstackSecretsKeys[resVersion] {
		osCred := secretValues.Data[value].(string)
		if osCred != "" {
			err := os.Setenv(value, osCred)
			if err != nil {
				return nil, ErrSetEnv
			}
		}
		osCreds[key] = osCred
	}
	return osCreds, nil
}

func (aL LauncherServer) MakeDockerCreds(keyName string, vaultClient *vaultapi.Client) (*utils.DockerCredentials, error) {
	secrets, err := vaultClient.Logical().Read(keyName)
	if err != nil {
		return nil, ErrCouchSecretsRead
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
			return ErrCouchSecretsRead
		}
		sshKey := secretValues.Data[utils.VaultSshKey].(string)
		f, err := os.Create(sshPath)
		if err != nil {
			return ErrCreate
		}
		err = os.Chmod(sshPath, 0777)
		if err != nil {
			return ErrChmod
		}
		_, err = f.WriteString(sshKey)
		if err != nil {
			return ErrWrite
		}
		err = f.Close()
		if err != nil {
			return ErrClose
		}
		err = os.Chmod(sshPath, 0400)
		if err != nil {
			return ErrChmod
		}
	}
	return nil
}
