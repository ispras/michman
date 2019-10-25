package database

import (
	"errors"
	"fmt"
	proto "gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/protobuf"
	"gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/utils"
	"gopkg.in/couchbase/gocb.v1"
)

const (
	couchPath          string = "couchbase://couchbase_ip"
	couchUsername      string = "couchbase_user"
	couchPassword      string = "couchbase_password"
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
	auth              *couchAuth
	couchCluster      *gocb.Cluster
	currentBucket     *gocb.Bucket
	clusterBucketName string
	templateBucketName string
	projectBucketName string
	VaultCommunicator  utils.SecretStorage
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

func (db *CouchDatabase) getBucket(name string) error {
	if db.couchCluster == nil {
		err := db.getCouchCluster()
		if err != nil {
			return err
		}
	}

	bucket, err := db.couchCluster.OpenBucket(name, "")
	if err != nil {
		return err
	}

	db.currentBucket = bucket
	return nil
}

func (db CouchDatabase) ReadProject(projectID string) (*proto.Project, error) {
	if db.currentBucket == nil {
		if db.projectBucketName == "" {
			db.projectBucketName = projectBucketName
		}
		if err := db.getBucket(db.projectBucketName); err != nil {
			return nil, err
		}
	}

	var project proto.Project
	db.currentBucket.Get(projectID, &project)

	return &project, nil
}

func (db CouchDatabase) ReadProjectByName(projectName string) (*proto.Project, error) {
	if db.currentBucket == nil {
		if db.projectBucketName == "" {
			db.projectBucketName = projectBucketName
		}
		if err := db.getBucket(db.projectBucketName); err != nil {
			return nil, err
		}
	}

	query := gocb.NewN1qlQuery("SELECT ID, Name, DisplayName, GroupID, Description FROM " + db.projectBucketName +
		" WHERE Name = '" + projectName + "'")
	result, err := db.currentBucket.ExecuteN1qlQuery(query, []interface{}{})
	if err != nil {
		return nil, err
	}

	var p proto.Project
	result.Next(&p)

	return &p, nil
}

func (db CouchDatabase) WriteProject(project *proto.Project) error {
	if db.currentBucket == nil {
		if db.projectBucketName == "" {
			db.projectBucketName = projectBucketName
		}
		if err := db.getBucket(db.projectBucketName); err != nil {
			return err
		}
	}

	db.currentBucket.Upsert(project.ID, project, 0)

	return nil
}

func (db CouchDatabase) ListProjects() ([]proto.Project, error) {
	if db.currentBucket == nil {
		if db.projectBucketName == "" {
			db.projectBucketName = projectBucketName
		}
		if err := db.getBucket(db.projectBucketName); err != nil {
			return nil, err
		}
	}

	query := gocb.NewN1qlQuery("SELECT ID, Name, DisplayName, GroupID, Description FROM " + db.projectBucketName)
	rows, err := db.currentBucket.ExecuteN1qlQuery(query, []interface{}{})
	if err != nil {
		return nil, err
	}

	var row proto.Project
	var result []proto.Project

	for rows.Next(&row) {
		result = append(result, row)
		row = proto.Project{}
	}

	return result, nil
}

func (db CouchDatabase) ReadProjectClusters(projectID string) ([]proto.Cluster, error) {
	if db.currentBucket == nil {
		if db.projectBucketName == "" {
			db.projectBucketName = projectBucketName
		}
		if err := db.getBucket(db.projectBucketName); err != nil {
			return nil, err
		}
	}

	q := "SELECT ID, Name, DisplayName, HostURL, ClusterType, NHosts, EntityStatus, services from " + clusterBucketName +
		" where ProjectID = '" + projectID + "'"
	query := gocb.NewN1qlQuery(q)

	rows, err := db.currentBucket.ExecuteN1qlQuery(query, []interface{}{})
	if err != nil {
		return nil, err
	}

	var row proto.Cluster
	var result []proto.Cluster

	for rows.Next(&row) {
		result = append(result, row)
		row = proto.Cluster{}
	}

	return result, nil
}

func (db CouchDatabase) WriteCluster(cluster *proto.Cluster) error {
	if db.currentBucket == nil {
		if db.clusterBucketName == "" {
			db.clusterBucketName = clusterBucketName
		}
		if err := db.getBucket(db.clusterBucketName); err != nil {
			return err
		}
	}

	err := db.getBucket(db.clusterBucketName)
	if err != nil {
		return err
	}

	db.currentBucket.Upsert(cluster.ID, cluster, 0)
	return nil
}

func (db CouchDatabase) ReadCluster(clusterID string) (*proto.Cluster, error) {
	if db.currentBucket == nil {
		if db.clusterBucketName == "" {
			db.clusterBucketName = clusterBucketName
		}
		if err := db.getBucket(db.clusterBucketName); err != nil {
			return nil, err
		}
	}

	err := db.getBucket(db.clusterBucketName)
	if err != nil {
		return nil, err
	}

	var cluster proto.Cluster
	db.currentBucket.Get(clusterID, &cluster)
	return &cluster, nil
}

func (db CouchDatabase) ReadClusterByName(projectID, clusterName string) (*proto.Cluster, error) {
	if db.currentBucket == nil {
		if db.projectBucketName == "" {
			db.projectBucketName = projectBucketName
		}
		if err := db.getBucket(db.projectBucketName); err != nil {
			return nil, err
		}
	}

	q := "SELECT ID, Name, DisplayName, HostURL, EntityStatus, ClusterType," +
		"services, NHosts, masterIP, ProjectID FROM " + clusterBucketName +
		" WHERE ProjectID = '" + projectID + "' and Name = '" + clusterName + "'"
	//println(q)
	query := gocb.NewN1qlQuery(q)
	result, err := db.currentBucket.ExecuteN1qlQuery(query, []interface{}{})
	if err != nil {
		return nil, err
	}

	var c proto.Cluster
	result.Next(&c)

	return &c, nil
}

func (db CouchDatabase) UpdateCluster(cluster *proto.Cluster) error {
	if db.currentBucket == nil {
		if db.clusterBucketName == "" {
			db.clusterBucketName = clusterBucketName
		}
		if err := db.getBucket(db.clusterBucketName); err != nil {
			return err
		}
	}

	var cas gocb.Cas
	cas, err := db.currentBucket.Replace(cluster.ID, cluster, cas, 0)

	return err
}

func (db CouchDatabase) DeleteCluster(clusterID string) error {
	if db.currentBucket == nil {
		if db.clusterBucketName == "" {
			db.clusterBucketName = clusterBucketName
		}
		if err := db.getBucket(db.clusterBucketName); err != nil {
			return err
		}
	}

	err := db.getBucket(db.clusterBucketName)
	if err != nil {
		return err
	}

	_, err = db.currentBucket.Remove(clusterID, 0)
	println(clusterID)
	if err != nil {
		return err
	}

	return nil
}

func (db CouchDatabase) WriteTemplate(template *proto.Template) error {
	if db.currentBucket == nil {
		if db.templateBucketName == "" {
			db.templateBucketName = templateBucketName
		}
		if err := db.getBucket(db.templateBucketName); err != nil {
			return err
		}
	}
	err := db.getBucket(db.templateBucketName)
	if err != nil {
		return err
	}
	_, err = db.currentBucket.Upsert(template.ID, template, 0)
	if err != nil {
		return err
	}
	return nil
}

func (db CouchDatabase) ReadTemplate(projectID, id string) (*proto.Template, error) {
	if db.currentBucket == nil {
		if db.templateBucketName == "" {
			db.templateBucketName = templateBucketName
		}
		if err := db.getBucket(db.templateBucketName); err != nil {
			return nil, err
		}
	}

	err := db.getBucket(db.templateBucketName)
	if err != nil {
		return nil, err
	}
	var template proto.Template
	_, err = db.currentBucket.Get(id, &template)
	if err != nil {
		return &proto.Template{}, nil
	}
	if projectID != template.ProjectID {
		return &proto.Template{}, nil
	}
	return &template, nil
}

func (db CouchDatabase) ReadTemplateByName(templateName string) (*proto.Template, error) {
	if db.currentBucket == nil {
		if db.templateBucketName == "" {
			db.templateBucketName = templateBucketName
		}
		if err := db.getBucket(db.templateBucketName); err != nil {
			return nil, err
		}
	}

	err := db.getBucket(db.templateBucketName)
	if err != nil {
		return nil, err
	}

	query := gocb.NewN1qlQuery(fmt.Sprintf("SELECT ID, ProjectID, Name, DisplayName, Services,"+
		" NHosts, Description FROM %v WHERE Name = '%v'",
		db.templateBucketName, templateName))
	rows, err := db.currentBucket.ExecuteN1qlQuery(query, []interface{}{})
	if err != nil {
		return nil, err
	}
	var template proto.Template

	if hasResult := rows.Next(template); !hasResult {
		template = proto.Template{}
	}

	return &template, nil
}

func (db CouchDatabase) ListTemplates(projectID string) ([]proto.Template, error) {
	if db.currentBucket == nil {
		if db.templateBucketName == "" {
			db.templateBucketName = templateBucketName
		}
		if err := db.getBucket(db.templateBucketName); err != nil {
			return nil, err
		}
	}

	query := gocb.NewN1qlQuery(fmt.Sprintf("SELECT ID, ProjectID, Name, DisplayName, Services,"+
		" NHosts, Description FROM %v WHERE ProjectID = '%v'",
		db.templateBucketName, projectID))
	rows, err := db.currentBucket.ExecuteN1qlQuery(query, []interface{}{})
	if err != nil {
		return nil, err
	}
	var row proto.Template
	var result []proto.Template

	for rows.Next(&row) {
		result = append(result, row)
	}
	return result, nil
}

func (db CouchDatabase) DeleteTemplate(id string) error {
	if db.currentBucket == nil {
		if db.templateBucketName == "" {
			db.templateBucketName = templateBucketName
		}
		if err := db.getBucket(db.templateBucketName); err != nil {
			return err
		}
	}
	_, err := db.currentBucket.Remove(id, 0)
	if err != nil {
		return err
	}
	return nil
}

func (db CouchDatabase) UpdateProject(project *proto.Project) error {
	if db.currentBucket == nil {
		if db.projectBucketName == "" {
			db.projectBucketName = projectBucketName
		}
		if err := db.getBucket(db.projectBucketName); err != nil {
			return err
		}
	}

	var cas gocb.Cas
	cas, err := db.currentBucket.Replace(project.ID, project, cas, 0)

	return err
}

func (db CouchDatabase) DeleteProject(name string) error {
	if db.currentBucket == nil {
		if db.projectBucketName == "" {
			db.projectBucketName = projectBucketName
		}
		if err := db.getBucket(db.projectBucketName); err != nil {
			return err
		}
	}
	db.currentBucket.Remove(name, 0)
	return nil
}
