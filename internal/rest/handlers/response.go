package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/ispras/michman/internal/database"
	"github.com/ispras/michman/internal/utils"
	"net/http"
	"strconv"
)

const (
	ResponseOkCode                         = 1
	ResponseCreatedCode                    = 2
	JSONerrorIncorrect                     = 11
	JSONerrorIncorrectMessage              = "Incorrect JSON"
	JSONerrorIncorrectField                = 12
	JSONerrorIncorrectFieldMessage         = "Bad name. You should use only alpha-numeric characters and '-' symbols and only alphabetic characters for leading symbol."
	JSONerrorMissField                     = 13
	JSONerrorMissFieldMessage              = "Required field is empty"
	DBerror                                = 21
	DBerrorMessage                         = "DB error"
	DBemptyField                           = "Major field is empty"
	DBemptyHealthCheck                     = "ServiceType.HealthCheck field is empty"
	LibErrorUUID                           = 31
	LibErrorUUIDMessage                    = "UUID generating error"
	LibErrorStructToJson                   = 32
	LibErrorStructToJsonMessage            = "Struct to JSON converting error"
	LibErrorStructToJsonResponse           = 32
	LibErrorStructToJsonResponseMessage    = "Struct to JSON converting error in Response Handler"
	UserErrorProjectUnmodField             = 41
	UserErrorProjectUnmodFieldMessage      = "This fields of project or cluster can't be modified"
	UserErrorProjectWithClustersDel        = 42
	UserErrorProjectWithClustersDelMessage = "Project has already had clusters. Delete them first"
	UserErrorClusterExisted                = 43
	UserErrorClusterExistedMessage         = "Cluster with this name has already existed in this project"
	UserErrorClusterStatus                 = 44
	UserErrorClusterStatusMessage          = "Cluster status must be 'CREATED' or 'FAILED' for UPDATE or DELETE"
	UserErrorBadServiceVersion             = 45
	UserErrorBadServiceVersionMessage      = "Incompatible versions between services"
	ImageExisted                           = 50
	ImageExistedMessage                    = "Image with this name already exists"
	ImageUsed                              = 51
	ImageUsedMessage                       = "Image already in use. It can't be modified or deleted"
	ImageUnmodField                        = 52
	ImageUnmodFieldMessage                 = "Some fields can't be modified"
	AuthorizationHeaderIncorrect           = 61
	AuthorizationHeaderIncorrectMessage    = "Authorization header is empty or doesn't contain access token"
	AuthorizationFailed                    = 62
	AuthorizationFailedMessage             = "Authentication failed"
	FlavorExisted                          = 70
	FlavorExistedMessage                   = "Flavor with this name already exists"
	FlavorUsed                             = 71
	FlavorUsedMessage                      = "Flavor already in use. It can't be modified or deleted"
	FlavorUnmodField                       = 72
	FlavorUnmodFieldMessage                = "Some fields can't be modified"
	FlavorValidation                       = 73
	FlavorValidationMesssage               = "Flavor validation error"
	ClusterValidation                      = 74
	ClusterValidationMesssage              = "Cluster validation error"
	ServiceValidation                      = 74
	ServiceValidationMesssage              = "Service validation error"
)

// ResponseHandler handling interface
type ResponseHandler interface {
	Handle(w http.ResponseWriter, code int, message string, err error) (string, error)
}

type DetailStruct struct {
	Message string
	Data    interface{}
}

type IDs struct {
	ID string
}

type ResponseStruct struct {
	Type   string
	Status int
	Title  string
	Detail DetailStruct
}

type HttpResponseHandler struct{}

func (httpEH HttpResponseHandler) Handle(w http.ResponseWriter, code int, message string, err error) (string, error) {
	var receivedError string
	if err != nil {
		receivedError = err.Error()
	} else {
		receivedError = ""
	}
	errMessage := fmt.Sprintf("Error #%d. Description: %s. Error message: %v", code, message, receivedError)

	w.WriteHeader(http.StatusBadRequest)
	errStruct := ResponseStruct{
		Type:  strconv.Itoa(code),
		Title: message,
		Detail: DetailStruct{
			Message: receivedError,
			Data:    "No data",
		},
	}
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", "application/json")
	errr := enc.Encode(errStruct)

	return errMessage, errr
}

// ResponseOK (The 200 (OK) status code indicates that the request has succeeded)
func ResponseOK(w http.ResponseWriter, msgStruct interface{}, requestName string) {
	respStruct := ResponseStruct{
		Type:   strconv.Itoa(ResponseOkCode),
		Status: http.StatusOK,
		Title:  "Request: " + requestName,
		Detail: DetailStruct{
			Message: "Request successful",
			Data:    "No data",
		},
	}

	if msgStruct != nil {
		respStruct.Detail.Data = msgStruct
	}

	w.WriteHeader(respStruct.Status)
	enc := json.NewEncoder(w)
	err := enc.Encode(respStruct)
	if err != nil {
		ResponseInternalError(w, ErrJsonEncode)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

// ResponseCreated (The 201 (Created) status code indicates that the request has been fulfilled
// and has resulted in one or more new resources being created.)
func ResponseCreated(w http.ResponseWriter, msgStruct interface{}, requestName string) {
	respStruct := ResponseStruct{
		Type:   strconv.Itoa(ResponseCreatedCode),
		Status: http.StatusCreated,
		Title:  "Request: " + requestName,
		Detail: DetailStruct{
			Message: "Request successful",
			Data:    "No data",
		},
	}

	if msgStruct != nil {
		respStruct.Detail.Data = msgStruct
	}

	w.WriteHeader(respStruct.Status)
	enc := json.NewEncoder(w)
	err := enc.Encode(respStruct)
	if err != nil {
		ResponseInternalError(w, ErrJsonEncode)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

// ResponseNoContent (The 204 (No Content) status code indicates that the server has successfully fulfilled the request
// and that there is no additional content to send in the response content.)
func ResponseNoContent(w http.ResponseWriter) {
	status := http.StatusNoContent

	w.WriteHeader(status)

	w.Header().Set("Content-Type", "application/json")
}

// ResponseNotModified (The 304 (Not Modified) status code indicates that a conditional GET or HEAD request has been received
// and would have resulted in a 200 (OK) response if it were not for the fact that the condition evaluated to false.)
func ResponseNotModified(w http.ResponseWriter) {
	status := http.StatusNotModified
	w.WriteHeader(status)

	w.Header().Set("Content-Type", "application/json")
}

func findErrorType(err error) int {
	if HandlerErrorsMap[err] != 0 {
		return HandlerErrorsMap[err]
	}
	if database.DbErrorsMap[err] != 0 {
		return database.DbErrorsMap[err]
	}
	return utils.UnknownError
}

// ResponseBadRequest The 400 (Bad Request) status code indicates that the server cannot
// or will not process the request due to something that is perceived to be a client error
func ResponseBadRequest(w http.ResponseWriter, respErr error) {
	respStruct := ResponseStruct{
		Type:   strconv.Itoa(findErrorType(respErr)),
		Status: http.StatusBadRequest,
		Title:  respErr.Error(),
		Detail: DetailStruct{
			Message: "Bad request",
			Data:    "No data",
		},
	}

	w.WriteHeader(respStruct.Status)
	enc := json.NewEncoder(w)
	err := enc.Encode(respStruct)
	if err != nil {
		ResponseInternalError(w, ErrJsonEncode)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

// ResponseNotFound (The 404 (Not Found) status code indicates that the origin server
// did not find a current representation for the target resource or is not willing to disclose that one exists)
func ResponseNotFound(w http.ResponseWriter, respErr error) {
	respStruct := ResponseStruct{
		Type:   strconv.Itoa(findErrorType(respErr)),
		Status: http.StatusNotFound,
		Title:  respErr.Error(),
		Detail: DetailStruct{
			Message: "Object not found",
			Data:    "No data",
		},
	}

	w.WriteHeader(respStruct.Status)
	enc := json.NewEncoder(w)
	err := enc.Encode(respStruct)
	if err != nil {
		ResponseInternalError(w, ErrJsonEncode)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

// ResponseInternalError (The 500 (Internal Server Error) status code indicates that
// the server encountered an unexpected condition that prevented it from fulfilling the request)
func ResponseInternalError(w http.ResponseWriter, respErr error) {
	respStruct := ResponseStruct{
		Type:   strconv.Itoa(findErrorType(respErr)),
		Status: http.StatusInternalServerError,
		Title:  respErr.Error(),
		Detail: DetailStruct{
			Message: "Object not found",
			Data:    "No data",
		},
	}

	w.WriteHeader(respStruct.Status)
	enc := json.NewEncoder(w)
	err := enc.Encode(respStruct)
	if err != nil {
		ResponseInternalError(w, ErrJsonEncode)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}
