package handlers

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	proto "github.com/ispras/michman/internal/protobuf"
	"net/http"
)

func validateImage(hs HttpServer, image *proto.Image) (bool, error) {
	hs.Logger.Print("Validating image...")
	if image.ID != "" {
		msg := "ERROR: image ID is generated field."
		hs.Logger.Print(msg)
		return false, errors.New(msg)
	}
	if image.Name == "" {
		msg := "ERROR: image Name can't be empty."
		hs.Logger.Print(msg)
		return false, errors.New(msg)
	}
	if image.AnsibleUser == "" {
		msg := "ERROR: image AnsibleUser can't be empty."
		hs.Logger.Print(msg)
		return false, errors.New(msg)
	}
	if image.CloudImageID == "" {
		msg := "ERROR: image ImageID can't be empty."
		hs.Logger.Print(msg)
		return false, errors.New(msg)
	}
	return true, nil
}

func isImageUsed(hs HttpServer, name string) bool {
	hs.Logger.Print("Checking is image used...")
	clusters, err := hs.Db.ListClusters()
	if err != nil {
		hs.Logger.Print(err)
		return false
	}
	for _, c := range clusters {
		if c.Image == name {
			return true
		}
	}
	projects, err := hs.Db.ListProjects()
	if err != nil {
		hs.Logger.Print(err)
		return false
	}
	for _, p := range projects {
		if p.DefaultImage == name {
			return true
		}
	}
	return false
}

func (hs HttpServer) ImagesGetList(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	hs.Logger.Print("Get /images GET")

	images, err := hs.Db.ListImages()
	if err != nil {
		msg, _ := hs.ErrHandler.Handle(w, DBerror, DBerrorMessage, err)
		hs.Logger.Print(msg)
		return
	}
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	err = enc.Encode(images)
	if err != nil {
		msg, _ := hs.ErrHandler.Handle(w, LibErrorStructToJson, LibErrorStructToJsonMessage, err)
		hs.Logger.Print(msg)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

func (hs HttpServer) ImageGet(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	hs.Logger.Print("Get /images/:imageName GET")
	image, err := hs.Db.ReadImage(params.ByName("imageName"))
	if err != nil {
		msg, _ := hs.ErrHandler.Handle(w, DBerror, DBerrorMessage, err)
		hs.Logger.Print(msg)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	err = enc.Encode(image)
	if err != nil {
		msg, _ := hs.ErrHandler.Handle(w, LibErrorStructToJson, LibErrorStructToJsonMessage, err)
		hs.Logger.Print(msg)
		return
	}
}

func (hs HttpServer) ImagesPost(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	hs.Logger.Print("Get /images POST")

	var image proto.Image
	err := json.NewDecoder(r.Body).Decode(&image)
	if err != nil {
		msg, _ := hs.ErrHandler.Handle(w, JSONerrorIncorrect, JSONerrorIncorrectMessage, err)
		hs.Logger.Print(msg)
		return
	}
	valid, err := validateImage(hs, &image)
	if !valid || err != nil {
		msg, _ := hs.ErrHandler.Handle(w, JSONerrorIncorrect, JSONerrorIncorrectMessage, err)
		hs.Logger.Print(msg)
		return
	}
	dbImg, _ := hs.Db.ReadImage(image.Name)
	if dbImg.ID != "" {
		msg, _ := hs.ErrHandler.Handle(w, ImageExisted, ImageExistedMessage, nil)
		hs.Logger.Print(msg)
		return
	}
	iUuid, err := uuid.NewRandom()
	if err != nil {
		msg, _ := hs.ErrHandler.Handle(w, LibErrorUUID, LibErrorUUIDMessage, err)
		hs.Logger.Print(msg)
		return
	}
	image.ID = iUuid.String()
	err = hs.Db.WriteImage(&image)
	if err != nil {
		msg, _ := hs.ErrHandler.Handle(w, DBerror, DBerrorMessage, err)
		hs.Logger.Print(msg)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	err = enc.Encode(image)
	if err != nil {
		msg, _ := hs.ErrHandler.Handle(w, LibErrorStructToJson, LibErrorStructToJsonMessage, err)
		hs.Logger.Print(msg)
		return
	}
}

func (hs HttpServer) ImagePut(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	hs.Logger.Print("Get /images/:imageName PUT")

	oldImg, err := hs.Db.ReadImage(params.ByName("imageName"))
	if err != nil {
		msg, _ := hs.ErrHandler.Handle(w, DBerror, DBerrorMessage, err)
		hs.Logger.Print(msg)
		return
	}
	var newImg proto.Image
	err = json.NewDecoder(r.Body).Decode(&newImg)
	if err != nil {
		msg, _ := hs.ErrHandler.Handle(w, JSONerrorIncorrect, JSONerrorIncorrectMessage, err)
		hs.Logger.Print(msg)
		return
	}
	used := isImageUsed(hs, oldImg.Name)
	if used {
		msg, _ := hs.ErrHandler.Handle(w, ImageUsed, ImageUsedMessage, nil)
		hs.Logger.Print(msg)
		return
	}
	resImg := oldImg
	if newImg.ID != "" {
		msg, _ := hs.ErrHandler.Handle(w, ImageUnmodField, ImageUnmodFieldMessage, nil)
		hs.Logger.Print(msg)
		return
	}
	if newImg.Name != "" {
		dbImg, _ := hs.Db.ReadImage(newImg.Name)
		if dbImg.ID != "" {
			msg, _ := hs.ErrHandler.Handle(w, ImageExisted, ImageExistedMessage, nil)
			hs.Logger.Print(msg)
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
	err = hs.Db.UpdateImage(oldImg.ID, resImg)
	if err != nil {
		msg, _ := hs.ErrHandler.Handle(w, DBerror, DBerrorMessage, err)
		hs.Logger.Print(msg)
		return
	}
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	err = enc.Encode(resImg)
	if err != nil {
		msg, _ := hs.ErrHandler.Handle(w, LibErrorStructToJson, LibErrorStructToJsonMessage, err)
		hs.Logger.Print(msg)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

func (hs HttpServer) ImageDelete(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	hs.Logger.Print("Get /image/:imageName DELETE")

	used := isImageUsed(hs, params.ByName("imageName"))
	if used {
		msg, _ := hs.ErrHandler.Handle(w, ImageUsed, ImageUsedMessage, nil)
		hs.Logger.Print(msg)
		return
	}
	err := hs.Db.DeleteImage(params.ByName("imageName"))
	if err != nil {
		msg, _ := hs.ErrHandler.Handle(w, DBerror, DBerrorMessage, err)
		hs.Logger.Print(msg)
		return
	}
	w.WriteHeader(http.StatusOK)
}
