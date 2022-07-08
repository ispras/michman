package grpc

import "errors"

const (
	errServerUnavailable = "gRPC server is currently unavailable"
	errGrpcConnection    = "gRPC client connection error"
	errCreate            = "error occurred while executing create request"
	errModify            = "error occurred while executing update request"
	errDestroy           = "error occurred while executing delete request"
)

var (
	ErrServerUnavailable = errors.New(errServerUnavailable)
	ErrGrpcConnection    = errors.New(errGrpcConnection)
	ErrCreate            = errors.New(errCreate)
	ErrModify            = errors.New(errModify)
	ErrDestroy           = errors.New(errDestroy)
)
