package handlers

import (
	"encoding/json"
	"github.com/ispras/michman/internal/database"
	"github.com/ispras/michman/internal/utils"
	"net/http"
	"strconv"
)

const (
	ResponseOkCode      = 1
	ResponseCreatedCode = 2
)

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
