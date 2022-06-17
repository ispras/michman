package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

const (
	ResponseOkCode                         = 0
	ResponseNoContentCode                  = 1
	JSONerrorIncorrect                     = 11
	JSONerrorIncorrectMessage              = "Incorrect JSON"
	JSONerrorIncorrectField                = 12
	JSONerrorIncorrectFieldMessage         = "Bad name. You should use only alpha-numeric characters and '-' symbols and only alphabetic characters for leading symbol."
	JSONerrorMissField                     = 13
	JSONerrorMissFieldMessage              = "Required field is empty"
	DBerror                                = 21
	DBerrorMessage                         = "DB error"
	DBemptyHealthCheck                     = "ServiceType.HealthCheck field is empty"
	LibErrorUUID                           = 31
	LibErrorUUIDMessage                    = "UUID generating error"
	LibErrorStructToJson                   = 32
	LibErrorStructToJsonMessage            = "Struct to JSON converting error"
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
	ResponseBadRequest(logger *logrus.Logger, w http.ResponseWriter, code int, message string, resp_err error)
	ResponseOK(logger *logrus.Logger, w http.ResponseWriter, msgStruct interface{}, requestName string)
	ResponseNoContent(logger *logrus.Logger, w http.ResponseWriter, msg string)
}

type DetailStruct struct {
	Message string
	Data    interface{}
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
		},
	}
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", "application/json")
	errr := enc.Encode(errStruct)

	return errMessage, errr
}

func (httpEH HttpResponseHandler) ResponseBadRequest(logger *logrus.Logger, w http.ResponseWriter, code int, title string, respErr error) {
	respStruct := ResponseStruct{
		Type:   strconv.Itoa(code),
		Status: http.StatusBadRequest,
		Title:  title,
		Detail: DetailStruct{
			Message: "No error message",
		},
	}

	if respErr != nil {
		respStruct.Detail.Message = respErr.Error()
	}

	w.WriteHeader(respStruct.Status)
	enc := json.NewEncoder(w)
	err := enc.Encode(respStruct)
	if err != nil {
		httpEH.ResponseBadRequest(logger, w, LibErrorStructToJson, LibErrorStructToJsonMessage, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	logger.Println(respStruct)
}

func (httpEH HttpResponseHandler) ResponseOK(logger *logrus.Logger, w http.ResponseWriter, msgStruct interface{}, requestName string) {
	respStruct := ResponseStruct{
		Type:   strconv.Itoa(ResponseOkCode),
		Status: http.StatusOK,
		Title:  requestName,
		Detail: DetailStruct{
			Message: "No Details",
		},
	}

	if msgStruct != nil {
		respStruct.Detail.Data = msgStruct
	}

	w.WriteHeader(respStruct.Status)
	enc := json.NewEncoder(w)
	err := enc.Encode(respStruct)
	if err != nil {
		httpEH.ResponseBadRequest(logger, w, LibErrorStructToJson, LibErrorStructToJsonMessage, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	logger.Info(respStruct)
}

func (httpEH HttpResponseHandler) ResponseNoContent(logger *logrus.Logger, w http.ResponseWriter, message string) {
	respStruct := ResponseStruct{
		Type:   strconv.Itoa(ResponseNoContentCode),
		Status: http.StatusNoContent,
		Title:  message,
		Detail: DetailStruct{
			Message: "No Content",
		},
	}

	w.WriteHeader(respStruct.Status)
	enc := json.NewEncoder(w)
	err := enc.Encode(respStruct)
	if err != nil {
		httpEH.ResponseBadRequest(logger, w, LibErrorStructToJson, LibErrorStructToJsonMessage, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	logger.Info(respStruct)
}
