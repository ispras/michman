package handlers

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (hS HttpServer) ImagesGetList(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	request := "/images GET"
	hS.Logger.Info("Get " + request)

	images, err := hS.Db.ListImages()
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	ResponseOK(w, images, request)
}

func (hS HttpServer) ImageGet(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	imageName := params.ByName("imageName")
	request := "/images/" + imageName + " GET"
	hS.Logger.Info("Get " + request)

	image, err := hS.Db.ReadImage(imageName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}
	if image.Name == "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusNotFound, ": ", ErrImageNotFound.Error())
		ResponseNotFound(w, ErrImageNotFound)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	ResponseOK(w, image, request)
}

func (hS HttpServer) ImageCreate(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	request := "/images POST"
	hS.Logger.Info("Get " + request)

	var image protobuf.Image
	err := json.NewDecoder(r.Body).Decode(&image)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrJsonIncorrect.Error())
		ResponseBadRequest(w, ErrJsonIncorrect)
		return
	}
	err = ValidateImage(hS, &image)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", err.Error())
		ResponseBadRequest(w, err)
		return
	}

	dbImg, err := hS.Db.ReadImage(image.Name)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}
	if dbImg.ID != "" {
		hS.Logger.Warn("Request ", request, " completed with status ", http.StatusBadRequest, ": ", ErrImageExisted.Error())
		ResponseBadRequest(w, ErrImageExisted)
		return
	}

	iUuid, err := uuid.NewRandom()
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", ErrUuidLibError.Error())
		ResponseInternalError(w, ErrUuidLibError)
		return
	}
	image.ID = iUuid.String()
	err = hS.Db.WriteImage(&image)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusCreated)
	ResponseCreated(w, image, request)
}

func (hS HttpServer) ImageUpdate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	imageName := params.ByName("imageName")
	request := "/images/" + imageName + " PUT"
	hS.Logger.Info("Get " + request)

	oldImg, err := hS.Db.ReadImage(imageName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}
	if oldImg.Name == "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusNotFound, ": ", ErrImageNotFound.Error())
		ResponseNotFound(w, ErrImageNotFound)
		return
	}

	var newImg protobuf.Image
	err = json.NewDecoder(r.Body).Decode(&newImg)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrJsonIncorrect.Error())
		ResponseBadRequest(w, ErrJsonIncorrect)
		return
	}

	used, err := IsImageUsed(hS, oldImg.Name)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
	}
	if used {
		hS.Logger.Warn("Request ", request, " completed with status ", http.StatusBadRequest, ": ", ErrImageUsed.Error())
		ResponseBadRequest(w, ErrImageUsed)
		return
	}

	if newImg.ID != "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrImageUnmodFields.Error())
		ResponseBadRequest(w, ErrImageUnmodFields)
		return
	}

	resImg := oldImg
	if newImg.Name != "" && oldImg.Name != newImg.Name {
		dbImg, err := hS.Db.ReadImage(newImg.Name)
		if err != nil {
			hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
			ResponseInternalError(w, err)
			return
		}
		if dbImg.ID != "" {
			hS.Logger.Warn("Request ", request, " completed with status ", http.StatusBadRequest, ": ", ErrImageExisted.Error())
			ResponseBadRequest(w, ErrImageExisted)
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
	err = hS.Db.UpdateImage(oldImg.ID, resImg)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	ResponseOK(w, resImg, request)
}

func (hS HttpServer) ImageDelete(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	imageName := params.ByName("imageName")
	request := "/images/" + imageName + " DELETE"
	hS.Logger.Info("Get " + request)

	image, err := hS.Db.ReadImage(imageName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}
	if image.Name == "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusNotFound, ": ", ErrImageNotFound.Error())
		ResponseNotFound(w, ErrImageNotFound)
		return
	}

	used, err := IsImageUsed(hS, imageName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}
	if used {
		hS.Logger.Warn("Request ", request, " completed with status ", http.StatusBadRequest, ": ", ErrImageUsed.Error())
		ResponseBadRequest(w, ErrImageUsed)
		return
	}

	err = hS.Db.DeleteImage(params.ByName("imageName"))
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusNoContent)
	ResponseNoContent(w)
}
