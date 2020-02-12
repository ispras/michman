package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

const (
	JSONerrorIncorrect                     = 11
	JSONerrorIncorrectMessage              = "Incorrect JSON"
	JSONerrorIncorrectField                = 12
	JSONerrorIncorrectFieldMessage         = "Bad name. You should use only alpha-numeric characters and '-' symbols and only alphabetic characters for leading symbol."
	JSONerrorMissField                     = 13
	JSONerrorMissFieldMessage              = "Required field is empty"
	DBerror                                = 21
	DBerrorMessage                         = "DB error"
	LibErrorUUID                           = 31
	LibErrorUUIDMessage                    = "UUID generating error"
	LibErrorStructToJson                   = 32
	LibErrorStructToJsonMessage            = "Struct to JSON converting error"
	UserErrorProjectUnmodField             = 41
	UserErrorProjectUnmodFieldMessage      = "This fields of project or cluster can't be modifed"
	UserErrorProjectWithClustersDel        = 42
	UserErrorProjectWithClustersDelMessage = "Project has already had clusters. Delete them first"
	UserErrorClusterExisted                = 43
	UserErrorClusterExistedMessage         = "Cluster with this name has already existed in this project"
	UserErrorClusterStatus                 = 44
	UserErrorClusterStatusMessage          = "Cluster status must be 'CREATED' or 'FAILED' for UPDATE or DELETE"
)

// ErrorHandler handling interface
type ErrorHandler interface {
	Handle(w http.ResponseWriter, code int, message string, err error) (string, error)
}

type ErrorStruct struct {
	ErrorCode    string
	Desription   string
	ErrorMessage string
}

type HttpErrorHandler struct{}

func (httpEH HttpErrorHandler) Handle(w http.ResponseWriter, code int, message string, err error) (string, error) {
	var received_error string
	if err != nil {
		received_error = err.Error()
	} else {
		received_error = ""
	}
	errMessage := fmt.Sprintf("Error #%d. Description: %s. Error message: %v", code, message, received_error)

	w.WriteHeader(http.StatusBadRequest)
	errStruct := ErrorStruct{ErrorCode: strconv.Itoa(code), Desription: message, ErrorMessage: received_error}
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", "application/json")
	errr := enc.Encode(errStruct)

	return errMessage, errr
}
