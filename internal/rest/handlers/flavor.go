package handlers

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (hS HttpServer) FlavorCreate(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	request := "/flavors POST"
	hS.Logger.Info("Get " + request)

	var flavor protobuf.Flavor
	err := json.NewDecoder(r.Body).Decode(&flavor)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrJsonIncorrect.Error())
		ResponseBadRequest(w, ErrJsonIncorrect)
		return
	}

	err = ValidateFlavor(hS, &flavor)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", err.Error())
		ResponseBadRequest(w, err)
		return
	}

	dbFlavor, err := FlavorGetByIdOrName(hS, flavor.Name)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}
	if dbFlavor.ID != "" {
		hS.Logger.Warn("Request ", request, " completed with status ", http.StatusBadRequest, ": ", ErrFlavorExisted.Error())
		ResponseBadRequest(w, ErrFlavorExisted)
		return
	}

	iUuid, err := uuid.NewRandom()
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", ErrUuidLibError.Error())
		ResponseInternalError(w, ErrUuidLibError)
		return
	}
	flavor.ID = iUuid.String()
	err = hS.Db.WriteFlavor(&flavor)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}
	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusCreated)
	ResponseCreated(w, flavor, request)
}

func (hS HttpServer) FlavorsGetList(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	request := "/flavors GET"
	hS.Logger.Info("Get " + request)

	flavors, err := hS.Db.ListFlavors()
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}
	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	ResponseOK(w, flavors, request)
}

func (hS HttpServer) FlavorGet(w http.ResponseWriter, _ *http.Request, params httprouter.Params) {
	flavorIdOrName := params.ByName("flavorIdOrName")
	hS.Logger.Info("Get /flavors/", flavorIdOrName, " GET")

	request := "/flavors/" + flavorIdOrName + " GET"

	flavor, err := FlavorGetByIdOrName(hS, flavorIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}
	if flavor.Name == "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusNotFound, ": ", ErrFlavorNotFound.Error())
		ResponseNotFound(w, ErrFlavorNotFound)
		return
	}
	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	ResponseOK(w, flavor, request)
}

func (hS HttpServer) FlavorUpdate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	flavorIdOrName := params.ByName("flavorIdOrName")
	hS.Logger.Info("Get /flavors/", flavorIdOrName, "PUT")

	request := "/flavors/" + flavorIdOrName + " PUT"

	oldFlavor, err := FlavorGetByIdOrName(hS, flavorIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}
	if oldFlavor.Name == "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusNotFound, ": ", ErrFlavorNotFound.Error())
		ResponseNotFound(w, ErrFlavorNotFound)
		return
	}

	var newFlavor protobuf.Flavor
	err = json.NewDecoder(r.Body).Decode(&newFlavor)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrJsonIncorrect.Error())
		ResponseBadRequest(w, ErrJsonIncorrect)
		return
	}

	used, err := IsFlavorUsed(hS, oldFlavor.Name)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", ErrJsonIncorrect.Error())
		ResponseInternalError(w, err)
		return
	}
	if used {
		hS.Logger.Warn("Request ", request, " completed with status ", http.StatusBadRequest, ": ", ErrFlavorUsed.Error())
		ResponseBadRequest(w, ErrFlavorUsed)
		return
	}

	if newFlavor.ID != "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrFlavorUnmodField.Error())
		ResponseBadRequest(w, ErrFlavorUnmodField)
		return
	}

	resFlavor := oldFlavor
	if newFlavor.Name != "" && newFlavor.Name != oldFlavor.Name {
		dbFlavor, err := FlavorGetByIdOrName(hS, newFlavor.Name)
		if err != nil {
			hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
			ResponseInternalError(w, err)
			return
		}
		if dbFlavor.ID != "" {
			hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrFlavorExisted.Error())
			ResponseBadRequest(w, ErrFlavorExisted)
			return
		}
		resFlavor.Name = newFlavor.Name
	}

	if newFlavor.VCPUs > 0 {
		resFlavor.VCPUs = newFlavor.VCPUs
	} else {
		hS.Logger.Warn(warnFlavorVCPU)
	}
	if newFlavor.RAM > 0 {
		resFlavor.RAM = newFlavor.RAM
	} else {
		hS.Logger.Warn(warnFlavorRAM)
	}
	if newFlavor.Disk > 0 {
		resFlavor.Disk = newFlavor.Disk
	} else {
		hS.Logger.Warn(warnFlavorDisk)
	}

	err = hS.Db.UpdateFlavor(oldFlavor.ID, resFlavor)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}
	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	ResponseOK(w, resFlavor, request)
}

func (hS HttpServer) FlavorDelete(w http.ResponseWriter, _ *http.Request, params httprouter.Params) {
	flavorIdOrName := params.ByName("flavorIdOrName")
	hS.Logger.Info("Get /flavors/", flavorIdOrName, "DELETE")

	request := "/flavors/" + flavorIdOrName + " DELETE"

	flavor, err := FlavorGetByIdOrName(hS, flavorIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}
	if flavor.Name == "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusNotFound, ": ", ErrFlavorNotFound.Error())
		ResponseNotFound(w, ErrFlavorNotFound)
		return
	}

	used, err := IsFlavorUsed(hS, flavor.Name)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}
	if used {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrFlavorUsed.Error())
		ResponseBadRequest(w, ErrFlavorUsed)
		return
	}

	err = hS.Db.DeleteFlavor(flavor.Name)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}
	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusNoContent)
	ResponseNoContent(w)
}
