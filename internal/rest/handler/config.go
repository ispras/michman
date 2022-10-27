package handler

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/rest/handler/check"
	"github.com/ispras/michman/internal/rest/handler/validate"
	response "github.com/ispras/michman/internal/rest/response"
	"github.com/ispras/michman/internal/utils"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

// service types:

// ConfigsServiceTypeGet processes a request to get a service type struct by id or name from database
func (hS HttpServer) ConfigsServiceTypeGet(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	serviceTypeIdOrName := params.ByName("serviceTypeIdOrName")
	request := "GET /configs/" + serviceTypeIdOrName
	queryViewValue := r.URL.Query().Get(QueryViewKey)

	if queryViewValue != "" {
		request += "?" + QueryViewKey + "=" + queryViewValue
	}
	hS.Logger.Info(request)

	if queryViewValue != "" {
		if queryViewValue != QueryViewTypeSummary && queryViewValue != QueryViewTypeFull {
			err := ErrGetQueryParams
			hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
			response.Error(w, err)
			return
		}
	}

	// read service type from database
	sType, err := hS.Db.ReadServiceType(serviceTypeIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	// if query view type is set to summary then output partial information (and it's a default value)
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
	response.Ok(w, resServiceType, request)
}

// ConfigsServiceTypesGetList processes a request to get a list of all service types in database
func (hS HttpServer) ConfigsServiceTypesGetList(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	request := "GET /configs"
	hS.Logger.Info(request)

	sTypes, err := hS.Db.ReadServicesTypesList()
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	response.Ok(w, sTypes, request)
}

// ConfigsServiceTypeCreate processes a request to create a service type struct in database
func (hS HttpServer) ConfigsServiceTypeCreate(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	request := "POST /configs"
	hS.Logger.Info(request)

	var sType protobuf.ServiceType
	err := json.NewDecoder(r.Body).Decode(&sType)
	if err != nil {
		err = ErrJsonIncorrect
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	// check service type does not exist in the database
	dbServiceType, err := hS.Db.ReadServiceType(sType.Type)
	if dbServiceType != nil {
		err = ErrObjectExists("service type", sType.Type)
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}
	if err != nil && response.ErrorClass(err) != utils.ObjectNotFound {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	hS.Logger.Info("Validating service type for creation...")
	err = validate.ServiceTypeCreate(hS.Db, &sType)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
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
			err = ErrUuidLibError
			hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
			response.Error(w, err)
			return
		}
		sType.Versions[i].ID = vUuid.String()
	}

	// generating UUID for new service type
	stUuid, err := uuid.NewRandom()
	if err != nil {
		err = ErrUuidLibError
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}
	sType.ID = stUuid.String()

	//saving new service type
	err = hS.Db.WriteServiceType(&sType)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	response.Created(w, sType, request)
}

// ConfigsServiceTypeUpdate updates only information about service type;
// versions could be updated in ConfigsServiceTypeVersionUpdate
// configs could be updated in ConfigsServiceTypeVersionConfigUpdate
func (hS HttpServer) ConfigsServiceTypeUpdate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	serviceTypeIdOrName := params.ByName("serviceTypeIdOrName")
	request := "PUT /configs/" + serviceTypeIdOrName
	hS.Logger.Info(request)

	// read service type from database
	oldServiceType, err := hS.Db.ReadServiceType(serviceTypeIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	var newServiceType protobuf.ServiceType
	err = json.NewDecoder(r.Body).Decode(&newServiceType)
	if err != nil {
		err = ErrJsonIncorrect
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	hS.Logger.Info("Validating service type values for update...")
	err = validate.ServiceTypeUpdate(oldServiceType, &newServiceType)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
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
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	response.Ok(w, oldServiceType, request)
}

// ConfigsServiceTypeDelete processes a request to delete a service type struct from database
func (hS HttpServer) ConfigsServiceTypeDelete(w http.ResponseWriter, _ *http.Request, params httprouter.Params) {
	serviceTypeIdOrName := params.ByName("serviceTypeIdOrName")
	request := "GET /configs/" + serviceTypeIdOrName
	hS.Logger.Info(request)

	// read service type from database
	sType, err := hS.Db.ReadServiceType(serviceTypeIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	hS.Logger.Info("Validating service type values for delete...")
	err = validate.ServiceTypeDelete(hS.Db, sType.Type)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	err = hS.Db.DeleteServiceType(serviceTypeIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusNoContent)
	response.NoContent(w)
}

// service type versions:

// ConfigsServiceTypeVersionGet processes a request to get a service type version struct by id or name from database
func (hS HttpServer) ConfigsServiceTypeVersionGet(w http.ResponseWriter, _ *http.Request, params httprouter.Params) {
	serviceTypeIdOrName := params.ByName("serviceTypeIdOrName")
	versionIdOrName := params.ByName("versionIdOrName")
	request := "GET /configs/" + serviceTypeIdOrName + "/versions/" + versionIdOrName
	hS.Logger.Info(request)

	// read service type from database to verify the existence
	_, err := hS.Db.ReadServiceType(serviceTypeIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	// read service type version from database
	version, err := hS.Db.ReadServiceTypeVersion(serviceTypeIdOrName, versionIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	response.Ok(w, version, request)
}

// ConfigsServiceTypeVersionsGetList processes a request to get a list of all service type versions in database
func (hS HttpServer) ConfigsServiceTypeVersionsGetList(w http.ResponseWriter, _ *http.Request, params httprouter.Params) {
	serviceTypeIdOrName := params.ByName("serviceTypeIdOrName")
	request := "GET /configs/" + serviceTypeIdOrName + "/versions"
	hS.Logger.Info(request)

	// read service type from database
	sType, err := hS.Db.ReadServiceType(serviceTypeIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	response.Ok(w, sType.Versions, request)
}

// ConfigsServiceTypeVersionCreate processes a request to create a service type version struct in database
func (hS HttpServer) ConfigsServiceTypeVersionCreate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	serviceTypeIdOrName := params.ByName("serviceTypeIdOrName")
	request := "POST /configs/" + serviceTypeIdOrName + "/versions"
	hS.Logger.Info(request)

	// read service type from database
	sType, err := hS.Db.ReadServiceType(serviceTypeIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	var newServiceTypeVersion protobuf.ServiceVersion
	err = json.NewDecoder(r.Body).Decode(&newServiceTypeVersion)
	if err != nil {
		err = ErrJsonIncorrect
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	// read service type version to check that there is no such object in database
	sVersion, err := hS.Db.ReadServiceTypeVersion(serviceTypeIdOrName, newServiceTypeVersion.Version)
	if sVersion != nil {
		err = ErrObjectExists("service type version", newServiceTypeVersion.Version)
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}
	if err != nil && response.ErrorClass(err) != utils.ObjectNotFound {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	hS.Logger.Info("Validating service type version for creation...")
	err = validate.ServiceTypeVersionCreate(hS.Db, sType.Versions, newServiceTypeVersion)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	// generate AnsibleVarName params
	if newServiceTypeVersion.Configs != nil {
		for i, config := range newServiceTypeVersion.Configs {
			newServiceTypeVersion.Configs[i].AnsibleVarName = serviceTypeIdOrName + "_" + config.ParameterName
		}
	}

	// generating UUID for new service version
	vUuid, err := uuid.NewRandom()
	if err != nil {
		err = ErrUuidLibError
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}
	newServiceTypeVersion.ID = vUuid.String()

	sType.Versions = append(sType.Versions, &newServiceTypeVersion)

	//saving updated service type
	err = hS.Db.UpdateServiceType(sType)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusCreated)
	response.Created(w, sType, request)
}

// ConfigsServiceTypeVersionUpdate processes a request to update a service type version struct in database
func (hS HttpServer) ConfigsServiceTypeVersionUpdate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	serviceTypeIdOrName := params.ByName("serviceTypeIdOrName")
	versionIdOrName := params.ByName("versionIdOrName")
	request := "PUT /configs/" + serviceTypeIdOrName + "/versions/" + versionIdOrName
	hS.Logger.Info(request)

	// read service type from database to verify the existence
	_, err := hS.Db.ReadServiceType(serviceTypeIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	// read service type version from database
	oldServiceTypeVersion, err := hS.Db.ReadServiceTypeVersion(serviceTypeIdOrName, versionIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	var newServiceTypeVersion protobuf.ServiceVersion
	err = json.NewDecoder(r.Body).Decode(&newServiceTypeVersion)
	if err != nil {
		err = ErrJsonIncorrect
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	hS.Logger.Info("Validating service type version values for update...")
	err = validate.ServiceTypeVersionUpdate(newServiceTypeVersion)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
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
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	response.Ok(w, oldServiceTypeVersion, request)
}

// ConfigsServiceTypeVersionDelete processes a request to delete a service type version struct from database
func (hS HttpServer) ConfigsServiceTypeVersionDelete(w http.ResponseWriter, _ *http.Request, params httprouter.Params) {
	serviceTypeIdOrName := params.ByName("serviceTypeIdOrName")
	versionIdOrName := params.ByName("versionIdOrName")
	request := "DELETE /configs/" + serviceTypeIdOrName + "/versions/" + versionIdOrName
	hS.Logger.Info(request)

	// read service type from database to verify the existence
	sType, err := hS.Db.ReadServiceType(serviceTypeIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	// read service type version from database
	serviceTypeVersion, err := hS.Db.ReadServiceTypeVersion(serviceTypeIdOrName, versionIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	hS.Logger.Info("Validating service type version values for delete...")
	err = validate.ServiceTypeVersionDelete(hS.Db, sType, serviceTypeVersion)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	err = hS.Db.DeleteServiceTypeVersion(serviceTypeIdOrName, versionIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusNoContent)
	response.NoContent(w)
}

// service type version configs:

// ConfigsServiceTypeVersionConfigGet processes a request to get a service type version config struct by parameter name from database
func (hS HttpServer) ConfigsServiceTypeVersionConfigGet(w http.ResponseWriter, _ *http.Request, params httprouter.Params) {
	serviceTypeIdOrName := params.ByName("serviceTypeIdOrName")
	versionIdOrName := params.ByName("versionIdOrName")
	parameterName := params.ByName("parameterName")
	request := "GET /configs/" + serviceTypeIdOrName + "/versions/" + versionIdOrName + "/configs/" + parameterName
	hS.Logger.Info(request)

	// read service type from database to verify the existence
	_, err := hS.Db.ReadServiceType(serviceTypeIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	// read service type version from database to verify the existence
	_, err = hS.Db.ReadServiceTypeVersion(serviceTypeIdOrName, versionIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	// read service type version config from database
	config, err := hS.Db.ReadServiceTypeVersionConfig(serviceTypeIdOrName, versionIdOrName, parameterName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	response.Ok(w, config, request)
}

// ConfigsServiceTypeVersionConfigsGetList processes a request to get a list of all service type version configs in database
func (hS HttpServer) ConfigsServiceTypeVersionConfigsGetList(w http.ResponseWriter, _ *http.Request, params httprouter.Params) {
	serviceTypeIdOrName := params.ByName("serviceTypeIdOrName")
	versionIdOrName := params.ByName("versionIdOrName")
	request := "GET /configs/" + serviceTypeIdOrName + "/versions/" + versionIdOrName + "/configs"
	hS.Logger.Info(request)

	// read service type from database to verify the existence
	_, err := hS.Db.ReadServiceType(serviceTypeIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	// read service type version from database
	version, err := hS.Db.ReadServiceTypeVersion(serviceTypeIdOrName, versionIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	response.Ok(w, version.Configs, request)
}

// ConfigsServiceTypeVersionConfigCreate processes a request to create a service type version config struct in database
func (hS HttpServer) ConfigsServiceTypeVersionConfigCreate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	serviceTypeIdOrName := params.ByName("serviceTypeIdOrName")
	versionIdOrName := params.ByName("versionIdOrName")
	request := "POST /configs/" + serviceTypeIdOrName + "/versions/" + versionIdOrName + "/configs"
	hS.Logger.Info(request)

	// read service type from database to verify the existence
	_, err := hS.Db.ReadServiceType(serviceTypeIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	// read service type version from database to verify the existence
	version, err := hS.Db.ReadServiceTypeVersion(serviceTypeIdOrName, versionIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	var newServiceTypeConfig *protobuf.ServiceConfig
	err = json.NewDecoder(r.Body).Decode(&newServiceTypeConfig)
	if err != nil {
		err = ErrJsonIncorrect
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	// check service type version config does not exist in the database
	dbConfig, err := hS.Db.ReadServiceTypeVersionConfig(serviceTypeIdOrName, versionIdOrName, newServiceTypeConfig.ParameterName)
	if dbConfig != nil {
		err = ErrObjectExists("service type version config", newServiceTypeConfig.ParameterName)
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}
	if err != nil && response.ErrorClass(err) != utils.ObjectNotFound {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	hS.Logger.Info("Validating service type version config for creation...")
	err = validate.ServiceTypeVersionConfigCreate(newServiceTypeConfig, version.Configs)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	newServiceTypeConfig.AnsibleVarName = serviceTypeIdOrName + "_" + newServiceTypeConfig.ParameterName
	version.Configs = append(version.Configs, newServiceTypeConfig)

	err = hS.Db.UpdateServiceTypeVersion(serviceTypeIdOrName, version)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusCreated)
	response.Created(w, newServiceTypeConfig, request)
}

// ConfigsServiceTypeVersionConfigUpdate processes a request to update a service type version config struct in database
func (hS HttpServer) ConfigsServiceTypeVersionConfigUpdate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	serviceTypeIdOrName := params.ByName("serviceTypeIdOrName")
	versionIdOrName := params.ByName("versionIdOrName")
	parameterName := params.ByName("parameterName")
	request := "PUT /configs/" + serviceTypeIdOrName + "/versions/" + versionIdOrName + "/configs/" + parameterName
	hS.Logger.Info(request)

	// read service type from database to verify the existence
	_, err := hS.Db.ReadServiceType(serviceTypeIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	// read service type version from database to verify the existence
	_, err = hS.Db.ReadServiceTypeVersion(serviceTypeIdOrName, versionIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	// read service type version config from database
	oldConfig, err := hS.Db.ReadServiceTypeVersionConfig(serviceTypeIdOrName, versionIdOrName, parameterName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	var newServiceTypeConfig *protobuf.ServiceConfig
	err = json.NewDecoder(r.Body).Decode(&newServiceTypeConfig)
	if err != nil {
		err = ErrJsonIncorrect
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	hS.Logger.Info("Validating service type version values for update...")
	err = validate.ServiceTypeVersionConfigUpdate(newServiceTypeConfig, oldConfig)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	if newServiceTypeConfig.Type != "" {
		oldConfig.Type = newServiceTypeConfig.Type
	}
	if newServiceTypeConfig.PossibleValues != nil {
		oldConfig.PossibleValues = newServiceTypeConfig.PossibleValues
	}
	if newServiceTypeConfig.DefaultValue != "" {
		oldConfig.DefaultValue = newServiceTypeConfig.DefaultValue
	}
	if newServiceTypeConfig.Description != "" {
		oldConfig.Description = newServiceTypeConfig.Description
	}

	err = hS.Db.UpdateServiceTypeVersionConfig(serviceTypeIdOrName, versionIdOrName, oldConfig)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	response.Ok(w, oldConfig, request)
}

// ConfigsServiceTypeVersionConfigDelete processes a request to delete a service type version config struct from database
func (hS HttpServer) ConfigsServiceTypeVersionConfigDelete(w http.ResponseWriter, _ *http.Request, params httprouter.Params) {
	serviceTypeIdOrName := params.ByName("serviceTypeIdOrName")
	versionIdOrName := params.ByName("versionIdOrName")
	parameterName := params.ByName("parameterName")
	request := "DELETE /configs/" + serviceTypeIdOrName + "/versions/" + versionIdOrName + "/configs/" + parameterName
	hS.Logger.Info(request)

	// read service type from database to verify the existence
	_, err := hS.Db.ReadServiceType(serviceTypeIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	// read service type version from database to verify the existence
	_, err = hS.Db.ReadServiceTypeVersion(serviceTypeIdOrName, versionIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	// read service type version from database
	sTypeVersionConfig, err := hS.Db.ReadServiceTypeVersionConfig(serviceTypeIdOrName, versionIdOrName, parameterName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	// TODO:
	//hS.Logger.Info("Validating service type version config values for delete...")
	//err, status := ValidateServiceTypeVersionConfigDelete(hS, sType, sTypeVersionConfig)
	//if err != nil {
	//	hS.Logger.Warn("Request ", request, " failed with status ", status, ": ", err.Error())
	//	switch status {
	//	case http.StatusBadRequest:
	//		response.BadRequest(w, err)
	//		return
	//	case http.StatusInternalServerError:
	//		response.InternalError(w, err)
	//		return
	//	}
	//}

	err = hS.Db.DeleteServiceTypeVersionConfig(serviceTypeIdOrName, versionIdOrName, sTypeVersionConfig.ParameterName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusNoContent)
	response.NoContent(w)
}

// service type version dependencies:

// ConfigsServiceTypeVersionDependencyGet processes a request to get a service type version dependency struct by dependency type from database
func (hS HttpServer) ConfigsServiceTypeVersionDependencyGet(w http.ResponseWriter, _ *http.Request, params httprouter.Params) {
	serviceTypeIdOrName := params.ByName("serviceTypeIdOrName")
	versionIdOrName := params.ByName("versionIdOrName")
	dependencyType := params.ByName("dependencyType")
	request := "GET /configs/" + serviceTypeIdOrName + "/versions/" + versionIdOrName + "/dependencies/" + dependencyType
	hS.Logger.Info(request)

	// read service type from database to verify the existence
	_, err := hS.Db.ReadServiceType(serviceTypeIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	// read service type version from database to verify the existence
	sTypeVersion, err := hS.Db.ReadServiceTypeVersion(serviceTypeIdOrName, versionIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	// read service type version config from database

	// TODO: ReadServiceTypeVersionDependency
	//sTypeVersionDependency, err := hS.Db.ReadServiceTypeVersionDependency(serviceTypeIdOrName, versionIdOrName, dependencyType)
	//if err != nil {
	//	hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
	//	response.InternalError(w, err)
	//	return
	//}
	//

	sTypeVersionDependency := new(protobuf.ServiceDependency)
	exists := false
	if sTypeVersion.Dependencies != nil {
		for _, dependency := range sTypeVersion.Dependencies {
			if dependency.ServiceType == dependencyType {
				sTypeVersionDependency = dependency
				exists = true
				break
			}
		}
	}

	if !exists {
		err = check.ErrDependencyServiceTypeNotExists(dependencyType, sTypeVersion.Version)
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	response.Ok(w, sTypeVersionDependency, request)
}

// ConfigsServiceTypeVersionDependenciesGetList processes a request to get a list of all service type version dependencies in database
func (hS HttpServer) ConfigsServiceTypeVersionDependenciesGetList(w http.ResponseWriter, _ *http.Request, params httprouter.Params) {
	serviceTypeIdOrName := params.ByName("serviceTypeIdOrName")
	versionIdOrName := params.ByName("versionIdOrName")
	request := "GET /configs/" + serviceTypeIdOrName + "/versions/" + versionIdOrName + "/dependencies"
	hS.Logger.Info(request)

	// read service type from database to verify the existence
	_, err := hS.Db.ReadServiceType(serviceTypeIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	// read service type version from database
	sTypeVersion, err := hS.Db.ReadServiceTypeVersion(serviceTypeIdOrName, versionIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	response.Ok(w, sTypeVersion.Dependencies, request)
}

// ConfigsServiceTypeVersionDependencyCreate processes a request to create a service type version dependency struct in database
func (hS HttpServer) ConfigsServiceTypeVersionDependencyCreate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	serviceTypeIdOrName := params.ByName("serviceTypeIdOrName")
	versionIdOrName := params.ByName("versionIdOrName")
	request := "POST /configs/" + serviceTypeIdOrName + "/versions/" + versionIdOrName + "/dependencies"
	hS.Logger.Info(request)

	// read service type from database to verify the existence
	_, err := hS.Db.ReadServiceType(serviceTypeIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	// read service type version from database to verify the existence
	sTypeVersion, err := hS.Db.ReadServiceTypeVersion(serviceTypeIdOrName, versionIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	var newServiceTypeDependency *protobuf.ServiceDependency
	err = json.NewDecoder(r.Body).Decode(&newServiceTypeDependency)
	if err != nil {
		err = ErrJsonIncorrect
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	// check service type version config does not exist in the database

	// TODO: ReadServiceTypeVersionDependency
	//dbDependency, err := hS.Db.ReadServiceTypeVersionDependency(serviceTypeIdOrName, versionIdOrName, newServiceTypeConfig.ParameterName)
	//if err != nil {
	//	hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
	//	response.InternalError(w, err)
	//	return
	//}

	dbDependency := new(protobuf.ServiceDependency)
	if sTypeVersion.Dependencies != nil {
		for _, dependency := range sTypeVersion.Dependencies {
			if dependency.ServiceType == newServiceTypeDependency.ServiceType {
				dbDependency = dependency
				break
			}
		}
	}

	if dbDependency.ServiceType != "" {
		err = errors.New("service type version config with this id or name already exists")
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	hS.Logger.Info("Validating service type version config for creation...")
	err = validate.ServiceTypeVersionDependencyCreate(hS.Db, newServiceTypeDependency, sTypeVersion.Dependencies)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	sTypeVersion.Dependencies = append(sTypeVersion.Dependencies, newServiceTypeDependency)

	err = hS.Db.UpdateServiceTypeVersion(serviceTypeIdOrName, sTypeVersion)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusCreated)
	response.Created(w, newServiceTypeDependency, request)
}

// ConfigsServiceTypeVersionDependencyUpdate processes a request to update a service type version dependency struct in database
func (hS HttpServer) ConfigsServiceTypeVersionDependencyUpdate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	serviceTypeIdOrName := params.ByName("serviceTypeIdOrName")
	versionIdOrName := params.ByName("versionIdOrName")
	dependencyType := params.ByName("dependencyType")
	request := "PUT /configs/" + serviceTypeIdOrName + "/versions/" + versionIdOrName + "/configs/" + dependencyType
	hS.Logger.Info(request)

	// read service type from database to verify the existence
	_, err := hS.Db.ReadServiceType(serviceTypeIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	// read service type version from database to verify the existence
	sTypeVersion, err := hS.Db.ReadServiceTypeVersion(serviceTypeIdOrName, versionIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	// read service type version dependency from database

	// TODO: ReadServiceTypeVersionDependency
	//oldSTypeVersionVersionDependency, err := hS.Db.ReadServiceTypeVersionDependency(serviceTypeIdOrName, versionIdOrName, dependencyType)
	//if err != nil {
	//	hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
	//	response.InternalError(w, err)
	//	return
	//}
	//

	oldSTypeVersionVersionDependency := new(protobuf.ServiceDependency)
	exists := false
	if sTypeVersion.Dependencies != nil {
		for _, dependency := range sTypeVersion.Dependencies {
			if dependency.ServiceType == dependencyType {
				oldSTypeVersionVersionDependency = dependency
				exists = true
				break
			}
		}
	}

	if !exists {
		err = check.ErrDependencyServiceTypeNotExists(dependencyType, sTypeVersion.Version)
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	var newServiceTypeDependency *protobuf.ServiceDependency
	err = json.NewDecoder(r.Body).Decode(&newServiceTypeDependency)
	if err != nil {
		err = ErrJsonIncorrect
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	hS.Logger.Info("Validating service type version values for update...")
	err = validate.ServiceTypeVersionDependencyUpdate(hS.Db, oldSTypeVersionVersionDependency, newServiceTypeDependency)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	if newServiceTypeDependency.ServiceVersions != nil {
		if oldSTypeVersionVersionDependency.ServiceVersions != nil {
			newSVLen := len(newServiceTypeDependency.ServiceVersions)
			for _, oldSV := range oldSTypeVersionVersionDependency.ServiceVersions {
				f := false
				for _, newSV := range newServiceTypeDependency.ServiceVersions[:newSVLen] {
					if oldSV == newSV {
						f = true
						break
					}
				}
				//add old dependency if it hasn't been updated
				if !f {
					newServiceTypeDependency.ServiceVersions = append(newServiceTypeDependency.ServiceVersions, oldSV)
				}
			}
		}
		oldSTypeVersionVersionDependency.ServiceVersions = newServiceTypeDependency.ServiceVersions
	}

	if newServiceTypeDependency.DefaultServiceVersion != "" {
		oldSTypeVersionVersionDependency.DefaultServiceVersion = newServiceTypeDependency.DefaultServiceVersion
	}

	if newServiceTypeDependency.Description != "" {
		oldSTypeVersionVersionDependency.Description = newServiceTypeDependency.Description
	}

	// TODO: UpdateServiceTypeVersionDependency

	idToUpdate := -1
	for i, curVersionDependency := range sTypeVersion.Dependencies {
		if curVersionDependency.ServiceType == oldSTypeVersionVersionDependency.ServiceType {
			idToUpdate = i
			break
		}
	}

	sTypeVersion.Dependencies[idToUpdate] = oldSTypeVersionVersionDependency

	err = hS.Db.UpdateServiceTypeVersion(serviceTypeIdOrName, sTypeVersion)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	response.Ok(w, oldSTypeVersionVersionDependency, request)
}

// ConfigsServiceTypeVersionDependencyDelete processes a request to delete a service type version dependency struct from database
func (hS HttpServer) ConfigsServiceTypeVersionDependencyDelete(w http.ResponseWriter, _ *http.Request, params httprouter.Params) {
	serviceTypeIdOrName := params.ByName("serviceTypeIdOrName")
	versionIdOrName := params.ByName("versionIdOrName")
	dependencyType := params.ByName("dependencyType")
	request := "DELETE /configs/" + serviceTypeIdOrName + "/versions/" + versionIdOrName + "/configs/" + dependencyType
	hS.Logger.Info(request)

	// read service type from database to verify the existence
	_, err := hS.Db.ReadServiceType(serviceTypeIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	// read service type version from database to verify the existence
	sTypeVersion, err := hS.Db.ReadServiceTypeVersion(serviceTypeIdOrName, versionIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	// read service type dependency from database

	// TODO: ReadServiceTypeVersionDependency
	//sTypeVersionVersionDependency, err := hS.Db.ReadServiceTypeVersionDependency(serviceTypeIdOrName, versionIdOrName, dependencyType)
	//if err != nil {
	//	hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
	//	response.InternalError(w, err)
	//	return
	//}
	//

	sTypeVersionVersionDependency := new(protobuf.ServiceDependency)
	if sTypeVersion.Dependencies != nil {
		for _, dependency := range sTypeVersion.Dependencies {
			if dependency.ServiceType == dependencyType {
				sTypeVersionVersionDependency = dependency
				break
			}
		}
	}

	if sTypeVersionVersionDependency.ServiceType == "" {
		err = errors.New("service type version config with this id or name not found")
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	// TODO:
	//hS.Logger.Info("Validating service type version config values for delete...")
	//err, status := ValidateServiceTypeVersionDependencyDelete(hS, sType, sTypeVersionConfig)
	//if err != nil {
	//	hS.Logger.Warn("Request ", request, " failed with status ", status, ": ", err.Error())
	//	switch status {
	//	case http.StatusBadRequest:
	//		response.BadRequest(w, err)
	//		return
	//	case http.StatusInternalServerError:
	//		response.InternalError(w, err)
	//		return
	//	}
	//}

	// TODO: DeleteServiceTypeVersionDependency

	idToDelete := -1
	for i, curDependency := range sTypeVersion.Dependencies {
		if curDependency.ServiceType == dependencyType {
			idToDelete = i
			break
		}
	}

	dependenciesLen := len(sTypeVersion.Dependencies)
	sTypeVersion.Dependencies[idToDelete] = sTypeVersion.Dependencies[dependenciesLen-1]
	sTypeVersion.Dependencies = sTypeVersion.Dependencies[:dependenciesLen-1]

	err = hS.Db.UpdateServiceTypeVersion(serviceTypeIdOrName, sTypeVersion)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusNoContent)
	response.NoContent(w)
}
