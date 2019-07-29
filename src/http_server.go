package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"google.golang.org/grpc"

	model "./model"

	grpc_ansible "./protobuf/ansible_runner"
)

const (
	address = "localhost:5000"
)

type grpcAnsibleClient struct {
	ansibleService grpc_ansible.AnsibleRunnerClient
}

func (gac *grpcAnsibleClient) getConnection(addr string) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())

	if err != nil {
		log.Fatalf("gRPC client connection error: %v", err)
	}

	gac.ansibleService = grpc_ansible.NewAnsibleRunnerClient(conn)
}

func (gac grpcAnsibleClient) sendToServer(c model.Cluster) (*model.Cluster, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	stream, err := gac.ansibleService.RunAnsible(ctx, &grpc_ansible.Cluster{Name: c.Name, Slaves: c.Slaves, Status: "WANTED"})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	message, err := stream.Recv()

	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Printf("Write to database new status (%s) for %s with %d slaves\n", message.Status, message.Name, message.Slaves)

	result := &model.Cluster{Name: message.Name, Slaves: message.Slaves, Status: message.Status}

	updated, err := stream.Recv()
	if err != nil {
		log.Println(err)
	}
	log.Printf("Write to database new status (%s) for %s with %d slaves\n", updated.Status, updated.Name, updated.Slaves)
	return result, nil
}

func (gac grpcAnsibleClient) clustersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		w.WriteHeader(http.StatusOK)
	case "POST":
		var c model.Cluster
		err := json.NewDecoder(r.Body).Decode(&c)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
		result, err := gac.sendToServer(c)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.Header().Set("Content-Type", "application/json")
			enc := json.NewEncoder(w)
			enc.Encode(result)
		}
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func main() {
	srv := grpcAnsibleClient{}
	srv.getConnection(address)

	http.HandleFunc("/clusters", srv.clustersHandler)
	log.Print("Server starts to work")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
