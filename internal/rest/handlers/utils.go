package handlers

import (
	"encoding/json"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/utils"
	"net/http"
	"os"
	"regexp"
	"strconv"
)

type serviceExists struct {
	exists  bool
	service *protobuf.Service
}

const (
	QueryViewTypeFull    = "full"
	QueryViewTypeSummary = "summary"
	QueryViewKey         = "view"
)

// CheckType list of supported types
func CheckType(_type string) error {
	switch _type {
	case
		"int",
		"float",
		"bool",
		"string":
		return nil
	}
	return ErrValidTypeParam(_type)
}

func IsFlavorUsed(hS HttpServer, flavorName string) (bool, error) {
	hS.Logger.Print("Checking is flavor used...")
	clusters, err := hS.Db.ReadClustersList()
	if err != nil {
		return false, err
	}
	for _, c := range clusters {
		if c.MasterFlavor == flavorName || c.StorageFlavor == flavorName ||
			c.SlavesFlavor == flavorName || c.MonitoringFlavor == flavorName {
			return true, nil
		}
	}
	projects, err := hS.Db.ReadProjectsList()
	if err != nil {
		return false, err
	}
	for _, p := range projects {
		if p.DefaultMasterFlavor == flavorName || p.DefaultStorageFlavor == flavorName ||
			p.DefaultSlavesFlavor == flavorName || p.DefaultMonitoringFlavor == flavorName {
			return true, nil
		}
	}
	return false, nil
}

func ValidateFlavor(hS HttpServer, flavor *protobuf.Flavor) error {
	hS.Logger.Info("Validating flavor...")
	if flavor.ID != "" {
		return ErrFlavorIdNotEmpty
	}
	if flavor.Name == "" {
		return ErrFlavorEmptyName
	}

	switch interface{}(flavor.VCPUs).(type) {
	case int32:
		if flavor.VCPUs <= 0 {
			return ErrFlavorParamVal("VCPUs")
		}
	default:
		return ErrFlavorParamType("VCPUs")
	}

	switch interface{}(flavor.RAM).(type) {
	case int32:
		if flavor.RAM <= 0 {
			return ErrFlavorParamVal("RAM")
		}
	default:
		return ErrFlavorParamType("RAM")
	}

	switch interface{}(flavor.Disk).(type) {
	case int32:
		if flavor.Disk <= 0 {
			return ErrFlavorParamVal("Disk")
		}
	default:
		return ErrFlavorParamType("RAM")
	}
	return nil
}

func ValidateCluster(hS HttpServer, cluster *protobuf.Cluster) (error, int) {
	hS.Logger.Info("Validating cluster...")
	if err, status := CheckValidName(cluster.DisplayName, utils.ClusterNamePattern, ErrClusterBadName); err != nil {
		return err, status
	}

	for _, service := range cluster.Services {
		if err, status := ValidateService(hS, service); err != nil {
			return err, status
		}
	}

	if cluster.NHosts < 0 {
		return ErrClusterNhostsZero, http.StatusBadRequest
	}

	if cluster.NHosts == 0 {
		res, err := CheckMSServices(hS, cluster)
		if err != nil {
			return err, http.StatusInternalServerError
		}
		if res {
			return ErrClustersNhostsMasterSlave, http.StatusBadRequest
		}
	}

	dbImg, err := hS.Db.ReadImage(cluster.Image)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if dbImg.Name == "" {
		return ErrClusterImageNotFound, http.StatusBadRequest
	}
	dbFlavor, err := hS.Db.ReadFlavor(cluster.MasterFlavor)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if dbFlavor.ID == "" {
		return ErrFlavorFieldValueNotFound("MasterFlavor"), http.StatusBadRequest
	}
	dbFlavor, err = hS.Db.ReadFlavor(cluster.SlavesFlavor)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if dbFlavor.ID == "" {
		return ErrFlavorFieldValueNotFound("SlavesFlavor"), http.StatusBadRequest
	}
	dbFlavor, err = hS.Db.ReadFlavor(cluster.StorageFlavor)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if dbFlavor.ID == "" {
		return ErrFlavorFieldValueNotFound("StorageFlavor"), http.StatusBadRequest
	}
	dbFlavor, err = hS.Db.ReadFlavor(cluster.MonitoringFlavor)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if dbFlavor.ID == "" {
		return ErrFlavorFieldValueNotFound("MonitoringFlavor"), http.StatusBadRequest
	}
	return nil, 0
}

func AddDependencies(hS HttpServer, cluster *protobuf.Cluster, curS *protobuf.Service) ([]*protobuf.Service, error, int) {
	var serviceToAdd *protobuf.Service = nil
	var servicesList []*protobuf.Service = nil

	sv, err := hS.Db.ReadServiceTypeVersion(curS.Type, curS.Version)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}

	//check if version has dependencies
	if sv.Dependencies != nil {
		for _, sd := range sv.Dependencies {
			//check if the service from dependencies has already listed in cluster and version is ok
			for _, clusterS := range cluster.Services {
				if clusterS.Type == sd.ServiceType {
					if !utils.ItemExists(sd.ServiceVersions, clusterS.Version) {
						return nil, ErrClusterDependenceServicesIncompatibleVersion(clusterS.Type, curS.Type), http.StatusBadRequest
					}
				}
			}

			//add service from dependencies with default configurations
			serviceToAdd = &protobuf.Service{
				Name:    curS.Name + "-dependent", //TODO: use better service name?
				Type:    sd.ServiceType,
				Version: sd.DefaultServiceVersion,
			}
			servicesList = append(servicesList, serviceToAdd)
		}
	}

	return servicesList, nil, http.StatusOK
}

// CheckMSServices returns true if master-slave service exists
func CheckMSServices(hS HttpServer, cluster *protobuf.Cluster) (bool, error) {
	for _, service := range cluster.Services {
		st, err := hS.Db.ReadServiceType(service.Type)
		if err != nil {
			return false, err
		}
		if st.Class == utils.ClassMasterSlave {
			return true, nil
		}
	}
	return false, nil
}

func ValidateImage(hs HttpServer, image *protobuf.Image) error {
	hs.Logger.Info("Validating image...")
	if image.ID != "" {
		return ErrImageIdNotEmpty
	}
	if image.Name == "" {
		return ErrImageValidationParam("Name")
	}
	if image.AnsibleUser == "" {
		return ErrImageValidationParam("AnsibleUser")
	}
	if image.CloudImageID == "" {
		return ErrImageValidationParam("ImageID")
	}
	return nil
}

func IsImageUsed(hs HttpServer, name string) (bool, error) {
	hs.Logger.Info("Checking is image used...")
	clusters, err := hs.Db.ReadClustersList()
	if err != nil {
		return false, err
	}
	for _, c := range clusters {
		if c.Image == name {
			return true, nil
		}
	}
	projects, err := hs.Db.ReadProjectsList()
	if err != nil {
		return false, err
	}
	for _, p := range projects {
		if p.DefaultImage == name {
			return true, nil
		}
	}
	return false, nil
}

func DeleteSpaces(valStr string) string {
	resStr := ""
	for _, ch := range valStr {
		if ch != ' ' {
			resStr += string(ch)
		}
	}
	return resStr
}

func CheckValuesAllowed(val string, posVal []string) bool {
	val = DeleteSpaces(val)
	for _, pv := range posVal {
		if val == pv {
			return true
		}
	}
	return false
}

func ValidateService(hS HttpServer, service *protobuf.Service) (error, int) {
	hS.Logger.Info("Validating service type and config params...")

	if service.Type == "" {
		return ErrClusterServiceTypeEmpty, http.StatusBadRequest
	}

	sTypes, err := hS.Db.ReadServicesTypesList()
	if err != nil {
		return err, http.StatusInternalServerError
	}

	stIdx, err := GetServiceTypeIdx(service, sTypes)
	if err != nil {
		return err, http.StatusBadRequest
	}

	//check service version
	if service.Version == "" && sTypes[stIdx].DefaultVersion != "" {
		service.Version = sTypes[stIdx].DefaultVersion
	} else if service.Version == "" && sTypes[stIdx].DefaultVersion == "" {
		return ErrClusterServiceVersionsEmpty(service.Type), http.StatusBadRequest
	}

	svIdx, err := GetServiceVersionIdx(service, sTypes, stIdx)
	if err != nil {
		return err, http.StatusBadRequest
	}

	if err, status := ValidateConfigs(service, sTypes[stIdx].Versions[svIdx].Configs); err != nil {
		return err, status
	}

	return nil, 0
}

func ValidateProjectCreate(hs HttpServer, project *protobuf.Project) (error, int) {
	hs.Logger.Info("Validating project...")
	if project.DisplayName == "" {
		return ErrProjectFieldEmpty("DisplayName"), http.StatusBadRequest
	}
	if project.ID != "" {
		return ErrProjectFieldIsGenerated("ID"), http.StatusBadRequest
	}
	if project.Name != "" {
		return ErrProjectFieldIsGenerated("Name"), http.StatusBadRequest
	}
	if project.GroupID != "" {
		return ErrProjectFieldIsGenerated("GroupID"), http.StatusBadRequest
	}
	if err, status := CheckValidName(project.DisplayName, utils.ProjectNamePattern, ErrProjectValidation); err != nil {
		return err, status
	}
	if project.DefaultImage == "" {
		return ErrProjectFieldEmpty("DefaultImage"), http.StatusBadRequest
	}
	if project.DefaultMasterFlavor == "" {
		return ErrProjectFieldEmpty("DefaultMasterFlavor"), http.StatusBadRequest
	}
	if project.DefaultSlavesFlavor == "" {
		return ErrProjectFieldEmpty("DefaultSlavesFlavor"), http.StatusBadRequest
	}
	if project.DefaultStorageFlavor == "" {
		return ErrProjectFieldEmpty("DefaultStorageFlavor"), http.StatusBadRequest
	}
	if project.DefaultMonitoringFlavor == "" {
		return ErrProjectFieldEmpty("DefaultMonitoringFlavor"), http.StatusBadRequest
	}

	dbRes, err := hs.Db.ReadProject(project.Name)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if dbRes.Name != "" {
		return ErrProjectExisted, http.StatusBadRequest
	}

	err, status := ValidateProjectFieldsDb(hs, project)
	if err != nil {
		return err, status
	}

	return nil, 0
}

func ValidateProjectUpdate(hs HttpServer, project *protobuf.Project) (error, int) {
	hs.Logger.Info("Validating updated values of the project fields...")
	if project.ID != "" || project.Name != "" || project.GroupID != "" {
		return ErrProjectUnmodFields, http.StatusBadRequest
	}
	err, status := ValidateProjectFieldsDb(hs, project)
	if err != nil {
		return err, status
	}

	return nil, 0
}

func ValidateProjectFieldsDb(hs HttpServer, project *protobuf.Project) (error, int) {
	if project.DefaultImage != "" {
		dbImg, err := hs.Db.ReadImage(project.DefaultImage)
		if err != nil {
			return err, http.StatusInternalServerError
		}
		if dbImg.Name == "" {
			return ErrProjectImageNotFound, http.StatusBadRequest
		}
	}
	if project.DefaultMasterFlavor != "" {
		flavor, err := hs.Db.ReadFlavor(project.DefaultMasterFlavor)
		if err != nil {
			return err, http.StatusInternalServerError
		}
		if flavor.Name == "" {
			return ErrFlavorFieldValueNotFound("DefaultMasterFlavor"), http.StatusBadRequest
		}
	}
	if project.DefaultSlavesFlavor != "" {
		flavor, err := hs.Db.ReadFlavor(project.DefaultSlavesFlavor)
		if err != nil {
			return err, http.StatusInternalServerError
		}
		if flavor.Name == "" {
			return ErrFlavorFieldValueNotFound("DefaultSlavesFlavor"), http.StatusBadRequest
		}
	}
	if project.DefaultStorageFlavor != "" {
		flavor, err := hs.Db.ReadFlavor(project.DefaultStorageFlavor)
		if err != nil {
			return err, http.StatusInternalServerError
		}
		if flavor.Name == "" {
			return ErrFlavorFieldValueNotFound("DefaultStorageFlavor"), http.StatusBadRequest
		}
	}
	if project.DefaultMonitoringFlavor != "" {
		flavor, err := hs.Db.ReadFlavor(project.DefaultMonitoringFlavor)
		if err != nil {
			return err, http.StatusInternalServerError
		}
		if flavor.Name == "" {
			return ErrFlavorFieldValueNotFound("DefaultMonitoringFlavor"), http.StatusBadRequest
		}
	}
	return nil, 0
}

func CheckFileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	return false, err
}

func MakeLogFilePath(filename string, LogsFilePath string) string {
	if LogsFilePath[0] == '/' {
		return LogsFilePath + "/" + filename
	}
	return "./" + LogsFilePath + "/" + filename
}

func CheckValidName(name string, pattern string, errorType error) (error, int) {
	validName := regexp.MustCompile(pattern).MatchString
	if !validName(name) {
		return errorType, http.StatusBadRequest
	}
	return nil, 0
}

func CheckClass(st *protobuf.ServiceType) error {
	if st.Class == utils.ClassMasterSlave || st.Class == utils.ClassStandAlone || st.Class == utils.ClassStorage {
		return nil
	}
	return ErrServiceTypeClass
}

func CheckPort(port int32) error {
	//TODO: add another checks for port?
	if port > 0 {
		return nil
	}
	return ErrServiceTypePort
}

func CheckDefaultVersion(sTypeVersions []*protobuf.ServiceVersion, defaultVersion string) error {
	for _, curVersion := range sTypeVersions {
		if curVersion.Version == defaultVersion {
			return nil
		}
	}
	return ErrServiceTypeDefaultVersion
}

func CheckVersionUnique(sTypeVersions []*protobuf.ServiceVersion, newVersion protobuf.ServiceVersion) error {
	for _, curVersion := range sTypeVersions {
		if curVersion.Version == newVersion.Version {
			return ErrServiceTypeVersionUnique(curVersion.Version)
		}
	}
	return nil
}

func CheckConfigsUnique(sTypeVersionConfigs []*protobuf.ServiceConfig, newConfig protobuf.ServiceConfig) error {
	for _, curConfig := range sTypeVersionConfigs {
		if curConfig.ParameterName == newConfig.ParameterName {
			return ErrServiceTypeVersionConfigUnique(curConfig.ParameterName)
		}
	}
	return nil
}

func CheckPossibleValuesUnique(possibleValues []string) error {
	for i, curVal := range possibleValues[:len(possibleValues)-1] {
		if curVal == "" {
			return ErrConfigPossibleValueEmpty
		}
		for _, otherVal := range possibleValues[i+1:] {
			if curVal == otherVal {
				return ErrServiceTypeVersionConfigPossibleValuesUnique(curVal)
			}
		}
	}
	return nil
}

func ValidateServiceTypeCreate(hS HttpServer, sType *protobuf.ServiceType) (error, int) {
	// check service class
	err := CheckClass(sType)
	if err != nil {
		return err, http.StatusBadRequest
	}

	// check service access port
	if sType.AccessPort != 0 { //0 if port not provided
		err = CheckPort(sType.AccessPort)
		if err != nil {
			return err, http.StatusBadRequest
		}
	}

	// check all ports
	if sType.Ports != nil {
		for _, p := range sType.Ports {
			err = CheckPort(p.Port)
			if err != nil {
				return err, http.StatusBadRequest
			}
		}
	}

	err, status := ValidateServiceTypeVersions(hS, sType.Versions)
	if err != nil {
		return err, status
	}

	//check default version
	err = CheckDefaultVersion(sType.Versions, sType.DefaultVersion)
	if err != nil {
		return err, http.StatusBadRequest
	}

	return nil, 0
}

func ValidateServiceTypeVersions(hS HttpServer, sTypeVersions []*protobuf.ServiceVersion) (error, int) {
	for i, serviceVersion := range sTypeVersions {
		// check service version is unique
		err := CheckVersionUnique(sTypeVersions[i+1:], *serviceVersion)
		if err != nil {
			return err, http.StatusBadRequest
		}

		//check service version config
		if serviceVersion.Configs != nil {
			err, status := CheckConfigs(serviceVersion.Configs)
			if err != nil {
				return err, status
			}
		}

		//check service version dependencies
		err, status := CheckDependencies(hS, serviceVersion.Dependencies)
		if err != nil {
			return err, status
		}
	}
	return nil, 0
}

func CheckPossibleValues(possibleValues []string, vType string, IsList bool) error {
	//check PossibleValues type
	if !IsList {
		switch vType {
		case "int":
			for _, value := range possibleValues {
				if _, err := strconv.ParseInt(value, 10, 32); err != nil {
					return ErrServiceTypeVersionConfigPossibleValues(value)
				}
			}
		case "float":
			for _, value := range possibleValues {
				if _, err := strconv.ParseFloat(value, 64); err != nil {
					return ErrServiceTypeVersionConfigPossibleValues(value)
				}
			}
		case "bool":
			for _, value := range possibleValues {
				if _, err := strconv.ParseBool(value); err != nil {
					return ErrServiceTypeVersionConfigPossibleValues(value)
				}
			}
		}
	} else {
		switch vType {
		case "int":
			var valList []int64
			for _, value := range possibleValues {
				if err := json.Unmarshal([]byte(value), &valList); err != nil {
					return ErrServiceTypeVersionConfigPossibleValues(value)
				}
			}
		case "float":
			var valList []float64
			for _, value := range possibleValues {
				if err := json.Unmarshal([]byte(value), &valList); err != nil {
					return ErrServiceTypeVersionConfigPossibleValues(value)
				}
			}
		case "bool":
			var valList []bool
			for _, value := range possibleValues {
				if err := json.Unmarshal([]byte(value), &valList); err != nil {
					return ErrServiceTypeVersionConfigPossibleValues(value)
				}
			}
		case "string":
			var valList []string
			for _, value := range possibleValues {
				if err := json.Unmarshal([]byte(value), &valList); err != nil {
					return ErrServiceTypeVersionConfigPossibleValues(value)
				}
			}
		}

		//format PossibleValue strings
		for i, pV := range possibleValues {
			possibleValues[i] = DeleteSpaces(pV)
		}
	}

	//check PossibleValues are unique
	err := CheckPossibleValuesUnique(possibleValues)
	if err != nil {
		return err
	}
	return nil
}

func CheckConfigs(versionConfigs []*protobuf.ServiceConfig) (error, int) {
	for i, curConfig := range versionConfigs {
		// check param type
		err := CheckType(curConfig.Type)
		if err != nil {
			return err, http.StatusBadRequest
		}

		// check param name is unique
		if curConfig.ParameterName == "" {
			return ErrServiceTypeVersionConfigParamEmpty("ParameterName"), http.StatusBadRequest
		}

		// check config is unique by parameter name
		err = CheckConfigsUnique(versionConfigs[i+1:], *curConfig)
		if err != nil {
			return err, http.StatusBadRequest
		}

		//check param possible values
		if curConfig.PossibleValues != nil {
			err = CheckPossibleValues(curConfig.PossibleValues, curConfig.Type, curConfig.IsList)
			if err != nil {
				return err, http.StatusBadRequest
			}
		}
	}
	return nil, 0
}

func CheckDependencies(hS HttpServer, serviceDependencies []*protobuf.ServiceDependency) (error, int) {
	for _, serviceDependency := range serviceDependencies {
		sType, err := hS.Db.ReadServiceType(serviceDependency.ServiceType)
		if err != nil {
			return err, http.StatusInternalServerError
		}

		if sType.Type == "" {
			return ErrServiceDependenciesNotExists(serviceDependency.ServiceType), http.StatusBadRequest
		}

		if serviceDependency.ServiceVersions == nil {
			return ErrConfigDependencyServiceVersionEmpty, http.StatusBadRequest
		}

		if serviceDependency.DefaultServiceVersion == "" {
			return ErrConfigDependencyServiceDefaultVersionEmpty, http.StatusBadRequest
		}

		//check correctness of versions list
		flagDefaultVersion := false
		for _, dependencyServiceVersion := range serviceDependency.ServiceVersions {
			flag := false

			for _, sv := range sType.Versions {
				if dependencyServiceVersion == sv.Version {
					flag = true
					break
				}
			}

			if !flag {
				return ErrConfigServiceDependencyVersionNotFound, http.StatusBadRequest
			}
			if dependencyServiceVersion == serviceDependency.DefaultServiceVersion {
				flagDefaultVersion = true
			}
		}

		if !flagDefaultVersion {
			return ErrConfigServiceDependencyDefaultVersionNotFound, http.StatusBadRequest
		}
	}
	return nil, 0
}

func ValidateServiceTypeUpdate(oldServiceType *protobuf.ServiceType, newServiceType *protobuf.ServiceType) (error, int) {
	if newServiceType.ID != "" || newServiceType.Type != "" {
		return ErrServiceTypeUnmodFields, http.StatusBadRequest
	}
	if newServiceType.Versions != nil {
		return ErrServiceTypeUnmodVersionsField, http.StatusBadRequest
	}

	if newServiceType.DefaultVersion != "" {
		err := CheckDefaultVersion(oldServiceType.Versions, newServiceType.DefaultVersion)
		if err != nil {
			return err, http.StatusBadRequest
		}
	}

	if newServiceType.Class != "" {
		err := CheckClass(newServiceType)
		if err != nil {
			return err, http.StatusBadRequest
		}
	}

	if newServiceType.AccessPort != 0 { //0 if port not provided
		err := CheckPort(newServiceType.AccessPort)
		if err != nil {
			return err, http.StatusBadRequest
		}
	}

	if newServiceType.Ports != nil {
		for _, port := range newServiceType.Ports {
			err := CheckPort(port.Port)
			if err != nil {
				return err, http.StatusBadRequest
			}
		}
	}
	return nil, 0
}

func ValidateServiceTypeVersionCreate(hS HttpServer, versions []*protobuf.ServiceVersion, newServiceTypeVersion protobuf.ServiceVersion) (error, int) {
	if newServiceTypeVersion.ID != "" {
		return ErrServiceTypeVersionUnmodFields, http.StatusBadRequest
	}

	if newServiceTypeVersion.Version == "" {
		return ErrServiceTypeVersionEmptyVersionField, http.StatusBadRequest
	}

	//check that version is unique
	if versions != nil {
		err := CheckVersionUnique(versions, newServiceTypeVersion)
		if err != nil {
			return err, http.StatusBadRequest
		}
	}

	//check service version config
	if newServiceTypeVersion.Configs != nil {
		err, status := CheckConfigs(newServiceTypeVersion.Configs)
		if err != nil {
			return err, status
		}
	}

	//check service version dependencies
	if newServiceTypeVersion.Dependencies != nil {
		err, status := CheckDependencies(hS, newServiceTypeVersion.Dependencies)
		if err != nil {
			return err, status
		}
	}
	return nil, 0
}

func ValidateServiceTypeVersionUpdate(newServiceTypeVersion protobuf.ServiceVersion) (error, int) {
	if newServiceTypeVersion.ID != "" || newServiceTypeVersion.Version != "" {
		return ErrServiceTypeVersionUnmodFields, http.StatusBadRequest
	}

	if newServiceTypeVersion.Configs != nil || newServiceTypeVersion.Dependencies != nil {
		return ErrServiceTypeUnmodVersionFields, http.StatusBadRequest
	}
	return nil, 0
}

func ValidateServiceTypeVersionDelete(hS HttpServer, serviceType *protobuf.ServiceType, serviceTypeVersion *protobuf.ServiceVersion) (error, int) {
	//check that this service version doesn't present in dependencies
	serviceTypes, err := hS.Db.ReadServicesTypesList()
	if err != nil {
		return err, http.StatusInternalServerError
	}
	for _, curServiceType := range serviceTypes {
		for _, serviceVersion := range curServiceType.Versions {
			for _, serviceVersionDependency := range serviceVersion.Dependencies {
				if serviceVersionDependency.ServiceType == serviceType.Type {
					for _, serviceVersionDependencyVersion := range serviceVersionDependency.ServiceVersions {
						if serviceVersionDependencyVersion == serviceTypeVersion.Version {
							return ErrConfigServiceTypeDependenceVersionExists(serviceVersionDependencyVersion, curServiceType.Type), http.StatusBadRequest
						}
					}
				}
			}
		}
	}
	if serviceType.DefaultVersion == serviceTypeVersion.Version {
		return ErrServiceTypeDeleteVersionDefault, http.StatusBadRequest
	}
	return nil, 0
}

func ValidateServiceTypeDelete(hS HttpServer, serviceType string) (error, int) {
	//check that service type doesn't exist in dependencies
	serviceTypes, err := hS.Db.ReadServicesTypesList()
	if err != nil {
		return err, http.StatusInternalServerError
	}
	for _, curServiceType := range serviceTypes {
		for _, serviceVersion := range curServiceType.Versions {
			for _, serviceVersionDependency := range serviceVersion.Dependencies {
				if serviceVersionDependency.ServiceType == serviceType {
					return ErrConfigServiceTypeDependenceExists, http.StatusBadRequest
				}
			}
		}
	}
	return nil, 0
}

func GetServiceTypeIdx(service *protobuf.Service, ServiceTypes []protobuf.ServiceType) (int, error) {
	for i, serviceType := range ServiceTypes {
		if serviceType.Type == service.Type {
			return i, nil
		}
	}
	return 0, ErrClusterServiceTypeNotSupported(service.Type)
}

func GetServiceVersionIdx(service *protobuf.Service, ServiceTypes []protobuf.ServiceType, stIdx int) (int, error) {
	for i, sv := range ServiceTypes[stIdx].Versions {
		if sv.Version == service.Version {
			return i, nil
		}
	}
	return 0, ErrClusterServiceVersionNotSupported(service.Version, service.Type)
}

func ValidateConfigs(service *protobuf.Service, Configs []*protobuf.ServiceConfig) (error, int) {
	for key, value := range service.Config {
		flagPN := false
		for _, sc := range Configs {
			if key == sc.ParameterName {
				flagPN = true

				//check type
				if !sc.IsList {
					switch sc.Type {
					case "int":
						if _, err := strconv.ParseInt(value, 10, 32); err != nil {
							return ErrClusterServiceConfigIncorrectType(key, service.Type), http.StatusBadRequest
						}
					case "float":
						if _, err := strconv.ParseFloat(value, 64); err != nil {
							return ErrClusterServiceConfigIncorrectType(key, service.Type), http.StatusBadRequest
						}
					case "bool":
						if _, err := strconv.ParseBool(value); err != nil {
							return ErrClusterServiceConfigIncorrectType(key, service.Type), http.StatusBadRequest
						}
					}
				} else {
					switch sc.Type {
					case "int":
						var valList []int64
						if err := json.Unmarshal([]byte(value), &valList); err != nil {
							return ErrClusterServiceConfigIncorrectType(key, service.Type), http.StatusBadRequest
						}
					case "float":
						var valList []float64
						if err := json.Unmarshal([]byte(value), &valList); err != nil {
							return ErrClusterServiceConfigIncorrectType(key, service.Type), http.StatusBadRequest
						}
					case "bool":
						var valList []bool
						if err := json.Unmarshal([]byte(value), &valList); err != nil {
							return ErrClusterServiceConfigIncorrectType(key, service.Type), http.StatusBadRequest
						}
					case "string":
						var valList []string
						if err := json.Unmarshal([]byte(value), &valList); err != nil {
							return ErrClusterServiceConfigIncorrectType(key, service.Type), http.StatusBadRequest
						}
					}
				}

				//check for possible values
				if sc.PossibleValues != nil {
					if !CheckValuesAllowed(value, sc.PossibleValues) {
						return ErrClusterServiceConfigNotPossibleValue(key, service.Type), http.StatusBadRequest
					}
				}

				break
			}
		}
		if !flagPN {
			return ErrClusterServiceConfigNotSupported(key, service.Type), http.StatusBadRequest
		}
	}

	return nil, 0
}
