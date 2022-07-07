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

func (hS HttpServer) FlavorCreate(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	request := "/flavors POST"
	hS.Logger.Info("Get " + request)

	var flavor protobuf.Flavor
	err := json.NewDecoder(r.Body).Decode(&flavor)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrJsonIncorrect.Error())
		response.BadRequest(w, ErrJsonIncorrect)
		return
	}

	err = validate.Flavor(hS.Logger, &flavor)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", err.Error())
		response.BadRequest(w, err)
		return
	}

	dbFlavor, err := hS.Db.ReadFlavor(flavor.Name)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		response.InternalError(w, err)
		return
	}
	if dbFlavor.ID != "" {
		hS.Logger.Warn("Request ", request, " completed with status ", http.StatusBadRequest, ": ", ErrFlavorExisted.Error())
		response.BadRequest(w, ErrFlavorExisted)
		return
	}

	iUuid, err := uuid.NewRandom()
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", ErrUuidLibError.Error())
		response.InternalError(w, ErrUuidLibError)
		return
	}
	flavor.ID = iUuid.String()
	err = hS.Db.WriteFlavor(&flavor)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		response.InternalError(w, err)
		return
	}
	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusCreated)
	response.Created(w, flavor, request)
}

func (hS HttpServer) FlavorsGetList(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	request := "/flavors GET"
	hS.Logger.Info("Get " + request)

	flavors, err := hS.Db.ReadFlavorsList()
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		response.InternalError(w, err)
		return
	}
	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	response.Ok(w, flavors, request)
}

func (hS HttpServer) FlavorGet(w http.ResponseWriter, _ *http.Request, params httprouter.Params) {
	flavorIdOrName := params.ByName("flavorIdOrName")
	hS.Logger.Info("Get /flavors/", flavorIdOrName, " GET")

	request := "/flavors/" + flavorIdOrName + " GET"

	flavor, err := hS.Db.ReadFlavor(flavorIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		response.InternalError(w, err)
		return
	}
	if flavor.Name == "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusNotFound, ": ", ErrFlavorNotFound.Error())
		response.NotFound(w, ErrFlavorNotFound)
		return
	}
	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	response.Ok(w, flavor, request)
}

func (hS HttpServer) FlavorUpdate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	flavorIdOrName := params.ByName("flavorIdOrName")
	hS.Logger.Info("Get /flavors/", flavorIdOrName, "PUT")

	request := "/flavors/" + flavorIdOrName + " PUT"

	oldFlavor, err := hS.Db.ReadFlavor(flavorIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		response.InternalError(w, err)
		return
	}
	if oldFlavor.Name == "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusNotFound, ": ", ErrFlavorNotFound.Error())
		response.NotFound(w, ErrFlavorNotFound)
		return
	}

	var newFlavor protobuf.Flavor
	err = json.NewDecoder(r.Body).Decode(&newFlavor)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrJsonIncorrect.Error())
		response.BadRequest(w, ErrJsonIncorrect)
		return
	}

	used, err := helpfunc.IsFlavorUsed(hS.Db, hS.Logger, oldFlavor.Name)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", ErrJsonIncorrect.Error())
		response.InternalError(w, err)
		return
	}
	if used {
		hS.Logger.Warn("Request ", request, " completed with status ", http.StatusBadRequest, ": ", ErrFlavorUsed.Error())
		response.BadRequest(w, ErrFlavorUsed)
		return
	}

	if newFlavor.ID != "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrFlavorUnmodField.Error())
		response.BadRequest(w, ErrFlavorUnmodField)
		return
	}

	resFlavor := oldFlavor
	if newFlavor.Name != "" && newFlavor.Name != oldFlavor.Name {
		dbFlavor, err := hS.Db.ReadFlavor(newFlavor.Name)
		if err != nil {
			hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
			response.InternalError(w, err)
			return
		}
		if dbFlavor.ID != "" {
			hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrFlavorExisted.Error())
			response.BadRequest(w, ErrFlavorExisted)
			return
		}
		resFlavor.Name = newFlavor.Name
	}

	if newFlavor.VCPUs > 0 {
		resFlavor.VCPUs = newFlavor.VCPUs
	} else {
		hS.Logger.Warn(WarnFlavorVCPU)
	}
	if newFlavor.RAM > 0 {
		resFlavor.RAM = newFlavor.RAM
	} else {
		hS.Logger.Warn(WarnFlavorRAM)
	}
	if newFlavor.Disk > 0 {
		resFlavor.Disk = newFlavor.Disk
	} else {
		hS.Logger.Warn(WarnFlavorDisk)
	}

	err = hS.Db.UpdateFlavor(oldFlavor.ID, resFlavor)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		response.InternalError(w, err)
		return
	}
	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	response.Ok(w, resFlavor, request)
}

func (hS HttpServer) FlavorDelete(w http.ResponseWriter, _ *http.Request, params httprouter.Params) {
	flavorIdOrName := params.ByName("flavorIdOrName")
	hS.Logger.Info("Get /flavors/", flavorIdOrName, "DELETE")

	request := "/flavors/" + flavorIdOrName + " DELETE"
	flavor, err := hS.Db.ReadFlavor(flavorIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		response.InternalError(w, err)
		return
	}
	if flavor.Name == "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusNotFound, ": ", ErrFlavorNotFound.Error())
		response.NotFound(w, ErrFlavorNotFound)
		return
	}

	used, err := helpfunc.IsFlavorUsed(hS.Db, hS.Logger, flavor.Name)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		response.InternalError(w, err)
		return
	}
	if used {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrFlavorUsed.Error())
		response.BadRequest(w, ErrFlavorUsed)
		return
	}

	err = hS.Db.DeleteFlavor(flavor.Name)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		response.InternalError(w, err)
		return
	}
	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusNoContent)
	response.NoContent(w)
}
