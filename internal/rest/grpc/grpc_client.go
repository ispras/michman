package grpc

import (
	"context"
	"github.com/ispras/michman/internal/database"
	protobuf "github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/utils"
	"google.golang.org/grpc"
	"log"
	"time"
)

const (
	EMPTY_BODY   = 0
	ERROR_NUM    = -1
	WAITING_TIME = 100
)

type GrpcClient struct {
	ansibleServiceClient protobuf.AnsibleRunnerClient
	logger *log.Logger
	Db     database.Database
}

func (gc *GrpcClient) SetLogger(l *log.Logger) {
	gc.logger = l
}

// SetConnection will set connection with both of ansible and db services
func (gc *GrpcClient) SetConnection(ansibleServiceAddr string) {
	connAnsible, errAnsible := grpc.Dial(ansibleServiceAddr, grpc.WithInsecure())
	if errAnsible != nil {
		gc.logger.Fatalf("gRPC client connection error: %v", errAnsible)
	}
	gc.ansibleServiceClient = protobuf.NewAnsibleRunnerClient(connAnsible)
}

// StartClusterCreation will send cluster struct to ansible-service for run ansible
func (gc GrpcClient) StartClusterCreation(c *protobuf.Cluster) {
	ctx, cancel := context.WithTimeout(context.Background(), WAITING_TIME*time.Minute)
	defer cancel()

	gc.logger.Print("Sending request to ansible-service")
	stream, err := gc.ansibleServiceClient.Create(ctx, c)
	if err != nil {
		gc.logger.Println(err)
		c.EntityStatus = utils.StatusFailed
		err = gc.Db.WriteCluster(c)
		if err != nil {
			gc.logger.Print(err)
		}
		return
	}

	message, err := stream.Recv()
	if err != nil {
		gc.logger.Println(err)
	}
	gc.logger.Printf("From ansible-service: %s", message.Status)

	if err != nil || message.Status != "OK" {
		if err != nil {
			gc.logger.Println(err)
		}
		// request to db-service about errors with ansible service
		c.EntityStatus = utils.StatusFailed
		err = gc.Db.WriteCluster(c)
		if err != nil {
			gc.logger.Print(err)
		}
		return
	}

	gc.logger.Printf("Sending to db-service new status for %s cluster\n", c.Name)
	newC, err := gc.Db.ReadCluster(c.ID)
	if err != nil {
		log.Fatalln(err)
	}
	if newC.Name == "" {
		log.Fatalf("Cluster with ID %v not found\n", c.ID)
	}

	newC.EntityStatus = utils.StatusActive
	err = gc.Db.WriteCluster(newC)
	if err != nil {
		gc.logger.Print(err)
	}
}

// StartClusterDestroying will send cluster struct to ansible-service for run ansible delete
func (gc GrpcClient) StartClusterDestroying(c *protobuf.Cluster) {
	ctx, cancel := context.WithTimeout(context.Background(), WAITING_TIME*time.Minute)
	defer cancel()

	gc.logger.Print("Sending request to ansible-service")
	stream, err := gc.ansibleServiceClient.Delete(ctx, c)
	if err != nil {
		gc.logger.Println(err)
		c.EntityStatus = utils.StatusFailed
		err = gc.Db.WriteCluster(c)
		if err != nil {
			gc.logger.Print(err)
		}
		return
	}

	message, err := stream.Recv()
	gc.logger.Printf("From ansible-service: %s", message.Status)

	if err != nil || message.Status != "OK" {
		if err != nil {
			gc.logger.Println(err)
		}
		// request to db-service about errors with ansible service
		c.EntityStatus = utils.StatusFailed
		err = gc.Db.WriteCluster(c)
		if err != nil {
			gc.logger.Print(err)
		}
		return
	}

	gc.logger.Printf("Sending to db-service delete request for %s cluster\n", c.Name)
	err = gc.Db.DeleteCluster(c.ID)
	if err != nil {
		gc.logger.Print(err)
	}
}

// StartClusterDestroying will send cluster struct to ansible-service for run ansible delete
func (gc GrpcClient) StartClusterModification(c *protobuf.Cluster) {
	ctx, cancel := context.WithTimeout(context.Background(), WAITING_TIME*time.Minute)
	defer cancel()

	gc.logger.Print("Sending request to ansible-service")
	stream, err := gc.ansibleServiceClient.Update(ctx, c)
	if err != nil {
		gc.logger.Println(err)
		c.EntityStatus = utils.StatusFailed
		err = gc.Db.WriteCluster(c)
		if err != nil {
			gc.logger.Print(err)
		}
		return
	}

	message, err := stream.Recv()
	gc.logger.Printf("From ansible-service: %s", message.Status)

	if err != nil || message.Status != "OK" {
		if err != nil {
			gc.logger.Println(err)
		}
		// request to db-service about errors with ansible service
		c.EntityStatus = utils.StatusFailed
		err = gc.Db.WriteCluster(c)
		if err != nil {
			gc.logger.Print(err)
		}
		return
	}
	newC, err := gc.Db.ReadCluster(c.ID)
	if err != nil {
		log.Fatalln(err)
	}
	if newC.Name == "" {
		log.Fatalf("Cluster with ID %v not found\n", c.ID)
	}
	gc.logger.Printf("Sending to db-service new status for %s cluster\n", c.Name)
	newC.EntityStatus = utils.StatusActive
	err = gc.Db.WriteCluster(newC)
	if err != nil {
		gc.logger.Print(err)
	}
}
