package main

import (
	"context"
	"log"
	"net"

	ansible_service "gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/app/ansible-pb"
	"google.golang.org/grpc"
)

const (
	port = ":5000"
)

// ansibleServer implements ansible service
type ansibleServer struct{}

func (s *ansibleServer) CreateCluster(ctx context.Context, in *ansible_service.ClusterDataRequest) (*ansible_service.ClusterReply, error) {
	log.Print("Get create cluster request...")
	log.Printf("Need to create cluster with name: %s of type: %s", in.Name, in.Type)
	return &ansible_service.ClusterReply{State: "Success", Template: "Created template"}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	ansible_service.RegisterGreeterServer(s, &ansibleServer{})

	log.Print("Ansible service start work...\n")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
