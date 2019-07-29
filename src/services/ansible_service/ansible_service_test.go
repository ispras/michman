package main

import (
	"context"
	"log"
	"net"
	"os"
	"testing"
	"time"

	protobuf "gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/protobuf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

type mockAnsibleRunner struct{}

func (mockAnsRunner mockAnsibleRunner) Run(c *protobuf.Cluster) error {
	log.Print("MOCK gRPC call for cluster creating")
	return nil
}

// creating server in goroutine
func init() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	testLogger := log.New(os.Stdout, "test_logger: ", log.Ldate|log.Ltime)
	testAnsibleRunner := mockAnsibleRunner{}
	protobuf.RegisterAnsibleRunnerServer(s, &ansibleService{testLogger, testAnsibleRunner})

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
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := protobuf.NewAnsibleRunnerClient(conn)

	respStream, err := client.RunAnsible(ctx, &protobuf.Cluster{Name: "test-cluster"})
	if err != nil {
		t.Fatalf("CreateCluster failed: %v", err)
	}

	message, err := respStream.Recv()

	if err != nil || message.Status != "OK" {
		t.Fatalf("Didn't get OK message: %v", err)
	}
}
