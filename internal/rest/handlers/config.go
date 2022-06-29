package handlers

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/jinzhu/copier"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

// service types:

func (hS HttpServer) ConfigsCreateServiceType(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	request := "/configs POST"
	hS.Logger.Info("Get " + request)

	var sType protobuf.ServiceType
	err := json.NewDecoder(r.Body).Decode(&sType)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrJsonIncorrect.Error())
		ResponseBadRequest(w, ErrJsonIncorrect)
		return
	}

	dbServiceType, err := hS.Db.ReadServiceType(sType.Type)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}
	if dbServiceType.Type != "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrServiceTypeExisted.Error())
		ResponseBadRequest(w, ErrServiceTypeExisted)
		return
	}

	hS.Logger.Info("Validating service type for creation...")
	err, status := ValidateServiceTypeCreate(hS, &sType)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", status, ": ", err.Error())
		switch status {
		case http.StatusBadRequest:
			ResponseBadRequest(w, err)
			return
		case http.StatusInternalServerError:
			ResponseInternalError(w, err)
			return
		}
	}

	// generate AnsibleVarName params in configs + generating UUID for new service version
	for i, sv := range sType.Versions {
		if sv.Configs != nil {
			for j, c := range sv.Configs {
				sType.Versions[i].Configs[j].AnsibleVarName = sType.Type + "_" + c.ParameterName
			}
		}

		vUuid, err := uuid.NewRandom()
		if err != nil {
			hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", ErrUuidLibError.Error())
			ResponseInternalError(w, ErrUuidLibError)
			return
		}
		sType.Versions[i].ID = vUuid.String()
	}

	// generating UUID for new service type
	stUuid, err := uuid.NewRandom()
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", ErrUuidLibError.Error())
		ResponseInternalError(w, ErrUuidLibError)
		return
	}
	sType.ID = stUuid.String()

	//saving new service type
	err = hS.Db.WriteServiceType(&sType)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	ResponseOK(w, sType, request)
}

func (hS HttpServer) ConfigsServiceTypesGetList(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	request := "/configs GET"
	hS.Logger.Info("Get " + request)

	sTypes, err := hS.Db.ReadServicesTypesList()
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	ResponseOK(w, sTypes, request)
}

func (hS HttpServer) ConfigsServiceTypeGet(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	serviceTypeIdOrName := params.ByName("serviceTypeIdOrName")
	queryViewValue := r.URL.Query().Get(QueryViewKey)

	request := "/configs/" + serviceTypeIdOrName
	if queryViewValue != "" {
		request += "?" + QueryViewKey + "=" + queryViewValue
	}
	request += " GET"
	hS.Logger.Info("Get " + request)

	if queryViewValue != "" {
		if queryViewValue != QueryViewTypeSummary && queryViewValue != QueryViewTypeFull {
			hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrGetQueryParams.Error())
			ResponseBadRequest(w, ErrGetQueryParams)
			return
		}
	}

	sType, err := hS.Db.ReadServiceType(serviceTypeIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}

	if sType.Type == "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrServiceTypeNotFound.Error())
		ResponseBadRequest(w, ErrServiceTypeNotFound)
		return
	}

	var resServiceType protobuf.ServiceType
	resServiceType.ID = sType.ID
	resServiceType.Type = sType.Type
	resServiceType.Description = sType.Description
	resServiceType.DefaultVersion = sType.DefaultVersion
	resServiceType.Class = sType.Class
	resServiceType.AccessPort = sType.AccessPort

	if queryViewValue == QueryViewTypeFull {
		resServiceType = *sType
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	ResponseOK(w, resServiceType, request)
}

// ConfigsServiceTypeUpdate updates only information about service type
// versions and config params could be updated in ConfigsServiceTypeVersionUpdate
func (hS HttpServer) ConfigsServiceTypeUpdate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	serviceTypeIdOrName := params.ByName("serviceTypeIdOrName")
	request := "/configs/" + serviceTypeIdOrName + " PUT"
	hS.Logger.Info("Get " + request)

	oldServiceType, err := hS.Db.ReadServiceType(serviceTypeIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}

	if oldServiceType.Type == "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusNotFound, ": ", ErrServiceTypeNotFound.Error())
		ResponseNotFound(w, ErrServiceTypeNotFound)
		return
	}

	var newServiceType protobuf.ServiceType
	err = json.NewDecoder(r.Body).Decode(&newServiceType)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrJsonIncorrect.Error())
		ResponseBadRequest(w, ErrJsonIncorrect)
		return
	}

	hS.Logger.Info("Validating service type values for update...")
	err, status := ValidateServiceTypeUpdate(oldServiceType, &newServiceType)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", status, ": ", err.Error())
		switch status {
		case http.StatusBadRequest:
			ResponseBadRequest(w, err)
			return
		case http.StatusInternalServerError:
			ResponseInternalError(w, err)
			return
		}
	}

	if newServiceType.Description != "" {
		oldServiceType.Description = newServiceType.Description
	}

	if newServiceType.DefaultVersion != "" {
		oldServiceType.DefaultVersion = newServiceType.DefaultVersion
	}

	if newServiceType.Class != "" {
		oldServiceType.Class = newServiceType.Class
	}

	if newServiceType.AccessPort != 0 { //0 if port not provided
		oldServiceType.AccessPort = newServiceType.AccessPort
	}

	if newServiceType.Ports != nil {
		if oldServiceType.Ports != nil {
			newPLen := len(newServiceType.Ports)
			for _, oldP := range oldServiceType.Ports {
				f := false
				for _, newP := range newServiceType.Ports[:newPLen] {
					if oldP.Port == newP.Port {
						f = true
						break
					}
				}
				//add old port if it hasn't been updated
				if !f {
					newServiceType.Ports = append(newServiceType.Ports, oldP)
				}
			}
		}
		oldServiceType.Ports = newServiceType.Ports
	}

	err = hS.Db.UpdateServiceType(oldServiceType)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	ResponseOK(w, oldServiceType, request)
}

func (hS HttpServer) ConfigsServiceTypeDelete(w http.ResponseWriter, _ *http.Request, params httprouter.Params) {
	serviceTypeIdOrName := params.ByName("serviceTypeIdOrName")
	request := "/configs/" + serviceTypeIdOrName + " GET"
	hS.Logger.Info("Get " + request)

	sType, err := hS.Db.ReadServiceType(serviceTypeIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}

	if sType.Type == "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusNotFound, ": ", ErrServiceTypeNotFound.Error())
		ResponseNotFound(w, ErrServiceTypeNotFound)
		return
	}

	hS.Logger.Info("Validating service type values for delete...")
	err, status := ValidateServiceTypeDelete(hS, sType.Type)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", status, ": ", err.Error())
		switch status {
		case http.StatusBadRequest:
			ResponseBadRequest(w, err)
			return
		case http.StatusInternalServerError:
			ResponseInternalError(w, err)
			return
		}
	}

	err = hS.Db.DeleteServiceType(serviceTypeIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusNoContent)
	ResponseNoContent(w)
}

// service type versions:

func (hS HttpServer) ConfigsServiceTypeVersionsGet(w http.ResponseWriter, _ *http.Request, params httprouter.Params) {
	serviceTypeIdOrName := params.ByName("serviceTypeIdOrName")
	request := "/configs/" + serviceTypeIdOrName + "/versions GET"
	hS.Logger.Info("Get " + request)

	sType, err := hS.Db.ReadServiceType(serviceTypeIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}

	if sType.Type == "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrServiceTypeNotFound.Error())
		ResponseBadRequest(w, ErrServiceTypeNotFound)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	ResponseOK(w, sType.Versions, request)
}

func (hS HttpServer) ConfigsServiceTypeVersionCreate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	serviceTypeIdOrName := params.ByName("serviceTypeIdOrName")
	request := "/configs/" + serviceTypeIdOrName + "/versions GET"
	hS.Logger.Info("Get " + request)

	sType, err := hS.Db.ReadServiceType(serviceTypeIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}

	if sType.Type == "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrServiceTypeNotFound.Error())
		ResponseBadRequest(w, ErrServiceTypeNotFound)
		return
	}

	var newServiceTypeVersion protobuf.ServiceVersion
	err = json.NewDecoder(r.Body).Decode(&newServiceTypeVersion)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrJsonIncorrect.Error())
		ResponseBadRequest(w, ErrJsonIncorrect)
		return
	}

	sVersion, err := hS.Db.ReadServiceTypeVersion(serviceTypeIdOrName, newServiceTypeVersion.Version)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}

	if sVersion.ID != "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrServiceTypeVersionExisted.Error())
		ResponseBadRequest(w, ErrServiceTypeVersionExisted)
		return
	}

	hS.Logger.Info("Validating service type version for creation...")
	err, status := ValidateServiceTypeVersionCreate(hS, sType.Versions, newServiceTypeVersion)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", status, ": ", err.Error())
		switch status {
		case http.StatusBadRequest:
			ResponseBadRequest(w, err)
			return
		case http.StatusInternalServerError:
			ResponseInternalError(w, err)
			return
		}
	}

	if newServiceTypeVersion.Configs != nil {
		for i, config := range newServiceTypeVersion.Configs {
			newServiceTypeVersion.Configs[i].AnsibleVarName = serviceTypeIdOrName + "_" + config.ParameterName
		}
	}

	// generating UUID for new service version
	vUuid, err := uuid.NewRandom()
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", ErrUuidLibError.Error())
		ResponseInternalError(w, ErrUuidLibError)
		return
	}
	newServiceTypeVersion.ID = vUuid.String()

	sType.Versions = append(sType.Versions, &newServiceTypeVersion)

	//saving updated service type
	err = hS.Db.UpdateServiceType(sType)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	ResponseOK(w, sType, request)
}

func (hS HttpServer) ConfigsServiceTypeVersionGet(w http.ResponseWriter, _ *http.Request, params httprouter.Params) {
	serviceTypeIdOrName := params.ByName("serviceTypeIdOrName")
	versionIdOrName := params.ByName("versionIdOrName")
	request := "/configs/" + serviceTypeIdOrName + "/versions/" + versionIdOrName + " GET"
	hS.Logger.Info("Get " + request)

	version, err := hS.Db.ReadServiceTypeVersion(serviceTypeIdOrName, versionIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}

	if version.ID == "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrServiceTypeVersionNotFound.Error())
		ResponseBadRequest(w, ErrServiceTypeVersionNotFound)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	ResponseOK(w, version, request)
}

func (hS HttpServer) ConfigsServiceTypeVersionUpdate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	serviceTypeIdOrName := params.ByName("serviceTypeIdOrName")
	versionIdOrName := params.ByName("versionIdOrName")
	request := "/configs/" + serviceTypeIdOrName + "/versions/" + versionIdOrName + " PUT"
	hS.Logger.Info("Get " + request)

	sType, err := hS.Db.ReadServiceType(serviceTypeIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}

	if sType.Type == "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusNotFound, ": ", ErrServiceTypeNotFound.Error())
		ResponseNotFound(w, ErrServiceTypeNotFound)
		return
	}

	oldServiceTypeVersion, err := hS.Db.ReadServiceTypeVersion(serviceTypeIdOrName, versionIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}

	if oldServiceTypeVersion.ID == "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrServiceTypeVersionNotFound.Error())
		ResponseBadRequest(w, ErrServiceTypeVersionNotFound)
		return
	}

	var newServiceTypeVersion protobuf.ServiceVersion
	err = json.NewDecoder(r.Body).Decode(&newServiceTypeVersion)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrJsonIncorrect.Error())
		ResponseBadRequest(w, ErrJsonIncorrect)
		return
	}

	hS.Logger.Info("Validating service type version values for update...")
	err, status := ValidateServiceTypeVersionUpdate(newServiceTypeVersion)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", status, ": ", err.Error())
		switch status {
		case http.StatusBadRequest:
			ResponseBadRequest(w, err)
			return
		case http.StatusInternalServerError:
			ResponseInternalError(w, err)
			return
		}
	}

	if newServiceTypeVersion.Description != "" {
		oldServiceTypeVersion.Description = newServiceTypeVersion.Description
	}

	if newServiceTypeVersion.DownloadURL != "" {
		oldServiceTypeVersion.DownloadURL = newServiceTypeVersion.DownloadURL
	}

	//saving updated service type
	err = hS.Db.UpdateServiceTypeVersion(serviceTypeIdOrName, oldServiceTypeVersion)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	ResponseOK(w, oldServiceTypeVersion, request)
}

func (hS HttpServer) ConfigsServiceTypeVersionDelete(w http.ResponseWriter, _ *http.Request, params httprouter.Params) {
	serviceTypeIdOrName := params.ByName("serviceTypeIdOrName")
	versionIdOrName := params.ByName("versionIdOrName")
	request := "/configs/" + serviceTypeIdOrName + "/versions/" + versionIdOrName + " DELETE"
	hS.Logger.Info("Get " + request)

	sType, err := hS.Db.ReadServiceType(serviceTypeIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}

	if sType.Type == "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusNotFound, ": ", ErrServiceTypeNotFound.Error())
		ResponseNotFound(w, ErrServiceTypeNotFound)
		return
	}

	serviceTypeVersion, err := hS.Db.ReadServiceTypeVersion(serviceTypeIdOrName, versionIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}

	if serviceTypeVersion.ID == "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusNotFound, ": ", ErrServiceTypeVersionNotFound.Error())
		ResponseNotFound(w, ErrServiceTypeVersionNotFound)
		return
	}

	hS.Logger.Info("Validating service type version values for delete...")
	err, status := ValidateServiceTypeVersionDelete(hS, sType, serviceTypeVersion)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", status, ": ", err.Error())
		switch status {
		case http.StatusBadRequest:
			ResponseBadRequest(w, err)
			return
		case http.StatusInternalServerError:
			ResponseInternalError(w, err)
			return
		}
	}

	err = hS.Db.DeleteServiceTypeVersion(serviceTypeIdOrName, versionIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusNoContent)
	ResponseNoContent(w)
}

func (hS HttpServer) ConfigsCreateConfigParam(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	sTypeName := params.ByName("serviceTypeIdOrName")
	vId := params.ByName("versionId")
	hS.Logger.Print("Get /configs/", sTypeName, "/versions/", vId, "/configs POST")

	var newStConfig *protobuf.ServiceConfig
	err := json.NewDecoder(r.Body).Decode(&newStConfig)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	newStConfig.AnsibleVarName = sTypeName + "_" + newStConfig.ParameterName

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

	//get service version idx in versions array
	flag := false
	var idToUpdate int
	var oldV protobuf.ServiceVersion
	for i, v := range st.Versions {
		if v.ID == vId {
			idToUpdate = i
			//used for deep copy
			copier.Copy(&oldV, &v)
			flag = true
			break
		}
	}

	if !flag {
		hS.Logger.Print("Service version not found")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	tmpC := make([]*protobuf.ServiceConfig, len(oldV.Configs))
	copy(tmpC, oldV.Configs)
	tmpC = append(tmpC, newStConfig)

	//check if configs array with new config param is ok
	if tmpC != nil {
		//check service version config
		err, _ = CheckConfigs(tmpC)
		if err != nil {
			hS.Logger.Print(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	st.Versions[idToUpdate].Configs = append(st.Versions[idToUpdate].Configs, newStConfig)
	//saving updated service type
	err = hS.Db.UpdateServiceType(st)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	err = enc.Encode(st.Versions[idToUpdate])
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
