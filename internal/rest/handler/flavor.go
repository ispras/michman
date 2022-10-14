package handler

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/rest/handler/validate"
	response "github.com/ispras/michman/internal/rest/response"
	"github.com/ispras/michman/internal/utils"
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
		err = ErrJsonIncorrect
		hS.Logger.Warn("Request ", request, "failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	// check flavor does not exist in the database
	_, err = hS.Db.ReadFlavor(flavor.Name)
	if err == nil && response.ErrorClass(err) != utils.ObjectNotFound {
		err = ErrObjectExists("flavor", flavor.Name)
		hS.Logger.Warn("Request ", request, "failed with an error: ", err.Error())
		response.Error(w, err)
		return
	} else if err != nil && response.ErrorClass(err) != utils.ObjectNotFound {
		hS.Logger.Warn("Request ", request, "failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	hS.Logger.Info("Validating flavor...")
	err = validate.FlavorCreate(&flavor)
	if err != nil {
		hS.Logger.Warn("Request ", request, "failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	// generating UUID for new flavor
	iUuid, err := uuid.NewRandom()
	if err != nil {
		err = ErrUuidLibError
		hS.Logger.Warn("Request ", request, "failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}
	flavor.ID = iUuid.String()

	err = hS.Db.WriteFlavor(&flavor)
	if err != nil {
		hS.Logger.Warn("Request ", request, "failed with an error: ", err.Error())
		response.Error(w, err)
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
		hS.Logger.Warn("Request ", request, "failed with an error: ", err.Error())
		response.Error(w, err)
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
		hS.Logger.Warn("Request ", request, "failed with an error: ", err.Error())
		response.Error(w, err)
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

	// read flavor from database
	oldFlavor, err := hS.Db.ReadFlavor(flavorIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, "failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	var newFlavor *protobuf.Flavor
	err = json.NewDecoder(r.Body).Decode(&newFlavor)
	if err != nil {
		err = ErrJsonIncorrect
		hS.Logger.Warn("Request ", request, "failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	err = validate.FlavorUpdate(hS.Db, oldFlavor, newFlavor)
	if err != nil {
		hS.Logger.Warn("Request ", request, "failed with an error: ", err.Error())
		response.Error(w, err)
		return
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
		hS.Logger.Warn("Request ", request, "failed with an error: ", err.Error())
		response.Error(w, err)
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

	// read flavor from database
	flavor, err := hS.Db.ReadFlavor(flavorIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, "failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	err = validate.FlavorDelete(hS.Db, flavor)
	if err != nil {
		hS.Logger.Warn("Request ", request, "failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	err = hS.Db.DeleteFlavor(flavor.Name)
	if err != nil {
		hS.Logger.Warn("Request ", request, "failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}
	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusNoContent)
	response.NoContent(w)
}
