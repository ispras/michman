package database

import (
	proto "gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/protobuf"
)

type Database interface {
	WriteCluster(cluster *proto.Cluster) error
	ReadCluster(name string) (*proto.Cluster, error)
	ListClusters() ([]proto.Cluster, error)
	DeleteCluster(name string) error

	ListProjects() ([]proto.Project, error)
	ReadProject(name string) (*proto.Project, error)
	WriteProject(project *proto.Project) error

	WriteServiceType(sType *proto.ServiceType) error
	ReadServiceType(sType string) (*proto.ServiceType, error)
	ListServicesTypes() ([]proto.ServiceType, error)
	DeleteServiceType(name string) error
}

