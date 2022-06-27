package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/utils"
	"regexp"
	"strconv"
)

type serviceExists struct {
	exists  bool
	service *protobuf.Service
}

const (
	respTypeFull    = "full"
	respTypeSummary = "summary"
	respTypeKey     = "view"
)

// IsValidType list of supported types
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

func FlavorGetByIdOrName(hS HttpServer, idOrName string) (*protobuf.Flavor, error) {
	isUuid := utils.IsUuid(idOrName)
	var flavor *protobuf.Flavor
	var err error
	if isUuid {
		flavor, err = hS.Db.ReadFlavorById(idOrName)
	} else {
		flavor, err = hS.Db.ReadFlavorByName(idOrName)
	}

	return flavor, err
}

func CheckVersionUnique(stVersions []*protobuf.ServiceVersion, newV protobuf.ServiceVersion) bool {
	for _, curV := range stVersions {
		if curV.Version == newV.Version {
			return false
		}
	}
	return true
}

func CheckDefaultVersion(stVersions []*protobuf.ServiceVersion, defaultV string) bool {
	for _, curV := range stVersions {
		if curV.Version == defaultV {
			return true
		}
	}
	return false
}

func CheckPossibleValues(vPossibleValues []string, vType string, IsList bool) bool {
	//check PossibleValues type
	if !IsList {
		switch vType {
		case "int":
			for _, pV := range vPossibleValues {
				if _, err := strconv.ParseInt(pV, 10, 32); err != nil {
					return false
				}
			}
		case "float":
			for _, pV := range vPossibleValues {
				if _, err := strconv.ParseFloat(pV, 64); err != nil {
					return false
				}
			}
		case "bool":
			for _, pV := range vPossibleValues {
				if _, err := strconv.ParseBool(pV); err != nil {
					return false
				}
			}
		}
	} else {
		switch vType {
		case "int":
			var valList []int64
			for _, pV := range vPossibleValues {
				if err := json.Unmarshal([]byte(pV), &valList); err != nil {
					return false
				}
			}
		case "float":
			var valList []float64
			for _, pV := range vPossibleValues {
				if err := json.Unmarshal([]byte(pV), &valList); err != nil {
					return false
				}
			}
		case "bool":
			var valList []bool
			for _, pV := range vPossibleValues {
				if err := json.Unmarshal([]byte(pV), &valList); err != nil {
					return false
				}
			}
		case "string":
			var valList []string
			for _, pV := range vPossibleValues {
				if err := json.Unmarshal([]byte(pV), &valList); err != nil {
					return false
				}
			}
		}

		//format PossibleValue strings
		for i, pV := range vPossibleValues {
			vPossibleValues[i] = DeleteSpaces(pV)
		}
	}

	//check PossibleValues are unique
	for i, curVal := range vPossibleValues[:len(vPossibleValues)-1] {
		if curVal == "" {
			return false
		}
		for _, otherVal := range vPossibleValues[i+1:] {
			if curVal == otherVal {
				return false
			}
		}
	}

	return true
}

func CheckConfigs(hS HttpServer, vConfigs []*protobuf.ServiceConfig) (bool, error) {
	for i, curC := range vConfigs {
		//check param type
		if !IsValidType(curC.Type) {
			hS.Logger.Print("ERROR: parameter type must be int, float, bool, string, error in param " + curC.ParameterName)
			return false, errors.New("ERROR: parameter type must be one of supported: int, float, bool, string")
		}

		//check param name is unique
		curName := curC.ParameterName
		if curName == "" {
			hS.Logger.Print("ERROR: parameter names must be set")
			return false, errors.New("ERROR: parameter names must be set")
		}
		for _, otherC := range vConfigs[i+1:] {
			if curName == otherC.ParameterName {
				hS.Logger.Print("ERROR: parameter names in service config must be uniques")
				return false, errors.New("ERROR: parameter names in service config must be uniques")
			}
		}

		//check param possible values
		if curC.PossibleValues != nil {
			if flag := CheckPossibleValues(curC.PossibleValues, curC.Type, curC.IsList); flag != true {
				hS.Logger.Print("ERROR: possible values are set incorrectly, check the value type or spelling")
				return false, errors.New("ERROR: possible values are set incorrectly, check the value type or spelling")
			}
		}
	}
	return true, nil
}

func CheckDependency(hS HttpServer, d *protobuf.ServiceDependency) (bool, error) {
	st, err := hS.Db.ReadServiceType(d.ServiceType)
	if err != nil {
		hS.Logger.Print(err)
		return false, err
	}

	if st.Type == "" {
		hS.Logger.Print("Service " + d.ServiceType + " from dependencies with this type doesn't exist")
		return false, errors.New("Service " + d.ServiceType + " from dependencies with this type doesn't exist")
	}

	if d.ServiceVersions == nil {
		hS.Logger.Print("Service versions list in dependencies can't be empty")
		return false, errors.New("Service versions list in dependencies can't be empty")
	}

	if d.DefaultServiceVersion == "" {
		hS.Logger.Print("Service default version in dependency can't be empty")
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
			hS.Logger.Print("Service version in dependency doesn't exist")
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

func CheckClass(st *protobuf.ServiceType) bool {
	if st.Class == utils.ClassMasterSlave || st.Class == utils.ClassStandAlone || st.Class == utils.ClassStorage {
		return true
	}
	return false
}

func CheckPort(port int32) bool {
	//TODO: add another checks for port?
	if port > 0 {
		return true
	}
	return false
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

func ValidateCluster(hS HttpServer, cluster *protobuf.Cluster) (bool, error) {
	validName := regexp.MustCompile(`^[A-Za-z][A-Za-z0-9-]+$`).MatchString

	if !validName(cluster.DisplayName) {
		return false, errors.New("bad name for cluster. You should use only alpha-numeric characters and '-' symbols and only alphabetic characters for leading symbol")
	}

	for _, service := range cluster.Services {
		if res, err := ValidateService(hS, service); !res {
			return false, err
		}
	}

	if cluster.NHosts < 0 {
		return false, errors.New("NHosts parameter must be number >= 0")
	}

	if cluster.NHosts == 0 {
		res, err := CheckMSServices(hS, cluster)
		if err != nil {
			return false, err
		}
		if res {
			return false, errors.New("NHosts parameter must be number >= 1 if you want to install master-slave services.")
		}
	}

	dbFlavor, err := FlavorGetByIdOrName(hS, cluster.MasterFlavor)
	if err != nil {
		return false, err
	}
	if dbFlavor.ID == "" {
		return false, errors.New(fmt.Sprintf("Flavor with name '%s' not found", cluster.MasterFlavor))
	}
	dbFlavor, err = FlavorGetByIdOrName(hS, cluster.SlavesFlavor)
	if err != nil {
		return false, err
	}
	if dbFlavor.ID == "" {
		return false, errors.New(fmt.Sprintf("Flavor with name '%s' not found", cluster.SlavesFlavor))
	}
	dbFlavor, err = FlavorGetByIdOrName(hS, cluster.StorageFlavor)
	if err != nil {
		return false, err
	}
	if dbFlavor.ID == "" {
		return false, errors.New(fmt.Sprintf("Flavor with name '%s' not found", cluster.StorageFlavor))
	}
	dbFlavor, err = FlavorGetByIdOrName(hS, cluster.MonitoringFlavor)
	if err != nil {
		return false, err
	}
	if dbFlavor.ID == "" {
		return false, errors.New(fmt.Sprintf("Flavor with name '%s' not found", cluster.MonitoringFlavor))
	}
	return true, nil
}

func AddDependencies(hS HttpServer, c *protobuf.Cluster, curS *protobuf.Service) ([]*protobuf.Service, error) {
	var err error = nil
	var serviceToAdd *protobuf.Service = nil
	var servicesList []*protobuf.Service = nil

	sv, err := hS.Db.ReadServiceTypeVersion(curS.Type, curS.Version)
	if err != nil {
		return nil, err
	}

	//check if version has dependencies
	if sv.Dependencies != nil {
		for _, sd := range sv.Dependencies {
			//check if the service from dependencies has already listed in cluster and version is ok
			flagAddS := true
			for _, clusterS := range c.Services {
				if clusterS.Type == sd.ServiceType {
					if !utils.ItemExists(sd.ServiceVersions, clusterS.Version) {
						//error: bad service version from user list
						err = errors.New("service " + clusterS.Type +
							" has incompatible version for service " + curS.Type + ".")
					}
					flagAddS = false
					break
				}
			}
			if flagAddS && err == nil {
				//add service from dependencies with default configurations
				serviceToAdd = &protobuf.Service{
					Name:    curS.Name + "-dependent", //TODO: use better service name?
					Type:    sd.ServiceType,
					Version: sd.DefaultServiceVersion,
				}
				servicesList = append(servicesList, serviceToAdd)
			}
		}
	}

	return servicesList, err
}

//returns true if master-slave service exists
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

func ValidateProject(project *protobuf.Project) bool {
	validName := regexp.MustCompile(`^[A-Za-z][A-Za-z0-9-]+$`).MatchString

	if !validName(project.DisplayName) {
		return false
	}
	return true
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

func ValidateService(hS HttpServer, service *protobuf.Service) (bool, error) {
	hS.Logger.Print("Validating service type and config params...")

	if service.Type == "" {
		return false, errors.New("service type can't be nil")
	}

	sTypes, err := hS.Db.ReadServicesTypesList()
	if err != nil {
		return false, err
	}

	//check that service type is supported
	stOk := false
	var stIdx int
	for i, st := range sTypes {
		if st.Type == service.Type {
			stOk = true
			stIdx = i
			break
		}
	}

	if !stOk {
		return false, errors.New("service type " + service.Type + " is not supported")
	}

	//check service version
	if service.Version == "" && sTypes[stIdx].DefaultVersion != "" {
		service.Version = sTypes[stIdx].DefaultVersion
	} else if service.Version == "" && sTypes[stIdx].DefaultVersion == "" {
		return false, errors.New("service version and default version for service type " + service.Type + " are nil")
	}

	//get idx of service version
	var svIdx int
	svOk := false
	for i, sv := range sTypes[stIdx].Versions {
		if sv.Version == service.Version {
			svIdx = i
			svOk = true
			break
		}
	}

	if !svOk {
		return false, errors.New("service version " + service.Version + " is not supported")
	}

	//validate configs
	for k, v := range service.Config {
		flagPN := false
		for _, sc := range sTypes[stIdx].Versions[svIdx].Configs {
			if k == sc.ParameterName {
				flagPN = true

				//check type
				if !sc.IsList {
					switch sc.Type {
					case "int":
						if _, err := strconv.ParseInt(v, 10, 32); err != nil {
							return false, err
						}
					case "float":
						if _, err := strconv.ParseFloat(v, 64); err != nil {
							return false, err
						}
					case "bool":
						if _, err := strconv.ParseBool(v); err != nil {
							return false, err
						}
					}
				} else {
					switch sc.Type {
					case "int":
						var valList []int64
						if err := json.Unmarshal([]byte(v), &valList); err != nil {
							return false, err
						}
					case "float":
						var valList []float64
						if err := json.Unmarshal([]byte(v), &valList); err != nil {
							return false, err
						}
					case "bool":
						var valList []bool
						if err := json.Unmarshal([]byte(v), &valList); err != nil {
							return false, err
						}
					case "string":
						var valList []string
						if err := json.Unmarshal([]byte(v), &valList); err != nil {
							return false, err
						}
					}
				}

				//check for possible values
				if sc.PossibleValues != nil {
					flagPV := CheckValuesAllowed(v, sc.PossibleValues)
					if !flagPV {
						return false, errors.New("service version " + v + " is not supported")
					}
				}

				break
			}
		}
		if !flagPN {
			return false, errors.New("service config param name " + k + " is not supported")
		}
	}

	return true, nil
}
