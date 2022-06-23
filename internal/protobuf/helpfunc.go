package protobuf

import (
	"github.com/sirupsen/logrus"
)

func (c Cluster) PrintClusterData(Logger *logrus.Logger) {
	Logger.Info("Cluster info:")
	Logger.Infof("----Cluster name: %s", c.Name)
	Logger.Infof("----Cluster ID: %s", c.ID)
	Logger.Infof("----Cluster status: %s", c.EntityStatus)
	Logger.Info("Services:")

	for i := 0; i < len(c.Services); i++ {
		c.Services[i].PrintServiceData(Logger)
	}
}

func (s Service) PrintServiceData(Logger *logrus.Logger) {
	Logger.Infof("----Service name: %s, type: %s", s.Name, s.Type)
}
