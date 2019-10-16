package handlers

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/database"
	protobuf "gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/protobuf"
	"gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/utils"
	"log"
	"net/http"
	"regexp"
)

const (
	NEW_CLUSTER         = -1
	CLUSTER_DIDNT_EXIST = -2
)

type GrpcClient interface {
	GetID(c *protobuf.Cluster) (int32, error)
	StartClusterCreation(c *protobuf.Cluster)
	StartClusterDestroying(c *protobuf.Cluster)
	StartClusterModification(c *protobuf.Cluster)
}

type HttpServer struct {
	Gc     GrpcClient
	Logger *log.Logger
	Db     database.Database
}

type serviceExists struct {
	exists  bool
	service *protobuf.Service
}

func ValidateCluster(cluster *protobuf.Cluster) bool {
	validName := regexp.MustCompile(`^[A-Za-z][A-Za-z0-9-]+$`).MatchString

	if !validName(cluster.Name) {
		log.Print("ERROR: bad name for cluster. You should use only alpha-numeric characters and '-' symbols and only alphabetic characters for leading symbol.")
		return false
	}

	if cluster.NHosts < 1 {
		log.Print(cluster.NHosts)
		log.Print("ERROR: NHosts parameter must be number >= 1.")
		return false
	}

	for _, service := range cluster.Services {
		if !ValidateService(service) {
			return false
		}
	}

	return true
}

func (hS HttpServer) ClustersGet(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	projectName := params.ByName("projectName")
	hS.Logger.Print("Get /projects/", projectName, "/clusters GET")

	project, err := hS.Db.ReadProject(projectName)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if project.Name == "" {
		hS.Logger.Printf("Project with name '%s' not found", projectName)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	//reading cluster info from database
	clusters, err := hS.Db.ReadProjectClusters(project.ID)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	err = enc.Encode(clusters)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (hS HttpServer) ClusterCreate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	projectName := params.ByName("projectName")
	hS.Logger.Print("Get /project/" + projectName + "/clusters POST")

	project, err := hS.Db.ReadProject(projectName)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if project.Name == "" {
		hS.Logger.Printf("Project with name '%s' not found", projectName)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var c protobuf.Cluster
	err = json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		hS.Logger.Print("ERROR:")
		hS.Logger.Print(err)
		hS.Logger.Print(r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// validate struct
	if !ValidateCluster(&c) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//check, that cluster with such name doesn't exist
	clusters, err := hS.Db.ReadProjectClusters(project.ID)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	existedClusterInd := CLUSTER_DIDNT_EXIST

	for i := 0; i < len(clusters); i++ {
		hS.Logger.Print(clusters[i].Name)
		if clusters[i].Name == c.Name {
			existedClusterInd = i
			if clusters[i].EntityStatus != utils.StatusFailed {
				hS.Logger.Print("Cluster with this name exists in this project")
				w.WriteHeader(http.StatusBadRequest)
				enc := json.NewEncoder(w)
				err := enc.Encode("Cluster with this name exists in this project")
				if err != nil {
					hS.Logger.Print(err)
					w.WriteHeader(http.StatusBadRequest)
				}
				return
			}
			break
		}
	}

	// If cluster was failed
	if existedClusterInd != CLUSTER_DIDNT_EXIST {
		c = clusters[existedClusterInd]
	} else {
		cUuid, err := uuid.NewRandom()
		if err != nil {
			hS.Logger.Print(err)
			w.WriteHeader(http.StatusBadRequest)
		}
		c.ID = cUuid.String()
		for _, s := range c.Services {
			sUuid, err := uuid.NewRandom()
			if err != nil {
				hS.Logger.Print(err)
				w.WriteHeader(http.StatusBadRequest)
			}
			s.ID = sUuid.String()
		}

		c.ProjectID = project.ID
	}

	c.EntityStatus = utils.StatusInited
	err = hS.Db.WriteCluster(&c)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
	}
	go hS.Gc.StartClusterCreation(&c)

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(c)
}

func (hS HttpServer) ClustersGetByName(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	projectName := params.ByName("projectName")
	clusterName := params.ByName("clusterName")
	hS.Logger.Print("Get /projects/"+projectName+"/clusters/", clusterName, " GET")

	project, err := hS.Db.ReadProject(projectName)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if project.Name == "" {
		hS.Logger.Printf("Project with name '%s' not found", projectName)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// reading project info from database
	clusters, err := hS.Db.ReadProjectClusters(project.ID)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	clusterInd := CLUSTER_DIDNT_EXIST

	for i := 0; i < len(clusters); i++ {
		if clusters[i].Name == clusterName {
			clusterInd = i
			break
		}
	}

	if clusterInd == CLUSTER_DIDNT_EXIST {
		hS.Logger.Printf("Cluster with name '%s' not found", clusterName)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	err = enc.Encode(clusters[clusterInd])
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (hS HttpServer) ClustersUpdate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	projectName := params.ByName("projectName")
	clusterName := params.ByName("clusterName")
	hS.Logger.Print("Get /projects/"+projectName+"/clusters/", clusterName, " PUT")

	project, err := hS.Db.ReadProject(projectName)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if project.Name == "" {
		hS.Logger.Printf("Project with name '%s' not found", projectName)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	//check that cluster exists
	hS.Logger.Print("Sending request to db-service to check that cluster exists...")

	clusters, err := hS.Db.ReadProjectClusters(project.ID)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	clusterInd := CLUSTER_DIDNT_EXIST

	for i := 0; i < len(clusters); i++ {
		if clusters[i].Name == clusterName {
			clusterInd = i
			if clusters[i].EntityStatus != utils.StatusCreated && clusters[i].EntityStatus != utils.StatusFailed {
				errMessage := "Status of cluster to update must be 'CREATED' or 'FAILED'"
				hS.Logger.Print(errMessage)
				w.WriteHeader(http.StatusBadRequest)
				enc := json.NewEncoder(w)
				err = enc.Encode(errMessage)
				return
			}
			break
		}
	}

	if clusterInd == CLUSTER_DIDNT_EXIST {
		errMessage := "Cluster didn't found"
		hS.Logger.Print(errMessage)
		w.WriteHeader(http.StatusBadRequest)
		enc := json.NewEncoder(w)
		err = enc.Encode(errMessage)
		return
	}

	// validate request struct
	var newC protobuf.Cluster
	err = json.NewDecoder(r.Body).Decode(&newC)
	if err != nil {
		errMessage := "Invalid JSON"
		hS.Logger.Print(errMessage)
		w.WriteHeader(http.StatusBadRequest)
		enc := json.NewEncoder(w)
		err = enc.Encode(errMessage)
		return
	}

	if !ValidateCluster(&newC) {
		errMessage := "JSON isn't correct"
		hS.Logger.Print(errMessage)
		w.WriteHeader(http.StatusBadRequest)
		enc := json.NewEncoder(w)
		err = enc.Encode(errMessage)
		return
	}

	//appending old services which does not exist in new cluster configuration
	var serviceTypesNew = map[string]serviceExists{
		utils.ServiceTypeCassandra: {
			exists:  false,
			service: nil,
		},
		utils.ServiceTypeSpark: {
			exists:  false,
			service: nil,
		},
		utils.ServiceTypeElastic: {
			exists:  false,
			service: nil,
		},
		utils.ServiceTypeJupyter: {
			exists:  false,
			service: nil,
		},
		utils.ServiceTypeIgnite: {
			exists:  false,
			service: nil,
		},
		utils.ServiceTypeJupyterhub: {
			exists:  false,
			service: nil,
		},
	}

	newC.ID = clusters[clusterInd].ID
	for _, s := range newC.Services {
		sUuid, err := uuid.NewRandom()
		if err != nil {
			hS.Logger.Print(err)
			w.WriteHeader(http.StatusBadRequest)
		}
		s.ID = sUuid.String()
		serviceTypesNew[s.Type] = serviceExists{
			exists:  true,
			service: s,
		}
	}

	for _, s := range clusters[clusterInd].Services {
		if serviceTypesNew[s.Type].exists == false {
			newC.Services = append(newC.Services, s)
		}
	}

	newC.EntityStatus = utils.StatusInited
	go hS.Gc.StartClusterModification(&newC)

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(newC)
}

func (hS HttpServer) ClustersDelete(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	projectName := params.ByName("projectName")
	clusterName := params.ByName("clusterName")
	hS.Logger.Print("Get /projects/"+projectName+"/clusters/", clusterName, " PUT")

	project, err := hS.Db.ReadProject(projectName)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if project.Name == "" {
		hS.Logger.Printf("Project with name '%s' not found", projectName)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	clusters, err := hS.Db.ReadProjectClusters(project.ID)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	clusterInd := CLUSTER_DIDNT_EXIST

	for i := 0; i < len(clusters); i++ {
		if clusters[i].Name == clusterName {
			clusterInd = i
			if clusters[i].EntityStatus != utils.StatusCreated && clusters[i].EntityStatus != utils.StatusFailed {
				errMessage := "Status of cluster to update must be 'CREATED' or 'FAILED'"
				hS.Logger.Print(errMessage)
				w.WriteHeader(http.StatusBadRequest)
				enc := json.NewEncoder(w)
				err = enc.Encode(errMessage)
				return
			}
			break
		}
	}

	if clusterInd == CLUSTER_DIDNT_EXIST {
		errMessage := "Cluster didn't found"
		hS.Logger.Print(errMessage)
		w.WriteHeader(http.StatusBadRequest)
		enc := json.NewEncoder(w)
		err = enc.Encode(errMessage)
		return
	}

	clusters[clusterInd].EntityStatus = utils.StatusStopping

	go hS.Gc.StartClusterDestroying(&clusters[clusterInd])
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(clusters[clusterInd])
}

