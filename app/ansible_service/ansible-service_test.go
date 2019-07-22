package main

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	ansible_service "gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/app/ansible-pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

// creating server in goroutine
func init() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	ansible_service.RegisterGreeterServer(s, &ansibleServer{})

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

func bufDialer(string, time.Duration) (net.Conn, error) {
	return lis.Dial()
}

func TestCreateCluster(t *testing.T) {
	// creating a connection to testi ng server
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := ansible_service.NewGreeterClient(conn)

	// test data
	clusterName := "MY_TEST_CLASTER"
	clusterType := "MASTER-SLAVE"

	resp, err := client.CreateCluster(ctx, &ansible_service.ClusterDataRequest{Name: clusterName, Type: clusterType})
	if err != nil {
		t.Fatalf("CreateCluster failed: %v", err)
	}

	// Testing output
	if resp.State != "Success" {
		t.Fatal("Cluster was not created")
	}
}
