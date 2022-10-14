package database

import (
	"errors"
	"fmt"
	"github.com/ispras/michman/internal/utils"
)

type Error struct {
	Message string
	Class   int
}

func (eS Error) Error() string {
	return eS.Message
}

func MakeError(err string, class int) error {
	return &Error{Message: err, Class: class}
}

const (
	errWriteObjectByKey     = "error occurred while writing object by key in database"
	errReadObjectByKey      = "error occurred while getting object by key from database"
	errQueryExecution       = "error occurred while performing a query"
	errUpdateObjectByKey    = "error occurred while replacing an object in database"
	errCloseQuerySession    = "error occurred while closing query session"
	errDeleteObjectByKey    = "error occurred while removing object from database"
	errTransactionExecution = "error occurred while handling query within transaction"
	errTransactionCommit    = "error occurred while committing transaction"
	errStartQueryConnection = "error occurred while starting query connection"
	errUnmarshalJson        = "error occurred while unmarshalling/marshalling json data of the object"
	errReadObjectList       = "error occurred while getting object list from database"
	errNewUuid              = "error occurred while generating uuid for new object"
	errScanRows             = "error occurred while scan SQL query result rows"
	errQueryRows            = "error occurred while handling query rows"

	// errors without class:
	errCouchSecretsRead             = "error occurred while reading couchbase secrets"
	errCouchbaseClusterConnection   = "error occurred while creating Cluster object for a specific couchbase cluster"
	errCouchbaseClusterAuthenticate = "couchbase cluster authentication error"

	// errors for MySQL
	errMySQLSecretsRead = "error occured while reading mysql secrets"
	errMySQLConnection  = "error occured while creating connection to MySQL Database"
	errMySQLPing        = "error occured while sending ping to MySQL Database"
)

func ErrObjectNotFound(object, value string) error {
	errMessage := fmt.Sprintf("%s with this name or id (%s) does not exist", object, value)
	return MakeError(errMessage, utils.ObjectNotFound)
}

func ErrReadIncludedObject(included, main, mainId string) error {
	errMessage := fmt.Sprintf("can't read %s with ref to %s (id: %s)", included, main, mainId)
	return MakeError(errMessage, utils.DatabaseError)
}

func ErrUpdateIncludedObject(included, main, mainId string) error {
	errMessage := fmt.Sprintf("error on operation with %s when updating %s (id: %s)", included, main, mainId)
	return MakeError(errMessage, utils.DatabaseError)
}

func ErrInsertIncludedObject(included, main, mainId string) error {
	errMessage := fmt.Sprintf("error on operation with %s when inserting %s (id: %s)", included, main, mainId)
	return MakeError(errMessage, utils.DatabaseError)
}

func ErrOpenParamBucket(bucket string) error {
	errMessage := fmt.Errorf("can't open %s bucket", bucket)
	return errMessage
}

var (
	ErrStartQueryConnection = MakeError(errStartQueryConnection, utils.DatabaseError)
	ErrUnmarshalJson        = MakeError(errUnmarshalJson, utils.DatabaseError)
	ErrReadObjectList       = MakeError(errReadObjectList, utils.DatabaseError)
	ErrNewUuid              = MakeError(errNewUuid, utils.DatabaseError)
	ErrScanRows             = MakeError(errScanRows, utils.DatabaseError)
	ErrQueryRows            = MakeError(errQueryRows, utils.DatabaseError)

	ErrReadObjectByKey   = MakeError(errReadObjectByKey, utils.DatabaseError)
	ErrWriteObjectByKey  = MakeError(errWriteObjectByKey, utils.DatabaseError)
	ErrQueryExecution    = MakeError(errQueryExecution, utils.DatabaseError)
	ErrUpdateObjectByKey = MakeError(errUpdateObjectByKey, utils.DatabaseError)
	ErrCloseQuerySession = MakeError(errCloseQuerySession, utils.DatabaseError)
	ErrDeleteObjectByKey = MakeError(errDeleteObjectByKey, utils.DatabaseError)
	ErrTransactionQuery  = MakeError(errTransactionExecution, utils.DatabaseError)
	ErrTransactionCommit = MakeError(errTransactionCommit, utils.DatabaseError)

	// errors for MySQL
	ErrMySQLSecretsRead = MakeError(errMySQLSecretsRead, utils.DatabaseError)
	ErrMySQLConnection  = MakeError(errMySQLConnection, utils.DatabaseError)
	ErrMySQLPing        = MakeError(errMySQLPing, utils.DatabaseError)

	// errors without class:
	ErrCouchSecretsRead             = errors.New(errCouchSecretsRead)
	ErrCouchbaseClusterConnection   = errors.New(errCouchbaseClusterConnection)
	ErrCouchbaseClusterAuthenticate = errors.New(errCouchbaseClusterAuthenticate)
)
