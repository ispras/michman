package main

import (
	"log"
	"net"
	"time"

	grpc_ansible "../../protobuf/ansible_runner"
	// cluster "../model"
	"google.golang.org/grpc"
)

const (
	inputPort = ":5000"
)

// ansibleServer implements ansible service
type ansibleServer struct{}

func (gas *ansibleServer) RunAnsible(in *grpc_ansible.Cluster, stream grpc_ansible.AnsibleRunner_RunAnsibleServer) error {
	log.Print("Get create cluster request...")
	log.Printf("Need to create cluster with name: %s with %d slaves with status: %s", in.Name, in.Slaves, in.Status)
	// var c = Cluster{ name: in.Name, status: Inited }
	//	return &ansible_service.Cluster{Name: in.Name, Slaves: in.Slaves, Status:"Inited"}, nil
	if err := stream.Send(&grpc_ansible.Cluster{Name: in.Name, Slaves: in.Slaves, Status: "Inited"}); err != nil {
		return err
	}
	time.Sleep(3 * time.Second)
	if err := stream.Send(&grpc_ansible.Cluster{Name: in.Name, Slaves: in.Slaves, Status: "Created"}); err != nil {
		return err
	}
//	if err := stream.Send(io.EOF); err != nil {
//		return err
//	}
	return nil
}

func main() {
	lis, err := net.Listen("tcp", inputPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	gas := grpc.NewServer()
	grpc_ansible.RegisterAnsibleRunnerServer(gas, &ansibleServer{})

	log.Print("Ansible runner start work...\n")
	if err := gas.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
