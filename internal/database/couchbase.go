package database

import (
	"errors"
	"fmt"
	proto "github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/utils"
	"gopkg.in/couchbase/gocb.v1"
)

const (
	clusterBucketName     string = "clusters"
	templateBucketName    string = "templates"
	projectBucketName     string = "projects"
	serviceTypeBucketName string = "service_types"
	imageBucketName       string = "images"
)

type CouchDatabase struct {
	auth               *utils.CbCredentials
	couchCluster       *gocb.Cluster
	clustersBucket     *gocb.Bucket
	projectsBucket     *gocb.Bucket
	templatesBucket    *gocb.Bucket
	serviceTypesBucket *gocb.Bucket
	imageBucket        *gocb.Bucket
	VaultCommunicator  utils.SecretStorage
}

func NewCouchBase(vaultCom utils.SecretStorage) (Database, error) {
	cb := new(CouchDatabase)
	cb.VaultCommunicator = vaultCom
	client, vaultCfg := cb.VaultCommunicator.ConnectVault()
	if client == nil {
		return nil, errors.New("Error: can't connect to vault secrets storage")
	}

	couchSecrets, err := client.Logical().Read(vaultCfg.CbKey)
	if err != nil {
		return nil, err
	}

	cb.auth = &utils.CbCredentials{
		Address:  couchSecrets.Data[utils.CouchbasePath].(string),
		Username: couchSecrets.Data[utils.CouchbaseUsername].(string),
		Password: couchSecrets.Data[utils.CouchbasePassword].(string),
	}
	cluster, err := gocb.Connect(cb.auth.Address)
	if err != nil {
		return nil, err
	}
	cluster.Authenticate(gocb.PasswordAuthenticator{
		Username: cb.auth.Username,
		Password: cb.auth.Password,
	})
	cb.couchCluster = cluster

	bucket, err := cb.couchCluster.OpenBucket(projectBucketName, "")
	if err != nil {
		return nil, err
	}
	cb.projectsBucket = bucket

	bucket, err = cb.couchCluster.OpenBucket(clusterBucketName, "")
	if err != nil {
		return nil, err
	}
	cb.clustersBucket = bucket

	bucket, err = cb.couchCluster.OpenBucket(templateBucketName, "")
	if err != nil {
		return nil, err
	}
	cb.templatesBucket = bucket

	bucket, err = cb.couchCluster.OpenBucket(serviceTypeBucketName, "")
	if err != nil {
		return nil, err
	}
	cb.serviceTypesBucket = bucket

	bucket, err = cb.couchCluster.OpenBucket(imageBucketName, "")
	if err != nil {
		return nil, err
	}
	cb.imageBucket = bucket

	return cb, nil
}

func (db *CouchDatabase) getCouchCluster() error {
	if db.auth == nil {
		client, vaultCfg := db.VaultCommunicator.ConnectVault()
		if client == nil {
			return errors.New("Error: can't connect to vault secrets storage")
		}

		couchSecrets, err := client.Logical().Read(vaultCfg.CbKey)
		if err != nil {
			return err
		}

		db.auth = &utils.CbCredentials{
			Address:  couchSecrets.Data[utils.CouchbasePath].(string),
			Username: couchSecrets.Data[utils.CouchbaseUsername].(string),
			Password: couchSecrets.Data[utils.CouchbasePassword].(string),
		}
	}

	cluster, err := gocb.Connect(db.auth.Address)
	if err != nil {
		return err
	}
	cluster.Authenticate(gocb.PasswordAuthenticator{
		Username: db.auth.Username,
		Password: db.auth.Password,
	})
	db.couchCluster = cluster
	return nil
}

func (db CouchDatabase) ReadProject(projectID string) (*proto.Project, error) {
	var project proto.Project
	_, err := db.projectsBucket.Get(projectID, &project)
	if err != nil {
		return nil, err
	}

	return &project, nil
}

func (db CouchDatabase) ReadProjectByName(projectName string) (*proto.Project, error) {
	query := gocb.NewN1qlQuery("SELECT ID, Name, DisplayName, GroupID, Description, DefaultImage FROM " + projectBucketName +
		" WHERE Name = '" + projectName + "'")
	result, err := db.couchCluster.ExecuteN1qlQuery(query, []interface{}{})
	if err != nil {
		return nil, err
	}

	var p proto.Project
	result.Next(&p)
	result.Close()
	return &p, nil
}

func (db CouchDatabase) WriteProject(project *proto.Project) error {
	_, err := db.projectsBucket.Upsert(project.ID, project, 0)
	return err
}

func (db CouchDatabase) ListProjects() ([]proto.Project, error) {
	query := gocb.NewN1qlQuery("SELECT ID, Name, DisplayName, GroupID, Description, DefaultImage FROM " + projectBucketName)
	rows, err := db.couchCluster.ExecuteN1qlQuery(query, []interface{}{})
	if err != nil {
		return nil, err
	}
	var row proto.Project
	var result []proto.Project

	for rows.Next(&row) {
		result = append(result, row)
		row = proto.Project{}
	}
	rows.Close()

	return result, nil
}

func (db CouchDatabase) ReadProjectClusters(projectID string) ([]proto.Cluster, error) {
	q := "SELECT ID, Name, DisplayName, HostURL, ClusterType, NHosts, EntityStatus, Image, Services, MasterIP, Description from " + clusterBucketName + " where ProjectID = '" + projectID + "'"
	query := gocb.NewN1qlQuery(q)
	rows, err := db.couchCluster.ExecuteN1qlQuery(query, []interface{}{})
	if err != nil {
		return nil, err
	}
	var row proto.Cluster
	var result []proto.Cluster
	for rows.Next(&row) {
		result = append(result, row)
		row = proto.Cluster{}
	}
	rows.Close()
	return result, nil
}

func (db CouchDatabase) WriteCluster(cluster *proto.Cluster) error {
	_, err := db.clustersBucket.Upsert(cluster.ID, cluster, 0)
	return err
}

func (db CouchDatabase) ReadCluster(clusterID string) (*proto.Cluster, error) {
	var cluster proto.Cluster
	_, err := db.clustersBucket.Get(clusterID, &cluster)
	if err != nil {
		return nil, err
	}
	return &cluster, nil
}

func (db CouchDatabase) ReadClusterByName(projectID, clusterName string) (*proto.Cluster, error) {
	q := "SELECT ID, Name, DisplayName, HostURL, EntityStatus, Image, ClusterType," +
		"Services, NHosts, MasterIP, ProjectID, Description FROM " + clusterBucketName +
		" WHERE ProjectID = '" + projectID + "' and Name = '" + clusterName + "'"
	query := gocb.NewN1qlQuery(q)
	result, err := db.couchCluster.ExecuteN1qlQuery(query, []interface{}{})
	if err != nil {
		return nil, err
	}
	var c proto.Cluster
	result.Next(&c)
	result.Close()
	return &c, nil
}

func (db CouchDatabase) UpdateCluster(cluster *proto.Cluster) error {
	var cas gocb.Cas
	_, err := db.clustersBucket.Replace(cluster.ID, cluster, cas, 0)
	return err
}

func (db CouchDatabase) DeleteCluster(clusterID string) error {
	_, err := db.clustersBucket.Remove(clusterID, 0)
	println(clusterID)
	if err != nil {
		return err
	}
	return nil
}

func (db CouchDatabase) WriteTemplate(template *proto.Template) error {
	_, err := db.templatesBucket.Upsert(template.ID, template, 0)
	if err != nil {
		return err
	}
	return nil
}

func (db CouchDatabase) ListClusters() ([]proto.Cluster, error) {
	query := gocb.NewN1qlQuery("SELECT ID, Name, DisplayName, HostURL, ClusterType, NHosts, EntityStatus, Image, Services, MasterIP, Description FROM " + clusterBucketName)
	rows, err := db.couchCluster.ExecuteN1qlQuery(query, []interface{}{})
	if err != nil {
		return nil, err
	}
	var row proto.Cluster
	var result []proto.Cluster

	for rows.Next(&row) {
		result = append(result, row)
		row = proto.Cluster{}
	}
	rows.Close()

	return result, nil
}

func (db CouchDatabase) ReadTemplate(projectID, id string) (*proto.Template, error) {
	var template proto.Template
	_, err := db.templatesBucket.Get(id, &template)
	if err != nil {
		return &proto.Template{}, nil
	}
	if projectID != template.ProjectID {
		return &proto.Template{}, nil
	}
	return &template, nil
}

func (db CouchDatabase) ReadTemplateByName(templateName string) (*proto.Template, error) {
	query := gocb.NewN1qlQuery(fmt.Sprintf("SELECT ID, ProjectID, Name, DisplayName, Services,"+
		" NHosts, Description FROM %v WHERE Name = '%v'",
		templateBucketName, templateName))
	rows, err := db.couchCluster.ExecuteN1qlQuery(query, []interface{}{})
	if err != nil {
		return nil, err
	}
	var template proto.Template

	if hasResult := rows.Next(template); !hasResult {
		template = proto.Template{}
	}
	rows.Close()
	return &template, nil
}

func (db CouchDatabase) ListTemplates(projectID string) ([]proto.Template, error) {
	query := gocb.NewN1qlQuery(fmt.Sprintf("SELECT ID, ProjectID, Name, DisplayName, Services,"+
		" NHosts, Description FROM %v WHERE ProjectID = '%v'",
		templateBucketName, projectID))
	rows, err := db.couchCluster.ExecuteN1qlQuery(query, []interface{}{})
	if err != nil {
		return nil, err
	}
	var row proto.Template
	var result []proto.Template

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

func (db CouchDatabase) UpdateProject(project *proto.Project) error {
	var cas gocb.Cas
	cas, err := db.projectsBucket.Replace(project.ID, project, cas, 0)

	return err
}

func (db CouchDatabase) DeleteProject(name string) error {
	db.projectsBucket.Remove(name, 0)
	return nil
}

func (db CouchDatabase) ReadServiceType(sTypeName string) (*proto.ServiceType, error) {
	var sType proto.ServiceType
	db.serviceTypesBucket.Get(sTypeName, &sType)
	return &sType, nil
}

func (db CouchDatabase) WriteServiceType(sType *proto.ServiceType) error {
	_, err := db.serviceTypesBucket.Upsert(sType.Type, sType, 0)
	return err
}

func (db CouchDatabase) ListServicesTypes() ([]proto.ServiceType, error) {
	query := gocb.NewN1qlQuery("SELECT ID, Type, Description, DefaultVersion, Class, AccessPort, Ports, Versions FROM " + serviceTypeBucketName)
	rows, err := db.serviceTypesBucket.ExecuteN1qlQuery(query, []interface{}{})
	if err != nil {
		return nil, err
	}
	var row proto.ServiceType
	var result []proto.ServiceType

	for rows.Next(&row) {
		result = append(result, row)
		row = proto.ServiceType{}
	}
	return result, nil
}

func (db CouchDatabase) DeleteServiceType(name string) error {
	_, err := db.serviceTypesBucket.Remove(name, 0)
	return err
}

func (db CouchDatabase) ReadServiceVersion(sType string, vId string) (*proto.ServiceVersion, error) {
	var st proto.ServiceType
	db.serviceTypesBucket.Get(sType, &st)

	var result *proto.ServiceVersion

	if st.Type == "" {
		return nil, errors.New("Error: service with this type doesn't exist")
	}

	flag := false
	for _, v := range st.Versions {
		if v.ID == vId {
			result = v
			flag = true
			break
		}
	}

	if flag {
		return result, nil
	}
	return nil, errors.New("Error: service version with this ID doesn't exist")
}

func (db CouchDatabase) DeleteServiceVersion(sType string, vId string) (*proto.ServiceVersion, error) {
	var st proto.ServiceType
	db.serviceTypesBucket.Get(sType, &st)

	if st.Type == "" {
		return nil, errors.New("Error: service with this type doesn't exist")
	}

	flag := false
	var idToDelete int
	var result *proto.ServiceVersion
	for i, v := range st.Versions {
		if v.ID == vId {
			idToDelete = i
			result = v
			flag = true
			break
		}
	}

	if !flag {
		return nil, errors.New("Error: service version with this ID doesn't exist")
	}

	st.Versions = st.Versions[:idToDelete+copy(st.Versions[idToDelete:], st.Versions[idToDelete+1:])]
	_, err := db.serviceTypesBucket.Upsert(st.Type, sType, 0)
	return result, err
}

func (db CouchDatabase) UpdateServiceType(st *proto.ServiceType) error {
	var cas gocb.Cas
	_, err := db.serviceTypesBucket.Replace(st.Type, st, cas, 0)
	return err
}

func (db CouchDatabase) ReadServiceVersionByName(sType string, version string) (*proto.ServiceVersion, error) {
	var st proto.ServiceType
	db.serviceTypesBucket.Get(sType, &st)

	var result *proto.ServiceVersion

	if st.Type == "" {
		return nil, errors.New("Error: service with this type doesn't exist")
	}

	flag := false
	for _, v := range st.Versions {
		if v.Version == version {
			result = v
			flag = true
			break
		}
	}

	if flag {
		return result, nil
	}
	return nil, errors.New("Error: service version with this ID doesn't exist")
}

func (db CouchDatabase) ReadImage(imageName string) (*proto.Image, error) {
	q := gocb.NewN1qlQuery(fmt.Sprintf("SELECT ID, Name, AnsibleUser, CloudImageID FROM %v WHERE Name = '%v'", imageBucketName, imageName))
	res, err := db.couchCluster.ExecuteN1qlQuery(q, []interface{}{})
	if err != nil {
		return nil, err
	}
	var img proto.Image
	res.Next(&img)
	return &img, nil
}

func (db CouchDatabase) WriteImage(image *proto.Image) error {
	_, err := db.imageBucket.Upsert(image.ID, image, 0)
	if err != nil {
		return err
	}
	return nil
}

func (db CouchDatabase) UpdateImage(id string, image *proto.Image) error {
	var cas gocb.Cas
	_, err := db.imageBucket.Replace(id, image, cas, 0)
	return err
}

func (db CouchDatabase) DeleteImage(imageName string) error {
	q := gocb.NewN1qlQuery(fmt.Sprintf("SELECT ID FROM %v WHERE Name = '%v'", imageBucketName, imageName))
	res, err := db.couchCluster.ExecuteN1qlQuery(q, []interface{}{})
	if err != nil {
		return err
	}
	var img proto.Image
	res.Next(&img)
	res.Close()
	_, err = db.imageBucket.Remove(img.ID, 0)
	return err
}

func (db CouchDatabase) ListImages() ([]proto.Image, error) {
	query := gocb.NewN1qlQuery("SELECT ID, Name, AnsibleUser, CloudImageID FROM " + imageBucketName)
	rows, err := db.couchCluster.ExecuteN1qlQuery(query, []interface{}{})
	if err != nil {
		return nil, err
	}
	var row proto.Image
	var result []proto.Image

	for rows.Next(&row) {
		result = append(result, row)
		row = proto.Image{}
	}
	rows.Close()

	return result, nil
}
