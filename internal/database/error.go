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
	errStartQueryConnection = "error occured while starting query connection"
	errUnmarshalJson = "error occured while unmarshalling/marshalling json data of the object"
	errReadObjectList = "error occured while getting object list from database"
	errNewUuid = "error occured while generating uuid for new object"

	// errors without class:
	errCouchSecretsRead             = "error occurred while reading couchbase secrets"
	errCouchbaseClusterConnection   = "error occurred while creating Cluster object for a specific couchbase cluster"
	errCouchbaseClusterAuthenticate = "couchbase cluster authentication error"

	// errors for MySQL
	errMySQLSecretsRead				= "error occured while reading mysql secrets"
	errMySQLConnection 				= "error occured while creating connection to MySQL Database"
	errMySQLPing 					= "error occured while sending ping to MySQL Database"

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
	ErrStartQueryConnection = errors.New(errStartQueryConnection)
	ErrUnmarshalJson = errors.New(errUnmarshalJson)
	ErrReadObjectList = errors.New(errReadObjectList)
	ErrNewUuid = errors.New(errNewUuid)

	// errors without class:
	ErrCouchSecretsRead             = errors.New(errCouchSecretsRead)
	ErrCouchbaseClusterConnection   = errors.New(errCouchbaseClusterConnection)
	ErrCouchbaseClusterAuthenticate = errors.New(errCouchbaseClusterAuthenticate)

	// errors for MySQL			
	ErrMySQLSecretsRead				= errors.New(errMySQLSecretsRead)
	ErrMySQLConnection				= errors.New(errMySQLConnection)
	ErrMySQLPing					= errors.New(errMySQLPing)
)

func init() {
	DbErrorsMap[ErrWriteObjectByKey] = utils.DatabaseError
	DbErrorsMap[ErrReadObjectByKey] = utils.DatabaseError
	DbErrorsMap[ErrQueryExecution] = utils.DatabaseError
	DbErrorsMap[ErrUpdateObjectByKey] = utils.DatabaseError
	DbErrorsMap[ErrCloseQuerySession] = utils.DatabaseError
	DbErrorsMap[ErrDeleteObjectByKey] = utils.DatabaseError
	DbErrorsMap[ErrStartQueryConnection] = utils.DatabaseError
	DbErrorsMap[ErrUnmarshalJson] = utils.DatabaseError
	DbErrorsMap[ErrReadObjectList] = utils.DatabaseError
	DbErrorsMap[ErrNewUuid] = utils.DatabaseError
}

func main() {}
