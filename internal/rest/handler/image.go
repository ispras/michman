package handler

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/rest/handler/helpfunc"
	"github.com/ispras/michman/internal/rest/handler/response"
	"github.com/ispras/michman/internal/rest/handler/validate"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (hS HttpServer) ImagesGetList(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	request := "/images GET"
	hS.Logger.Info("Get " + request)

	images, err := hS.Db.ReadImagesList()
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		response.InternalError(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	response.Ok(w, images, request)
}

func (hS HttpServer) ImageGet(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	imageName := params.ByName("imageName")
	request := "/images/" + imageName + " GET"
	hS.Logger.Info("Get " + request)

	image, err := hS.Db.ReadImage(imageName)
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

func (hS HttpServer) ImageCreate(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	request := "/images POST"
	hS.Logger.Info("Get " + request)

	var image protobuf.Image
	err := json.NewDecoder(r.Body).Decode(&image)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrJsonIncorrect.Error())
		response.BadRequest(w, ErrJsonIncorrect)
		return
	}
	err = validate.Image(hS.Logger, &image)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", err.Error())
		response.BadRequest(w, err)
		return
	}

	dbImg, err := hS.Db.ReadImage(image.Name)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		response.InternalError(w, err)
		return
	}
	if dbImg.ID != "" {
		hS.Logger.Warn("Request ", request, " completed with status ", http.StatusBadRequest, ": ", ErrImageExisted.Error())
		response.BadRequest(w, ErrImageExisted)
		return
	}

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

func (hS HttpServer) ImageUpdate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	imageName := params.ByName("imageName")
	request := "/images/" + imageName + " PUT"
	hS.Logger.Info("Get " + request)

	oldImg, err := hS.Db.ReadImage(imageName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		response.InternalError(w, err)
		return
	}
	if oldImg.Name == "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusNotFound, ": ", ErrImageNotFound.Error())
		response.NotFound(w, ErrImageNotFound)
		return
	}

	var newImg protobuf.Image
	err = json.NewDecoder(r.Body).Decode(&newImg)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrJsonIncorrect.Error())
		response.BadRequest(w, ErrJsonIncorrect)
		return
	}

	used, err := helpfunc.IsImageUsed(hS.Db, hS.Logger, oldImg.Name)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		response.InternalError(w, err)
	}
	if used {
		hS.Logger.Warn("Request ", request, " completed with status ", http.StatusBadRequest, ": ", ErrImageUsed.Error())
		response.BadRequest(w, ErrImageUsed)
		return
	}

	if newImg.ID != "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrImageUnmodFields.Error())
		response.BadRequest(w, ErrImageUnmodFields)
		return
	}

	resImg := oldImg
	if newImg.Name != "" && oldImg.Name != newImg.Name {
		dbImg, err := hS.Db.ReadImage(newImg.Name)
		if err != nil {
			hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
			response.InternalError(w, err)
			return
		}
		if dbImg.ID != "" {
			hS.Logger.Warn("Request ", request, " completed with status ", http.StatusBadRequest, ": ", ErrImageExisted.Error())
			response.BadRequest(w, ErrImageExisted)
			return
		}
		resImg.Name = newImg.Name
	}
	if newImg.AnsibleUser != "" {
		resImg.AnsibleUser = newImg.AnsibleUser
	}
	if newImg.CloudImageID != "" {
		resImg.CloudImageID = newImg.CloudImageID
	}

	err = hS.Db.UpdateImage(resImg)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		response.InternalError(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	response.Ok(w, resImg, request)
}

func (hS HttpServer) ImageDelete(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	imageName := params.ByName("imageName")
	request := "/images/" + imageName + " DELETE"
	hS.Logger.Info("Get " + request)

	image, err := hS.Db.ReadImage(imageName)
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

	used, err := helpfunc.IsImageUsed(hS.Db, hS.Logger, imageName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		response.InternalError(w, err)
		return
	}
	if used {
		hS.Logger.Warn("Request ", request, " completed with status ", http.StatusBadRequest, ": ", ErrImageUsed.Error())
		response.BadRequest(w, ErrImageUsed)
		return
	}

	err = hS.Db.DeleteImage(params.ByName("imageName"))
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		response.InternalError(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusNoContent)
	response.NoContent(w)
}
