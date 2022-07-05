package database

import (
	"errors"
	"fmt"
	"github.com/ispras/michman/internal/utils"
)

var DbErrorsMap = make(map[error]int)

const (
	errWriteObjectByKey  = "error occurred while writing object by key in database"
	errReadObjectByKey   = "error occurred while getting object by key from database"
	errQueryExecution    = "error occurred while performing a query"
	errUpdateObjectByKey = "error occurred while replacing an object in database"
	errCloseQuerySession = "error occurred while closing query session"
	errDeleteObjectByKey = "error occurred while removing object from database"

	// errors without class:
	errCouchSecretsRead             = "error occurred while reading couchbase secrets"
	errCouchbaseClusterConnection   = "error occurred while creating Cluster object for a specific couchbase cluster"
	errCouchbaseClusterAuthenticate = "couchbase cluster authentication error"
)

func ErrObjectParamNotExist(param string) error {
	ErrParamType := fmt.Errorf("object with this name or id (%s) does not exist", param)
	DbErrorsMap[ErrParamType] = utils.DatabaseError
	return ErrParamType
}

func ErrOpenParamBucket(param string) error {
	ErrParamType := fmt.Errorf("can't open %s bucket", param)
	return ErrParamType
}

var (
	ErrWriteObjectByKey  = errors.New(errWriteObjectByKey)
	ErrReadObjectByKey   = errors.New(errReadObjectByKey)
	ErrQueryExecution    = errors.New(errQueryExecution)
	ErrUpdateObjectByKey = errors.New(errUpdateObjectByKey)
	ErrCloseQuerySession = errors.New(errCloseQuerySession)
	ErrDeleteObjectByKey = errors.New(errDeleteObjectByKey)

	// errors without class:
	ErrCouchSecretsRead             = errors.New(errCouchSecretsRead)
	ErrCouchbaseClusterConnection   = errors.New(errCouchbaseClusterConnection)
	ErrCouchbaseClusterAuthenticate = errors.New(errCouchbaseClusterAuthenticate)
)

func init() {
	DbErrorsMap[ErrWriteObjectByKey] = utils.DatabaseError
	DbErrorsMap[ErrReadObjectByKey] = utils.DatabaseError
	DbErrorsMap[ErrQueryExecution] = utils.DatabaseError
	DbErrorsMap[ErrUpdateObjectByKey] = utils.DatabaseError
	DbErrorsMap[ErrCloseQuerySession] = utils.DatabaseError
	DbErrorsMap[ErrDeleteObjectByKey] = utils.DatabaseError
}

func main() {}
