package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	proto "github.com/ispras/michman/internal/protobuf"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func isFlavorUsed(hs HttpServer, flavorName string) bool {
	hs.Logger.Print("Checking is flavor used...")
	clusters, err := hs.Db.ListClusters()
	if err != nil {
		hs.Logger.Print(err)
		return false
	}
	for _, c := range clusters {
		if c.MasterFlavor == flavorName || c.StorageFlavor == flavorName ||
			c.SlavesFlavor == flavorName || c.MonitoringFlavor == flavorName {
			return true
		}
	}
	projects, err := hs.Db.ListProjects()
	if err != nil {
		hs.Logger.Print(err)
		return false
	}
	for _, p := range projects {
		if p.DefaultMasterFlavor == flavorName || p.DefaultStorageFlavor == flavorName ||
			p.DefaultSlavesFlavor == flavorName || p.DefaultMonitoringFlavor == flavorName {
			return true
		}
	}
	return false
}

func validateFlavor(hs HttpServer, flavor *proto.Flavor) (bool, error) {
	hs.Logger.Print("Validating flavor...")
	if flavor.ID != "" {
		msg := "ERROR: flavor ID is generated field."
		hs.Logger.Print(msg)
		return false, errors.New(msg)
	}
	if flavor.Name == "" {
		msg := "ERROR: flavor Name can't be empty."
		hs.Logger.Print(msg)
		return false, errors.New(msg)
	}

	switch v := interface{}(flavor.VCPUs).(type) {
	case int32:
		if flavor.VCPUs <= 0 {
			msg := "ERROR: flavor VCPUs can't be less than or equal to zero"
			hs.Logger.Print(msg)
			return false, errors.New(msg)
		}
	default:
		msg := fmt.Sprintf("ERROR: flavor VCPUs can't be %T type!\n", v)
		hs.Logger.Print(msg)
		return false, errors.New(msg)
	}

	switch v := interface{}(flavor.RAM).(type) {
	case int32:
		if flavor.RAM <= 0 {
			msg := "ERROR: flavor RAM can't be less than or equal to zero"
			hs.Logger.Print(msg)
			return false, errors.New(msg)
		}
	default:
		msg := fmt.Sprintf("ERROR: flavor RAM can't be %T type!\n", v)
		hs.Logger.Print(msg)
		return false, errors.New(msg)
	}

	switch v := interface{}(flavor.Disk).(type) {
	case int32:
		if flavor.Disk <= 0 {
			msg := "ERROR: flavor Disk can't be less than or equal to zero"
			hs.Logger.Print(msg)
			return false, errors.New(msg)
		}
	default:
		msg := fmt.Sprintf("ERROR: flavor Disk can't be %T type!\n", v)
		hs.Logger.Print(msg)
		return false, errors.New(msg)
	}
	return true, nil
}

func (hs HttpServer) FlavorsPost(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	hs.Logger.Print("Get /flavors POST")

	var flavor proto.Flavor
	err := json.NewDecoder(r.Body).Decode(&flavor)
	if err != nil {
		msg, _ := hs.ErrHandler.Handle(w, JSONerrorIncorrect, JSONerrorIncorrectMessage, err)
		hs.Logger.Print(msg)
		return
	}
	valid, err := validateFlavor(hs, &flavor)
	if !valid || err != nil {
		msg, _ := hs.ErrHandler.Handle(w, JSONerrorIncorrect, JSONerrorIncorrectMessage, err)
		hs.Logger.Print(msg)
		return
	}

	dbFlavor, _ := hs.FlavorGetByIdOrName(flavor.Name)
	if dbFlavor.ID != "" {
		msg, _ := hs.ErrHandler.Handle(w, FlavorExisted, FlavorExistedMessage, nil)
		hs.Logger.Print(msg)
		return
	}

	iUuid, err := uuid.NewRandom()
	if err != nil {
		msg, _ := hs.ErrHandler.Handle(w, LibErrorUUID, LibErrorUUIDMessage, err)
		hs.Logger.Print(msg)
		return
	}
	flavor.ID = iUuid.String()
	err = hs.Db.WriteFlavor(&flavor)
	if err != nil {
		msg, _ := hs.ErrHandler.Handle(w, DBerror, DBerrorMessage, err)
		hs.Logger.Print(msg)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	err = enc.Encode(flavor)
	if err != nil {
		msg, _ := hs.ErrHandler.Handle(w, LibErrorStructToJson, LibErrorStructToJsonMessage, err)
		hs.Logger.Print(msg)
		return
	}
}

func (hs HttpServer) FlavorsGetList(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	hs.Logger.Print("Get /flavors GET")

	flavors, err := hs.Db.ListFlavors()
	if err != nil {
		msg, _ := hs.ErrHandler.Handle(w, DBerror, DBerrorMessage, err)
		hs.Logger.Print(msg)
		return
	}
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	err = enc.Encode(flavors)
	if err != nil {
		msg, _ := hs.ErrHandler.Handle(w, LibErrorStructToJson, LibErrorStructToJsonMessage, err)
		hs.Logger.Print(msg)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

func (hs HttpServer) FlavorGetByIdOrName(idOrName string) (*proto.Flavor, error) {
	isUuid := true
	_, err := uuid.Parse(idOrName)
	if err != nil {
		isUuid = false
	}

	var flavor *proto.Flavor
	if isUuid {
		flavor, err = hs.Db.ReadFlavor(idOrName)
	} else {
		flavor, err = hs.Db.ReadFlavorByName(idOrName)
	}

	return flavor, err
}

func (hs HttpServer) FlavorGet(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	hs.Logger.Print("Get /flavors/:flavorIdOrName GET")
	flavorIdOrName := params.ByName("flavorIdOrName")

	flavor, err := hs.FlavorGetByIdOrName(flavorIdOrName)
	if err != nil {
		msg, _ := hs.ErrHandler.Handle(w, DBerror, DBerrorMessage, err)
		hs.Logger.Print(msg)
		return
	}
	if flavor.Name == "" {
		hs.Logger.Printf("Flavor with name '%s' not found", flavorIdOrName)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	err = enc.Encode(flavor)
	if err != nil {
		msg, _ := hs.ErrHandler.Handle(w, LibErrorStructToJson, LibErrorStructToJsonMessage, err)
		hs.Logger.Print(msg)
		return
	}
}

func (hs HttpServer) FlavorPut(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	hs.Logger.Print("Get /flavors/:flavorIdOrName PUT")
	flavorIdOrName := params.ByName("flavorIdOrName")
	oldFlavor, err := hs.FlavorGetByIdOrName(flavorIdOrName)
	if err != nil {
		msg, _ := hs.ErrHandler.Handle(w, DBerror, DBerrorMessage, err)
		hs.Logger.Print(msg)
		return
	}
	if oldFlavor.Name == "" {
		hs.Logger.Printf("Flavor with name '%s' not found", flavorIdOrName)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var newFlavor proto.Flavor
	err = json.NewDecoder(r.Body).Decode(&newFlavor)
	if err != nil {
		msg, _ := hs.ErrHandler.Handle(w, JSONerrorIncorrect, JSONerrorIncorrectMessage, err)
		hs.Logger.Print(msg)
		return
	}

	used := isFlavorUsed(hs, oldFlavor.Name)
	if used {
		msg, _ := hs.ErrHandler.Handle(w, FlavorUsed, FlavorUsedMessage, nil)
		hs.Logger.Print(msg)
		return
	}

	if newFlavor.ID != "" {
		msg, _ := hs.ErrHandler.Handle(w, FlavorUnmodField, FlavoUnmodFieldMessage, nil)
		hs.Logger.Print(msg)
		return
	}

	resFlavor := oldFlavor
	if newFlavor.Name != "" && newFlavor.Name != oldFlavor.Name {
		dbFlavor, _ := hs.FlavorGetByIdOrName(newFlavor.Name)
		if dbFlavor.ID != "" {
			msg, _ := hs.ErrHandler.Handle(w, FlavorExisted, FlavorExistedMessage, nil)
			hs.Logger.Print(msg)
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

	err = hs.Db.UpdateFlavor(oldFlavor.ID, resFlavor)
	if err != nil {
		msg, _ := hs.ErrHandler.Handle(w, DBerror, DBerrorMessage, err)
		hs.Logger.Print(msg)
		return
	}

	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	err = enc.Encode(resFlavor)
	if err != nil {
		msg, _ := hs.ErrHandler.Handle(w, LibErrorStructToJson, LibErrorStructToJsonMessage, err)
		hs.Logger.Print(msg)
		return
	}
	w.Header().Set("Content-Type", "application/json")

}

func (hs HttpServer) FlavorDelete(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	hs.Logger.Print("Get /flavors/:flavorIdOrName DELETE")
	flavorIdOrName := params.ByName("flavorIdOrName")

	flavor, err := hs.FlavorGetByIdOrName(flavorIdOrName)
	if err != nil {
		msg, _ := hs.ErrHandler.Handle(w, DBerror, DBerrorMessage, err)
		hs.Logger.Print(msg)
		return
	}

	if flavor.Name == "" {
		hs.Logger.Printf("Flavor with name '%s' not found", flavorIdOrName)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	used := isFlavorUsed(hs, flavor.Name)
	if used {
		msg, _ := hs.ErrHandler.Handle(w, FlavorUsed, FlavorUsedMessage, nil)
		hs.Logger.Print(msg)
		return
	}

	err = hs.Db.DeleteFlavor(flavor.Name)
	if err != nil {
		msg, _ := hs.ErrHandler.Handle(w, DBerror, DBerrorMessage, err)
		hs.Logger.Print(msg)
		return
	}
	w.WriteHeader(http.StatusOK)
}
