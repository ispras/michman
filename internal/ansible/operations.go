package ansible

import (
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/utils"
)

func (aL *LauncherServer) Delete(in *protobuf.Cluster, stream protobuf.AnsibleRunner_DeleteServer) error {
	aL.Logger.Print("Getting delete cluster request...")
	aL.Logger.Print("Cluster info:")
	in.PrintClusterData(aL.Logger)

	aL.Logger.Print("Getting vault secrets...")

	vaultClient, vaultCfg := aL.VaultCommunicator.ConnectVault()
	if vaultClient == nil {
		aL.Logger.Fatalln("Error: can't connect to vault secrets storage")
		return nil
	}

	keyName := vaultCfg.OsKey

	osCreds := aL.MakeOsCreds(keyName, vaultClient, aL.Config.OsVersion)
	if osCreds == nil {
		return nil
	}

	err := aL.CheckSshKey(vaultCfg.SshKey, vaultClient)
	if err != nil {
		aL.Logger.Fatalln(err)
		return nil
	}

	var dockRegCreds *utils.DockerCredentials
	if aL.Config.SelfignedRegistry || aL.Config.GitlabRegistry {
		dockRegCreds = aL.MakeDockerCreds(vaultCfg.RegistryKey, vaultClient)
	}

	ansibleStatus := aL.Run(in, aL.Logger, osCreds, dockRegCreds, &aL.Config, utils.ActionDelete)

	if err := stream.Send(&protobuf.TaskStatus{Status: ansibleStatus}); err != nil {
		return err
	}

	return nil
}

func (aL *LauncherServer) Update(in *protobuf.Cluster, stream protobuf.AnsibleRunner_UpdateServer) error {
	aL.Logger.Print("Getting update cluster request...")
	aL.Logger.Print("Cluster info:")
	in.PrintClusterData(aL.Logger)

	aL.Logger.Print("Getting vault secrets...")

	vaultClient, vaultCfg := aL.VaultCommunicator.ConnectVault()
	if vaultClient == nil {
		aL.Logger.Fatalln("Error: can't connect to vault secrets storage")
		return nil
	}

	osCreds := aL.MakeOsCreds(vaultCfg.OsKey, vaultClient, aL.Config.OsVersion)
	if osCreds == nil {
		return nil
	}

	err := aL.CheckSshKey(vaultCfg.SshKey, vaultClient)
	if err != nil {
		aL.Logger.Fatalln(err)
		return nil
	}

	var dockRegCreds *utils.DockerCredentials
	if aL.Config.SelfignedRegistry || aL.Config.GitlabRegistry {
		dockRegCreds = aL.MakeDockerCreds(vaultCfg.RegistryKey, vaultClient)
	}

	ansibleStatus := aL.Run(in, aL.Logger, osCreds, dockRegCreds, &aL.Config, utils.ActionUpdate)

	if err := stream.Send(&protobuf.TaskStatus{Status: ansibleStatus}); err != nil {
		return err
	}
	return nil
}

func (aL *LauncherServer) Create(in *protobuf.Cluster, stream protobuf.AnsibleRunner_CreateServer) error {
	aL.Logger.Print("Getting create cluster request...")
	aL.Logger.Print("Cluster info:")
	in.PrintClusterData(aL.Logger)

	aL.Logger.Print("Getting vault secrets...")
	vaultClient, vaultCfg := aL.VaultCommunicator.ConnectVault()
	if vaultClient == nil {
		aL.Logger.Fatalln("Error: can't connect to vault secrets storage")
		return nil
	}

	osCreds := aL.MakeOsCreds(vaultCfg.OsKey, vaultClient, aL.Config.OsVersion)
	if osCreds == nil {
		return nil
	}

	err := aL.CheckSshKey(vaultCfg.SshKey, vaultClient)
	if err != nil {
		aL.Logger.Fatalln(err)
		return nil
	}
	var dockRegCreds *utils.DockerCredentials
	if aL.Config.SelfignedRegistry || aL.Config.GitlabRegistry {
		dockRegCreds = aL.MakeDockerCreds(vaultCfg.RegistryKey, vaultClient)
	}

	ansibleStatus := aL.Run(in, aL.Logger, osCreds, dockRegCreds, &aL.Config, utils.ActionCreate)

	if err := stream.Send(&protobuf.TaskStatus{Status: ansibleStatus}); err != nil {
		return err
	}

	return nil
}
