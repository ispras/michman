package check

import (
	"github.com/ispras/michman/internal/database"
	"github.com/ispras/michman/internal/protobuf"
)

// FlavorUsed checks whether the transferred flavor is used in other created clusters or projects
func FlavorUsed(db database.Database, flavorName string) (bool, error) {
	clusters, err := db.ReadClustersList()
	if err != nil {
		return false, err
	}
	for _, c := range clusters {
		if c.MasterFlavor == flavorName || c.StorageFlavor == flavorName ||
			c.SlavesFlavor == flavorName || c.MonitoringFlavor == flavorName {
			return true, nil
		}
	}
	projects, err := db.ReadProjectsList()
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

// FlavorFields checks type and value of the VCPUs, RAM and Disk fields of the flavor structure
func FlavorFields(flavor *protobuf.Flavor) error {
	switch interface{}(flavor.VCPUs).(type) {
	case int32:
		if flavor.VCPUs < 0 {
			return ErrFlavorParamVal("VCPUs")
		}
	default:
		return ErrFlavorParamType("VCPUs")
	}

	switch interface{}(flavor.RAM).(type) {
	case int32:
		if flavor.RAM < 0 {
			return ErrFlavorParamVal("RAM")
		}
	default:
		return ErrFlavorParamType("RAM")
	}

	switch interface{}(flavor.Disk).(type) {
	case int32:
		if flavor.Disk < 0 {
			return ErrFlavorParamVal("Disk")
		}
	default:
		return ErrFlavorParamType("Disk")
	}
	return nil
}
