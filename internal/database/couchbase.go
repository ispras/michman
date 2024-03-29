package database

import (
	"fmt"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/utils"
	"gopkg.in/couchbase/gocb.v1"
)

const (
	clusterBucketName     string = "clusters"
	templateBucketName    string = "templates"
	projectBucketName     string = "projects"
	serviceTypeBucketName string = "service_types"
	imageBucketName       string = "images"
	flavorBucketName      string = "flavors"
)

type CouchDatabase struct {
	auth               *utils.CbCredentials
	couchCluster       *gocb.Cluster
	clustersBucket     *gocb.Bucket
	projectsBucket     *gocb.Bucket
	templatesBucket    *gocb.Bucket
	serviceTypesBucket *gocb.Bucket
	imageBucket        *gocb.Bucket
	flavorBucket       *gocb.Bucket
	VaultCommunicator  utils.SecretStorage
}

func NewCouchBase(vaultCom utils.SecretStorage) (Database, error) {
	couchbase := new(CouchDatabase)
	couchbase.VaultCommunicator = vaultCom
	client, vaultCfg, err := couchbase.VaultCommunicator.ConnectVault()
	if client == nil || err != nil {
		return nil, err
	}

	couchSecrets, err := client.Logical().Read(vaultCfg.CbKey)
	if err != nil {
		return nil, ErrCouchSecretsRead
	}

	couchbase.auth = &utils.CbCredentials{
		Address:  couchSecrets.Data[utils.CouchbasePath].(string),
		Username: couchSecrets.Data[utils.CouchbaseUsername].(string),
		Password: couchSecrets.Data[utils.CouchbasePassword].(string),
	}
	cluster, err := gocb.Connect(couchbase.auth.Address)
	if err != nil {
		return nil, ErrCouchbaseClusterConnection
	}
	err = cluster.Authenticate(gocb.PasswordAuthenticator{
		Username: couchbase.auth.Username,
		Password: couchbase.auth.Password,
	})
	if err != nil {
		return nil, ErrCouchbaseClusterAuthenticate
	}

	couchbase.couchCluster = cluster

	bucket, err := couchbase.couchCluster.OpenBucket(projectBucketName, "")
	if err != nil {
		return nil, ErrOpenParamBucket("project")
	}
	couchbase.projectsBucket = bucket

	bucket, err = couchbase.couchCluster.OpenBucket(clusterBucketName, "")
	if err != nil {
		return nil, ErrOpenParamBucket("cluster")
	}
	couchbase.clustersBucket = bucket

	bucket, err = couchbase.couchCluster.OpenBucket(templateBucketName, "")
	if err != nil {
		return nil, ErrOpenParamBucket("template")
	}
	couchbase.templatesBucket = bucket

	bucket, err = couchbase.couchCluster.OpenBucket(serviceTypeBucketName, "")
	if err != nil {
		return nil, ErrOpenParamBucket("service")
	}
	couchbase.serviceTypesBucket = bucket

	bucket, err = couchbase.couchCluster.OpenBucket(imageBucketName, "")
	if err != nil {
		return nil, ErrOpenParamBucket("image")
	}
	couchbase.imageBucket = bucket

	bucket, err = couchbase.couchCluster.OpenBucket(flavorBucketName, "")
	if err != nil {
		return nil, ErrOpenParamBucket("flavor")
	}
	couchbase.flavorBucket = bucket

	return couchbase, nil
}

func (db *CouchDatabase) getCouchCluster() error {
	if db.auth == nil {
		client, vaultCfg, err := db.VaultCommunicator.ConnectVault()
		if client == nil || err != nil {
			return err
		}

		couchSecrets, err := client.Logical().Read(vaultCfg.CbKey)
		if err != nil {
			return ErrCouchSecretsRead
		}

		db.auth = &utils.CbCredentials{
			Address:  couchSecrets.Data[utils.CouchbasePath].(string),
			Username: couchSecrets.Data[utils.CouchbaseUsername].(string),
			Password: couchSecrets.Data[utils.CouchbasePassword].(string),
		}
	}

	cluster, err := gocb.Connect(db.auth.Address)
	if err != nil {
		return ErrCouchbaseClusterConnection
	}
	err = cluster.Authenticate(gocb.PasswordAuthenticator{
		Username: db.auth.Username,
		Password: db.auth.Password,
	})
	if err != nil {
		return ErrCouchbaseClusterAuthenticate
	}
	db.couchCluster = cluster
	return nil
}

// project:

func readProjectById(db CouchDatabase, projectID string) (*protobuf.Project, error) {
	var project protobuf.Project
	_, err := db.projectsBucket.Get(projectID, &project)
	if err != nil {
		return nil, ErrObjectNotFound("project", projectID)
	}
	return &project, nil
}

func readProjectByName(db CouchDatabase, projectName string) (*protobuf.Project, error) {
	q := fmt.Sprintf("SELECT b.* FROM %s b WHERE Name = '%s'", projectBucketName, projectName)
	query := gocb.NewN1qlQuery(q)
	rows, err := db.couchCluster.ExecuteN1qlQuery(query, []interface{}{})
	if err != nil {
		return nil, ErrQueryExecution
	}
	var project protobuf.Project
	rows.Next(&project)
	err = rows.Close()
	if err != nil {
		return nil, ErrCloseQuerySession
	}
	if project.ID == "" {
		return nil, ErrObjectNotFound("project", projectName)
	}
	return &project, nil
}

func deleteProjectById(db CouchDatabase, projectID string) error {
	_, err := db.projectsBucket.Remove(projectID, 0)
	if err != nil {
		return ErrDeleteObjectByKey
	}
	return nil
}

func deleteProjectByName(db CouchDatabase, projectName string) error {
	project, err := readProjectByName(db, projectName)
	if err != nil {
		return err
	}

	err = deleteProjectById(db, project.ID)
	return err
}

func (db CouchDatabase) ReadProject(projectIdOrName string) (*protobuf.Project, error) {
	isUuid := utils.IsUuid(projectIdOrName)
	var project *protobuf.Project
	var err error
	if isUuid {
		project, err = readProjectById(db, projectIdOrName)
	} else {
		project, err = readProjectByName(db, projectIdOrName)
	}
	if err != nil {
		return nil, err
	}
	return project, nil
}

func (db CouchDatabase) ReadProjectsList() ([]protobuf.Project, error) {
	q := fmt.Sprintf("SELECT b.* FROM %s b", projectBucketName)
	query := gocb.NewN1qlQuery(q)
	rows, err := db.couchCluster.ExecuteN1qlQuery(query, []interface{}{})
	if err != nil {
		return nil, ErrQueryExecution
	}
	var row protobuf.Project
	var result []protobuf.Project

	for rows.Next(&row) {
		result = append(result, row)
		row = protobuf.Project{}
	}
	err = rows.Close()
	if err != nil {
		return nil, ErrCloseQuerySession
	}

	return result, nil
}

func (db CouchDatabase) ReadProjectClusters(projectIdOrName string) ([]protobuf.Cluster, error) {
	isUuid := utils.IsUuid(projectIdOrName)
	q := fmt.Sprintf("SELECT b.* FROM %s b WHERE Name = '%s'", clusterBucketName, projectIdOrName)
	if isUuid {
		q = fmt.Sprintf("SELECT b.* FROM %s b WHERE ProjectID = '%s'", clusterBucketName, projectIdOrName)
	}
	query := gocb.NewN1qlQuery(q)
	rows, err := db.couchCluster.ExecuteN1qlQuery(query, []interface{}{})
	if err != nil {
		return nil, ErrQueryExecution
	}
	var row protobuf.Cluster
	var result []protobuf.Cluster

	for rows.Next(&row) {
		result = append(result, row)
		row = protobuf.Cluster{}
	}
	err = rows.Close()
	if err != nil {
		return nil, ErrCloseQuerySession
	}
	return result, nil
}

func (db CouchDatabase) WriteProject(project *protobuf.Project) error {
	_, err := db.projectsBucket.Upsert(project.ID, project, 0)
	if err != nil {
		return ErrWriteObjectByKey
	}
	return nil
}

func (db CouchDatabase) UpdateProject(project *protobuf.Project) error {
	var cas gocb.Cas
	_, err := db.projectsBucket.Replace(project.ID, project, cas, 0)
	if err != nil {
		return ErrUpdateObjectByKey
	}
	return nil
}

func (db CouchDatabase) DeleteProject(projectIdOrName string) error {
	isUuid := utils.IsUuid(projectIdOrName)
	var err error
	if isUuid {
		err = deleteProjectById(db, projectIdOrName)
	} else {
		err = deleteProjectByName(db, projectIdOrName)
	}
	return err
}

// cluster:

func readClusterById(db CouchDatabase, clusterID string) (*protobuf.Cluster, error) {
	var cluster protobuf.Cluster
	_, err := db.clustersBucket.Get(clusterID, &cluster)
	if err != nil {
		if err == gocb.ErrKeyNotFound {
			return nil, ErrObjectNotFound("cluster", clusterID)
		}
		return nil, ErrReadObjectByKey
	}
	return &cluster, nil
}

func readClusterByName(db CouchDatabase, projectID string, clusterName string) (*protobuf.Cluster, error) {
	q := fmt.Sprintf("SELECT b.* FROM %s b WHERE ProjectID = '%s' and Name = '%s'", clusterBucketName, projectID, clusterName)
	query := gocb.NewN1qlQuery(q)
	var cluster protobuf.Cluster
	rows, err := db.couchCluster.ExecuteN1qlQuery(query, []interface{}{})
	if err != nil {
		if err == gocb.ErrKeyNotFound {
			return nil, ErrObjectNotFound("cluster", clusterName)
		}
		return nil, ErrQueryExecution
	}
	rows.Next(&cluster)
	err = rows.Close()
	if err != nil {
		return nil, ErrCloseQuerySession
	}
	if cluster.ID == "" {
		return nil, ErrObjectNotFound("cluster", clusterName)
	}
	return &cluster, nil
}

func deleteClusterById(db CouchDatabase, clusterID string) error {
	_, err := db.clustersBucket.Remove(clusterID, 0)
	if err != nil {
		return ErrDeleteObjectByKey
	}
	return nil
}

func deleteClusterByName(db CouchDatabase, projectIdOrName string, clusterName string) error {
	project, err := db.ReadProject(projectIdOrName)
	if err != nil {
		return err
	}

	cluster, err := readClusterByName(db, project.ID, clusterName)
	if err != nil {
		return err
	}
	err = deleteClusterById(db, cluster.ID)
	return err
}

func (db CouchDatabase) ReadCluster(projectIdOrName string, clusterIdOrName string) (*protobuf.Cluster, error) {
	project, err := db.ReadProject(projectIdOrName)
	if err != nil {
		return nil, err
	}

	isUuid := utils.IsUuid(clusterIdOrName)
	var cluster *protobuf.Cluster
	if isUuid {
		cluster, err = readClusterById(db, clusterIdOrName)
	} else {
		cluster, err = readClusterByName(db, project.ID, clusterIdOrName)
	}
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

func (db CouchDatabase) ReadClustersList() ([]protobuf.Cluster, error) {
	q := fmt.Sprintf("SELECT b.* FROM %s b", clusterBucketName)
	query := gocb.NewN1qlQuery(q)
	rows, err := db.couchCluster.ExecuteN1qlQuery(query, []interface{}{})
	if err != nil {
		return nil, ErrQueryExecution
	}
	var row protobuf.Cluster
	var result []protobuf.Cluster

	for rows.Next(&row) {
		result = append(result, row)
		row = protobuf.Cluster{}
	}
	err = rows.Close()
	if err != nil {
		return nil, ErrCloseQuerySession
	}
	return result, nil
}

func (db CouchDatabase) WriteCluster(cluster *protobuf.Cluster) error {
	_, err := db.clustersBucket.Upsert(cluster.ID, cluster, 0)
	if err != nil {
		return ErrWriteObjectByKey
	}
	return nil
}

func (db CouchDatabase) UpdateCluster(cluster *protobuf.Cluster) error {
	var cas gocb.Cas
	_, err := db.clustersBucket.Replace(cluster.ID, cluster, cas, 0)
	if err != nil {
		return ErrUpdateObjectByKey
	}
	return nil
}

func (db CouchDatabase) DeleteCluster(projectIdOrName, clusterIdOrName string) error {
	isUuid := utils.IsUuid(clusterIdOrName)
	var err error
	if isUuid {
		err = deleteClusterById(db, clusterIdOrName)
	} else {
		err = deleteClusterByName(db, projectIdOrName, clusterIdOrName)
	}
	return err
}

// service type:

func readServiceTypeById(db CouchDatabase, serviceTypeID string) (*protobuf.ServiceType, error) {
	var sType protobuf.ServiceType
	_, err := db.serviceTypesBucket.Get(serviceTypeID, &sType)
	if err != nil {
		if err == gocb.ErrKeyNotFound {
			return nil, ErrObjectNotFound("service type", serviceTypeID)
		} else {
			return nil, ErrReadObjectByKey
		}
	}
	return &sType, nil
}

func readServiceTypeByName(db CouchDatabase, serviceTypeName string) (*protobuf.ServiceType, error) {
	q := fmt.Sprintf("SELECT b.* FROM %s b WHERE Type = '%s'", serviceTypeBucketName, serviceTypeName)
	query := gocb.NewN1qlQuery(q)
	rows, err := db.couchCluster.ExecuteN1qlQuery(query, []interface{}{})
	if err != nil {
		return nil, ErrQueryExecution
	}
	defer rows.Close()
	var sType protobuf.ServiceType
	rows.Next(&sType)
	if sType.ID == "" {
		return nil, ErrObjectNotFound("service type", serviceTypeName)
	}
	return &sType, nil
}

func deleteServiceTypeById(db CouchDatabase, serviceTypeID string) error {
	_, err := db.serviceTypesBucket.Remove(serviceTypeID, 0)
	if err != nil {
		return ErrDeleteObjectByKey
	}
	return nil
}

func deleteServiceTypeByName(db CouchDatabase, serviceTypeName string) error {
	sType, err := readServiceTypeByName(db, serviceTypeName)
	if err != nil {
		return err
	}

	err = deleteServiceTypeById(db, sType.ID)
	return err
}

func (db CouchDatabase) ReadServiceType(serviceTypeIdOrName string) (*protobuf.ServiceType, error) {
	isUuid := utils.IsUuid(serviceTypeIdOrName)
	var sType *protobuf.ServiceType
	var err error
	if isUuid {
		sType, err = readServiceTypeById(db, serviceTypeIdOrName)
	} else {
		sType, err = readServiceTypeByName(db, serviceTypeIdOrName)
	}
	if err != nil {
		return nil, err
	}
	return sType, nil
}

func (db CouchDatabase) ReadServicesTypesList() ([]protobuf.ServiceType, error) {
	q := fmt.Sprintf("SELECT b.* FROM %s b", serviceTypeBucketName)
	query := gocb.NewN1qlQuery(q)
	rows, err := db.serviceTypesBucket.ExecuteN1qlQuery(query, []interface{}{})
	if err != nil {
		return nil, ErrQueryExecution
	}
	var row protobuf.ServiceType
	var result []protobuf.ServiceType

	for rows.Next(&row) {
		result = append(result, row)
		row = protobuf.ServiceType{}
	}
	err = rows.Close()
	if err != nil {
		return nil, ErrCloseQuerySession
	}
	return result, nil
}

func (db CouchDatabase) WriteServiceType(sType *protobuf.ServiceType) error {
	_, err := db.serviceTypesBucket.Upsert(sType.ID, sType, 0)
	if err != nil {
		return ErrWriteObjectByKey
	}
	return err
}

func (db CouchDatabase) UpdateServiceType(sType *protobuf.ServiceType) error {
	var cas gocb.Cas
	_, err := db.serviceTypesBucket.Replace(sType.ID, sType, cas, 0)
	if err != nil {
		return ErrUpdateObjectByKey
	}
	return nil
}

func (db CouchDatabase) DeleteServiceType(serviceTypeIdOrName string) error {
	isUuid := utils.IsUuid(serviceTypeIdOrName)
	var err error
	if isUuid {
		err = deleteServiceTypeById(db, serviceTypeIdOrName)
	} else {
		err = deleteServiceTypeByName(db, serviceTypeIdOrName)
	}
	return err
}

// service type version:

func readServiceTypeVersionById(sType *protobuf.ServiceType, versionId string) (*protobuf.ServiceVersion, error) {
	for _, curVersion := range sType.Versions {
		if curVersion.ID == versionId {
			return curVersion, nil
		}
	}

	return nil, ErrObjectNotFound("version", versionId)
}

func readServiceTypeVersionByName(sType *protobuf.ServiceType, versionName string) (*protobuf.ServiceVersion, error) {
	for _, curVersion := range sType.Versions {
		if curVersion.Version == versionName {
			return curVersion, nil
		}
	}
	return nil, ErrObjectNotFound("version", versionName)
}

func deleteServiceTypeVersionById(sType *protobuf.ServiceType, versionId string) (int, error) {
	idToDelete := -1
	for i, curVersion := range sType.Versions {
		if curVersion.ID == versionId {
			idToDelete = i
			break
		}
	}
	if idToDelete == -1 {
		return -1, ErrObjectNotFound("version", versionId)
	}
	return idToDelete, nil
}

func deleteServiceTypeVersionByName(sType *protobuf.ServiceType, versionName string) (int, error) {
	idToDelete := -1
	for i, curVersion := range sType.Versions {
		if curVersion.Version == versionName {
			idToDelete = i
			break
		}
	}
	if idToDelete == -1 {
		return -1, ErrObjectNotFound("version", versionName)
	}
	return idToDelete, nil
}

func (db CouchDatabase) ReadServiceTypeVersion(serviceTypeIdOrName string, versionIdOrName string) (*protobuf.ServiceVersion, error) {
	sType, err := db.ReadServiceType(serviceTypeIdOrName)
	if err != nil {
		return nil, err
	}

	var sTypeVersion *protobuf.ServiceVersion
	isUuid := utils.IsUuid(versionIdOrName)
	if isUuid {
		sTypeVersion, err = readServiceTypeVersionById(sType, versionIdOrName)
	} else {
		sTypeVersion, err = readServiceTypeVersionByName(sType, versionIdOrName)
	}
	if err != nil {
		return nil, err
	}
	return sTypeVersion, nil
}

func (db CouchDatabase) DeleteServiceTypeVersion(serviceTypeIdOrName string, versionIdOrName string) error {
	sType, err := db.ReadServiceType(serviceTypeIdOrName)
	if err != nil {
		return err
	}

	idToDelete := -1
	isUuid := utils.IsUuid(versionIdOrName)
	if isUuid {
		idToDelete, err = deleteServiceTypeVersionById(sType, versionIdOrName)
	} else {
		idToDelete, err = deleteServiceTypeVersionByName(sType, versionIdOrName)
	}

	versionsLen := len(sType.Versions)
	sType.Versions[idToDelete] = sType.Versions[versionsLen-1]
	sType.Versions = sType.Versions[:versionsLen-1]

	err = db.UpdateServiceType(sType)
	if err != nil {
		return err
	}
	return nil
}

func (db CouchDatabase) UpdateServiceTypeVersion(serviceTypeIdOrName string, version *protobuf.ServiceVersion) error {
	sType, err := db.ReadServiceType(serviceTypeIdOrName)
	if err != nil {
		return err
	}

	idToUpdate := -1
	for i, curVersion := range sType.Versions {
		if curVersion.ID == version.ID {
			idToUpdate = i
			break
		}
	}
	if idToUpdate == -1 {
		return ErrObjectNotFound("service type version", version.Version)
	}

	sType.Versions[idToUpdate] = version

	err = db.UpdateServiceType(sType)
	if err != nil {
		return err
	}
	return nil
}

// service type version config:

func (db CouchDatabase) ReadServiceTypeVersionConfig(serviceTypeIdOrName string, versionIdOrName string, parameterName string) (*protobuf.ServiceConfig, error) {
	sType, err := db.ReadServiceType(serviceTypeIdOrName)
	if err != nil {
		return nil, err
	}

	var sTypeVersion *protobuf.ServiceVersion
	isUuid := utils.IsUuid(versionIdOrName)
	if isUuid {
		sTypeVersion, err = readServiceTypeVersionById(sType, versionIdOrName)
	} else {
		sTypeVersion, err = readServiceTypeVersionByName(sType, versionIdOrName)
	}

	if err != nil {
		return nil, err
	}

	sTypeVersionConfig := new(protobuf.ServiceConfig)
	if sTypeVersion.Configs != nil {
		for _, config := range sTypeVersion.Configs {
			if config.ParameterName == parameterName {
				sTypeVersionConfig = config
				break
			}
		}
	}
	return sTypeVersionConfig, nil
}

func (db CouchDatabase) UpdateServiceTypeVersionConfig(serviceTypeIdOrName string, versionIdOrName string, config *protobuf.ServiceConfig) error {
	sTypeVersion, err := db.ReadServiceTypeVersion(serviceTypeIdOrName, versionIdOrName)
	if err != nil {
		return err
	}

	idToUpdate := -1
	for i, curVersionConfig := range sTypeVersion.Configs {
		if curVersionConfig.ParameterName == config.ParameterName {
			idToUpdate = i
			break
		}
	}
	if idToUpdate == -1 {
		return ErrObjectNotFound("service type version config", config.ParameterName)
	}

	sTypeVersion.Configs[idToUpdate] = config

	err = db.UpdateServiceTypeVersion(serviceTypeIdOrName, sTypeVersion)
	if err != nil {
		return err
	}
	return nil
}

func (db CouchDatabase) DeleteServiceTypeVersionConfig(serviceTypeIdOrName string, versionIdOrName string, parameterName string) error {
	sTypeVersion, err := db.ReadServiceTypeVersion(serviceTypeIdOrName, versionIdOrName)
	if err != nil {
		return err
	}

	idToDelete := -1
	for i, curConfig := range sTypeVersion.Configs {
		if curConfig.ParameterName == parameterName {
			idToDelete = i
			break
		}
	}
	if idToDelete == -1 {
		return ErrObjectNotFound("service type version config", parameterName)
	}

	configsLen := len(sTypeVersion.Configs)
	sTypeVersion.Configs[idToDelete] = sTypeVersion.Configs[configsLen-1]
	sTypeVersion.Configs = sTypeVersion.Configs[:configsLen-1]

	err = db.UpdateServiceTypeVersion(serviceTypeIdOrName, sTypeVersion)
	if err != nil {
		return err
	}
	return nil
}

// image:

func readImageById(db CouchDatabase, imageID string) (*protobuf.Image, error) {
	var image protobuf.Image
	_, err := db.imageBucket.Get(imageID, &image)
	if err != nil {
		if err == gocb.ErrKeyNotFound {
			return nil, ErrObjectNotFound("image", imageID)
		}
		return nil, ErrReadObjectByKey
	}
	return &image, nil
}

func readImageByName(db CouchDatabase, imageName string) (*protobuf.Image, error) {
	q := fmt.Sprintf("SELECT b.* FROM %s b WHERE Name = '%s'", imageBucketName, imageName)
	query := gocb.NewN1qlQuery(q)
	rows, err := db.couchCluster.ExecuteN1qlQuery(query, []interface{}{})
	if err != nil {
		return nil, ErrQueryExecution
	}
	defer rows.Close()
	var image protobuf.Image
	rows.Next(&image)
	if image.ID == "" {
		return nil, ErrObjectNotFound("image", imageName)
	}
	return &image, nil
}

func deleteImageById(db CouchDatabase, imageID string) error {
	_, err := db.imageBucket.Remove(imageID, 0)
	if err != nil {
		return ErrDeleteObjectByKey
	}
	return nil
}

func deleteImageByName(db CouchDatabase, imageName string) error {
	image, err := readImageByName(db, imageName)
	if err != nil {
		return err
	}

	err = deleteImageById(db, image.ID)
	return err
}

func (db CouchDatabase) ReadImage(imageIdOrName string) (*protobuf.Image, error) {
	isUuid := utils.IsUuid(imageIdOrName)
	var image *protobuf.Image
	var err error
	if isUuid {
		image, err = readImageById(db, imageIdOrName)
	} else {
		image, err = readImageByName(db, imageIdOrName)
	}
	if err != nil {
		return nil, err
	}
	return image, nil
}

func (db CouchDatabase) ReadImagesList() ([]protobuf.Image, error) {
	q := fmt.Sprintf("SELECT b.* FROM %s b", imageBucketName)
	query := gocb.NewN1qlQuery(q)
	rows, err := db.couchCluster.ExecuteN1qlQuery(query, []interface{}{})
	if err != nil {
		return nil, ErrQueryExecution
	}
	var row protobuf.Image
	var result []protobuf.Image

	for rows.Next(&row) {
		result = append(result, row)
		row = protobuf.Image{}
	}
	err = rows.Close()
	if err != nil {
		return nil, ErrCloseQuerySession
	}

	return result, nil
}

func (db CouchDatabase) WriteImage(image *protobuf.Image) error {
	_, err := db.imageBucket.Upsert(image.ID, image, 0)
	if err != nil {
		return err
	}
	return nil
}

func (db CouchDatabase) UpdateImage(image *protobuf.Image) error {
	var cas gocb.Cas
	_, err := db.imageBucket.Replace(image.ID, image, cas, 0)
	if err != nil {
		return ErrUpdateObjectByKey
	}
	return nil
}

func (db CouchDatabase) DeleteImage(imageIdOrName string) error {
	isUuid := utils.IsUuid(imageIdOrName)
	var err error
	if isUuid {
		err = deleteImageById(db, imageIdOrName)
	} else {
		err = deleteImageByName(db, imageIdOrName)
	}
	return err
}

// flavors

func readFlavorById(db CouchDatabase, flavorID string) (*protobuf.Flavor, error) {
	var flavor protobuf.Flavor
	_, err := db.flavorBucket.Get(flavorID, &flavor)
	if err != nil {
		if err == gocb.ErrKeyNotFound {
			return nil, ErrObjectNotFound("flavor", flavorID)
		}
		return nil, ErrReadObjectByKey
	}
	return &flavor, nil
}

func readFlavorByName(db CouchDatabase, flavorName string) (*protobuf.Flavor, error) {
	q := fmt.Sprintf("SELECT b.* FROM %s b WHERE Name = '%s'", flavorBucketName, flavorName)
	query := gocb.NewN1qlQuery(q)
	rows, err := db.couchCluster.ExecuteN1qlQuery(query, []interface{}{})
	if err != nil {
		return nil, ErrQueryExecution
	}
	defer rows.Close()
	var flavor protobuf.Flavor
	rows.Next(&flavor)
	if flavor.ID == "" {
		return nil, ErrObjectNotFound("flavor", flavorName)
	}
	return &flavor, nil
}

func deleteFlavorById(db CouchDatabase, flavorID string) error {
	_, err := db.flavorBucket.Remove(flavorID, 0)
	if err != nil {
		return ErrDeleteObjectByKey
	}
	return nil
}

func deleteFlavorByName(db CouchDatabase, flavorName string) error {
	flavor, err := readFlavorByName(db, flavorName)
	if err != nil {
		return err
	}

	err = deleteFlavorById(db, flavor.ID)
	return err
}

func (db CouchDatabase) ReadFlavor(flavorIdOrName string) (*protobuf.Flavor, error) {
	isUuid := utils.IsUuid(flavorIdOrName)
	var flavor *protobuf.Flavor
	var err error
	if isUuid {
		flavor, err = readFlavorById(db, flavorIdOrName)
	} else {
		flavor, err = readFlavorByName(db, flavorIdOrName)
	}
	if err != nil {
		return nil, err
	}
	return flavor, nil
}

func (db CouchDatabase) ReadFlavorsList() ([]protobuf.Flavor, error) {
	q := fmt.Sprintf("SELECT b.* FROM %s b", flavorBucketName)
	query := gocb.NewN1qlQuery(q)
	rows, err := db.couchCluster.ExecuteN1qlQuery(query, []interface{}{})
	if err != nil {
		return nil, ErrQueryExecution
	}
	var row protobuf.Flavor
	var result []protobuf.Flavor

	for rows.Next(&row) {
		result = append(result, row)
		row = protobuf.Flavor{}
	}
	err = rows.Close()
	if err != nil {
		return nil, ErrCloseQuerySession
	}
	return result, nil
}

func (db CouchDatabase) WriteFlavor(flavor *protobuf.Flavor) error {
	_, err := db.flavorBucket.Upsert(flavor.ID, flavor, 0)
	if err != nil {
		return ErrWriteObjectByKey
	}
	return nil
}

func (db CouchDatabase) UpdateFlavor(id string, flavor *protobuf.Flavor) error {
	var cas gocb.Cas
	_, err := db.flavorBucket.Replace(id, flavor, cas, 0)
	if err != nil {
		return ErrUpdateObjectByKey
	}
	return nil
}

func (db CouchDatabase) DeleteFlavor(flavorIdOrName string) error {
	isUuid := utils.IsUuid(flavorIdOrName)
	var err error
	if isUuid {
		err = deleteFlavorById(db, flavorIdOrName)
	} else {
		err = deleteFlavorByName(db, flavorIdOrName)
	}
	return err
}

// template:

func (db CouchDatabase) WriteTemplate(template *protobuf.Template) error {
	_, err := db.templatesBucket.Upsert(template.ID, template, 0)
	if err != nil {
		return err
	}
	return nil
}

func (db CouchDatabase) ReadTemplate(id string) (*protobuf.Template, error) {
	var template protobuf.Template
	_, err := db.templatesBucket.Get(id, &template)
	if err != nil {
		return &protobuf.Template{}, nil
	}
	return &template, nil
}

func (db CouchDatabase) ReadTemplateByName(templateName string) (*protobuf.Template, error) {
	query := gocb.NewN1qlQuery(fmt.Sprintf("SELECT ID, ProjectID, Name, DisplayName, Services,"+
		" NSlaves, Description FROM %v WHERE Name = '%v'",
		templateBucketName, templateName))
	rows, err := db.couchCluster.ExecuteN1qlQuery(query, []interface{}{})
	if err != nil {
		return nil, ErrQueryExecution
	}
	var template protobuf.Template

	if hasResult := rows.Next(template); !hasResult {
		template = protobuf.Template{}
	}
	rows.Close()
	return &template, nil
}

func (db CouchDatabase) ListTemplates(projectID string) ([]protobuf.Template, error) {
	query := gocb.NewN1qlQuery(fmt.Sprintf("SELECT ID, ProjectID, Name, DisplayName, Services,"+
		" NSlaves, Description FROM %v WHERE ProjectID = '%v'",
		templateBucketName, projectID))
	rows, err := db.couchCluster.ExecuteN1qlQuery(query, []interface{}{})
	if err != nil {
		return nil, ErrQueryExecution
	}
	var row protobuf.Template
	var result []protobuf.Template

	for rows.Next(&row) {
		result = append(result, row)
	}
	rows.Close()
	return result, nil
}

func (db CouchDatabase) DeleteTemplate(id string) error {
	_, err := db.templatesBucket.Remove(id, 0)
	if err != nil {
		return err
	}
	return nil
}
