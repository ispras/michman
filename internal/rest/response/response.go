package response

import (
	"encoding/json"
	"net/http"
)

const (
	okCode      = 1
	createdCode = 2
)

type details struct {
	Message string
	Data    interface{}
}

type responseBody struct {
	Type   int
	Status int
	Title  string
	Detail details
}

// Ok (The 200 (OK) status code indicates that the request has succeeded)
func Ok(w http.ResponseWriter, msgStruct interface{}, requestName string) {
	respStruct := responseBody{
		Type:   okCode,
		Status: http.StatusOK,
		Title:  "Request successful",
		Detail: details{
			Message: "Request: " + requestName,
			Data:    "No data",
		},
	}

	if msgStruct != nil {
		respStruct.Detail.Data = msgStruct
	}

	w.WriteHeader(respStruct.Status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	err := enc.Encode(respStruct)
	if err != nil {
		Error(w, ErrJsonEncode)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

// Created (The 201 (Created) status code indicates that the request has been fulfilled
// and has resulted in one or more new resources being created.)
func Created(w http.ResponseWriter, msgStruct interface{}, requestName string) {
	respStruct := responseBody{
		Type:   createdCode,
		Status: http.StatusCreated,
		Title:  "Request successful",
		Detail: details{
			Message: "Request: " + requestName,
			Data:    "No data",
		},
	}

	if msgStruct != nil {
		respStruct.Detail.Data = msgStruct
	}

	w.WriteHeader(respStruct.Status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	err := enc.Encode(respStruct)
	if err != nil {
		Error(w, ErrJsonEncode)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

// NoContent (The 204 (No Content) status code indicates that the server has successfully fulfilled the request
// and that there is no additional content to send in the response content.)
func NoContent(w http.ResponseWriter) {
	status := http.StatusNoContent

	w.WriteHeader(status)

	w.Header().Set("Content-Type", "application/json")
}

// NotModified (The 304 (Not Modified) status code indicates that a conditional GET or HEAD request has been received
// and would have resulted in a 200 (OK) response if it were not for the fact that the condition evaluated to false.)
func NotModified(w http.ResponseWriter) {
	status := http.StatusNotModified
	w.WriteHeader(status)

	w.Header().Set("Content-Type", "application/json")
}

// BadRequest The 400 (Bad Request) status code indicates that the server cannot
// or will not process the request due to something that is perceived to be a client error
func BadRequest(w http.ResponseWriter, errMsg string, class int) {
	respStruct := responseBody{
		Type:   class,
		Status: http.StatusBadRequest,
		Title:  "Bad request",
		Detail: details{
			Message: errMsg,
			Data:    "No data",
		},
	}

	w.WriteHeader(respStruct.Status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	err := enc.Encode(respStruct)
	if err != nil {
		Error(w, ErrJsonEncode)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

// Unauthorized The 401 (Unauthorized) status code indicates that the request has not been applied
// because it lacks valid authentication credentials for the target resource.
func Unauthorized(w http.ResponseWriter, errMsg string, class int) {
	respStruct := responseBody{
		Type:   class,
		Status: http.StatusUnauthorized,
		Title:  "Unauthorized",
		Detail: details{
			Message: errMsg,
			Data:    "No data",
		},
	}

	w.WriteHeader(respStruct.Status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	err := enc.Encode(respStruct)
	if err != nil {
		Error(w, ErrJsonEncode)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

// Forbidden The 403 (Forbidden) status code indicates that the server understood the request,
// but refuses to fulfill it
func Forbidden(w http.ResponseWriter, errMsg string, class int) {
	respStruct := responseBody{
		Type:   class,
		Status: http.StatusForbidden,
		Title:  "Forbidden",
		Detail: details{
			Message: errMsg,
			Data:    "No data",
		},
	}

	w.WriteHeader(respStruct.Status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	err := enc.Encode(respStruct)
	if err != nil {
		Error(w, ErrJsonEncode)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

// NotFound (The 404 (Not Found) status code indicates that the origin server
// did not find a current representation for the target resource or is not willing to disclose that one exists)
func NotFound(w http.ResponseWriter, errMsg string, class int) {
	respStruct := responseBody{
		Type:   class,
		Status: http.StatusNotFound,
		Title:  "Object not found",
		Detail: details{
			Message: errMsg,
			Data:    "No data",
		},
	}

	w.WriteHeader(respStruct.Status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	err := enc.Encode(respStruct)
	if err != nil {
		Error(w, ErrJsonEncode)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

// InternalError (The 500 (Internal Server Error) status code indicates that
// the server encountered an unexpected condition that prevented it from fulfilling the request)
func InternalError(w http.ResponseWriter, errMsg string, class int) {
	respStruct := responseBody{
		Type:   class,
		Status: http.StatusInternalServerError,
		Title:  "Internal Server Error",
		Detail: details{
			Message: errMsg,
			Data:    "No data",
		},
	}

	w.WriteHeader(respStruct.Status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	err := enc.Encode(respStruct)
	if err != nil {
		Error(w, ErrJsonEncode)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}
