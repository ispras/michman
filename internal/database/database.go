package database

import (
	"github.com/ispras/michman/internal/protobuf"
)

type Database interface {
	ReadCluster(projectIdOrName string, clusterIdOrName string) (*protobuf.Cluster, error)
	WriteCluster(cluster *protobuf.Cluster) error
	DeleteCluster(projectIdOrName, clusterIdOrName string) error
	UpdateCluster(cluster *protobuf.Cluster) error
	ReadClustersList() ([]protobuf.Cluster, error)

	ReadProject(projectIdOrName string) (*protobuf.Project, error)
	ReadProjectsList() ([]protobuf.Project, error)
	ReadProjectClusters(projectIdOrName string) ([]protobuf.Cluster, error)
	WriteProject(project *protobuf.Project) error
	UpdateProject(project *protobuf.Project) error
	DeleteProject(projectIdOrName string) error

	ReadTemplate(projectID, id string) (*protobuf.Template, error)
	ReadTemplateByName(templateName string) (*protobuf.Template, error)
	WriteTemplate(template *protobuf.Template) error
	DeleteTemplate(id string) error
	ListTemplates(projectID string) ([]protobuf.Template, error)

	ReadServiceType(serviceTypeIdOrName string) (*protobuf.ServiceType, error)
	ReadServicesTypesList() ([]protobuf.ServiceType, error)
	WriteServiceType(sType *protobuf.ServiceType) error
	UpdateServiceType(sType *protobuf.ServiceType) error
	DeleteServiceType(serviceTypeIdOrName string) error

	ReadServiceTypeVersion(serviceTypeIdOrName string, versionIdOrName string) (*protobuf.ServiceVersion, error)
	DeleteServiceTypeVersion(serviceTypeIdOrName string, versionId string) error

	ReadImage(imageIdOrName string) (*protobuf.Image, error)
	WriteImage(image *protobuf.Image) error
	DeleteImage(imageIdOrName string) error
	UpdateImage(image *protobuf.Image) error
	ReadImagesList() ([]protobuf.Image, error)

	ReadFlavorById(flavorID string) (*protobuf.Flavor, error)
	ReadFlavorByName(flavorName string) (*protobuf.Flavor, error)
	WriteFlavor(flavor *protobuf.Flavor) error
	DeleteFlavor(flavorName string) error
	UpdateFlavor(name string, Flavor *protobuf.Flavor) error
	ListFlavors() ([]protobuf.Flavor, error)
}
