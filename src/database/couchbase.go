package database

import (
	"errors"
	proto "gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/protobuf"
	"gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/utils"
	"gopkg.in/couchbase/gocb.v1"
)

const (
	couchPath         string = "couchbase://couchbase_ip"
	couchUsername     string = "couchbase_user"
	couchPassword     string = "couchbase_password"
	clusterBucketName string = "clusters"
	projectBucketName string = "projects"
	serviceTypeBucketName string = "service_types"

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
	clusterBucket     *gocb.Bucket
	clusterBucketName string
	projectBucketName string
	serviceTypeBucketName string
	VaultCommunicator utils.SecretStorage
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
	db.currentBucket.Upsert(cluster.Name, cluster, 0)
	return nil
}

func (db CouchDatabase) ReadCluster(name string) (*proto.Cluster, error) {
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
	db.currentBucket.Get(name, &cluster)
	return &cluster, nil
}

func (db CouchDatabase) ListClusters() ([]proto.Cluster, error) {
	if db.currentBucket == nil {
		if db.clusterBucketName == "" {
			db.clusterBucketName = clusterBucketName
		}
		if err := db.getBucket(db.clusterBucketName); err != nil {
			return nil, err
		}
	}

	query := gocb.NewN1qlQuery("SELECT ID, Name, DisplayName, HostURL, ClusterType, NHosts, EntityStatus, Services, MasterIP FROM " + db.clusterBucketName)
	rows, err := db.clusterBucket.ExecuteN1qlQuery(query, []interface{}{})
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

func (db CouchDatabase) DeleteCluster(name string) error {
	if db.currentBucket == nil {
		if db.clusterBucketName == "" {
			db.clusterBucketName = clusterBucketName
		}
		if err := db.getBucket(db.clusterBucketName); err != nil {
			return err
		}
	}
	db.currentBucket.Remove(name, 0)
	return nil
}

func (db CouchDatabase) ListProjects() ([]proto.Project, error) {
	if db.currentBucket == nil {
		if db.clusterBucketName == "" {
			db.clusterBucketName = clusterBucketName
		}
		if err := db.getBucket(db.clusterBucketName); err != nil {
			return nil, err
		}
	}

	query := gocb.NewN1qlQuery("SELECT ID, Name, DisplayName, GroupID, Description FROM " + db.clusterBucketName)
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

func (db CouchDatabase) ReadProject(projectName string) (*proto.Project, error) {
	if db.currentBucket == nil {
		if db.projectBucketName == "" {
			db.projectBucketName = projectBucketName
		}
		if err := db.getBucket(db.projectBucketName); err != nil {
			return nil, err
		}
	}

	var project proto.Project
	db.currentBucket.Get(projectName, &project)

	return &project, nil
}


func (db CouchDatabase) WriteProject(project *proto.Project) error {
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
	db.currentBucket.Upsert(project.Name, project, 0)
	return nil
}

func (db CouchDatabase) ReadServiceType(sTypeName string) (*proto.ServiceType, error) {
	if db.currentBucket == nil {
		if db.serviceTypeBucketName == "" {
			db.serviceTypeBucketName = serviceTypeBucketName
		}
		if err := db.getBucket(db.serviceTypeBucketName); err != nil {
			return nil, err
		}
	}

	err := db.getBucket(db.serviceTypeBucketName)
	if err != nil {
		return nil, err
	}
	var sType proto.ServiceType
	db.currentBucket.Get(sTypeName, &sType)
	return &sType, nil
}

func (db CouchDatabase) WriteServiceType(sType *proto.ServiceType) error {
	if db.currentBucket == nil {
		if db.serviceTypeBucketName == "" {
			db.serviceTypeBucketName = serviceTypeBucketName
		}
		if err := db.getBucket(db.serviceTypeBucketName); err != nil {
			return err
		}
	}
	err := db.getBucket(db.serviceTypeBucketName)
	if err != nil {
		return err
	}
	db.currentBucket.Upsert(sType.Type, sType, 0)
	return nil
}

func (db CouchDatabase) ListServicesTypes() ([]proto.ServiceType, error) {
	if db.currentBucket == nil {
		if db.serviceTypeBucketName == "" {
			db.serviceTypeBucketName = serviceTypeBucketName
		}
		if err := db.getBucket(db.serviceTypeBucketName); err != nil {
			return nil, err
		}
	}

	query := gocb.NewN1qlQuery("SELECT ID, Type, Description FROM " + db.serviceTypeBucketName)
	rows, err := db.currentBucket.ExecuteN1qlQuery(query, []interface{}{})
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
	if db.currentBucket == nil {
		if db.serviceTypeBucketName == "" {
			db.serviceTypeBucketName = serviceTypeBucketName
		}
		if err := db.getBucket(db.serviceTypeBucketName); err != nil {
			return err
		}
	}
	db.currentBucket.Remove(name, 0)
	return nil
}