package handlers

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (hS HttpServer) FlavorCreate(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	hS.Logger.Print("Get /flavors POST")

	var flavor protobuf.Flavor
	err := json.NewDecoder(r.Body).Decode(&flavor)
	if err != nil {
		msg, _ := hS.RespHandler.Handle(w, JSONerrorIncorrect, JSONerrorIncorrectMessage, err)
		hS.Logger.Print(msg)
		return
	}
	valid, err := ValidateFlavor(hS, &flavor)
	if !valid || err != nil {
		msg, _ := hS.RespHandler.Handle(w, JSONerrorIncorrect, JSONerrorIncorrectMessage, err)
		hS.Logger.Print(msg)
		return
	}

	dbFlavor, _ := FlavorGetByIdOrName(hS, flavor.Name)
	if dbFlavor.ID != "" {
		msg, _ := hS.RespHandler.Handle(w, FlavorExisted, FlavorExistedMessage, nil)
		hS.Logger.Print(msg)
		return
	}

	iUuid, err := uuid.NewRandom()
	if err != nil {
		msg, _ := hS.RespHandler.Handle(w, LibErrorUUID, LibErrorUUIDMessage, err)
		hS.Logger.Print(msg)
		return
	}
	flavor.ID = iUuid.String()
	err = hS.Db.WriteFlavor(&flavor)
	if err != nil {
		msg, _ := hS.RespHandler.Handle(w, DBerror, DBerrorMessage, err)
		hS.Logger.Print(msg)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	err = enc.Encode(flavor)
	if err != nil {
		msg, _ := hS.RespHandler.Handle(w, LibErrorStructToJson, LibErrorStructToJsonMessage, err)
		hS.Logger.Print(msg)
		return
	}
}

func (hS HttpServer) FlavorsGetList(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	hS.Logger.Print("Get /flavors GET")

	flavors, err := hS.Db.ListFlavors()
	if err != nil {
		msg, _ := hS.RespHandler.Handle(w, DBerror, DBerrorMessage, err)
		hS.Logger.Print(msg)
		return
	}
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	err = enc.Encode(flavors)
	if err != nil {
		msg, _ := hS.RespHandler.Handle(w, LibErrorStructToJson, LibErrorStructToJsonMessage, err)
		hS.Logger.Print(msg)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

func (hS HttpServer) FlavorGet(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	hS.Logger.Print("Get /flavors/:flavorIdOrName GET")
	flavorIdOrName := params.ByName("flavorIdOrName")

	flavor, err := FlavorGetByIdOrName(hS, flavorIdOrName)
	if err != nil {
		msg, _ := hS.RespHandler.Handle(w, DBerror, DBerrorMessage, err)
		hS.Logger.Print(msg)
		return
	}
	if flavor.Name == "" {
		hS.Logger.Printf("Flavor with name '%s' not found", flavorIdOrName)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	err = enc.Encode(flavor)
	if err != nil {
		msg, _ := hS.RespHandler.Handle(w, LibErrorStructToJson, LibErrorStructToJsonMessage, err)
		hS.Logger.Print(msg)
		return
	}
}

func (hS HttpServer) FlavorUpdate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	hS.Logger.Print("Get /flavors/:flavorIdOrName PUT")
	flavorIdOrName := params.ByName("flavorIdOrName")
	oldFlavor, err := FlavorGetByIdOrName(hS, flavorIdOrName)
	if err != nil {
		msg, _ := hS.RespHandler.Handle(w, DBerror, DBerrorMessage, err)
		hS.Logger.Print(msg)
		return
	}
	if oldFlavor.Name == "" {
		hS.Logger.Printf("Flavor with name '%s' not found", flavorIdOrName)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var newFlavor protobuf.Flavor
	err = json.NewDecoder(r.Body).Decode(&newFlavor)
	if err != nil {
		msg, _ := hS.RespHandler.Handle(w, JSONerrorIncorrect, JSONerrorIncorrectMessage, err)
		hS.Logger.Print(msg)
		return
	}

	used, _ := IsFlavorUsed(hS, oldFlavor.Name)
	if used {
		msg, _ := hS.RespHandler.Handle(w, FlavorUsed, FlavorUsedMessage, nil)
		hS.Logger.Print(msg)
		return
	}

	if newFlavor.ID != "" {
		msg, _ := hS.RespHandler.Handle(w, FlavorUnmodField, FlavorUnmodFieldMessage, nil)
		hS.Logger.Print(msg)
		return
	}

	resFlavor := oldFlavor
	if newFlavor.Name != "" && newFlavor.Name != oldFlavor.Name {
		dbFlavor, _ := FlavorGetByIdOrName(hS, newFlavor.Name)
		if dbFlavor.ID != "" {
			msg, _ := hS.RespHandler.Handle(w, FlavorExisted, FlavorExistedMessage, nil)
			hS.Logger.Print(msg)
			return
		}
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
		msg, _ := hS.RespHandler.Handle(w, DBerror, DBerrorMessage, err)
		hS.Logger.Print(msg)
		return
	}

	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	err = enc.Encode(resFlavor)
	if err != nil {
		msg, _ := hS.RespHandler.Handle(w, LibErrorStructToJson, LibErrorStructToJsonMessage, err)
		hS.Logger.Print(msg)
		return
	}
	w.Header().Set("Content-Type", "application/json")

}

func (hS HttpServer) FlavorDelete(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	hS.Logger.Print("Get /flavors/:flavorIdOrName DELETE")
	flavorIdOrName := params.ByName("flavorIdOrName")

	flavor, err := FlavorGetByIdOrName(hS, flavorIdOrName)
	if err != nil {
		msg, _ := hS.RespHandler.Handle(w, DBerror, DBerrorMessage, err)
		hS.Logger.Print(msg)
		return
	}

	if flavor.Name == "" {
		hS.Logger.Printf("Flavor with name '%s' not found", flavorIdOrName)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	used, _ := IsFlavorUsed(hS, flavor.Name)
	if used {
		msg, _ := hS.RespHandler.Handle(w, FlavorUsed, FlavorUsedMessage, nil)
		hS.Logger.Print(msg)
		return
	}

	err = hS.Db.DeleteFlavor(flavor.Name)
	if err != nil {
		msg, _ := hS.RespHandler.Handle(w, DBerror, DBerrorMessage, err)
		hS.Logger.Print(msg)
		return
	}
	w.WriteHeader(http.StatusOK)
}
