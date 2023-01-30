package ansible

import (
	"context"
	clusterlogger "github.com/ispras/michman/internal/logger"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/utils"
)

func (aL *LauncherServer) Delete(ctx context.Context, cluster *protobuf.Cluster) (*protobuf.TaskStatus, error) {
	aL.Logger.Info("Getting delete cluster request...")
	cluster.PrintClusterData(aL.Logger)

	dockRegCreds, err := aL.GetDockerCreds()
	if err != nil {
		aL.Logger.Warn(err)
		return nil, err
	}

	aL.Logger.Info("Updating cluster in db...")
	err = aL.Db.UpdateCluster(cluster)
	if err != nil {
		return nil, err
	}

	cLogger, err := clusterlogger.MakeNewClusterLogger(aL.Config, cluster.ID, utils.ActionDelete)
	if err != nil {
		return nil, err
	}

	cLogsWriter, err := cLogger.PrepClusterLogsWriter()
	if err != nil {
		return nil, err
	}

	ansibleStatus, err := aL.RunInstances(cluster, dockRegCreds, utils.ActionDelete, cLogsWriter)
	if err != nil {
		aL.Logger.Warn(err)
		return nil, err
	}

	sTypes, err := aL.Db.ReadServicesTypesList()
	if err != nil {
		return nil, err
	}

	ansibleStatus, err = aL.RunServices(cluster, dockRegCreds, utils.ActionDelete, cLogsWriter, sTypes)
	if err != nil {
		aL.Logger.Warn(err)
		return nil, err
	}

	err = cLogger.FinClusterLogsWriter()
	if err != nil {
		return nil, err
	}

	res := new(protobuf.TaskStatus)
	res.Status = ansibleStatus
	return res, nil
}

func (aL *LauncherServer) Update(ctx context.Context, cluster *protobuf.Cluster) (*protobuf.TaskStatus, error) {
	aL.Logger.Info("Getting update cluster request...")
	cluster.PrintClusterData(aL.Logger)

	dockRegCreds, err := aL.GetDockerCreds()
	if err != nil {
		aL.Logger.Warn(err)
		return nil, err
	}

	aL.Logger.Info("Updating cluster in db...")
	err = aL.Db.UpdateCluster(cluster)
	if err != nil {
		return nil, err
	}

	cLogger, err := clusterlogger.MakeNewClusterLogger(aL.Config, cluster.ID, utils.ActionUpdate)
	if err != nil {
		return nil, err
	}

	cLogsWriter, err := cLogger.PrepClusterLogsWriter()
	if err != nil {
		return nil, err
	}

	ansibleStatus, err := aL.RunInstances(cluster, dockRegCreds, utils.ActionUpdate, cLogsWriter)
	if err != nil {
		aL.Logger.Warn(err)
		return nil, err
	}

	sTypes, err := aL.Db.ReadServicesTypesList()
	if err != nil {
		return nil, err
	}
	ansibleStatus, err = aL.RunServices(cluster, dockRegCreds, utils.ActionUpdate, cLogsWriter, sTypes)
	if err != nil {
		aL.Logger.Warn(err)
		return nil, err
	}

	err = cLogger.FinClusterLogsWriter()
	if err != nil {
		return nil, err
	}

	aL.Logger.Info("Saving IPs and URLs for services...")
	err = aL.Db.UpdateCluster(cluster)
	if err != nil {
		return nil, err
	}

	res := new(protobuf.TaskStatus)
	res.Status = ansibleStatus
	return res, nil
}

func (aL *LauncherServer) Create(ctx context.Context, cluster *protobuf.Cluster) (*protobuf.TaskStatus, error) {
	aL.Logger.Info("Getting create cluster request...")
	cluster.PrintClusterData(aL.Logger)

	dockRegCreds, err := aL.GetDockerCreds()
	if err != nil {
		aL.Logger.Warn(err)
		return nil, err
	}

	aL.Logger.Info("Writing new cluster to db...")
	err = aL.Db.UpdateCluster(cluster)
	if err != nil {
		aL.Logger.Warn(err)
		return nil, err
	}

	cLogger, err := clusterlogger.MakeNewClusterLogger(aL.Config, cluster.ID, utils.ActionCreate)
	if err != nil {
		aL.Logger.Warn(err)
		return nil, err
	}

	cLogsWriter, err := cLogger.PrepClusterLogsWriter()
	if err != nil {
		aL.Logger.Warn(err)
		return nil, err
	}

	ansibleStatus, err := aL.RunInstances(cluster, dockRegCreds, utils.ActionCreate, cLogsWriter)
	if err != nil {
		aL.Logger.Warn(err)
		return nil, err
	}

	sTypes, err := aL.Db.ReadServicesTypesList()
	if err != nil {
		aL.Logger.Warn(err)
		return nil, err
	}

	ansibleStatus, err = aL.RunServices(cluster, dockRegCreds, utils.ActionCreate, cLogsWriter, sTypes)
	if err != nil {
		aL.Logger.Warn(err)
		return nil, err
	}

	err = cLogger.FinClusterLogsWriter()
	if err != nil {
		aL.Logger.Warn(err)
		return nil, err
	}

	aL.Logger.Info("Saving IPs and URLs for services...")
	err = aL.Db.UpdateCluster(cluster)
	if err != nil {
		aL.Logger.Warn(err)
		return nil, err
	}

	res := new(protobuf.TaskStatus)
	res.Status = ansibleStatus
	return res, nil
}
