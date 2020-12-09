package handlers

import (
	"github.com/ispras/michman/database"
	proto "github.com/ispras/michman/protobuf"
	"github.com/ispras/michman/utils"
	"github.com/ispras/michman/auth"
	"log"
)

type GrpcClient interface {
	StartClusterCreation(c *proto.Cluster)
	StartClusterDestroying(c *proto.Cluster)
	StartClusterModification(c *proto.Cluster)
}

type HttpServer struct {
	Gc         GrpcClient
	Logger     *log.Logger
	Db         database.Database
	ErrHandler ErrorHandler
	Auth       auth.Authenticate
	Config     utils.Config
}
