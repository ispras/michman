package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"google.golang.org/grpc"

	ansible_service "gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/app/ansible-pb"
)

const (
	address = "localhost:5000"
)

type server struct {
	ansibleService ansible_service.GreeterClient
}

func (s *server) createConnToAnsibleServer(addr string) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	s.ansibleService = ansible_service.NewGreeterClient(conn)
}

func (s server) createCluster() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	clusterName := "MY_SUPER_CLASTER"
	clusterType := "MASTER-SLAVE"

	r, err := s.ansibleService.CreateCluster(ctx, &ansible_service.ClusterDataRequest{Name: clusterName, Type: clusterType})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Cluster creating state: %s", r.State)
}

func (s server) clusterCreateHandler(w http.ResponseWriter, r *http.Request) {
	// parsing off comming data from JSON to GO objects
	// for RPC call to ansible-service

	log.Print("Hello from http handler")
	log.Print("Calling ansible-service RPC method")
	s.createCluster()
}

func main() {
	srv := server{}
	srv.createConnToAnsibleServer(address)

	http.HandleFunc("/clusters", srv.clusterCreateHandler)
	log.Print("Server starts to work")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
