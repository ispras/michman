package database

import (
	"errors"
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
)

var (
	ErrWriteObjectByKey  = errors.New(errWriteObjectByKey)
	ErrReadObjectByKey   = errors.New(errReadObjectByKey)
	ErrQueryExecution    = errors.New(errQueryExecution)
	ErrUpdateObjectByKey = errors.New(errUpdateObjectByKey)
	ErrCloseQuerySession = errors.New(errCloseQuerySession)
	ErrDeleteObjectByKey = errors.New(errDeleteObjectByKey)
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
