package protobuf

import (
	"github.com/sirupsen/logrus"
)

func (c Cluster) PrintClusterData(Logger *logrus.Logger) {
	Logger.Printf("Cluster with name: %s, ID: %s,\n", c.Name, c.ID)
	Logger.Printf("status: %s, type: %s and services:\n", c.EntityStatus, c.ClusterType)

	for i := 0; i < len(c.Services); i++ {
		c.Services[i].PrintServiceData(Logger)
	}

	Logger.Print("\n")
}

func (s Service) PrintServiceData(Logger *logrus.Logger) {
	Logger.Printf("----Service with name: %s, state: %s\n", s.Name, s.ServiceState)
}
