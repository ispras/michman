package check

import (
	"encoding/json"
	"github.com/ispras/michman/internal/database"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/utils"
	"strconv"
)

// SupportedType list of supported types
func SupportedType(_type string) error {
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

// MSServices returns true if master-slave service exists
func MSServices(db database.Database, cluster *protobuf.Cluster) (bool, error) {
	for _, service := range cluster.Services {
		st, err := db.ReadServiceType(service.Type)
		if err != nil {
			return false, err
		}
		if st.Class == utils.ClassMasterSlave {
			return true, nil
		}
	}
	return false, nil
}

// PossibleValuesUnique checks whether the values in the transmitted list are unique
func PossibleValuesUnique(possibleValues []string) error {
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

// PossibleValues checks possible values field on correctness
func PossibleValues(possibleValues []string, vType string, IsList bool) error {
	// check PossibleValues type correct
	for _, value := range possibleValues {
		if err := CorrectType(value, vType, IsList); err != nil {
			return ErrPossibleValues(value)
		}
	}

	// format PossibleValue strings
	if IsList {
		for i, pV := range possibleValues {
			possibleValues[i] = utils.DeleteSpaces(pV)
		}
	}

	// check PossibleValues are unique
	err := PossibleValuesUnique(possibleValues)
	if err != nil {
		return err
	}
	return nil
}

// CorrectType checks whether the value matches the required type
func CorrectType(value string, Type string, IsList bool) error {
	if !IsList {
		switch Type {
		case "int":
			if _, err := strconv.ParseInt(value, 10, 32); err != nil {
				return err
			}
		case "float":
			if _, err := strconv.ParseFloat(value, 64); err != nil {
				return err
			}
		case "bool":
			if _, err := strconv.ParseBool(value); err != nil {
				return err
			}
		}
	} else {
		switch Type {
		case "int":
			var valList []int64
			if err := json.Unmarshal([]byte(value), &valList); err != nil {
				return err
			}
		case "float":
			var valList []float64
			if err := json.Unmarshal([]byte(value), &valList); err != nil {
				return err
			}
		case "bool":
			var valList []bool
			if err := json.Unmarshal([]byte(value), &valList); err != nil {
				return err
			}
		case "string":
			var valList []string
			if err := json.Unmarshal([]byte(value), &valList); err != nil {
				return err
			}
		}
	}

	return nil
}
