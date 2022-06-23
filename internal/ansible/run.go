package ansible

import (
	"bytes"
	"encoding/json"
	clusterlogger "github.com/ispras/michman/internal/logger"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/utils"
	"github.com/sirupsen/logrus"
)

func (aL LauncherServer) Run(cluster *protobuf.Cluster, logger *logrus.Logger, osCreds *utils.OsCredentials, dockRegCreds *utils.DockerCredentials, osConfig *utils.Config, action string) string {
	// creating ansible-playbook commands according to cluster object

	//constructing ansible-playbook command
	newExtraVars, err := aL.MakeExtraVars(aL.Db, cluster, osCreds, osConfig, action)
	if err != nil {
		logger.Warn(err)
		return utils.RunFail
	}

	newAnsibleArgs, err := json.Marshal(newExtraVars)
	if err != nil {
		logger.Warn(err)
		return utils.RunFail
	}

	cmdArgs := []string{"-vvv", utils.AnsibleMainRole, "--extra-vars", string(newAnsibleArgs)}
	//saving cluster to database
	logger.Info("Writing new cluster to db...")
	err = aL.Db.WriteCluster(cluster)
	if err != nil {
		logger.Warn(err)
		return utils.RunFail
	}
	// initialize output log
	cLogger, err := clusterlogger.MakeNewClusterLogger(*osConfig, cluster.ID, action)
	if err != nil {
		logger.Warn(err)
		return utils.RunFail
	}
	buf, err := cLogger.PrepClusterLogsWriter()
	logger.Info("Running ansible...")
	res, err := aL.RunAnsible(utils.AnsiblePlaybookCmd, cmdArgs, buf, buf)
	//write cluster logs
	err = cLogger.FinClusterLogsWriter()
	if err != nil {
		logger.Warn(err)
		return utils.RunFail
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
				logger.Warn(err)
				return utils.RunFail
			}

			args := []string{"-v", utils.AnsibleMasterIpRole, "--extra-vars", string(ipExtraVars)}

			logger.Info("Running ansible for getting master IP...")
			var outb bytes.Buffer
			aL.RunAnsible(utils.AnsiblePlaybookCmd, args, &outb, nil)
			masterIp = FindIP(outb.String())
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
				logger.Warn(err)
				return utils.RunFail
			}
			args := []string{"-v", utils.AnsibleIpRole, "--extra-vars", string(ipExtraVars)}
			logger.Info("Running ansible for getting storage IP...")
			var outb bytes.Buffer
			aL.RunAnsible(utils.AnsiblePlaybookCmd, args, &outb, nil)
			storageIp = FindIP(outb.String())
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
				logger.Warn(err)
				return utils.RunFail
			}
			args := []string{"-v", utils.AnsibleIpRole, "--extra-vars", string(ipExtraVars)}
			logger.Info("Running ansible for getting monitoring IP...")
			var outb bytes.Buffer
			aL.RunAnsible(utils.AnsiblePlaybookCmd, args, &outb, nil)
			monitoringIp = FindIP(outb.String())
		}

		//filling services URLs:
		sTypes, err := aL.Db.ListServicesTypes()
		if err != nil {
			logger.Warn(err)
			return utils.RunFail
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
			logger.Warn(err)
			return utils.RunFail
		}

	}

	if res {
		logger.Info("Launch: OK")
		return utils.AnsibleOk
	} else {
		logger.Info("Ansible has failed, check logs for more information.")
		return utils.AnsibleFail
	}
}
