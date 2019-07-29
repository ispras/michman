package main

import (
	"log"
	"net"
	"os"

	protobuf "gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/protobuf"
	"google.golang.org/grpc"
)

const (
	inputPort = ":5000"
)

type ansibleLaunche interface {
	Run(c *protobuf.Cluster) error
}

// ansibleService implements ansible service
type ansibleService struct {
	logger        *log.Logger
	ansibleRunner ansibleLaunche
}

func (aS *ansibleService) RunAnsible(in *protobuf.Cluster, stream protobuf.AnsibleRunner_RunAnsibleServer) error {
	aS.logger.Print("Getting create cluster request...")
	aS.logger.Print("Cluster info:")
	in.PrintClusterData()

	// here ansible will run
	aS.ansibleRunner.Run(in)

	if err := stream.Send(&protobuf.TaskStatus{Status: "OK"}); err != nil {
		return err
	}

	//	if err := stream.Send(io.EOF); err != nil {
	//		return err
	//	}

	return nil
}

func main() {
	ansibleServiceLogger := log.New(os.Stdout, "ANSIBLE_SERVICE: ", log.Ldate|log.Ltime)
	ansibleLaunche := AnsibleLauncher{}

	lis, err := net.Listen("tcp", inputPort)
	if err != nil {
		ansibleServiceLogger.Fatalf("failed to listen: %v", err)
	}

	gas := grpc.NewServer()
	protobuf.RegisterAnsibleRunnerServer(gas, &ansibleService{ansibleServiceLogger, ansibleLaunche})

	ansibleServiceLogger.Print("Ansible runner start work...\n")
	if err := gas.Serve(lis); err != nil {
		ansibleServiceLogger.Fatalf("failed to serve: %v", err)
	}
}
