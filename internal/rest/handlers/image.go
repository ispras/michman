package handlers

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (hS HttpServer) ImagesGetList(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	hS.Logger.Print("Get /images GET")

	images, err := hS.Db.ListImages()
	if err != nil {
		msg, _ := hS.RespHandler.Handle(w, DBerror, DBerrorMessage, err)
		hS.Logger.Print(msg)
		return
	}
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	err = enc.Encode(images)
	if err != nil {
		msg, _ := hS.RespHandler.Handle(w, LibErrorStructToJson, LibErrorStructToJsonMessage, err)
		hS.Logger.Print(msg)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

func (hS HttpServer) ImageGet(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	hS.Logger.Print("Get /images/:imageName GET")
	image, err := hS.Db.ReadImage(params.ByName("imageName"))
	if err != nil {
		msg, _ := hS.RespHandler.Handle(w, DBerror, DBerrorMessage, err)
		hS.Logger.Print(msg)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	err = enc.Encode(image)
	if err != nil {
		msg, _ := hS.RespHandler.Handle(w, LibErrorStructToJson, LibErrorStructToJsonMessage, err)
		hS.Logger.Print(msg)
		return
	}
}

func (hS HttpServer) ImageCreate(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	hS.Logger.Print("Get /images POST")

	var image protobuf.Image
	err := json.NewDecoder(r.Body).Decode(&image)
	if err != nil {
		msg, _ := hS.RespHandler.Handle(w, JSONerrorIncorrect, JSONerrorIncorrectMessage, err)
		hS.Logger.Print(msg)
		return
	}
	valid, err := ValidateImage(hS, &image)
	if !valid || err != nil {
		msg, _ := hS.RespHandler.Handle(w, JSONerrorIncorrect, JSONerrorIncorrectMessage, err)
		hS.Logger.Print(msg)
		return
	}
	dbImg, _ := hS.Db.ReadImage(image.Name)
	if dbImg.ID != "" {
		msg, _ := hS.RespHandler.Handle(w, ImageExisted, ImageExistedMessage, nil)
		hS.Logger.Print(msg)
		return
	}
	iUuid, err := uuid.NewRandom()
	if err != nil {
		msg, _ := hS.RespHandler.Handle(w, LibErrorUUID, LibErrorUUIDMessage, err)
		hS.Logger.Print(msg)
		return
	}
	image.ID = iUuid.String()
	err = hS.Db.WriteImage(&image)
	if err != nil {
		msg, _ := hS.RespHandler.Handle(w, DBerror, DBerrorMessage, err)
		hS.Logger.Print(msg)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	err = enc.Encode(image)
	if err != nil {
		msg, _ := hS.RespHandler.Handle(w, LibErrorStructToJson, LibErrorStructToJsonMessage, err)
		hS.Logger.Print(msg)
		return
	}
}

func (hS HttpServer) ImageUpdate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	hS.Logger.Print("Get /images/:imageName PUT")

	oldImg, err := hS.Db.ReadImage(params.ByName("imageName"))
	if err != nil {
		msg, _ := hS.RespHandler.Handle(w, DBerror, DBerrorMessage, err)
		hS.Logger.Print(msg)
		return
	}
	var newImg protobuf.Image
	err = json.NewDecoder(r.Body).Decode(&newImg)
	if err != nil {
		msg, _ := hS.RespHandler.Handle(w, JSONerrorIncorrect, JSONerrorIncorrectMessage, err)
		hS.Logger.Print(msg)
		return
	}
	used := IsImageUsed(hS, oldImg.Name)
	if used {
		msg, _ := hS.RespHandler.Handle(w, ImageUsed, ImageUsedMessage, nil)
		hS.Logger.Print(msg)
		return
	}
	resImg := oldImg
	if newImg.ID != "" {
		msg, _ := hS.RespHandler.Handle(w, ImageUnmodField, ImageUnmodFieldMessage, nil)
		hS.Logger.Print(msg)
		return
	}
	if newImg.Name != "" {
		dbImg, _ := hS.Db.ReadImage(newImg.Name)
		if dbImg.ID != "" {
			msg, _ := hS.RespHandler.Handle(w, ImageExisted, ImageExistedMessage, nil)
			hS.Logger.Print(msg)
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
		msg, _ := hS.RespHandler.Handle(w, DBerror, DBerrorMessage, err)
		hS.Logger.Print(msg)
		return
	}
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	err = enc.Encode(resImg)
	if err != nil {
		msg, _ := hS.RespHandler.Handle(w, LibErrorStructToJson, LibErrorStructToJsonMessage, err)
		hS.Logger.Print(msg)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

func (hS HttpServer) ImageDelete(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	hS.Logger.Print("Get /images/:imageName DELETE")

	used := IsImageUsed(hS, params.ByName("imageName"))
	if used {
		msg, _ := hS.RespHandler.Handle(w, ImageUsed, ImageUsedMessage, nil)
		hS.Logger.Print(msg)
		return
	}
	err := hS.Db.DeleteImage(params.ByName("imageName"))
	if err != nil {
		msg, _ := hS.RespHandler.Handle(w, DBerror, DBerrorMessage, err)
		hS.Logger.Print(msg)
		return
	}
	w.WriteHeader(http.StatusOK)
}
