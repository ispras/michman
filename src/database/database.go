package database

import (
	proto "github.com/ispras/michman/src/protobuf"
)

type Database interface {
	ReadCluster(clusterID string) (*proto.Cluster, error)
	ReadClusterByName(projectID, clusterName string) (*proto.Cluster, error)
	WriteCluster(cluster *proto.Cluster) error
	DeleteCluster(clusterID string) error
	UpdateCluster(*proto.Cluster) error

	ListProjects() ([]proto.Project, error)
	ReadProject(projectID string) (*proto.Project, error)
	ReadProjectByName(projectName string) (*proto.Project, error)
	ReadProjectClusters(projectID string) ([]proto.Cluster, error)
	WriteProject(project *proto.Project) error
	UpdateProject(*proto.Project) error
	DeleteProject(projectID string) error

	ReadTemplate(projectID, id string) (*proto.Template, error)
	ReadTemplateByName(templateName string) (*proto.Template, error)
	WriteTemplate(template *proto.Template) error
	DeleteTemplate(id string) error
	ListTemplates(projectID string) ([]proto.Template, error)

	WriteServiceType(sType *proto.ServiceType) error
	UpdateServiceType(st *proto.ServiceType) error
	ReadServiceType(sType string) (*proto.ServiceType, error)
	ListServicesTypes() ([]proto.ServiceType, error)
	DeleteServiceType(name string) error
	ReadServiceVersion(sType string, vId string) (*proto.ServiceVersion, error)
	ReadServiceVersionByName(sType string, version string) (*proto.ServiceVersion, error)
	DeleteServiceVersion(sType string, vId string) (*proto.ServiceVersion, error)
}
