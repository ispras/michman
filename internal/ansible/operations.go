package ansible

import (
	"context"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/utils"
)

func (aL *LauncherServer) Delete(ctx context.Context, cluster *protobuf.Cluster) (*protobuf.TaskStatus, error) {
	aL.Logger.Info("Getting delete cluster request...")
	cluster.PrintClusterData(aL.Logger)

	aL.Logger.Info("Getting vault secrets...")

	vaultClient, vaultCfg, err := aL.VaultCommunicator.ConnectVault()
	if vaultClient == nil || err != nil {
		aL.Logger.Warn(err)
		return nil, err
	}

	keyName := vaultCfg.OsKey

	osCreds, err := aL.MakeOsCreds(keyName, vaultClient, aL.Config.OsVersion)
	if osCreds == nil || err != nil {
		aL.Logger.Warn(err)
		return nil, err
	}

	err = aL.CheckSshKey(vaultCfg.SshKey, vaultClient)
	if err != nil {
		aL.Logger.Warn(err)
		return nil, err
	}

	var dockRegCreds *utils.DockerCredentials
	if aL.Config.SelfignedRegistry || aL.Config.GitlabRegistry {
		dockRegCreds, err = aL.MakeDockerCreds(vaultCfg.RegistryKey, vaultClient)
		if err != nil {
			aL.Logger.Warn(err)
			return nil, err
		}
	}

	ansibleStatus, err := aL.Run(cluster, aL.Logger, osCreds, dockRegCreds, &aL.Config, utils.ActionDelete)
	if err != nil {
		aL.Logger.Warn(err)
		return nil, err
	}

	res := new(protobuf.TaskStatus)
	res.Status = ansibleStatus
	return res, nil
}

func (aL *LauncherServer) Update(ctx context.Context, cluster *protobuf.Cluster) (*protobuf.TaskStatus, error) {
	aL.Logger.Info("Getting update cluster request...")
	cluster.PrintClusterData(aL.Logger)

	aL.Logger.Info("Getting vault secrets...")

	vaultClient, vaultCfg, err := aL.VaultCommunicator.ConnectVault()
	if vaultClient == nil || err != nil {
		aL.Logger.Warn(err)
		return nil, err
	}

	osCreds, err := aL.MakeOsCreds(vaultCfg.OsKey, vaultClient, aL.Config.OsVersion)
	if osCreds == nil || err != nil {
		aL.Logger.Warn(err)
		return nil, err
	}

	err = aL.CheckSshKey(vaultCfg.SshKey, vaultClient)
	if err != nil {
		aL.Logger.Warn(err)
		return nil, err
	}

	var dockRegCreds *utils.DockerCredentials
	if aL.Config.SelfignedRegistry || aL.Config.GitlabRegistry {
		dockRegCreds, err = aL.MakeDockerCreds(vaultCfg.RegistryKey, vaultClient)
		if err != nil {
			aL.Logger.Warn(err)
			return nil, err
		}
	}

	ansibleStatus, err := aL.Run(cluster, aL.Logger, osCreds, dockRegCreds, &aL.Config, utils.ActionUpdate)
	if err != nil {
		aL.Logger.Warn(err)
		return nil, err
	}

	res := new(protobuf.TaskStatus)
	res.Status = ansibleStatus
	return res, nil
}

func (aL *LauncherServer) Create(ctx context.Context, cluster *protobuf.Cluster) (*protobuf.TaskStatus, error) {
	aL.Logger.Info("Getting create cluster request...")
	cluster.PrintClusterData(aL.Logger)

	aL.Logger.Info("Getting vault secrets...")
	vaultClient, vaultCfg, err := aL.VaultCommunicator.ConnectVault()
	if vaultClient == nil || err != nil {
		aL.Logger.Warn(err)
		return nil, err
	}

	osCreds, err := aL.MakeOsCreds(vaultCfg.OsKey, vaultClient, aL.Config.OsVersion)
	if osCreds == nil || err != nil {
		aL.Logger.Warn(err)
		return nil, err
	}

	err = aL.CheckSshKey(vaultCfg.SshKey, vaultClient)
	if err != nil {
		aL.Logger.Warn(err)
		return nil, err
	}
	var dockRegCreds *utils.DockerCredentials
	if aL.Config.SelfignedRegistry || aL.Config.GitlabRegistry {
		dockRegCreds, err = aL.MakeDockerCreds(vaultCfg.RegistryKey, vaultClient)
		if err != nil {
			aL.Logger.Warn(err)
			return nil, err
		}
	}

	ansibleStatus, err := aL.Run(cluster, aL.Logger, osCreds, dockRegCreds, &aL.Config, utils.ActionCreate)
	if err != nil {
		aL.Logger.Warn(err)
		return nil, err
	}

	res := new(protobuf.TaskStatus)
	res.Status = ansibleStatus
	return res, nil
}
