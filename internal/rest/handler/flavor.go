package handler

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/rest/handler/check"
	"github.com/ispras/michman/internal/rest/handler/response"
	"github.com/ispras/michman/internal/rest/handler/validate"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

// FlavorCreate processes a request to create a flavor struct in database
func (hS HttpServer) FlavorCreate(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	request := "POST /flavors"
	hS.Logger.Info(request)

	var flavor protobuf.Flavor
	err := json.NewDecoder(r.Body).Decode(&flavor)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrJsonIncorrect.Error())
		response.BadRequest(w, ErrJsonIncorrect)
		return
	}

	hS.Logger.Info("Validating flavor...")
	err = validate.FlavorCreate(&flavor)
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

// FlavorsGetList processes a request to get a list of all flavors in database
func (hS HttpServer) FlavorsGetList(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	request := "GET /flavors"
	hS.Logger.Info(request)

	flavors, err := hS.Db.ReadFlavorsList()
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		response.InternalError(w, err)
		return
	}
	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	response.Ok(w, flavors, request)
}

// FlavorGet processes a request to get a flavor struct by id or name from database
func (hS HttpServer) FlavorGet(w http.ResponseWriter, _ *http.Request, params httprouter.Params) {
	flavorIdOrName := params.ByName("flavorIdOrName")
	request := "GET /flavors/" + flavorIdOrName
	hS.Logger.Info(request)

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

// FlavorUpdate processes a request to update a flavor struct in database
func (hS HttpServer) FlavorUpdate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	flavorIdOrName := params.ByName("flavorIdOrName")
	request := "PUT /flavors/" + flavorIdOrName
	hS.Logger.Info(request)

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

	var newFlavor *protobuf.Flavor
	err = json.NewDecoder(r.Body).Decode(&newFlavor)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrJsonIncorrect.Error())
		response.BadRequest(w, ErrJsonIncorrect)
		return
	}

	if newFlavor.ID != "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrFlavorUnmodField.Error())
		response.BadRequest(w, ErrFlavorUnmodField)
		return
	}

	used, err := check.FlavorUsed(hS.Db, hS.Logger, oldFlavor.Name)
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

	err, status := validate.FlavorUpdate(hS.Db, oldFlavor, newFlavor)
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

	resFlavor := oldFlavor
	if newFlavor.Name != "" {
		resFlavor.Name = newFlavor.Name
	}
	if newFlavor.VCPUs > 0 {
		resFlavor.VCPUs = newFlavor.VCPUs
	}
	if newFlavor.RAM > 0 {
		resFlavor.RAM = newFlavor.RAM
	}
	if newFlavor.Disk > 0 {
		resFlavor.Disk = newFlavor.Disk
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

// FlavorDelete processes a request to delete a flavor struct from database
func (hS HttpServer) FlavorDelete(w http.ResponseWriter, _ *http.Request, params httprouter.Params) {
	flavorIdOrName := params.ByName("flavorIdOrName")
	request := "DELETE /flavors/" + flavorIdOrName
	hS.Logger.Info(request)

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

	used, err := check.FlavorUsed(hS.Db, hS.Logger, flavor.Name)
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
