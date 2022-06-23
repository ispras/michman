package database

import (
	proto "github.com/ispras/michman/internal/protobuf"
)

type Database interface {
	ReadCluster(clusterID string) (*proto.Cluster, error)
	ReadClusterByName(projectID, clusterName string) (*proto.Cluster, error)
	WriteCluster(cluster *proto.Cluster) error
	DeleteCluster(clusterID string) error
	UpdateCluster(*proto.Cluster) error
	ListClusters() ([]proto.Cluster, error)

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

	ReadImage(imageName string) (*proto.Image, error)
	WriteImage(image *proto.Image) error
	DeleteImage(imageName string) error
	UpdateImage(name string, image *proto.Image) error
	ListImages() ([]proto.Image, error)

	ReadFlavorById(flavorID string) (*proto.Flavor, error)
	ReadFlavorByName(flavorName string) (*proto.Flavor, error)
	WriteFlavor(flavor *proto.Flavor) error
	DeleteFlavor(flavorName string) error
	UpdateFlavor(name string, Flavor *proto.Flavor) error
	ListFlavors() ([]proto.Flavor, error)
}
