package grpc

import (
	"context"
	"github.com/ispras/michman/internal/database"
	protobuf "github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/utils"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		return ErrGrpcConnection
	}
	gc.ansibleServiceClient = protobuf.NewAnsibleRunnerClient(connAnsible)
	return nil
}

// StartClusterCreation will send cluster struct to ansible-service for run ansible
func (gc GrpcClient) StartClusterCreation(c *protobuf.Cluster) {
	ctx, cancel := context.WithTimeout(context.Background(), WAITING_TIME*time.Minute)
	defer cancel()

	gc.logger.Info("Sending request to ansible-service")
	message, err := gc.ansibleServiceClient.Create(ctx, c)

	if err != nil {
		errStatus, _ := status.FromError(err)
		if errStatus.Code() == codes.Unavailable {
			gc.logger.Warn(ErrServerUnavailable)
		} else {
			gc.logger.Warn(ErrCreate)
		}
		c.EntityStatus = utils.StatusFailed
		err = gc.Db.UpdateCluster(c)
		if err != nil {
			gc.logger.Warn(err)
		}
		return
	}

	gc.logger.Infof("From ansible-service: %s", message.Status)

	if message.Status != utils.AnsibleOk {
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
		gc.logger.Warn(err)
		return
	}
	if newC.Name == "" {
		gc.logger.Warn("Cluster with ID %v not found", c.ID)
		return
	}

	gc.logger.Infof("Sending to db-service new status for %s cluster", c.Name)
	newC.EntityStatus = utils.StatusActive
	err = gc.Db.UpdateCluster(newC)
	if err != nil {
		gc.logger.Warn(err)
	}
	return
}

// StartClusterDestroying will send cluster struct to ansible-service for run ansible delete
func (gc GrpcClient) StartClusterDestroying(c *protobuf.Cluster) {
	ctx, cancel := context.WithTimeout(context.Background(), WAITING_TIME*time.Minute)
	defer cancel()

	gc.logger.Print("Sending request to ansible-service")
	message, err := gc.ansibleServiceClient.Delete(ctx, c)
	if err != nil {
		errStatus, _ := status.FromError(err)
		if errStatus.Code() == codes.Unavailable {
			gc.logger.Warn(ErrServerUnavailable)
		} else {
			gc.logger.Warn(ErrDestroy)
		}
		c.EntityStatus = utils.StatusFailed
		err = gc.Db.UpdateCluster(c)
		if err != nil {
			gc.logger.Warn(err)
		}
		return
	}

	gc.logger.Infof("From ansible-service: %s", message.Status)

	if message.Status != utils.AnsibleOk {
		// request to db-service about errors with ansible service
		c.EntityStatus = utils.StatusFailed
		err = gc.Db.UpdateCluster(c)
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
	return
}

// StartClusterDestroying will send cluster struct to ansible-service for run ansible delete
func (gc GrpcClient) StartClusterModification(c *protobuf.Cluster) {
	ctx, cancel := context.WithTimeout(context.Background(), WAITING_TIME*time.Minute)
	defer cancel()

	gc.logger.Info("Sending request to ansible-service")
	message, err := gc.ansibleServiceClient.Update(ctx, c)
	if err != nil {
		errStatus, _ := status.FromError(err)
		if errStatus.Code() == codes.Unavailable {
			gc.logger.Warn(ErrServerUnavailable)
		} else {
			gc.logger.Warn(ErrModify)
		}
		c.EntityStatus = utils.StatusFailed
		err = gc.Db.UpdateCluster(c)
		if err != nil {
			gc.logger.Warn(err)
		}
		return
	}
	
	gc.logger.Infof("From ansible-service: %s", message.Status)

	if message.Status != utils.AnsibleOk {
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
		gc.logger.Warn(err)
	}
	if newC.Name == "" {
		gc.logger.Warn("Cluster with ID %v not found", c.ID)
	}

	gc.logger.Infof("Sending to db-service new status for %s cluster", c.Name)
	newC.EntityStatus = utils.StatusActive
	err = gc.Db.UpdateCluster(newC)
	if err != nil {
		gc.logger.Warn(err)
	}
}
