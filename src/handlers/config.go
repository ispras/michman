package handlers

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	protobuf "github.com/ispras/michman/src/protobuf"
	"github.com/jinzhu/copier"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

const (
	respTypeFull    = "full"
	respTypeSummary = "summary"
	respTypeKey     = "view"
)

//list of supported types
func IsValidType(t string) bool {
	switch t {
	case
		"int",
		"float",
		"bool",
		"string":
		return true
	}
	return false
}

func checkVersionUnique(stVersions []*protobuf.ServiceVersion, newV protobuf.ServiceVersion) bool {
	for _, curV := range stVersions {
		if curV.Version == newV.Version {
			return false
		}
	}
	return true
}

func checkDefaultVersion(stVersions []*protobuf.ServiceVersion, defaultV string) bool {
	for _, curV := range stVersions {
		if curV.Version == defaultV {
			return true
		}
	}
	return false
}

func (hS HttpServer) checkConfigs(vConfigs []*protobuf.ServiceConfig) (bool, error) {
	for i, curC := range vConfigs {
		//check param type
		if !IsValidType(curC.Type) {
			hS.Logger.Print("ERROR: parameter type must be int, float, bool, string, error in param" + curC.ParameterName)
			return false, errors.New("ERROR: parameter type must be one of supported: int, float, bool, string")
		}

		//check param name is unique
		curName := curC.ParameterName
		if curName == "" {
			return false, errors.New("ERROR: parameter names must be set")
		}
		for _, otherC := range vConfigs[i+1:] {
			if curName == otherC.ParameterName {
				return false, errors.New("ERROR: parameter names in service config must be uniques")
			}
		}
	}
	return true, nil
}

func (hS HttpServer) checkDependency(d *protobuf.ServiceDependency) (bool, error) {
	st, err := hS.Db.ReadServiceType(d.ServiceType)
	if err != nil {
		hS.Logger.Print(err)
		return false, err
	}

	if st.Type == "" {
		return false, errors.New("Service " + d.ServiceType + " from dependencies with this type doesn't exist")
	}

	if d.ServiceVersions == nil {
		return false, errors.New("Service versions list in dependencies can't be empty")
	}

	if d.DefaultServiceVersion == "" {
		return false, errors.New("Service default version in dependency can't be empty")
	}

	//check correctness of versions list
	flagDefaultV := false
	for _, dSv := range d.ServiceVersions {
		flag := false
		for _, sv := range st.Versions {
			if dSv == sv.Version {
				flag = true
				break
			}
		}
		if !flag {
			return false, errors.New("Service version in dependency doesn't exist")
		}
		if dSv == d.DefaultServiceVersion {
			flagDefaultV = true
		}
	}

	if !flagDefaultV {
		return false, errors.New("Service default version in dependencies doesn't exist")
	}

	return true, nil
}

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

	//check deafault version
	res := checkDefaultVersion(st.Versions[:], st.DefaultVersion)
	if !res {
		hS.Logger.Print("ERROR: default service version doesn't exists in this service type")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for i, sv := range st.Versions {
		if !checkVersionUnique(st.Versions[i+1:], *sv) {
			hS.Logger.Print("ERROR: service version exists in this service type")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		//check service version config
		if sv.Configs != nil {
			res, err := hS.checkConfigs(sv.Configs)
			if !res {
				hS.Logger.Print(err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			for j, c := range sv.Configs {
				st.Versions[i].Configs[j].AnsibleVarName = st.Type + "_" + c.ParameterName
			}
		}

		//check service version dependencies
		for _, sd := range sv.Dependencies {
			if res, err := hS.checkDependency(sd); !res {
				hS.Logger.Print(err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		// generating UUID for new service version
		vUuid, err := uuid.NewRandom()
		if err != nil {
			hS.Logger.Print(err)
			w.WriteHeader(http.StatusBadRequest)
		}
		st.Versions[i].ID = vUuid.String()
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
	err = enc.Encode(st)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
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
	w.Header().Set("Content-Type", "application/json")
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
	hS.Logger.Print("Get /configs/", sTypeName, " GET")
	queryValues := r.URL.Query()
	respType := respTypeSummary

	if t := queryValues.Get(respTypeKey); t != "" {
		if t == respTypeSummary || t == respTypeFull {
			respType = t
		} else {
			hS.Logger.Print("Error: bad view param. Supported query variables for view parameter are 'full' and 'summary', 'summary' is default.")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

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

	var respBody protobuf.ServiceType
	if respType == respTypeSummary {
		respBody.ID = st.ID
		respBody.Type = st.Type
		respBody.Description = st.Description
		respBody.DefaultVersion = st.DefaultVersion
	} else {
		respBody = *st
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	err = enc.Encode(respBody)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

//updates only information about service type
//versions and config params could be updated in ConfigsUpdateVersion
func (hS HttpServer) ConfigsUpdateService(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	sTypeName := params.ByName("serviceType")
	hS.Logger.Print("Get /configs/", sTypeName, " PUT")

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

	var newSt protobuf.ServiceType
	err = json.NewDecoder(r.Body).Decode(&newSt)
	if err != nil {
		hS.Logger.Print("ERROR:")
		hS.Logger.Print(err)
		hS.Logger.Print(r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//update service type description
	if newSt.Description != "" {
		st.Description = newSt.Description
	}

	//update service type default version
	if newSt.DefaultVersion != "" {
		st.DefaultVersion = newSt.DefaultVersion
	}

	//saving new service type
	err = hS.Db.WriteServiceType(st)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
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
	hS.Logger.Print("Get /configs/", sTypeName, " GET")

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

	//check that service type doesn't exist in dependencies
	sts, err := hS.Db.ListServicesTypes()
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	for _, st := range sts {
		for _, sv := range st.Versions {
			for _, sd := range sv.Dependencies {
				if sd.ServiceType == sTypeName {
					hS.Logger.Print("Error: this service type presents in dependencies for service " + st.Type + ".")
					w.WriteHeader(http.StatusBadRequest)
					return
				}
			}
		}
	}

	err = hS.Db.DeleteServiceType(sTypeName)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	err = enc.Encode(st)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (hS HttpServer) ConfigsCreateVersion(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	stName := params.ByName("serviceType")
	hS.Logger.Print("Get /configs/", stName, "/versions POST")

	hS.Logger.Print("Reading service types information from db...")
	st, err := hS.Db.ReadServiceType(stName)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if st.Type == "" {
		hS.Logger.Print("Service type not found")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var newStVersion protobuf.ServiceVersion
	err = json.NewDecoder(r.Body).Decode(&newStVersion)
	if err != nil {
		hS.Logger.Print("ERROR:")
		hS.Logger.Print(err)
		hS.Logger.Print(r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//check that version is unique
	if st.Versions != nil && !checkVersionUnique(st.Versions, newStVersion) {
		hS.Logger.Print("ERROR: service version exists in this service type")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//check service version config
	if newStVersion.Configs != nil {
		res, err := hS.checkConfigs(newStVersion.Configs)
		if !res {
			hS.Logger.Print(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		for i, c := range newStVersion.Configs {
			newStVersion.Configs[i].AnsibleVarName = stName + "_" + c.ParameterName
		}
	}

	//check service version dependencies
	if newStVersion.Dependencies != nil {
		for _, sd := range newStVersion.Dependencies {
			if res, err := hS.checkDependency(sd); !res {
				hS.Logger.Print(err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}
	}

	// generating UUID for new service version
	vUuid, err := uuid.NewRandom()
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
	}
	newStVersion.ID = vUuid.String()

	st.Versions = append(st.Versions, &newStVersion)

	//saving updated service type
	err = hS.Db.UpdateServiceType(st)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	err = enc.Encode(newStVersion)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (hS HttpServer) ConfigsGetVersions(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	sTypeName := params.ByName("serviceType")
	hS.Logger.Print("Get /configs/", sTypeName, "/versions GET")

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
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	err = enc.Encode(st.Versions)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (hS HttpServer) ConfigsGetVersion(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	sTypeName := params.ByName("serviceType")
	vId := params.ByName("versionId")
	hS.Logger.Print("Get /configs/", sTypeName, "/versions/", vId, " GET")

	//reading service type info from database
	hS.Logger.Print("Reading service version information from db...")
	version, err := hS.Db.ReadServiceVersion(sTypeName, vId)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if version.ID == "" {
		hS.Logger.Print("Service version not found")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	err = enc.Encode(version)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (hS HttpServer) ConfigsUpdateVersion(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	//TODO: updating of service version dependencies is not supported
	sTypeName := params.ByName("serviceType")
	vId := params.ByName("versionId")
	hS.Logger.Print("Get /configs/", sTypeName, "/versions/", vId, " PUT")

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

	st.Versions = st.Versions[:idToUpdate+copy(st.Versions[idToUpdate:], st.Versions[idToUpdate+1:])]

	var newStVersion protobuf.ServiceVersion
	err = json.NewDecoder(r.Body).Decode(&newStVersion)
	if err != nil {
		hS.Logger.Print("ERROR:")
		hS.Logger.Print(err)
		hS.Logger.Print(r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//update description
	if newStVersion.Description != "" {
		oldV.Description = newStVersion.Description
	}

	//update download url
	if newStVersion.DownloadURL != "" {
		oldV.DownloadURL = newStVersion.DownloadURL
	}

	//update version configs
	if newStVersion.Configs != nil {
		//check service version config
		res, err := hS.checkConfigs(newStVersion.Configs)
		if !res {
			hS.Logger.Print(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		for i, c := range newStVersion.Configs {
			newStVersion.Configs[i].AnsibleVarName = sTypeName + "_" + c.ParameterName
		}
		//oldV.Configs = nil
		oldV.Configs = make([]*protobuf.ServiceConfig, len(newStVersion.Configs))
		copy(oldV.Configs, newStVersion.Configs)
	}

	st.Versions = append(st.Versions, &oldV)

	//saving updated service type
	err = hS.Db.UpdateServiceType(st)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	err = enc.Encode(oldV)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (hS HttpServer) ConfigsDeleteVersion(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	sTypeName := params.ByName("serviceType")
	vId := params.ByName("versionId")
	hS.Logger.Print("Get /configs/", sTypeName, "/versions/", vId, " DELETE")

	var result *protobuf.ServiceVersion
	result, err := hS.Db.ReadServiceVersion(sTypeName, vId)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if result.ID == "" {
		hS.Logger.Print("Service version not found")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	//check that this service version doesn't present in dependencies
	sts, err := hS.Db.ListServicesTypes()
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	for _, st := range sts {
		for _, sv := range st.Versions {
			for _, sd := range sv.Dependencies {
				for _, sdv := range sd.ServiceVersions {
					if sd.ServiceType == sTypeName && sdv == result.Version {
						hS.Logger.Print("Error: this service version presents in dependencies for service " + st.Type + ".")
						w.WriteHeader(http.StatusBadRequest)
						return
					}
				}
			}
		}
	}

	_, err = hS.Db.DeleteServiceVersion(sTypeName, vId)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	err = enc.Encode(result)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (hS HttpServer) ConfigsCreateConfigParam(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	sTypeName := params.ByName("serviceType")
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
		res, err := hS.checkConfigs(tmpC)
		if !res {
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
