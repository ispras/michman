package validate

import (
	"encoding/json"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/rest/handler/check"
	"net/http"
	"strconv"
)

func Configs(service *protobuf.Service, Configs []*protobuf.ServiceConfig) (error, int) {
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
					if !check.ValuesAllowed(value, sc.PossibleValues) {
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
