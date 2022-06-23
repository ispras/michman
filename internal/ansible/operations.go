package ansible

import (
	"errors"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/utils"
)

func (aL *LauncherServer) Delete(in *protobuf.Cluster, stream protobuf.AnsibleRunner_DeleteServer) error {
	aL.Logger.Info("Getting delete cluster request...")
	in.PrintClusterData(aL.Logger)

	aL.Logger.Info("Getting vault secrets...")

	vaultClient, vaultCfg, err := aL.VaultCommunicator.ConnectVault()
	if vaultClient == nil || err != nil {
		aL.Logger.Warn("Error: can't connect to vault secrets storage")
		return errors.New("Can't connect to vault secrets storage")
	}

	keyName := vaultCfg.OsKey

	osCreds, err := aL.MakeOsCreds(keyName, vaultClient, aL.Config.OsVersion)
	if osCreds == nil || err != nil {
		aL.Logger.Warn(err)
		return err
	}

	err = aL.CheckSshKey(vaultCfg.SshKey, vaultClient)
	if err != nil {
		aL.Logger.Warn(err)
		return err
	}

	var dockRegCreds *utils.DockerCredentials
	if aL.Config.SelfignedRegistry || aL.Config.GitlabRegistry {
		dockRegCreds, err = aL.MakeDockerCreds(vaultCfg.RegistryKey, vaultClient)
		if err != nil {
			aL.Logger.Warn(err)
			return err
		}
	}

	ansibleStatus := aL.Run(in, aL.Logger, osCreds, dockRegCreds, &aL.Config, utils.ActionDelete)

	if err := stream.Send(&protobuf.TaskStatus{Status: ansibleStatus}); err != nil {
		return err
	}

	return nil
}

func (aL *LauncherServer) Update(in *protobuf.Cluster, stream protobuf.AnsibleRunner_UpdateServer) error {
	aL.Logger.Info("Getting update cluster request...")
	in.PrintClusterData(aL.Logger)

	aL.Logger.Info("Getting vault secrets...")

	vaultClient, vaultCfg, err := aL.VaultCommunicator.ConnectVault()
	if vaultClient == nil || err != nil {
		aL.Logger.Warn("Error: can't connect to vault secrets storage")
		return errors.New("Can't connect to vault secrets storage")
	}

	osCreds, err := aL.MakeOsCreds(vaultCfg.OsKey, vaultClient, aL.Config.OsVersion)
	if osCreds == nil || err != nil {
		aL.Logger.Warn(err)
		return err
	}

	err = aL.CheckSshKey(vaultCfg.SshKey, vaultClient)
	if err != nil {
		aL.Logger.Warn(err)
		return err
	}

	var dockRegCreds *utils.DockerCredentials
	if aL.Config.SelfignedRegistry || aL.Config.GitlabRegistry {
		dockRegCreds, err = aL.MakeDockerCreds(vaultCfg.RegistryKey, vaultClient)
		if err != nil {
			aL.Logger.Warn(err)
			return err
		}
	}

	ansibleStatus := aL.Run(in, aL.Logger, osCreds, dockRegCreds, &aL.Config, utils.ActionUpdate)

	if err := stream.Send(&protobuf.TaskStatus{Status: ansibleStatus}); err != nil {
		return err
	}
	return nil
}

func (aL *LauncherServer) Create(in *protobuf.Cluster, stream protobuf.AnsibleRunner_CreateServer) error {
	aL.Logger.Info("Getting create cluster request...")
	in.PrintClusterData(aL.Logger)

	aL.Logger.Info("Getting vault secrets...")
	vaultClient, vaultCfg, err := aL.VaultCommunicator.ConnectVault()
	if vaultClient == nil || err != nil {
		aL.Logger.Warn("Error: can't connect to vault secrets storage")
		return errors.New("Can't connect to vault secrets storage")
	}

	osCreds, err := aL.MakeOsCreds(vaultCfg.OsKey, vaultClient, aL.Config.OsVersion)
	if osCreds == nil || err != nil {
		aL.Logger.Warn(err)
		return err
	}

	err = aL.CheckSshKey(vaultCfg.SshKey, vaultClient)
	if err != nil {
		aL.Logger.Warn(err)
		return err
	}
	var dockRegCreds *utils.DockerCredentials
	if aL.Config.SelfignedRegistry || aL.Config.GitlabRegistry {
		dockRegCreds, err = aL.MakeDockerCreds(vaultCfg.RegistryKey, vaultClient)
		if err != nil {
			aL.Logger.Warn(err)
			return err
		}
	}

	ansibleStatus := aL.Run(in, aL.Logger, osCreds, dockRegCreds, &aL.Config, utils.ActionCreate)

	if err := stream.Send(&protobuf.TaskStatus{Status: ansibleStatus}); err != nil {
		return err
	}

	return nil
}
