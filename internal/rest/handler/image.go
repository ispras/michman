package handler

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/rest/handler/response"
	"github.com/ispras/michman/internal/rest/handler/validate"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

// ImagesGetList processes a request to get a list of all images in database
func (hS HttpServer) ImagesGetList(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	request := "GET /images"
	hS.Logger.Info(request)

	images, err := hS.Db.ReadImagesList()
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		response.InternalError(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	response.Ok(w, images, request)
}

// ImageGet processes a request to get an image struct by id or name from database
func (hS HttpServer) ImageGet(w http.ResponseWriter, _ *http.Request, params httprouter.Params) {
	imageIdOrName := params.ByName("imageIdOrName")
	request := "GET /images/" + imageIdOrName
	hS.Logger.Info(request)

	image, err := hS.Db.ReadImage(imageIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		response.InternalError(w, err)
		return
	}
	if image.Name == "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusNotFound, ": ", ErrImageNotFound.Error())
		response.NotFound(w, ErrImageNotFound)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	response.Ok(w, image, request)
}

// ImageCreate processes a request to create an image struct in database
func (hS HttpServer) ImageCreate(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	request := "POST /images"
	hS.Logger.Info(request)

	var image protobuf.Image
	err := json.NewDecoder(r.Body).Decode(&image)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrJsonIncorrect.Error())
		response.BadRequest(w, ErrJsonIncorrect)
		return
	}

	hS.Logger.Info("Validating image...")
	err, status := validate.ImageCreate(hS.Db, &image)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", status, ": ", err.Error())
		switch status {
		case http.StatusBadRequest:
			response.BadRequest(w, err)
			return
		case http.StatusInternalServerError:
			response.InternalError(w, err)
			return
		}
	}

	// generating UUID for new image
	iUuid, err := uuid.NewRandom()
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", ErrUuidLibError.Error())
		response.InternalError(w, ErrUuidLibError)
		return
	}
	image.ID = iUuid.String()

	err = hS.Db.WriteImage(&image)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		response.InternalError(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusCreated)
	response.Created(w, image, request)
}

// ImageUpdate processes a request to update an image struct in database
func (hS HttpServer) ImageUpdate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	imageIdOrName := params.ByName("imageIdOrName")
	request := "PUT /images/" + imageIdOrName
	hS.Logger.Info(request)

	oldImage, err := hS.Db.ReadImage(imageIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		response.InternalError(w, err)
		return
	}
	if oldImage.Name == "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusNotFound, ": ", ErrImageNotFound.Error())
		response.NotFound(w, ErrImageNotFound)
		return
	}

	var newImage protobuf.Image
	err = json.NewDecoder(r.Body).Decode(&newImage)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrJsonIncorrect.Error())
		response.BadRequest(w, ErrJsonIncorrect)
		return
	}

	hS.Logger.Info("Validating updated values of the image fields...")
	err, status := validate.ImageUpdate(hS.Db, oldImage, &newImage)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", status, ": ", err.Error())
		switch status {
		case http.StatusBadRequest:
			response.BadRequest(w, err)
			return
		case http.StatusInternalServerError:
			response.InternalError(w, err)
			return
		}
	}

	resImage := oldImage
	resImage.Name = newImage.Name

	if newImage.AnsibleUser != "" {
		resImage.AnsibleUser = newImage.AnsibleUser
	}
	if newImage.CloudImageID != "" {
		resImage.CloudImageID = newImage.CloudImageID
	}

	err = hS.Db.UpdateImage(resImage)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		response.InternalError(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	response.Ok(w, resImage, request)
}

// ImageDelete processes a request to delete an image struct from database
func (hS HttpServer) ImageDelete(w http.ResponseWriter, _ *http.Request, params httprouter.Params) {
	imageIdOrName := params.ByName("imageIdOrName")
	request := "DELETE /images/" + imageIdOrName
	hS.Logger.Info(request)

	image, err := hS.Db.ReadImage(imageIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		response.InternalError(w, err)
		return
	}
	if image.Name == "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusNotFound, ": ", ErrImageNotFound.Error())
		response.NotFound(w, ErrImageNotFound)
		return
	}

	err, status := validate.ImageDelete(hS.Db, image)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", status, ": ", err.Error())
		switch status {
		case http.StatusBadRequest:
			response.BadRequest(w, err)
			return
		case http.StatusInternalServerError:
			response.InternalError(w, err)
			return
		}
	}

	err = hS.Db.DeleteImage(imageIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		response.InternalError(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusNoContent)
	response.NoContent(w)
}
