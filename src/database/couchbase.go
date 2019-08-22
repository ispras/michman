package database

import (
	"errors"
	vaultapi "github.com/hashicorp/vault/api"
	proto "gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/protobuf"
	"gopkg.in/couchbase/gocb.v1"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

const (
	couchPath     string = "couchbase://couchbase_ip"
	couchUsername string = "couchbase_user"
	couchPassword string = "couchbase_password"
	clusterBucketName string = "clusters"
)

type vaultAuth struct {
	Token     string `yaml:"token"`
	VaultAddr string `yaml:"vault_addr"`
	cbKey     string `yaml:"cb_key"`
}

type couchAuth struct {
	Address string `yaml:"cb_address"`
	Username string `yaml:"cb_username"`
	Password string `yaml:"cb_password"`
}

type CouchDatabase struct {
	auth *couchAuth
	couchCluster *gocb.Cluster
	clusterBucket *gocb.Bucket
	clusterBucketName string
}

func (db *CouchDatabase) getCouchCluster() error {
	if db.auth == nil {
		path, err := os.Getwd() //file must be executed from spark-openstack directory
		if err != nil {
			log.Fatalln(err)
		}

		workDir := filepath.Base(path)
		if workDir != "spark-openstack" {
			return errors.New("Error: working directory must be spark-openstack")
		}

		vaultFile, err := ioutil.ReadFile(filepath.Join(workDir, "vault.yaml"))
		if err != nil {
			return errors.New("Error: cannot read vault file")
		}

		var vault vaultAuth
		if err := yaml.Unmarshal(vaultFile, &vault); err != nil {
			return err
		}

		client, err := vaultapi.NewClient(&vaultapi.Config{
			Address: vault.VaultAddr,
		})
		if err != nil {
			return err
		}
		client.SetToken(vault.Token)
		couchSecrets, err := client.Logical().Read(vault.cbKey)
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

	db.clusterBucket = bucket
	db.clusterBucketName = name
	return nil
}

func (db *CouchDatabase) WriteCluster(cluster *proto.Cluster) error {
	if db.clusterBucket == nil {
		if err := db.getBucket(db.clusterBucketName); err != nil {
			return err
		}
	}
	err := db.getBucket(db.clusterBucketName)
	if err != nil {
		return err
	}
	db.clusterBucket.Upsert(cluster.Name, cluster, 0)
	return nil
}

func (db *CouchDatabase) ReadCluster(name string) (*proto.Cluster, error) {
	if db.clusterBucket == nil {
		if err := db.getBucket(db.clusterBucketName); err != nil {
			return nil, err
		}
	}
	err := db.getBucket(db.clusterBucketName)
	if err != nil {
		return nil, err
	}
	var cluster proto.Cluster
	db.clusterBucket.Get(name, &cluster)
	return &cluster, nil
}

func (db *CouchDatabase) ListClusters() ([]proto.Cluster, error) {
	if db.clusterBucket == nil {
		if err := db.getBucket(db.clusterBucketName); err != nil {
			return nil, err
		}
	}
	query := gocb.NewN1qlQuery("SELECT ID, Name, DisplayName, HostURL, ClusterType, NHosts, EntityStatus, services FROM " + db.clusterBucketName)
	rows, err := db.clusterBucket.ExecuteN1qlQuery(query, []interface{}{})
	if err != nil {
		return nil, err
	}
	var row proto.Cluster
	var result []proto.Cluster

	for rows.Next(&row) {
		result = append(result, row)
	}
	return result, nil
}

func (db *CouchDatabase) DeleteCluster(name string) error {
	if db.clusterBucket == nil {
		if err := db.getBucket(db.clusterBucketName); err != nil{
			return err
		}
	}
	db.clusterBucket.Remove(name, 0)
	return nil
}
