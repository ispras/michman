package grpc

import (
	"context"
	"errors"
	"fmt"
	"github.com/ispras/michman/internal/database"
	protobuf "github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/utils"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"time"
)

const (
	WAITING_TIME = 100
)

type GrpcClient struct {
	ansibleServiceClient protobuf.AnsibleRunnerClient
	logger               *logrus.Logger
	Db                   database.Database
}

func (gc *GrpcClient) SetLogger(l *logrus.Logger) {
	gc.logger = l
}

// SetConnection will set connection with both of ansible and db services
func (gc *GrpcClient) SetConnection(ansibleServiceAddr string) error {
	connAnsible, errAnsible := grpc.Dial(ansibleServiceAddr, grpc.WithInsecure())
	if errAnsible != nil {
		return errors.New(fmt.Sprintf("gRPC client connection error: %v", errAnsible))
	}
	gc.ansibleServiceClient = protobuf.NewAnsibleRunnerClient(connAnsible)
	return nil
}

// StartClusterCreation will send cluster struct to ansible-service for run ansible
func (gc GrpcClient) StartClusterCreation(c *protobuf.Cluster) {
	ctx, cancel := context.WithTimeout(context.Background(), WAITING_TIME*time.Minute)
	defer cancel()

	gc.logger.Info("Sending request to ansible-service")
	stream, err := gc.ansibleServiceClient.Create(ctx, c)
	if err != nil {
		gc.logger.Warn(err)
		c.EntityStatus = utils.StatusFailed
		err = gc.Db.WriteCluster(c)
		if err != nil {
			gc.logger.Warn(err)
		}
		return
	}

	message, err := stream.Recv()
	if err != nil {
		gc.logger.Warn(err)
	}
	gc.logger.Infof("From ansible-service: %s", message.Status)

	if err != nil || message.Status != utils.AnsibleOk {
		if err != nil {
			gc.logger.Warn(err)
		}
		// request to db-service about errors with ansible service
		c.EntityStatus = utils.StatusFailed
		err = gc.Db.WriteCluster(c)
		if err != nil {
			gc.logger.Warn(err)
		}
		return
	}

	gc.logger.Infof("Sending to db-service new status for %s cluster", c.Name)
	newC, err := gc.Db.ReadCluster(c.ProjectID, c.ID)
	if err != nil {
		gc.logger.Fatal(err)
	}
	if newC.Name == "" {
		gc.logger.Fatalf("Cluster with ID %v not found", c.ID)
	}

	newC.EntityStatus = utils.StatusActive
	err = gc.Db.WriteCluster(newC)
	if err != nil {
		gc.logger.Warn(err)
	}
}

// StartClusterDestroying will send cluster struct to ansible-service for run ansible delete
func (gc GrpcClient) StartClusterDestroying(c *protobuf.Cluster) {
	ctx, cancel := context.WithTimeout(context.Background(), WAITING_TIME*time.Minute)
	defer cancel()

	gc.logger.Print("Sending request to ansible-service")
	stream, err := gc.ansibleServiceClient.Delete(ctx, c)
	if err != nil {
		gc.logger.Warn(err)
		c.EntityStatus = utils.StatusFailed
		err = gc.Db.WriteCluster(c)
		if err != nil {
			gc.logger.Warn(err)
		}
		return
	}

	message, err := stream.Recv()
	gc.logger.Infof("From ansible-service: %s", message.Status)

	if err != nil || message.Status != utils.AnsibleOk {
		if err != nil {
			gc.logger.Warn(err)
		}
		// request to db-service about errors with ansible service
		c.EntityStatus = utils.StatusFailed
		err = gc.Db.WriteCluster(c)
		if err != nil {
			gc.logger.Warn(err)
		}
		return
	}

	gc.logger.Infof("Sending to db-service delete request for %s cluster", c.Name)
	err = gc.Db.DeleteCluster(c.ProjectID, c.ID)
	if err != nil {
		gc.logger.Warn(err)
	}
}

// StartClusterDestroying will send cluster struct to ansible-service for run ansible delete
func (gc GrpcClient) StartClusterModification(c *protobuf.Cluster) {
	ctx, cancel := context.WithTimeout(context.Background(), WAITING_TIME*time.Minute)
	defer cancel()

	gc.logger.Info("Sending request to ansible-service")
	stream, err := gc.ansibleServiceClient.Update(ctx, c)
	if err != nil {
		gc.logger.Warn(err)
		c.EntityStatus = utils.StatusFailed
		err = gc.Db.UpdateCluster(c)
		if err != nil {
			gc.logger.Warn(err)
		}
		return
	}

	message, err := stream.Recv()
	gc.logger.Infof("From ansible-service: %s", message.Status)

	if err != nil || message.Status != utils.AnsibleOk {
		if err != nil {
			gc.logger.Warn(err)
		}
		// request to db-service about errors with ansible service
		c.EntityStatus = utils.StatusFailed
		err = gc.Db.UpdateCluster(c)
		if err != nil {
			gc.logger.Warn(err)
		}
		return
	}
	newC, err := gc.Db.ReadCluster(c.ProjectID, c.ID)
	if err != nil {
		gc.logger.Fatal(err)
	}
	if newC.Name == "" {
		gc.logger.Fatalf("Cluster with ID %v not found", c.ID)
	}
	gc.logger.Infof("Sending to db-service new status for %s cluster", c.Name)
	newC.EntityStatus = utils.StatusActive
	err = gc.Db.UpdateCluster(newC)
	if err != nil {
		gc.logger.Warn(err)
	}
}
