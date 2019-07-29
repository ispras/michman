package grpcclient

import (
	"context"
	"log"
	"time"

	protobuf "gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/protobuf"

	"google.golang.org/grpc"
)

const (
	EMPTY_BODY   = 0
	ERROR_NUM    = -1
	WAITING_TIME = 20
)

type GrpcClient struct {
	ansibleServiceClient protobuf.AnsibleRunnerClient
	dbServiceClient      protobuf.DBClient
	logger               *log.Logger
}

func (gc *GrpcClient) SetLogger(l *log.Logger) {
	gc.logger = l
}

// SetConnection will set connection with both of ansible and db services
func (gc *GrpcClient) SetConnection(ansibleServiceAddr string, dbServiceAddr string) {
	connAnsible, errAnsible := grpc.Dial(ansibleServiceAddr, grpc.WithInsecure())
	connDB, errDB := grpc.Dial(dbServiceAddr, grpc.WithInsecure())

	if errAnsible != nil {
		gc.logger.Fatalf("gRPC client connection error: %v", errAnsible)
	}

	if errDB != nil {
		gc.logger.Fatalf("gRPC client connection error: %v", errDB)
	}

	gc.ansibleServiceClient = protobuf.NewAnsibleRunnerClient(connAnsible)
	gc.dbServiceClient = protobuf.NewDBClient(connDB)
}

// GetID will send new cluster struct to db-service and return ID for new cluster
// that can be used to request information about cluster state and etc
func (gc GrpcClient) GetID(c *protobuf.Cluster) (int32, error) {
	ctx, cancel := context.WithTimeout(context.Background(), WAITING_TIME*time.Minute)
	defer cancel()

	gc.logger.Print("Sending request to db-service for clusterID")
	newID, err := gc.dbServiceClient.GetID(ctx, c)
	if err != nil {
		gc.logger.Println(err)
		return ERROR_NUM, err
	}

	return newID.ID, nil
}

// StartClusterCreation will send cluster struct to ansible-service for run ansible
func (gc GrpcClient) StartClusterCreation(c *protobuf.Cluster) {
	ctx, cancel := context.WithTimeout(context.Background(), WAITING_TIME*time.Minute)
	defer cancel()

	gc.logger.Print("Sending request to ansible-service")
	stream, err := gc.ansibleServiceClient.RunAnsible(ctx, c)
	if err != nil {
		gc.logger.Println(err)
		c.EntityStatus = "FAILED"
		gc.updateClusterState(c)
		return
	}

	message, err := stream.Recv()
	gc.logger.Printf("From ansible-service: %s", message.Status)

	if err != nil || message.Status != "OK" {
		gc.logger.Println(err)
		// request to db-service about errors with ansible service
		c.EntityStatus = "FAILED"
		gc.updateClusterState(c)
		return
	}

	gc.logger.Printf("Sending to db-service new status for %s cluster\n", c.Name)
	c.EntityStatus = "CREATED"
	gc.updateClusterState(c)
}

// updateClusterState will send cluster struct to db-service for update it's state
func (gc GrpcClient) updateClusterState(c *protobuf.Cluster) error {
	ctx, cancel := context.WithTimeout(context.Background(), WAITING_TIME*time.Minute)
	defer cancel()

	gc.logger.Print("Sending update cluster state request to db-service")

	// todo: error handling
	gc.dbServiceClient.UpdateClusterState(ctx, c)

	return nil
}
