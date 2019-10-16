package handlers

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"net/http"
	protobuf "gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/protobuf"
	"github.com/google/uuid"
)

func (hS HttpServer) ConfigsCreateService(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	hS.Logger.Print("Get /configs POST")
	var st protobuf.ServiceType
	err := json.NewDecoder(r.Body).Decode(&st)
	if err != nil {
		hS.Logger.Print("ERROR:")
		hS.Logger.Print(err)
		hS.Logger.Print(r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//check, that service type with such type doesn't exist
	dbRes, err := hS.Db.ReadServiceType(st.Type)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if dbRes.Type != "" {
		hS.Logger.Print("Service with this type exists")
		w.WriteHeader(http.StatusBadRequest)
		enc := json.NewEncoder(w)
		err := enc.Encode("Service with this type exists")
		if err != nil {
			hS.Logger.Print(err)
			w.WriteHeader(http.StatusBadRequest)
		}
		return
	}

	// generating UUID for new service type
	stUuid, err := uuid.NewRandom()
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
	}
	st.ID = stUuid.String()

	//saving new service type
	err = hS.Db.WriteServiceType(&st)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(st)
}

func (hS HttpServer) ConfigsGetServices(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	hS.Logger.Print("Get /configs GET")
	//reading service types info from database
	hS.Logger.Print("Reading information about services types from db...")

	sTypes, err := hS.Db.ListServicesTypes()
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	err = enc.Encode(sTypes)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

}

func (hS HttpServer) ConfigsGetService(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	sTypeName := params.ByName("serviceType")
	hS.Logger.Print("Get /cconfigs/", sTypeName, " GET")

	//reading service type info from database
	hS.Logger.Print("Reading service types information from db...")
	st, err := hS.Db.ReadServiceType(sTypeName)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if st.Type == "" {
		hS.Logger.Print("Service type not found")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	err = enc.Encode(st)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (hS HttpServer) ConfigsDeleteService(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	sTypeName := params.ByName("serviceType")
	hS.Logger.Print("Get /cconfigs/", sTypeName, " GET")

	//reading service type info from database
	hS.Logger.Print("Reading service types information from db...")
	st, err := hS.Db.ReadServiceType(sTypeName)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if st.Type == "" {
		hS.Logger.Print("Service type not found")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	//TODO: Delete all information about this service type (all versions with configs) from configs bucket

	err = hS.Db.DeleteServiceType(sTypeName)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (hS HttpServer) ConfigsCreateVersion(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

}

func (hS HttpServer) ConfigsGetVersions(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

}

func (hS HttpServer) ConfigsGetVersion(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

}

func (hS HttpServer) ConfigsUpdateVersion(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

}

func (hS HttpServer) ConfigsDeleteVersion(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

}