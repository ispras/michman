package ansible

import (
	"bytes"
	"encoding/json"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/utils"
	"io"
)

func (aL LauncherServer) RunGetIP(cluster *protobuf.Cluster, extendedRole string) (string, error) {
	var outb bytes.Buffer
	v := map[string]string{
		"cluster_name":  cluster.Name,
		"extended_role": extendedRole,
	}
	ipExtraVars, err := json.Marshal(v)
	if err != nil {
		return utils.RunFail, ErrMarshal
	}
	args := []string{"-v", utils.AnsibleIpRole, "--extra-vars", string(ipExtraVars)}

	aL.Logger.Info("Running ansible for getting IP...")
	_, err = aL.RunAnsible(utils.AnsiblePlaybookCmd, args, &outb, nil)
	if err != nil {
		return utils.RunFail, err
	}
	return FindIP(outb.String()), nil
}

func (aL LauncherServer) RunServices(cluster *protobuf.Cluster, dockRegCreds *utils.DockerCredentials, action string, clusterLogsWriter io.Writer, serviceTypes []protobuf.ServiceType) (string, error) {
	newExtraVars, err := aL.MakeExtraVars(aL.Db, cluster, &aL.Config, dockRegCreds, action)
	if err != nil {
		return utils.RunFail, err
	}

	newAnsibleArgs, err := json.Marshal(newExtraVars)
	if err != nil {
		return utils.RunFail, ErrMarshal
	}

	cmdArgs := []string{"-vvv", utils.AnsibleServicesRole, "--extra-vars", string(newAnsibleArgs)}

	aL.Logger.Info("Running ansible...")
	res, err := aL.RunAnsible(utils.AnsiblePlaybookCmd, cmdArgs, clusterLogsWriter, clusterLogsWriter)
	if err != nil {
		return utils.RunFail, err
	}

	if res && (action == utils.ActionCreate || action == utils.ActionUpdate) {
		for i, service := range cluster.Services {
			for _, st := range serviceTypes {
				if st.Type == service.Type {
					storageIp := ""
					//check if cluster has storage
					if newExtraVars["create_storage"] == true {
						ip, err := aL.RunGetIP(cluster, "storage")
						if err != nil {
							return utils.RunFail, err
						}
						storageIp = ip
					}

					monitoringIp := ""
					//check if cluster has monitoring
					if newExtraVars["create_monitoring"] == true {
						ip, err := aL.RunGetIP(cluster, "monitoring")
						if err != nil {
							return utils.RunFail, err
						}
						monitoringIp = ip
					}

					if storageIp != "" {
						aL.Logger.Info("Storage IP is: ", storageIp)
					}
					if monitoringIp != "" {
						aL.Logger.Info("Monitoring IP is: ", monitoringIp)
					}

					//set service url for services on master host
					if (st.Class == utils.ClassStandAlone || st.Class == utils.ClassMasterSlave) && cluster.MasterIP != "" {
						if st.AccessPort > 0 { //service has access port
							cluster.Services[i].URL = SetServiceUrl(cluster.MasterIP, st.AccessPort)
						} else {
							cluster.Services[i].URL = cluster.MasterIP
						}
						//set service url for services on storage host
					} else if st.Class == utils.ClassStorage && storageIp != "" {
						if st.AccessPort > 0 { //service has access port
							cluster.Services[i].URL = SetServiceUrl(storageIp, st.AccessPort)
						} else {
							cluster.Services[i].URL = storageIp
						}
					}
					break
				}
			}
		}
	}

	if res {
		aL.Logger.Info("Launch: OK")
		return utils.AnsibleOk, nil
	} else {
		aL.Logger.Info("Ansible has failed, check logs for more information.")
		return utils.AnsibleFail, nil
	}
}

func (aL LauncherServer) RunInstances(cluster *protobuf.Cluster, dockRegCreds *utils.DockerCredentials, action string, clusterLogsWriter io.Writer) (string, error) {
	newExtraVars, err := aL.MakeExtraVars(aL.Db, cluster, &aL.Config, dockRegCreds, action)
	if err != nil {
		return utils.RunFail, err
	}

	newAnsibleArgs, err := json.Marshal(newExtraVars)
	if err != nil {
		return utils.RunFail, ErrMarshal
	}

	cmdArgs := []string{"-vvv", utils.AnsibleInstancesRole, "--extra-vars", string(newAnsibleArgs)}

	aL.Logger.Info("Running ansible...")
	res, err := aL.RunAnsible(utils.AnsiblePlaybookCmd, cmdArgs, clusterLogsWriter, clusterLogsWriter)
	if err != nil {
		return utils.RunFail, err
	}

	// get master IP
	if res && (action == utils.ActionCreate || action == utils.ActionUpdate) {
		masterIp := ""
		if newExtraVars["create_master"] == true || newExtraVars["create_master_slave"] == true {
			ip, err := aL.RunGetIP(cluster, "master")
			if err != nil {
				return utils.RunFail, err
			}
			masterIp = ip
		}

		if masterIp != "" {
			aL.Logger.Info("Master IP is: ", masterIp)
			cluster.MasterIP = masterIp
		}
	}

	if res {
		aL.Logger.Info("Launch: OK")
		return utils.AnsibleOk, nil
	} else {
		aL.Logger.Info("Ansible has failed, check logs for more information.")
		return utils.AnsibleFail, nil
	}
}
