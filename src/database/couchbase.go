package database

import (
	"errors"
	"fmt"
	proto "gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/protobuf"
	"gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/utils"
	"gopkg.in/couchbase/gocb.v1"
)

const (
	clusterBucketName  string = "clusters"
	templateBucketName string = "templates"
	projectBucketName  string = "projects"
)

type vaultAuth struct {
	Token     string `yaml:"token"`
	VaultAddr string `yaml:"vault_addr"`
	cbKey     string `yaml:"cb_key"`
}

type couchAuth struct {
	Address  string `yaml:"cb_address"`
	Username string `yaml:"cb_username"`
	Password string `yaml:"cb_password"`
}

type CouchDatabase struct {
	auth               *couchAuth
	couchCluster       *gocb.Cluster
	clustersBucket     *gocb.Bucket
	projectsBucket     *gocb.Bucket
	templatesBucket    *gocb.Bucket
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

	cb.auth = &couchAuth{
		Address:  couchSecrets.Data["path"].(string),
		Username: couchSecrets.Data["username"].(string),
		Password: couchSecrets.Data["password"].(string),
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

		db.auth = &couchAuth{
			Address:  couchSecrets.Data["path"].(string),
			Username: couchSecrets.Data["username"].(string),
			Password: couchSecrets.Data["password"].(string),
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
	query := gocb.NewN1qlQuery("SELECT ID, Name, DisplayName, GroupID, Description FROM " + projectBucketName +
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
	query := gocb.NewN1qlQuery("SELECT ID, Name, DisplayName, GroupID, Description FROM " + projectBucketName)
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

func (db CouchDatabase) ReadProjectClusters(projectID string) ([]proto.Cluster, error) {q := "SELECT ID, Name, DisplayName, HostURL, ClusterType, NHosts, EntityStatus, Services, MasterIP, Description from " + clusterBucketName +
		" where ProjectID = '" + projectID + "'"
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
	q := "SELECT ID, Name, DisplayName, HostURL, EntityStatus, ClusterType," +
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
