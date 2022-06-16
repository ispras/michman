package handlers

import (
	"github.com/ispras/michman/internal/auth"
	"github.com/ispras/michman/internal/database"
	proto "github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/utils"
	"github.com/julienschmidt/httprouter"
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
	Router     *httprouter.Router
	Auth       auth.Authenticate
	Config     utils.Config
}
