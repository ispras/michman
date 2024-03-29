package handler

import (
	"github.com/ispras/michman/internal/database"
	proto "github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/utils"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

type GrpcClient interface {
	StartClusterCreation(c *proto.Cluster)
	StartClusterDestroying(c *proto.Cluster)
	StartClusterModification(c *proto.Cluster)
}

type HttpServer struct {
	Gc     GrpcClient
	Logger *logrus.Logger
	Db     database.Database
	Router *httprouter.Router
	Config utils.Config
}
