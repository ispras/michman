package ansible

import (
	"bytes"
	"encoding/json"
	clusterlogger "github.com/ispras/michman/internal/logger"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/utils"
	"github.com/sirupsen/logrus"
)

func (aL LauncherServer) Run(cluster *protobuf.Cluster, logger *logrus.Logger, osCreds *utils.OsCredentials, dockRegCreds *utils.DockerCredentials, osConfig *utils.Config, action string) (string, error) {
	//constructing ansible-playbook command
	newExtraVars, err := aL.MakeExtraVars(aL.Db, cluster, osCreds, osConfig, action)
	if err != nil {
		return utils.RunFail, err
	}

	newAnsibleArgs, err := json.Marshal(newExtraVars)
	if err != nil {
		return utils.RunFail, ErrMarshal
	}

	cmdArgs := []string{"-vvv", utils.AnsibleMainRole, "--extra-vars", string(newAnsibleArgs)}

	//saving cluster to database
	logger.Info("Writing new cluster to db...")
	err = aL.Db.WriteCluster(cluster)
	if err != nil {
		return utils.RunFail, err
	}

	// initialize output log
	cLogger, err := clusterlogger.MakeNewClusterLogger(*osConfig, cluster.ID, action)
	if err != nil {
		return utils.RunFail, err
	}

	buf, err := cLogger.PrepClusterLogsWriter()
	if err != nil {
		return utils.RunFail, err
	}

	logger.Info("Running ansible...")
	res, err := aL.RunAnsible(utils.AnsiblePlaybookCmd, cmdArgs, buf, buf)
	if err != nil {
		return utils.RunFail, err
	}

	//write cluster logs
	err = cLogger.FinClusterLogsWriter()
	if err != nil {
		return utils.RunFail, err
	}

	//post-deploy actions: get ip for master and storage nodes for Cluster create or update action
	if res && (action == utils.ActionCreate || action == utils.ActionUpdate) {
		masterIp := ""
		if newExtraVars["create_master"] == true || newExtraVars["create_master_slave"] == true {
			var outb bytes.Buffer
			v := map[string]string{
				"cluster_name": cluster.Name,
			}
			ipExtraVars, err := json.Marshal(v)
			if err != nil {
				return utils.RunFail, ErrMarshal
			}
			args := []string{"-v", utils.AnsibleMasterIpRole, "--extra-vars", string(ipExtraVars)}

			logger.Info("Running ansible for getting master IP...")
			_, err = aL.RunAnsible(utils.AnsiblePlaybookCmd, args, &outb, nil)
			if err != nil {
				return utils.RunFail, err
			}
			masterIp = FindIP(outb.String())
		}

		storageIp := ""
		//check if cluster has storage
		if newExtraVars["create_storage"] == true {
			var outb bytes.Buffer
			v := map[string]string{
				"cluster_name":  cluster.Name,
				"extended_role": "storage",
			}
			ipExtraVars, err := json.Marshal(v)
			if err != nil {
				return utils.RunFail, ErrMarshal
			}
			args := []string{"-v", utils.AnsibleIpRole, "--extra-vars", string(ipExtraVars)}

			logger.Info("Running ansible for getting storage IP...")
			_, err = aL.RunAnsible(utils.AnsiblePlaybookCmd, args, &outb, nil)
			if err != nil {
				return utils.RunFail, err
			}
			storageIp = FindIP(outb.String())
		}
		monitoringIp := ""
		//check if cluster has monitoring
		if newExtraVars["create_monitoring"] == true {
			var outb bytes.Buffer
			v := map[string]string{
				"cluster_name":  cluster.Name,
				"extended_role": "monitoring",
			}
			ipExtraVars, err := json.Marshal(v)
			if err != nil {
				return utils.RunFail, ErrMarshal
			}
			args := []string{"-v", utils.AnsibleIpRole, "--extra-vars", string(ipExtraVars)}

			logger.Info("Running ansible for getting monitoring IP...")
			_, err = aL.RunAnsible(utils.AnsiblePlaybookCmd, args, &outb, nil)
			if err != nil {
				return "", err
			}
			monitoringIp = FindIP(outb.String())
		}

		//filling services URLs:
		sTypes, err := aL.Db.ReadServicesTypesList()
		if err != nil {
			return utils.RunFail, err
		}

		for i, service := range cluster.Services {
			for _, st := range sTypes {
				if st.Type == service.Type {
					//set service url for services on master host
					if (st.Class == utils.ClassStandAlone || st.Class == utils.ClassMasterSlave) && masterIp != "" {
						if st.AccessPort > 0 { //service has access port
							cluster.Services[i].URL = SetServiceUrl(masterIp, st.AccessPort)
						} else {
							cluster.Services[i].URL = masterIp
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

		if monitoringIp != "" {
			logger.Info("Monitoring IP is: ", monitoringIp)
		}

		if masterIp != "" {
			logger.Info("Master IP is: ", masterIp)
			cluster.MasterIP = masterIp
		}

		if storageIp != "" {
			logger.Info("Storage IP is: ", storageIp)
		}

		logger.Info("Saving IPs and URLs for services...")
		err = aL.Db.WriteCluster(cluster)
		if err != nil {
			return utils.RunFail, err
		}

	}

	if res {
		logger.Info("Launch: OK")
		return utils.AnsibleOk, nil
	} else {
		logger.Info("Ansible has failed, check logs for more information.")
		return utils.AnsibleFail, nil
	}
}
