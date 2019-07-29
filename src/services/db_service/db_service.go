package main

import (
	"context"
	"log"
	"net"
	"os"

	protobuf "gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/protobuf"
	"google.golang.org/grpc"
)

const (
	inputPort = ":5001"
)

type dbService struct {
	logger *log.Logger
}

func (dbs *dbService) GetID(ctx context.Context, c *protobuf.Cluster) (*protobuf.NewID, error) {
	dbs.logger.Print("Getting id request...")
	dbs.logger.Print("Got cluster:")
	c.PrintClusterData()
	var id int32
	//id, err = getIDFromDB()
	id = 32

	return &protobuf.NewID{ID: id}, nil
}

func (dbs *dbService) UpdateClusterState(ctx context.Context, c *protobuf.Cluster) (*protobuf.TaskStatus, error) {
	dbs.logger.Print("Getting update cluster state request...")
	dbs.logger.Print("Got cluster:")
	c.PrintClusterData()

	// going to DB and update data according to request

	return &protobuf.TaskStatus{Status: "OK"}, nil
}

func main() {
	dbLogger := log.New(os.Stdout, "DB_SERVICE: ", log.Ldate|log.Ltime)

	lis, err := net.Listen("tcp", inputPort)
	if err != nil {
		dbLogger.Fatalf("failed to listen: %v", err)
	}

	gas := grpc.NewServer()
	protobuf.RegisterDBServer(gas, &dbService{dbLogger})

	dbLogger.Print("DB service start work...\n")
	if err := gas.Serve(lis); err != nil {
		dbLogger.Fatalf("failed to serve: %v", err)
	}
}
