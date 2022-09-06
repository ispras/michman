package database

import (
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/utils"
)

type MySqlDatabase struct {
	connection        *sql.DB
	VaultCommunicator utils.SecretStorage
}

type MySqlCredentials struct {
	Address  string
	User     string
	Password string
	Database string
}

func NewMySQL(vaultCom utils.SecretStorage) (Database, error) {
	db := new(MySqlDatabase)
	db.VaultCommunicator = vaultCom
	client, vaultCfg, err := db.VaultCommunicator.ConnectVault()
	if client == nil || err != nil {
		return nil, err
	}

	mySqlSecrets, err := client.Logical().Read(vaultCfg.MySqlKey)
	if err != nil {
		return nil, ErrMySQLSecretsRead
	}

	creds := MySqlCredentials{
		Address:  mySqlSecrets.Data[utils.MySqlAddress].(string),
		User:     mySqlSecrets.Data[utils.MySqlUser].(string),
		Password: mySqlSecrets.Data[utils.MySqlPassword].(string),
		Database: mySqlSecrets.Data[utils.MySqlDatabase].(string),
	}

	connection, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s", creds.User, creds.Password, creds.Address, creds.Database))
	if err != nil {
		return nil, ErrMySQLConnection
	}
	if err := connection.Ping(); err != nil {
		return nil, ErrMySQLPing
	}
	db.connection = connection
	return db, nil
}

func (db MySqlDatabase) ReadCluster(projectIdOrName string, clusterIdOrName string) (*protobuf.Cluster, error) {
	isUuid := utils.IsUuid(clusterIdOrName)
	var cluster *protobuf.Cluster
	var err error
	if isUuid {
		cluster, err = readClusterbyId(db, clusterIdOrName)
	} else {
		cluster, err = readClusterbyName(db, projectIdOrName, clusterIdOrName)
	}
	return cluster, err
}

func readClusterbyId(db MySqlDatabase, clusterIdOrName string) (*protobuf.Cluster, error) {
	//read cluster by Id
	q := `SELECT ID, Name, DisplayName, HostURL, EntityStatus, ClusterType, NHosts, MasterIP, ProjectID, Description, Image, Monitoring, MasterFlavor, SlavesFlavor, StorageFlavor, Monitoring, SSH_Keys FROM cluster WHERE ID = ?`

	c := protobuf.Cluster{ID: "", Name: "", DisplayName: ""}
	var ssh_keys []byte
	res := db.connection.QueryRow(q, clusterIdOrName)
	if err := res.Scan(&c.ID, &c.Name, &c.DisplayName, &c.HostURL, &c.EntityStatus, &c.ClusterType,
		&c.NHosts, &c.MasterIP, &c.ProjectID, &c.Description, &c.Image, &c.Monitoring,
		&c.MasterFlavor, &c.SlavesFlavor, &c.StorageFlavor, &c.MonitoringFlavor, &ssh_keys); err != nil {
		if err == sql.ErrNoRows {
			return &c, nil
		}
		return nil, ErrReadObjectByKey
	}

	if len(ssh_keys) > 0 {
		err := json.Unmarshal(ssh_keys, &c.Keys)
		if err != nil {
			return nil, ErrUnmarshalJson
		}
	}
	//get service for cluster
	sq := `SELECT ID, Name, Type, ClusterRef, COALESCE(Config,''), DisplayName, 
		COALESCE(EntityStatus,''),  Version, COALESCE(URL, ''),  
		COALESCE(Description, '')  FROM service WHERE ClusterRef = ?`
	srows, err := db.connection.Query(sq, c.ID)
	if err != nil {
		return nil, ErrReadObjectByKey
	}

	var ss []*protobuf.Service
	for srows.Next() {
		var s protobuf.Service
		var config string
		if err := srows.Scan(&s.ID, &s.Name, &s.Type, &s.ClusterRef, &config, &s.DisplayName,
			&s.EntityStatus, &s.Version, &s.URL, &s.Description); err != nil {
			return nil, ErrReadObjectByKey
		}
		err = json.Unmarshal([]byte(config), &s.Config)
		if err != nil {
			return nil, ErrUnmarshalJson
		}
		//add service to array
		ss = append(ss, &s)
	}
	if err := srows.Err(); err != nil {
		return nil, ErrReadObjectByKey
	}
	srows.Close()

	//add srvice array to cluster structure
	c.Services = ss

	return &c, nil
}

func readClusterbyName(db MySqlDatabase, projectID, clusterIdOrName string) (*protobuf.Cluster, error) {
	//read cluster by name
	q := `SELECT ID, Name, DisplayName, HostURL, EntityStatus, ClusterType, NHosts, MasterIP, ProjectID, Description, Image, Monitoring, MasterFlavor, SlavesFlavor, StorageFlavor, Monitoring, SSH_Keys FROM cluster WHERE Name = ?`

	c := protobuf.Cluster{ID: "", Name: "", DisplayName: ""}
	var ssh_keys []byte
	res := db.connection.QueryRow(q, clusterIdOrName)
	if err := res.Scan(&c.ID, &c.Name, &c.DisplayName, &c.HostURL, &c.EntityStatus, &c.ClusterType,
		&c.NHosts, &c.MasterIP, &c.ProjectID, &c.Description, &c.Image, &c.Monitoring,
		&c.MasterFlavor, &c.SlavesFlavor, &c.StorageFlavor, &c.MonitoringFlavor, &ssh_keys); err != nil {
		if err == sql.ErrNoRows {
			return &c, nil
		}
		return nil, ErrReadObjectByKey
	}

	if len(ssh_keys) > 0 {
		err := json.Unmarshal(ssh_keys, &c.Keys)
		if err != nil {
			return nil, ErrUnmarshalJson
		}
	}

	//get service for cluster
	sq := `SELECT ID, Name, Type, ClusterRef, COALESCE(Config,''), DisplayName, 
		COALESCE(EntityStatus,''),  Version, COALESCE(URL, ''),  
		COALESCE(Description, '')  FROM service WHERE ClusterRef = ?`
	srows, err := db.connection.Query(sq, c.ID)
	if err != nil {
		return nil, ErrReadObjectByKey
	}

	var ss []*protobuf.Service
	for srows.Next() {
		var s protobuf.Service
		var config string
		if err := srows.Scan(&s.ID, &s.Name, &s.Type, &s.ClusterRef, &config, &s.DisplayName,
			&s.EntityStatus, &s.Version, &s.URL, &s.Description); err != nil {
			return nil, ErrReadObjectByKey
		}
		err = json.Unmarshal([]byte(config), &s.Config)
		if err != nil {
			return nil, ErrUnmarshalJson
		}
		//add service to array
		ss = append(ss, &s)
	}
	if err := srows.Err(); err != nil {
		return nil, ErrReadObjectByKey
	}
	srows.Close()

	//add srvice array to cluster structure
	c.Services = ss

	return &c, nil
}

func (db MySqlDatabase) WriteCluster(cluster *protobuf.Cluster) error {
	tx, err := db.connection.Begin()
	if err != nil {
		return ErrStartQueryConnection
	}

	//rollback in case of error
	defer tx.Rollback()
	q := "INSERT INTO cluster (ID, Name, DisplayName, HostURL, EntityStatus, ClusterType, NHosts, MasterIP, ProjectID, Description, Image, Monitoring, MasterFlavor, SlavesFlavor, StorageFlavor, MonitoringFlavor, SSH_Keys) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"

	ssh_keys, err := json.Marshal(cluster.Keys)
	if err != nil {
		return ErrWriteObjectByKey
	}

	_, err = tx.Exec(q, cluster.ID, cluster.Name, cluster.DisplayName, cluster.HostURL, cluster.EntityStatus, cluster.ClusterType, cluster.NHosts, cluster.MasterIP, cluster.ProjectID, cluster.Description, cluster.Image,
		cluster.Monitoring, cluster.MasterFlavor, cluster.SlavesFlavor, cluster.StorageFlavor, cluster.MonitoringFlavor, ssh_keys)
	if err != nil {
		return err
	}
	for _, s := range cluster.Services {
		sq := "INSERT INTO service (ID, Name, Type, ClusterRef, Config, DisplayName, EntityStatus,  Version, URL, Description) VALUES (?,?,?,?,?,?,?,?,?,?)"

		sConfig, err := json.Marshal(s.Config)
		if err != nil {
			return ErrWriteObjectByKey
		}

		_, err = tx.Exec(sq, s.ID, s.Name, s.Type, cluster.ID, string(sConfig), s.DisplayName, s.EntityStatus, s.Version, s.URL, s.Description)
		if err != nil {
			return ErrWriteObjectByKey
		}
	}

	if err = tx.Commit(); err != nil {
		return ErrWriteObjectByKey
	}

	return nil
}

func (db MySqlDatabase) DeleteCluster(projectIdOrName string, clusterIdOrName string) error {
	isUuid := utils.IsUuid(clusterIdOrName)
	var err error
	if isUuid {
		err = deleteClusterbyId(db, clusterIdOrName)
	} else {
		err = deleteClusterbyName(db, projectIdOrName, clusterIdOrName)
	}
	return err
}

func deleteClusterbyId(db MySqlDatabase, clusterIdOrName string) error {
	q := "DELETE FROM cluster WHERE ID = ?"

	_, err := db.connection.Exec(q, clusterIdOrName)
	if err != nil {
		return ErrDeleteObjectByKey
	}

	return nil
}

func deleteClusterbyName(db MySqlDatabase, projectIdOrName string, clusterIdOrName string) error {
	q := "DELETE FROM cluster WHERE Name = ?"

	_, err := db.connection.Exec(q, clusterIdOrName)
	if err != nil {
		return ErrDeleteObjectByKey
	}

	return nil

}

func (db MySqlDatabase) UpdateCluster(cluster *protobuf.Cluster) error {
	tx, err := db.connection.Begin()
	if err != nil {
		return ErrStartQueryConnection
	}

	//rollback in case of error
	defer tx.Rollback()
	for _, s := range cluster.Services { //replace because there might be new services for cluster
		sq := "SELECT Name FROM service WHERE ID = ?"
		res := db.connection.QueryRow(sq, s.ID)
		var sId string
		if err := res.Scan(&sId); err != nil {
			if err == sql.ErrNoRows {
				scq := "INSERT INTO service (ID, Name, Type, ClusterRef, Config, DisplayName, EntityStatus,  Version, URL, Description) VALUES (?,?,?,?,?,?,?,?,?,?)"

				sId, err := uuid.NewRandom()
				if err != nil {
					return ErrNewUuid
				}

				sConfig, err := json.Marshal(s.Config)
				if err != nil {
					return ErrUnmarshalJson
				}

				_, err = tx.Exec(scq, sId.String(), s.Name, s.Type, s.ClusterRef, string(sConfig), s.DisplayName, s.EntityStatus, s.Version, s.URL, s.Description)
				if err != nil {
					return ErrUpdateObjectByKey
				}
			} else {
				suq := "UPDATE service SET Name = ?, Type = ?, ClusterRef = ?, DisplayName = ?, EntityStatus = ?, Version = ?, URL = ?, Description = ? WHERE ID = ?"
				_, err = tx.Exec(suq, s.Name, s.Type, s.ClusterRef, s.DisplayName, s.EntityStatus, s.Version, s.URL, s.Description, s.ID)
				if err != nil {
					return ErrUpdateObjectByKey
				}
			}
		}
	}

	q := "UPDATE cluster SET Name = ?, DisplayName = ?, HostURL = ?, EntityStatus = ?, ClusterType = ?, NHosts = ?, Description = ?,  Image = ?, MasterFlavor = ?, SlavesFlavor = ?, StorageFlavor = ?, SSH_Keys = ? WHERE ID = ?"

	ssh_keys, err := json.Marshal(cluster.Keys)
	if err != nil {
		return ErrUnmarshalJson
	}

	_, err = tx.Exec(q, cluster.Name, cluster.DisplayName, cluster.HostURL, cluster.EntityStatus, cluster.ClusterType, cluster.NHosts, cluster.Description, cluster.Image, cluster.MasterFlavor, cluster.SlavesFlavor,
		cluster.StorageFlavor, ssh_keys, cluster.ID)
	if err != nil {
		return ErrUpdateObjectByKey
	}

	if err = tx.Commit(); err != nil {
		return ErrUpdateObjectByKey
	}

	return nil
}

func (db MySqlDatabase) ReadClustersList() ([]protobuf.Cluster, error) {
	//make a query to select all clusters
	q := `SELECT ID, Name, DisplayName, HostURL, EntityStatus, ClusterType,
	NHosts, MasterIP, ProjectID, Description, Image, Monitoring,
	MasterFlavor, SlavesFlavor, StorageFlavor, MonitoringFlavor, SSH_Keys FROM cluster`

	rows, err := db.connection.Query(q)
	if err != nil {
		return nil, ErrReadObjectByKey
	}
	defer rows.Close()

	var result []protobuf.Cluster
	for rows.Next() {
		var c protobuf.Cluster
		var ssh_keys []byte
		//select one cluster
		if err := rows.Scan(&c.ID, &c.Name, &c.DisplayName, &c.HostURL, &c.EntityStatus, &c.ClusterType,
			&c.NHosts, &c.MasterIP, &c.ProjectID, &c.Description, &c.Image, &c.Monitoring,
			&c.MasterFlavor, &c.SlavesFlavor, &c.StorageFlavor, &c.MonitoringFlavor, &ssh_keys); err != nil {
			return nil, ErrReadObjectList
		}

		if len(ssh_keys) > 0 {
			err = json.Unmarshal(ssh_keys, &c.Keys)
			if err != nil {
				return nil, ErrUnmarshalJson
			}
		}

		//select list of services for particular cluster
		sq := `SELECT ID, Name, Type, ClusterRef, COALESCE(Config,''), DisplayName, 
			COALESCE(EntityStatus,''),  Version, COALESCE(URL, ''),  
			COALESCE(Description, '')  FROM service WHERE ClusterRef = ?`
		srows, err := db.connection.Query(sq, c.ID)
		if err != nil {
			return nil, ErrReadObjectByKey
		}
		//make list of services for particular cluster
		var ss []*protobuf.Service
		for srows.Next() {
			var s protobuf.Service
			var config string
			//select one cluster
			if err := srows.Scan(&s.ID, &s.Name, &s.Type, &s.ClusterRef, &config, &s.DisplayName,
				&s.EntityStatus, &s.Version, &s.URL, &s.Description); err != nil {
				return nil, ErrReadObjectByKey
			}
			err = json.Unmarshal([]byte(config), &s.Config)
			if err != nil {
				return nil, ErrUnmarshalJson
			}
			//add particular cluster to array
			ss = append(ss, &s)
		}
		if err := srows.Err(); err != nil {
			return nil, ErrReadObjectByKey
		}
		srows.Close()

		//add service array to cluster structure
		c.Services = ss

		//add particular cluster to cluster array
		result = append(result, c)
	}
	if err := rows.Err(); err != nil {
		return nil, ErrReadObjectByKey
	}
	return result, nil
}

func (db MySqlDatabase) ReadProject(projectIdOrName string) (*protobuf.Project, error) {
	isUuid := utils.IsUuid(projectIdOrName)
	var project *protobuf.Project
	var err error
	if isUuid {
		project, err = readProjectbyId(db, projectIdOrName)
	} else {
		project, err = readProjectbyName(db, projectIdOrName)
	}
	return project, err
}

func readProjectbyId(db MySqlDatabase, projectIdOrName string) (*protobuf.Project, error) {
	q := `SELECT ID, Name, DisplayName, COALESCE(GroupID, ''), 
			DefaultImage, COALESCE(Description, ''), DefaultMasterFlavor, DefaultSlavesFlavor,
			DefaultStorageFlavor, DefaultMonitoringFlavor FROM project WHERE ID = ?`

	pr := protobuf.Project{ID: "", Name: "", DisplayName: ""}
	res := db.connection.QueryRow(q, projectIdOrName)
	if err := res.Scan(&pr.ID, &pr.Name, &pr.DisplayName, &pr.GroupID, &pr.Description, &pr.DefaultImage, &pr.DefaultMasterFlavor, &pr.DefaultSlavesFlavor, &pr.DefaultStorageFlavor, &pr.DefaultMonitoringFlavor); err != nil {
		if err == sql.ErrNoRows {
			return &pr, nil
		}
		return nil, ErrReadObjectByKey
	}

	return &pr, nil
}

func readProjectbyName(db MySqlDatabase, projectName string) (*protobuf.Project, error) {
	q := `SELECT ID, Name, DisplayName, COALESCE(GroupID, ''), COALESCE(Description, ''), 
			DefaultImage, DefaultMasterFlavor, DefaultSlavesFlavor,
			DefaultStorageFlavor, DefaultMonitoringFlavor FROM project WHERE Name = ?`

	pr := protobuf.Project{ID: "", Name: "", DisplayName: ""}
	res := db.connection.QueryRow(q, projectName)
	if err := res.Scan(&pr.ID, &pr.Name, &pr.DisplayName, &pr.GroupID, &pr.Description, &pr.DefaultImage, &pr.DefaultMasterFlavor, &pr.DefaultSlavesFlavor, &pr.DefaultStorageFlavor, &pr.DefaultMonitoringFlavor); err != nil {
		if err == sql.ErrNoRows {
			return &pr, nil
		}
		return nil, ErrReadObjectByKey
	}

	return &pr, nil
}

func (db MySqlDatabase) ReadProjectsList() ([]protobuf.Project, error) {
	q := `SELECT ID, Name, DisplayName, COALESCE(GroupID, ''), COALESCE(Description, ''), 
	DefaultImage, DefaultMasterFlavor, DefaultSlavesFlavor,
	DefaultStorageFlavor, DefaultMonitoringFlavor  FROM project`
	rows, err := db.connection.Query(q)
	if err != nil {
		return nil, ErrReadObjectList
	}
	defer rows.Close()

	var result []protobuf.Project
	for rows.Next() {
		var row protobuf.Project
		if err := rows.Scan(&row.ID, &row.Name, &row.DisplayName, &row.GroupID, &row.Description, &row.DefaultImage, &row.DefaultMasterFlavor, &row.DefaultSlavesFlavor, &row.DefaultStorageFlavor, &row.DefaultMonitoringFlavor); err != nil {
			return nil, ErrReadObjectByKey
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, ErrReadObjectList
	}
	return result, nil
}

func (db MySqlDatabase) ReadProjectClusters(projectID string) ([]protobuf.Cluster, error) {
	q := `SELECT ID, Name, DisplayName, HostURL, EntityStatus, ClusterType, NHosts, MasterIP, Description, ProjectID, Image, Monitoring, 
			MasterFlavor, SlavesFlavor, StorageFlavor, MonitoringFlavor, SSH_Keys FROM cluster WHERE ProjectID = ?`

	rows, err := db.connection.Query(q, projectID)
	if err != nil {
		return nil, ErrReadObjectList
	}
	defer rows.Close()

	var result []protobuf.Cluster
	for rows.Next() {
		var c protobuf.Cluster
		var ssh_keys []byte
		if err := rows.Scan(&c.ID, &c.Name, &c.DisplayName, &c.HostURL, &c.EntityStatus, &c.ClusterType, &c.NHosts, &c.MasterIP,
			&c.Description, &c.ProjectID, &c.Image, &c.Monitoring, &c.MasterFlavor, &c.SlavesFlavor, &c.StorageFlavor, &c.MonitoringFlavor, &ssh_keys); err != nil {
			return nil, ErrReadObjectByKey
		}

		if len(ssh_keys) > 0 {
			err = json.Unmarshal(ssh_keys, &c.Keys)
			if err != nil {
				return nil, ErrUnmarshalJson
			}
		}

		sq := `SELECT ID, Name, Type, COALESCE(Config,''), DisplayName, COALESCE(EntityStatus,''), Version, 
				COALESCE(URL, ''), COALESCE(Description, '') FROM service WHERE ClusterRef = ?`
		srows, err := db.connection.Query(sq, c.ID)
		if err != nil {
			return nil, ErrReadObjectByKey
		}

		var ss []*protobuf.Service
		for srows.Next() {
			var s protobuf.Service
			var config string
			if err := srows.Scan(&s.ID, &s.Name, &s.Type, &config, &s.DisplayName, &s.EntityStatus, &s.Version,
				&s.URL, &s.Description); err != nil {
				return nil, ErrReadObjectByKey
			}
			err = json.Unmarshal([]byte(config), &s.Config)
			if err != nil {
				return nil, ErrUnmarshalJson
			}
			ss = append(ss, &s)
		}
		if err := srows.Err(); err != nil {
			return nil, ErrReadObjectByKey
		}
		srows.Close()

		c.Services = ss

		result = append(result, c)
	}
	if err := rows.Err(); err != nil {
		return nil, ErrReadObjectByKey
	}
	return result, nil
}

func (db MySqlDatabase) WriteProject(project *protobuf.Project) error {
	q := "INSERT INTO project (ID, Name, DisplayName, GroupID, Description, DefaultImage, DefaultMasterFlavor, DefaultSlavesFlavor, DefaultStorageFlavor, DefaultMonitoringFlavor) VALUES (?,?,?,?,?,?,?,?,?,?)"

	_, err := db.connection.Exec(q, project.ID, project.Name, project.DisplayName, project.GroupID, project.Description, project.DefaultImage, project.DefaultMasterFlavor, project.DefaultSlavesFlavor, project.DefaultStorageFlavor, project.DefaultMonitoringFlavor)
	if err != nil {
		return ErrWriteObjectByKey
	}
	return nil
}

func (db MySqlDatabase) UpdateProject(project *protobuf.Project) error {
	q := "UPDATE project SET Name = ?, DisplayName = ?,  GroupID = ?, Description = ?, DefaultImage = ?, DefaultMasterFlavor = ?, DefaultSlavesFlavor = ?, DefaultStorageFlavor = ?, DefaultMonitoringFlavor = ? WHERE ID = ?"
	_, err := db.connection.Exec(q, project.Name, project.DisplayName, project.GroupID, project.Description, project.DefaultImage,
		project.DefaultMasterFlavor, project.DefaultSlavesFlavor, project.DefaultStorageFlavor, project.DefaultMonitoringFlavor, project.ID)
	if err != nil {
		return ErrUpdateObjectByKey
	}
	return nil
}

func (db MySqlDatabase) DeleteProject(projectIdOrName string) error {
	isUuid := utils.IsUuid(projectIdOrName)
	var err error
	if isUuid {
		err = deleteProjectbyId(db, projectIdOrName)
	} else {
		err = deleteProjectbyName(db, projectIdOrName)
	}
	return err
}

func deleteProjectbyName(db MySqlDatabase, projectIdOrName string) error {
	q := "DELETE FROM project WHERE Name = ?;"
	_, err := db.connection.Exec(q, projectIdOrName)
	if err != nil {
		return ErrDeleteObjectByKey
	}
	return nil
}

func deleteProjectbyId(db MySqlDatabase, projectIdOrName string) error {
	q := "DELETE FROM project WHERE ID = ?;"
	_, err := db.connection.Exec(q, projectIdOrName)
	if err != nil {
		return ErrDeleteObjectByKey
	}
	return nil
}

func (db MySqlDatabase) ReadTemplate(TemplateId string) (*protobuf.Template, error) {
	//todo: get services
	q := "SELECT ID, ProjectID, Name, DisplayName, NHosts, Description FROM template WHERE ID = ?"
	template := protobuf.Template{ID: "", ProjectID: "", Name: ""}
	res := db.connection.QueryRow(q, TemplateId)
	if err := res.Scan(&template.ID, &template.ProjectID, &template.Name, &template.DisplayName, &template.NHosts, &template.Description); err != nil {
		if err == sql.ErrNoRows {
			return &template, nil
		}
		return nil, ErrReadObjectByKey
	}
	return &template, nil
}

func (db MySqlDatabase) ReadTemplateByName(TemplateName string) (*protobuf.Template, error) {
	//todo: get services
	q := "SELECT ID, ProjectID, Name, DisplayName, NHosts, Description FROM template WHERE Name = ?"
	template := protobuf.Template{ID: "", ProjectID: "", Name: ""}
	res := db.connection.QueryRow(q, TemplateName)
	if err := res.Scan(&template.ID, &template.ProjectID, &template.Name, &template.DisplayName, &template.NHosts, &template.Description); err != nil {
		if err == sql.ErrNoRows {
			return &template, nil
		}
		return nil, ErrReadObjectByKey
	}
	return &template, nil
}

func (db MySqlDatabase) WriteTemplate(template *protobuf.Template) error {
	//todo: add services
	q := "INSERT INTO template (ID, ProjectID, Name, DisplayName, Services, NHosts, Description) VALUES (?,?,?,?,?,?,?)"

	_, err := db.connection.Exec(q, template.ID, template.ProjectID, template.Name, template.DisplayName, template.Services, template.NHosts, template.Description)
	if err != nil {
		return ErrWriteObjectByKey
	}
	return nil
}

func (db MySqlDatabase) DeleteTemplate(TemplateId string) error {
	q := "DELETE FROM template WHERE ID = ?"
	_, err := db.connection.Exec(q, TemplateId)
	if err != nil {
		return ErrDeleteObjectByKey
	}
	return nil
}

func (db MySqlDatabase) ListTemplates(projectID string) ([]protobuf.Template, error) {
	//todo: add services
	q := "SELECT ID, ProjectID, Name, DisplayName, NHosts, Description FROM template"
	rows, err := db.connection.Query(q)
	if err != nil {
		return nil, ErrReadObjectList
	}
	defer rows.Close()

	templates := []protobuf.Template{}
	template := protobuf.Template{}
	for rows.Next() {
		if err := rows.Scan(&template.ID, &template.ProjectID, &template.Name, &template.DisplayName, &template.NHosts, &template.Description); err != nil {
			return nil, ErrReadObjectByKey
		}
		templates = append(templates, template)
	}
	if err := rows.Err(); err != nil {
		return nil, ErrReadObjectList
	}
	return templates, nil
}

func (db MySqlDatabase) DeleteServiceType(serviceTypeIdOrName string) error {
	isUuid := utils.IsUuid(serviceTypeIdOrName)
	var err error
	if isUuid {
		err = deleteServiceTypebyId(db, serviceTypeIdOrName)
	} else {
		err = deleteServiceTypebyName(db, serviceTypeIdOrName)
	}
	return err
}

func deleteServiceTypebyId(db MySqlDatabase, serviceTypeIdOrName string) error {
	q := "DELETE FROM service_type WHERE ID = ?;"
	_, err := db.connection.Exec(q, serviceTypeIdOrName)
	if err != nil {
		return ErrDeleteObjectByKey
	}
	return nil
}

func deleteServiceTypebyName(db MySqlDatabase, serviceTypeIdOrName string) error {
	q := "DELETE FROM service_type WHERE Type = ?;"
	_, err := db.connection.Exec(q, serviceTypeIdOrName)
	if err != nil {
		return ErrDeleteObjectByKey
	}
	return nil
}

func (db MySqlDatabase) ReadImage(imageIdOrName string) (*protobuf.Image, error) {
	isUuid := utils.IsUuid(imageIdOrName)
	var image *protobuf.Image
	var err error
	if isUuid {
		image, err = readImagebyId(db, imageIdOrName)
	} else {
		image, err = readImagebyName(db, imageIdOrName)
	}
	return image, err
}

func readImagebyName(db MySqlDatabase, imageIdOrName string) (*protobuf.Image, error) {
	q := "SELECT ID, Name, AnsibleUser, CloudImageId FROM image WHERE Name = ?"
	image := protobuf.Image{ID: "", Name: "", AnsibleUser: "", CloudImageID: ""}
	res := db.connection.QueryRow(q, imageIdOrName)
	if err := res.Scan(&image.ID, &image.Name, &image.AnsibleUser, &image.CloudImageID); err != nil {
		if err == sql.ErrNoRows {
			return &image, nil
		}
		return nil, ErrReadObjectByKey
	}
	return &image, nil
}

func readImagebyId(db MySqlDatabase, imageIdOrName string) (*protobuf.Image, error) {
	q := "SELECT ID, Name, AnsibleUser, CloudImageId FROM image WHERE ID = ?"
	image := protobuf.Image{ID: "", Name: "", AnsibleUser: "", CloudImageID: ""}
	res := db.connection.QueryRow(q, imageIdOrName)
	if err := res.Scan(&image.ID, &image.Name, &image.AnsibleUser, &image.CloudImageID); err != nil {
		if err == sql.ErrNoRows {
			return &image, nil
		}
		return nil, ErrReadObjectByKey
	}
	return &image, nil
}

func (db MySqlDatabase) WriteImage(image *protobuf.Image) error {
	q := "INSERT INTO image (ID, Name, AnsibleUser, CloudImageId) VALUES (?,?,?,?)"

	_, err := db.connection.Exec(q, image.ID, image.Name, image.AnsibleUser, image.CloudImageID)
	if err != nil {
		return ErrWriteObjectByKey
	}
	return nil
}

func (db MySqlDatabase) DeleteImage(imageIdOrName string) error {
	isUuid := utils.IsUuid(imageIdOrName)
	var err error
	if isUuid {
		err = deleteImagebyId(db, imageIdOrName)
	} else {
		err = deleteImagebyName(db, imageIdOrName)
	}
	return err
}

func deleteImagebyName(db MySqlDatabase, imageIdOrName string) error {
	q := "DELETE FROM image WHERE Name = ?"
	_, err := db.connection.Exec(q, imageIdOrName)
	if err != nil {
		return ErrDeleteObjectByKey
	}
	return nil
}

func deleteImagebyId(db MySqlDatabase, imageIdOrName string) error {
	q := "DELETE FROM image WHERE ID = ?"
	_, err := db.connection.Exec(q, imageIdOrName)
	if err != nil {
		return ErrDeleteObjectByKey
	}
	return nil
}

func (db MySqlDatabase) UpdateImage(image *protobuf.Image) error {
	q := "UPDATE image SET Name = ?, AnsibleUser = ?, CloudImageId = ? WHERE ID = ?"
	_, err := db.connection.Exec(q, image.Name, image.AnsibleUser, image.CloudImageID, image.ID)
	if err != nil {
		return ErrUpdateObjectByKey
	}

	return nil
}

func (db MySqlDatabase) ReadImagesList() ([]protobuf.Image, error) {
	q := "SELECT ID, Name, AnsibleUser, CloudImageId FROM image"
	rows, err := db.connection.Query(q)
	if err != nil {
		return nil, ErrReadObjectList
	}
	defer rows.Close()

	images := []protobuf.Image{}
	for rows.Next() {
		var image protobuf.Image
		if err := rows.Scan(&image.ID, &image.Name, &image.AnsibleUser, &image.CloudImageID); err != nil {
			return nil, ErrReadObjectByKey
		}
		images = append(images, image)
	}
	if err := rows.Err(); err != nil {
		return nil, ErrReadObjectList
	}
	return images, nil
}

func (db MySqlDatabase) ReadFlavor(flavorIdOrName string) (*protobuf.Flavor, error) {
	isUuid := utils.IsUuid(flavorIdOrName)
	var flavor *protobuf.Flavor
	var err error
	if isUuid {
		flavor, err = readFlavorbyId(db, flavorIdOrName)
	} else {
		flavor, err = readFlavorbyName(db, flavorIdOrName)
	}
	return flavor, err
}

func readFlavorbyName(db MySqlDatabase, flavorIdOrName string) (*protobuf.Flavor, error) {
	q := "SELECT ID, Name, VCPUs, RAM, Disk FROM flavor WHERE Name = ?"
	flavor := protobuf.Flavor{ID: "", Name: "", VCPUs: 0, RAM: 0, Disk: 0}
	res := db.connection.QueryRow(q, flavorIdOrName)
	if err := res.Scan(&flavor.ID, &flavor.Name, &flavor.VCPUs, &flavor.RAM, &flavor.Disk); err != nil {
		if err == sql.ErrNoRows {
			return &flavor, nil
		}
		return nil, ErrReadObjectByKey
	}
	return &flavor, nil
}

func readFlavorbyId(db MySqlDatabase, flavorIdOrName string) (*protobuf.Flavor, error) {
	q := "SELECT ID, Name, VCPUs, RAM, Disk FROM flavor WHERE ID = ?"
	flavor := protobuf.Flavor{ID: "", Name: "", VCPUs: 0, RAM: 0, Disk: 0}
	res := db.connection.QueryRow(q, flavorIdOrName)
	if err := res.Scan(&flavor.ID, &flavor.Name, &flavor.VCPUs, &flavor.RAM, &flavor.Disk); err != nil {
		if err == sql.ErrNoRows {
			return &flavor, nil
		}
		return nil, ErrReadObjectByKey
	}
	return &flavor, nil
}

func (db MySqlDatabase) WriteFlavor(flavor *protobuf.Flavor) error {
	q := "INSERT INTO flavor (ID, Name, VCPUs, RAM, Disk) VALUES (?,?,?,?,?)"

	_, err := db.connection.Exec(q, flavor.ID, flavor.Name, flavor.VCPUs, flavor.RAM, flavor.Disk)
	if err != nil {
		return ErrWriteObjectByKey
	}
	return nil
}

func (db MySqlDatabase) DeleteFlavor(flavorIdOrName string) error {
	isUuid := utils.IsUuid(flavorIdOrName)
	var err error
	if isUuid {
		err = deleteFlavorbyId(db, flavorIdOrName)
	} else {
		err = deleteFlavorbyName(db, flavorIdOrName)
	}
	return err
}

func deleteFlavorbyName(db MySqlDatabase, flavorIdOrName string) error {
	q := "DELETE FROM flavor WHERE Name = ?"
	_, err := db.connection.Exec(q, flavorIdOrName)
	if err != nil {
		return ErrDeleteObjectByKey
	}
	return nil
}

func deleteFlavorbyId(db MySqlDatabase, flavorIdOrName string) error {
	q := "DELETE FROM flavor WHERE ID = ?"
	_, err := db.connection.Exec(q, flavorIdOrName)
	if err != nil {
		return ErrDeleteObjectByKey
	}
	return nil
}

func (db MySqlDatabase) UpdateFlavor(name string, flavor *protobuf.Flavor) error {
	q := "UPDATE flavor SET Name = ?, VCPUs = ?, RAM = ?, Disk = ? WHERE ID = ?"
	_, err := db.connection.Exec(q, flavor.Name, flavor.VCPUs, flavor.RAM, flavor.Disk, flavor.ID)
	if err != nil {
		return ErrUpdateObjectByKey
	}

	return nil
}

func (db MySqlDatabase) ReadFlavorsList() ([]protobuf.Flavor, error) {
	q := "SELECT ID, Name, VCPUs, RAM, Disk FROM flavor"
	rows, err := db.connection.Query(q)
	if err != nil {
		return nil, ErrStartQueryConnection
	}
	defer rows.Close()

	flavors := []protobuf.Flavor{}
	for rows.Next() {
		var flavor protobuf.Flavor
		if err := rows.Scan(&flavor.ID, &flavor.Name, &flavor.VCPUs, &flavor.RAM, &flavor.Disk); err != nil {
			return nil, ErrReadObjectByKey
		}
		flavors = append(flavors, flavor)
	}
	if err := rows.Err(); err != nil {
		return nil, ErrReadObjectList
	}
	return flavors, nil
}

func (db MySqlDatabase) ReadServiceType(serviceTypeIdOrName string) (*protobuf.ServiceType, error) {
	isUuid := utils.IsUuid(serviceTypeIdOrName)
	var sType *protobuf.ServiceType
	var err error
	if isUuid {
		sType, err = readServiceTypebyId(db, serviceTypeIdOrName)
	} else {
		sType, err = readServiceTypebyName(db, serviceTypeIdOrName)
	}

	return sType, err
}

func readServiceTypebyName(db MySqlDatabase, serviceTypeIdOrName string) (*protobuf.ServiceType, error) {
	q := `SELECT ID, Type, COALESCE(Description,''), DefaultVersion, Class, COALESCE(AccessPort,'')
			FROM service_type WHERE Type = ?`
	st := protobuf.ServiceType{ID: "", Type: ""}
	res := db.connection.QueryRow(q, serviceTypeIdOrName)
	if err := res.Scan(&st.ID, &st.Type, &st.Description, &st.DefaultVersion, &st.Class, &st.AccessPort); err != nil {
		if err == sql.ErrNoRows {
			return &st, nil
		}
		return nil, ErrReadObjectByKey
	}

	err := db.readServiceTypeInfo(&st)
	if err != nil {
		return nil, ErrReadObjectByKey
	}

	return &st, nil
}

func readServiceTypebyId(db MySqlDatabase, serviceTypeIdOrName string) (*protobuf.ServiceType, error) {
	q := `SELECT ID, Type, COALESCE(Description,''), DefaultVersion, Class, COALESCE(AccessPort,'')
			FROM service_type WHERE ID = ?`
	st := protobuf.ServiceType{ID: "", Type: ""}
	res := db.connection.QueryRow(q, serviceTypeIdOrName)
	if err := res.Scan(&st.ID, &st.Type, &st.Description, &st.DefaultVersion, &st.Class, &st.AccessPort); err != nil {
		if err == sql.ErrNoRows {
			return &st, nil
		}
		return nil, ErrReadObjectByKey
	}

	err := db.readServiceTypeInfo(&st)
	if err != nil {
		return nil, err
	}

	return &st, nil
}

func (db MySqlDatabase) readServiceTypeInfo(st *protobuf.ServiceType) error {
	//read all versions
	//make query for service_type versions
	vq := `SELECT ID, Version, COALESCE(Description,''), COALESCE(DownloadURL,'')
			 FROM service_version WHERE ServiceTypeID = ?`
	//get all rows
	vrows, err := db.connection.Query(vq, st.ID)
	if err != nil {
		return ErrReadObjectList
	}

	svv := []*protobuf.ServiceVersion{}
	for vrows.Next() {
		//read version rows one by one
		var sv protobuf.ServiceVersion
		if err := vrows.Scan(&sv.ID, &sv.Version, &sv.Description, &sv.DownloadURL); err != nil {
			return ErrReadObjectByKey

		}
		//add configs and dependencies
		err := db.readServiceVersionInfo(&sv)
		if err != nil {
			return err
		}
		//add version to array
		svv = append(svv, &sv)
	}
	if err := vrows.Err(); err != nil {
		return ErrReadObjectList
	}
	vrows.Close()

	//add all versions to service_type
	st.Versions = svv

	//read health_check
	hq := `SELECT ID, CheckType
			 FROM health_check WHERE ServiceTypeID = ?`
	//get all rows
	hrows, err := db.connection.Query(hq, st.ID)
	if err != nil {
		return ErrReadObjectList
	}

	shh := []*protobuf.ServiceHealthCheck{}
	for hrows.Next() {
		//read health check rows one by one
		var sh protobuf.ServiceHealthCheck
		if err := hrows.Scan(&sh.ID, &sh.CheckType); err != nil {
			return ErrReadObjectByKey
		}
		//add configs

		err := db.readHealthCheckInfo(&sh)
		if err != nil {
			return err
		}
		//add health check to array
		shh = append(shh, &sh)
	}
	if err := hrows.Err(); err != nil {
		return ErrReadObjectByKey
	}
	hrows.Close()

	//add all health checks to service_type
	st.HealthCheck = shh

	//read ports
	//make the query
	pq := `SELECT Port, COALESCE(Description,'') FROM service_port WHERE ServiceTypeID = ?`
	//get all rows of ports according to particular service_type
	prows, err := db.connection.Query(pq, st.ID)
	if err != nil {
		return ErrReadObjectByKey
	}

	sports := []*protobuf.ServicePort{}
	for prows.Next() {
		// add all ports one by one to array
		var sp protobuf.ServicePort
		if err := prows.Scan(&sp.Port, &sp.Description); err != nil {
			return ErrReadObjectByKey
		}
		sports = append(sports, &sp)
	}
	if err := prows.Err(); err != nil {
		return ErrReadObjectByKey
	}
	prows.Close()
	//add port array to service_type structure
	st.Ports = sports
	return nil
}

func (db MySqlDatabase) ReadServicesTypesList() ([]protobuf.ServiceType, error) {
	//make a query to read all service types
	q := `SELECT ID, Type, COALESCE(Description,''), DefaultVersion, Class, COALESCE(AccessPort,'')
 			 FROM service_type`
	rows, err := db.connection.Query(q)
	if err != nil {
		return nil, ErrReadObjectList
	}
	defer rows.Close()

	sTypes := []protobuf.ServiceType{}
	for rows.Next() {
		var st protobuf.ServiceType
		if err := rows.Scan(&st.ID, &st.Type, &st.Description, &st.DefaultVersion, &st.Class, &st.AccessPort); err != nil {
			return nil, ErrReadObjectByKey
		}
		err := db.readServiceTypeInfo(&st)
		if err != nil {
			return nil, err
		}
		sTypes = append(sTypes, st)
	}
	if err := rows.Err(); err != nil {
		return nil, ErrReadObjectList
	}

	return sTypes, nil
}

func (db MySqlDatabase) UpdateServiceType(st *protobuf.ServiceType) error {
	tx, err := db.connection.Begin()
	if err != nil {
		return ErrStartQueryConnection
	}

	//rollback in case of error
	defer tx.Rollback()

	csq := "SELECT ID FROM health_check WHERE ServiceTypeID = ?"
	res := db.connection.QueryRow(csq, st.ID)
	var hId string
	hc_exist := 1
	err = res.Scan(&hId)
	if err == sql.ErrNoRows {
		hc_exist = 0
	} else if err != nil {
		return ErrUpdateObjectByKey
	} else {
		dhq := "DELETE FROM health_check WHERE ServiceTypeID = ?"
		_, err = tx.Exec(dhq, st.ID)
		if err != nil {
			return ErrUpdateObjectByKey
		}
	}

	//update service type info
	q := "UPDATE service_type SET Type = ?, DefaultVersion = ?, Class = ?, AccessPort = ?, Description = ? WHERE ID = ?"
	_, err = tx.Exec(q, st.Type, st.DefaultVersion, st.Class, st.AccessPort, st.Description, st.ID)
	if err != nil {
		return ErrUpdateObjectByKey
	}

	//save health chek info
	if hc_exist != 0 {
		for _, sh := range st.HealthCheck {
			hq := "INSERT INTO health_check (ID, CheckType, ServiceTypeID) VALUES (?,?,?)"
			shId, err := uuid.NewRandom()
			if err != nil {
				return ErrNewUuid
			}
			_, err = tx.Exec(hq, shId.String(), sh.CheckType, st.ID)
			if err != nil {
				return ErrUpdateObjectByKey
			}
			for _, shc := range sh.Configs {
				q := `INSERT INTO health_configs (ID, ParameterName, AnsibleVarName, Type, DefaultValue, Required, 
				IsList, Description, CheckType) VALUES (?,?,?,?,?,?,?,?,?)`

				scId, err := uuid.NewRandom()
				if err != nil {
					return ErrNewUuid
				}

				_, err = tx.Exec(q, scId.String(), shc.ParameterName, shc.AnsibleVarName, shc.Type, shc.DefaultValue,
					shc.Required, shc.IsList, shc.Description, shId.String())
				if err != nil {
					return ErrUpdateObjectByKey
				}
			}
		}
	}
	//save versions info
	for _, sv := range st.Versions {
		veq := "SELECT ID FROM service_version WHERE Version = ? AND ServiceTypeID = ?"
		res := db.connection.QueryRow(veq, sv.Version, st.ID)
		var svId string
		if err := res.Scan(&svId); err != nil {
			if err == sql.ErrNoRows {
				//add new version
				vq := "INSERT INTO service_version (ID, Version, DownloadURL, ServiceTypeID, Description) VALUES (?,?,?,?,?)"
				svId, err := uuid.NewRandom()
				if err != nil {
					return ErrNewUuid
				}
				_, err = tx.Exec(vq, svId, sv.Version, sv.DownloadURL, st.ID, sv.Description)
				if err != nil {
					return ErrUpdateObjectByKey
				}
				for _, sc := range sv.Configs {
					q := `INSERT INTO service_config (ID, ParameterName, Type, PossibleValues, DefaultValue, Required,   
			  Description, AnsibleVarName, IsList, VersionID) VALUES (?,?,?,?,?,?,?,?,?,?)`
					pv, err := json.Marshal(sc.PossibleValues)
					if err != nil {
						return ErrUnmarshalJson
					}
					scId, err := uuid.NewRandom()
					if err != nil {
						return ErrNewUuid
					}
					_, err = tx.Exec(q, scId, sc.ParameterName, sc.Type, string(pv), sc.DefaultValue,
						sc.Required, sc.Description, sc.AnsibleVarName, sc.IsList, svId)
					if err != nil {
						return ErrUpdateObjectByKey
					}
				}
				for _, sd := range sv.Dependencies {
					q = `INSERT INTO service_dependency (ID, ServiceType, DefaultServiceVersion, Description, ServiceVersionID)
			VALUES (?,?,?,?,?)`

					sdId, err := uuid.NewRandom()
					if err != nil {
						return ErrNewUuid
					}

					_, err = tx.Exec(q, sdId, sd.ServiceType, sd.DefaultServiceVersion, sd.Description, svId)
					if err != nil {
						return ErrUpdateObjectByKey
					}
					for _, v := range sd.ServiceVersions {
						vq := `SELECT service_version.ID FROM service_version INNER JOIN service_type ON 
					  service_type.ID = service_version.ServiceTypeID WHERE service_type.Type = ? AND service_version.Version = ?`
						res := db.connection.QueryRow(vq, sd.ServiceType, v)
						var svId string
						if err := res.Scan(&svId); err != nil {
							if err == sql.ErrNoRows {
								return ErrObjectParamNotExist(svId)
							}
							return ErrUpdateObjectByKey
						}
						dtvq := "INSERT INTO dependency_to_version (ServiceDependencyID, DependentVersionID) VALUES (?,?)"
						_, err = tx.Exec(dtvq, sdId, svId)
						if err != nil {
							return ErrUpdateObjectByKey
						}
					}
				}
			} else {
				err = db.UpdateServiceTypeVersion(st.ID, sv)
				if err != nil {
					return ErrUpdateObjectByKey
				}
			}
		}
	}

	psq := "SELECT ID FROM service_port WHERE ServiceTypeID = ?"
	res = db.connection.QueryRow(psq, st.ID)
	var pId string
	p_exist := 1
	err = res.Scan(&pId)
	if err == sql.ErrNoRows {
		p_exist = 0
	} else if err != nil {
		return ErrUpdateObjectByKey
	} else {
		dpq := "DELETE FROM service_port WHERE ServiceTypeID = ?"
		_, err = tx.Exec(dpq, st.ID)
		if err != nil {
			return ErrUpdateObjectByKey
		}
	}
	if p_exist != 0 {
		for _, p := range st.Ports {
			pq := "INSERT INTO service_port (ID, Port, Description, ServiceTypeID) VALUES (?,?,?,?)"
			pId, err := uuid.NewRandom()
			if err != nil {
				return ErrNewUuid
			}
			_, err = tx.Exec(pq, pId.String(), p.Port, p.Description, st.ID)
			if err != nil {
				return ErrUpdateObjectByKey
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return ErrUpdateObjectByKey
	}

	return nil
}

func (db MySqlDatabase) ReadServiceTypeVersion(serviceTypeIdOrName string, versionIdOrName string) (*protobuf.ServiceVersion, error) {
	//read ID for particular service_type
	SisUuid := utils.IsUuid(serviceTypeIdOrName)
	var st *protobuf.ServiceType
	var err error
	if SisUuid {
		st, err = readServiceTypebyId(db, serviceTypeIdOrName)
		if err != nil {
			return nil, err
		}
	} else {
		st, err = readServiceTypebyName(db, serviceTypeIdOrName)
		if err != nil {
			return nil, err
		}
	}

	VisUuid := utils.IsUuid(versionIdOrName)
	var sv *protobuf.ServiceVersion
	if VisUuid {
		sv, err = readVersionbyId(db, st.ID, versionIdOrName)
		if err != nil {
			return nil, err
		}
	} else {
		sv, err = readVersionbyName(db, st.ID, versionIdOrName)
		if err != nil {
			return nil, err
		}
	}

	err = db.readServiceVersionInfo(sv)
	if err != nil {
		return nil, err
	}

	return sv, nil
}

func readVersionbyName(db MySqlDatabase, serviceTypeIdOrName string, versionIdOrName string) (*protobuf.ServiceVersion, error) {
	q := `SELECT ID, Version, COALESCE(Description,''), COALESCE(DownloadURL,'')
	FROM service_version WHERE ServiceTypeID = ? and Version = ?`
	sv := protobuf.ServiceVersion{ID: "", Version: ""}
	res := db.connection.QueryRow(q, serviceTypeIdOrName, versionIdOrName)
	if err := res.Scan(&sv.ID, &sv.Version, &sv.Description, &sv.DownloadURL); err != nil {
		if err == sql.ErrNoRows {
			return &sv, nil
		}
		return nil, ErrReadObjectByKey
	}
	return &sv, nil
}

func readVersionbyId(db MySqlDatabase, serviceTypeIdOrName string, versionIdOrName string) (*protobuf.ServiceVersion, error) {
	q := `SELECT ID, Version, COALESCE(Description,''), COALESCE(DownloadURL,'')
	FROM service_version WHERE ServiceTypeID = ? and ID = ?`
	sv := protobuf.ServiceVersion{ID: "", Version: ""}
	res := db.connection.QueryRow(q, serviceTypeIdOrName, versionIdOrName)
	if err := res.Scan(&sv.ID, &sv.Version, &sv.Description, &sv.DownloadURL); err != nil {
		if err == sql.ErrNoRows {
			return &sv, nil
		}
		return nil, ErrReadObjectByKey
	}
	return &sv, nil
}

func readServiceTypeVersionId(db MySqlDatabase, serviceTypeIdOrName string, versionIdOrName string) (string, error) {
	VisUuid := utils.IsUuid(versionIdOrName)
	if VisUuid {
		return versionIdOrName, nil
	} else {
		SisUuid := utils.IsUuid(serviceTypeIdOrName)
		var st *protobuf.ServiceType
		var sv *protobuf.ServiceVersion
		var err error
		if SisUuid {
			st.ID = serviceTypeIdOrName
		} else {
			st, err = readServiceTypebyName(db, serviceTypeIdOrName)
			if err != nil {
				return "", err
			}
		}
		sv, err = readVersionbyName(db, st.ID, versionIdOrName)
		if err != nil {
			return "", err
		}
		return sv.ID, nil
	}
}

func (db MySqlDatabase) DeleteServiceTypeVersion(serviceTypeIdOrName string, versionIdOrName string) error {

	VersionId, err := readServiceTypeVersionId(db, serviceTypeIdOrName, versionIdOrName)
	if err != nil {
		return err
	}

	q := "DELETE FROM service_version WHERE ID = ?;"
	_, err = db.connection.Exec(q, VersionId)
	if err != nil {
		return ErrDeleteObjectByKey
	}
	return nil
}

func (db MySqlDatabase) UpdateServiceTypeVersion(serviceTypeIdOrName string, version *protobuf.ServiceVersion) error {

	csq := "SELECT ID FROM service_config WHERE VersionID = ?"
	res := db.connection.QueryRow(csq, version.ID)
	var scId string
	err := res.Scan(&scId)
	if err != nil && err != sql.ErrNoRows {
		return ErrUpdateObjectByKey
	} else {
		cdq := "DELETE FROM service_config WHERE VersionID = ?"
		_, err = db.connection.Exec(cdq, version.ID)
		if err != nil {
			return ErrUpdateObjectByKey
		}
	}

	dsq := "SELECT ID FROM service_dependency WHERE ServiceVersionID = ?"
	res = db.connection.QueryRow(dsq, version.ID)
	var sdId string
	err = res.Scan(&sdId)
	if err != nil && err != sql.ErrNoRows {
		return ErrUpdateObjectByKey
	} else {
		ddq := "DELETE FROM service_dependency WHERE ServiceVersionID = ?"
		_, err = db.connection.Exec(ddq, version.ID)
		if err != nil {
			return ErrUpdateObjectByKey
		}
	}

	q := "UPDATE service_version SET Description = ?, DownloadURL = ? WHERE ID = ?"
	_, err = db.connection.Exec(q, version.Description, version.DownloadURL, version.ID)
	if err != nil {
		return ErrUpdateObjectByKey
	}

	for _, sc := range version.Configs {
		q := `INSERT INTO service_config (ID, ParameterName, Type, PossibleValues, DefaultValue, Required,   
		Description, AnsibleVarName, IsList, VersionID) VALUES (?,?,?,?,?,?,?,?,?,?)`

		pv, err := json.Marshal(sc.PossibleValues)
		if err != nil {
			return ErrUnmarshalJson
		}

		scId, err := uuid.NewRandom()
		if err != nil {
			return ErrNewUuid
		}

		_, err = db.connection.Exec(q, scId.String(), sc.ParameterName, sc.Type, string(pv), sc.DefaultValue,
			sc.Required, sc.Description, sc.AnsibleVarName, sc.IsList, version.ID)
		if err != nil {
			return ErrUpdateObjectByKey
		}
	}

	for _, sd := range version.Dependencies {
		q = `INSERT INTO service_dependency (ID, ServiceType, DefaultServiceVersion, Description, ServiceVersionID)
	  VALUES (?,?,?,?,?)`

		sdId, err := uuid.NewRandom()
		if err != nil {
			return ErrNewUuid
		}

		_, err = db.connection.Exec(q, sdId, sd.ServiceType, sd.DefaultServiceVersion, sd.Description, version.ID)
		if err != nil {
			return ErrUpdateObjectByKey
		}
		for _, v := range sd.ServiceVersions {
			var svId string
			isUuid := utils.IsUuid(v)
			if isUuid {
				svId = v
			} else {
				vq := `SELECT service_version.ID FROM service_version INNER JOIN service_type ON 
							service_type.ID = service_version.ServiceTypeID WHERE service_type.Type = ? AND service_version.Version = ?`
				res := db.connection.QueryRow(vq, sd.ServiceType, v)
				if err := res.Scan(&svId); err != nil {
					if err == sql.ErrNoRows {
						return ErrUpdateObjectByKey
					}
					return ErrUpdateObjectByKey
				}
			}

			dtvq := "REPLACE INTO dependency_to_version (ServiceDependencyID, DependentVersionID) VALUES (?,?)"
			_, err = db.connection.Exec(dtvq, sdId, svId)

			if err != nil {
				return ErrUpdateObjectByKey
			}
		}
	}
	return nil
}

func (db MySqlDatabase) ReadServiceTypeVersionConfig(serviceTypeIdOrName string, versionIdOrName string, parameterName string) (*protobuf.ServiceConfig, error) {
	VersionId, err := readServiceTypeVersionId(db, serviceTypeIdOrName, versionIdOrName)
	if err != nil {
		return nil, err
	}

	cq := `SELECT ID, ParameterName, Type,  COALESCE(PossibleValues, ''), DefaultValue,  Required, 
			COALESCE(Description, ''), AnsibleVarName,  IsList FROM service_config WHERE VersionID = ? and ParameterName = ?`
	c := protobuf.ServiceConfig{ParameterName: "", Type: ""}
	res := db.connection.QueryRow(cq, VersionId, parameterName)
	var posVals string
	if err := res.Scan(&c.ID, &c.ParameterName, &c.Type, &posVals, &c.DefaultValue, &c.Required, &c.Description,
		&c.AnsibleVarName, &c.IsList); err != nil {
		if err == sql.ErrNoRows {
			return &c, nil
		}

		err = json.Unmarshal([]byte(posVals), &c.PossibleValues)
		if err != nil {
			return nil, ErrUnmarshalJson
		}

		return nil, ErrReadObjectByKey
	}
	return &c, nil

}

func (db MySqlDatabase) UpdateServiceTypeVersionConfig(serviceTypeIdOrName string, versionIdOrName string, config *protobuf.ServiceConfig) error {
	VersionId, err := readServiceTypeVersionId(db, serviceTypeIdOrName, versionIdOrName)
	if err != nil {
		return err
	}
	q := "UPDATE service_config SET Type = ?, PossibleValues = ?, DefaultValue = ?, Required = ?, Description = ?, IsList = ?  WHERE VersionID = ? and ParameterName = ?"

	pv, err := json.Marshal(config.PossibleValues)
	if err != nil {
		return ErrUnmarshalJson
	}
	_, err = db.connection.Exec(q, config.Type, string(pv), config.DefaultValue, config.Required, config.Description, config.IsList, VersionId, config.ParameterName)
	if err != nil {
		return ErrUpdateObjectByKey
	}

	return nil

}

func (db MySqlDatabase) DeleteServiceTypeVersionConfig(serviceTypeIdOrName string, versionIdOrName string, parameterName string) error {
	VersionId, err := readServiceTypeVersionId(db, serviceTypeIdOrName, versionIdOrName)
	if err != nil {
		return err
	}

	q := "DELETE FROM service_config WHERE VersionID = ? and ParameterName = ?;"
	_, err = db.connection.Exec(q, VersionId, parameterName)
	if err != nil {
		return ErrDeleteObjectByKey
	}
	return nil

}

func (db MySqlDatabase) readHealthCheckInfo(sh *protobuf.ServiceHealthCheck) error {
	//read configs for health check
	hq := `SELECT  ParameterName, Type, Description, DefaultValue, Required, AnsibleVarName, IsList FROM health_configs WHERE CheckType = ?`
	hrows, err := db.connection.Query(hq, sh.ID)
	if err != nil {
		return ErrReadObjectByKey
	}
	shc := []*protobuf.HealthConfigs{}
	for hrows.Next() {
		//add all config info to array one by one
		var sc protobuf.HealthConfigs
		if err := hrows.Scan(&sc.ParameterName, &sc.Type, &sc.Description, &sc.DefaultValue, &sc.Required,
			&sc.AnsibleVarName, &sc.IsList); err != nil {
			return ErrReadObjectByKey
		}

		shc = append(shc, &sc)
	}
	if err := hrows.Err(); err != nil {
		return ErrReadObjectByKey
	}
	hrows.Close()
	//add config array to service_version structure
	sh.Configs = shc

	return nil
}

func (db MySqlDatabase) readServiceVersionInfo(sv *protobuf.ServiceVersion) error {
	// read configs for version
	cq := `SELECT ID, ParameterName, Type,  COALESCE(PossibleValues, ''), DefaultValue,  Required, 
			COALESCE(Description, ''), AnsibleVarName,  IsList FROM service_config WHERE VersionID = ?`
	// read all config rows
	crows, err := db.connection.Query(cq, sv.ID)
	if err != nil {
		return ErrReadObjectByKey
	}

	scc := []*protobuf.ServiceConfig{}
	for crows.Next() {
		//add all config info to array one by one
		var sc protobuf.ServiceConfig
		var posVals string
		if err := crows.Scan(&sc.ID, &sc.ParameterName, &sc.Type, &posVals, &sc.DefaultValue, &sc.Required, &sc.Description,
			&sc.AnsibleVarName, &sc.IsList); err != nil {
			return ErrReadObjectByKey
		}
		err = json.Unmarshal([]byte(posVals), &sc.PossibleValues)
		if err != nil {
			return ErrUnmarshalJson
		}

		scc = append(scc, &sc)
	}
	if err := crows.Err(); err != nil {
		return ErrReadObjectByKey
	}
	crows.Close()
	//add config array to service_version structure
	sv.Configs = scc

	//read dependencies for version
	dq := `SELECT ID, ServiceType, DefaultServiceVersion, COALESCE(Description, '') FROM service_dependency WHERE ServiceVersionID = ?`
	//select all rows
	drows, err := db.connection.Query(dq, sv.ID)
	if err != nil {
		return ErrReadObjectByKey
	}

	sdd := []*protobuf.ServiceDependency{}
	for drows.Next() {
		var sd protobuf.ServiceDependency
		var sdId string
		if err := drows.Scan(&sdId, &sd.ServiceType, &sd.DefaultServiceVersion, &sd.Description); err != nil {
			return ErrReadObjectByKey
		}

		//select version of dependent service
		dtvq := "SELECT DependentVersionID FROM dependency_to_version WHERE ServiceDependencyID = ?"
		dtvrows, err := db.connection.Query(dtvq, sdId)
		if err != nil {
			return ErrReadObjectByKey
		}

		depVersions := []string{}
		for dtvrows.Next() {
			//read all versions of dependent service and add them to array one by one
			var depV string
			if err := dtvrows.Scan(&depV); err != nil {
				return ErrReadObjectByKey
			}
			depVersions = append(depVersions, depV)
		}
		if err := dtvrows.Err(); err != nil {
			return ErrReadObjectByKey
		}
		dtvrows.Close()
		//add version array of dependent service to service_dependency structure
		sd.ServiceVersions = depVersions
		sdd = append(sdd, &sd)
	}
	if err := drows.Err(); err != nil {
		return ErrReadObjectByKey
	}
	drows.Close()
	//add all depenndencies to service_version
	sv.Dependencies = sdd
	return nil
}

func (db MySqlDatabase) WriteServiceType(sType *protobuf.ServiceType) error {
	tx, err := db.connection.Begin()
	if err != nil {
		return ErrStartQueryConnection
	}

	//rollback in case of error
	defer tx.Rollback()

	//save service type info
	q := "INSERT INTO service_type (ID, Type, DefaultVersion, Class, AccessPort, Description) VALUES (?,?,?,?,?,?)"
	_, err = tx.Exec(q, sType.ID, sType.Type, sType.DefaultVersion, sType.Class, sType.AccessPort, sType.Description)
	if err != nil {
		return ErrWriteObjectByKey
	}

	//save health check info
	for _, sh := range sType.HealthCheck {
		//save version
		hq := `INSERT INTO health_check (ID, CheckType, ServiceTypeID) VALUES (?,?,?)`
		shId, err := uuid.NewRandom()
		if err != nil {
			return ErrNewUuid
		}
		_, err = tx.Exec(hq, shId, sh.CheckType, sType.ID)
		if err != nil {
			return ErrWriteObjectByKey
		}

		//save configs info
		for _, sc := range sh.Configs {
			q := `INSERT INTO health_configs (ID, ParameterName, Description, Type, DefaultValue,  Required, AnsibleVarName,   
					IsList,  CheckType) VALUES (?,?,?,?,?,?,?,?,?)`

			scId, err := uuid.NewRandom()
			if err != nil {
				return ErrNewUuid
			}

			_, err = tx.Exec(q, scId, sc.ParameterName, sc.Description, sc.Type, sc.DefaultValue, sc.Required, sc.AnsibleVarName, sc.IsList, shId)
			if err != nil {
				return ErrWriteObjectByKey
			}
		}
	}

	//save versions info
	for _, sv := range sType.Versions {
		//save version
		vq := `INSERT INTO service_version (ID, Version, DownloadURL, ServiceTypeID, Description) VALUES (?,?,?,?,?)`
		_, err = tx.Exec(vq, sv.ID, sv.Version, sv.DownloadURL, sType.ID, sv.Description)
		if err != nil {
			return ErrWriteObjectByKey
		}

		//save configs info
		for _, sc := range sv.Configs {
			q := `INSERT INTO service_config (ID, ParameterName, AnsibleVarName, Type, DefaultValue, PossibleValues, Required, 
					IsList, Description, VersionID) VALUES (?,?,?,?,?,?,?,?,?,?)`

			pv, err := json.Marshal(sc.PossibleValues)
			if err != nil {
				return ErrUnmarshalJson
			}

			scId, err := uuid.NewRandom()
			if err != nil {
				return ErrNewUuid
			}

			_, err = tx.Exec(q, scId, sc.ParameterName, sc.AnsibleVarName, sc.Type, sc.DefaultValue, string(pv), sc.Required, sc.IsList, sc.Description, sv.ID)
			if err != nil {
				return ErrWriteObjectByKey
			}
		}
		//save dependencies info
		for _, sd := range sv.Dependencies {
			dq := `INSERT INTO service_dependency (ID, ServiceType, DefaultServiceVersion, Description, 
					ServiceVersionID) VALUES (?,?,?,?,?)`

			sdId, err := uuid.NewRandom()
			if err != nil {
				return ErrNewUuid
			}

			_, err = tx.Exec(dq, sdId, sd.ServiceType, sd.DefaultServiceVersion, sd.Description, sv.ID)
			if err != nil {
				return ErrWriteObjectByKey
			}

			for _, v := range sd.ServiceVersions {
				//get dependent sv ID
				vq := `SELECT service_version.ID FROM service_version INNER JOIN service_type ON 
							service_type.ID = service_version.ServiceTypeID WHERE service_type.Type = ? AND service_version.Version = ?`
				res := db.connection.QueryRow(vq, sd.ServiceType, v)
				var svId string
				if err := res.Scan(&svId); err != nil {
					if err == sql.ErrNoRows {
						return ErrObjectParamNotExist(svId)
					}
					return ErrWriteObjectByKey
				}

				dtvq := "REPLACE INTO dependency_to_version (ServiceDependencyID, DependentVersionID) VALUES (?,?)"
				_, err = tx.Exec(dtvq, sdId, svId)
				if err != nil {
					return ErrWriteObjectByKey
				}
			}
		}
	}
	//save ports info
	for _, p := range sType.Ports {
		pq := "REPLACE INTO service_port (ID, PORT, Description, ServiceTypeID) VALUES (?,?,?,?)"

		pId, err := uuid.NewRandom()
		if err != nil {
			return ErrNewUuid
		}
		_, err = tx.Exec(pq, pId, p.Port, p.Description, sType.ID)
		if err != nil {
			return ErrWriteObjectByKey
		}
	}

	if err = tx.Commit(); err != nil {
		return ErrWriteObjectByKey
	}

	return nil
}
