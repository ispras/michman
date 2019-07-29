package ansible

import (
	fmt "fmt"
)

func (c Cluster) PrintClusterData() {
	fmt.Printf("Cluster with name: %s, ID: %d,\n", c.Name, c.ID)
	fmt.Printf("status: %s, type: %s and services:\n", c.EntityStatus, c.ClusterType)

	for i := 0; i < len(c.Services); i++ {
		c.Services[i].PrintServiceData()
	}

	fmt.Print("\n")
}

func (s Service) PrintServiceData() {
	fmt.Printf("----Service with name: %s, state: %s\n", s.Name, s.ServiceState)
}
