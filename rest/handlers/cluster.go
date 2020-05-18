package handlers

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/ispras/michman/database"
	proto "github.com/ispras/michman/protobuf"
	"github.com/ispras/michman/utils"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"regexp"
)

const (
	NEW_CLUSTER         = -1
	CLUSTER_DIDNT_EXIST = -2
)

type GrpcClient interface {
	StartClusterCreation(c *proto.Cluster)
	StartClusterDestroying(c *proto.Cluster)
	StartClusterModification(c *proto.Cluster)
}

type HttpServer struct {
	Gc         GrpcClient
	Logger     *log.Logger
	Db         database.Database
	ErrHandler ErrorHandler
}

type serviceExists struct {
	exists  bool
	service *proto.Service
}

func ValidateCluster(hS HttpServer, cluster *proto.Cluster) bool {
	validName := regexp.MustCompile(`^[A-Za-z][A-Za-z0-9-]+$`).MatchString

	if !validName(cluster.DisplayName) {
		log.Print("ERROR: bad name for cluster. You should use only alpha-numeric characters and '-' symbols and only alphabetic characters for leading symbol.")
		return false
	}

	if cluster.NHosts < 1 {
		log.Print(cluster.NHosts)
		log.Print("ERROR: NHosts parameter must be number >= 1.")
		return false
	}

	for _, service := range cluster.Services {
		if res, err := ValidateService(hS, service); !res {
			log.Print(err)
			return false
		}
	}
	return true
}

func (hS HttpServer) AddDependencies(c *proto.Cluster, curS *proto.Service) ([]*proto.Service, error) {
	var err error = nil
	var serviceToAdd *proto.Service = nil
	var servicesList []*proto.Service = nil

	sv, err := hS.Db.ReadServiceVersionByName(curS.Type, curS.Version)
	if err != nil {
		return nil, err
	}

	//check if version has dependencies
	if sv.Dependencies != nil {
		for _, sd := range sv.Dependencies {
			//check if the service from dependencies has already listed in cluster and version is ok
			flagAddS := true
			for _, clusterS := range c.Services {
				if clusterS.Type == sd.ServiceType {
					if !utils.ItemExists(sd.ServiceVersions, clusterS.Version) {
						//error: bad service version from user list
						err = errors.New("Error: service " + clusterS.Type +
							" has incompatible version for service " + curS.Type + ".")
					}
					flagAddS = false
					break
				}
			}
			if flagAddS && err == nil {
				//add service from dependencies with default configurations
				serviceToAdd = &proto.Service{
					Name:    curS.Name + "-dependent", //TODO: use better service name?
					Type:    sd.ServiceType,
					Version: sd.DefaultServiceVersion,
				}
				servicesList = append(servicesList, serviceToAdd)
			}
		}
	}

	return servicesList, err
}

func (hS HttpServer) getCluster(projectID, idORname string) (*proto.Cluster, error) {
	is_uuid := true
	_, err := uuid.Parse(idORname)
	if err != nil {
		is_uuid = false
	}

	var cluster *proto.Cluster

	if is_uuid {
		cluster, err = hS.Db.ReadCluster(idORname)
	} else {
		cluster, err = hS.Db.ReadClusterByName(projectID, idORname)
	}

	return cluster, err
}

func (hS HttpServer) ClustersGet(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	projectIdOrName := params.ByName("projectIdOrName")
	hS.Logger.Print("Get /projects/", projectIdOrName, "/clusters GET")

	project, err := hS.getProject(projectIdOrName)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, DBerror, DBerrorMessage, err)
		hS.Logger.Print(mess)
		return
	}

	if project.Name == "" {
		hS.Logger.Printf("Project with name '%s' not found", projectIdOrName)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	//reading cluster info from database
	clusters, err := hS.Db.ReadProjectClusters(project.ID)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, DBerror, DBerrorMessage, err)
		hS.Logger.Print(mess)
		return
	}

	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	err = enc.Encode(clusters)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, LibErrorStructToJson, LibErrorStructToJsonMessage, err)
		hS.Logger.Print(mess)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

func (hS HttpServer) ClusterCreate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	projectIdOrName := params.ByName("projectIdOrName")
	hS.Logger.Print("Get /project/" + projectIdOrName + "/clusters POST")

	project, err := hS.getProject(projectIdOrName)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, DBerror, DBerrorMessage, err)
		hS.Logger.Print(mess)
		return
	}

	if project.Name == "" {
		hS.Logger.Printf("Project with name '%s' not found", projectIdOrName)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var c *proto.Cluster
	err = json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, JSONerrorIncorrect, JSONerrorIncorrectMessage, err)
		hS.Logger.Print(mess)
		return
	}

	// validate struct
	if !ValidateCluster(hS, c) {
		mess, _ := hS.ErrHandler.Handle(w, JSONerrorIncorrectField, JSONerrorIncorrectFieldMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	//check, that cluster with such name doesn't exist
	searchedName := c.DisplayName + "-" + project.Name
	cluster, err := hS.Db.ReadClusterByName(project.ID, searchedName)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, DBerror, DBerrorMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	clusterExists := false

	if cluster.Name != "" {
		clusterExists = true
		if cluster.EntityStatus != utils.StatusFailed {
			mess, _ := hS.ErrHandler.Handle(w, UserErrorClusterExisted, UserErrorClusterExistedMessage, nil)
			hS.Logger.Print(mess)
			return
		}
	}

	// If cluster was failed
	if clusterExists {
		c = cluster
	} else {
		cUuid, err := uuid.NewRandom()
		if err != nil {
			mess, _ := hS.ErrHandler.Handle(w, LibErrorUUID, LibErrorUUIDMessage, nil)
			hS.Logger.Print(mess)
			return
		}
		c.ID = cUuid.String()
		//add services from user request and from dependencies
		if c.Services != nil {

			retryFlag := true
			startIdx := 0

			//first for cycle is used for updating range values with appended services
			for retryFlag {
				for i, s := range c.Services[startIdx:] {
					st, err := hS.Db.ReadServiceType(s.Type)
					if err != nil {
						mess, _ := hS.ErrHandler.Handle(w, DBerror, DBerrorMessage, nil)
						hS.Logger.Print(mess)
						return
					}

					if s.Version == "" {
						s.Version = st.DefaultVersion
					}

					//add services from dependencies
					sToAdd, err := hS.AddDependencies(c, s)
					if err != nil {
						hS.Logger.Println(err)
						mess, _ := hS.ErrHandler.Handle(w, UserErrorBadServiceVersion, UserErrorBadServiceVersionMessage, nil)
						hS.Logger.Print(mess)
						//w.WriteHeader(http.StatusBadRequest)
						return
					}
					hS.Logger.Println(sToAdd)

					changesFlag := false
					if sToAdd != nil {
						for _, curS := range sToAdd {
							c.Services = append(c.Services, curS)
						}
						changesFlag = true
					}

					if !changesFlag {
						retryFlag = false
					} else {
						//update range values if new services has been added and start new iteration from the next value
						startIdx = i + 1
						break
					}

				}
			}
		}

		c.ProjectID = project.ID
		c.Name = c.DisplayName + "-" + project.Name
		//set uuids for all cluster services
		for _, s := range c.Services {
			sUuid, err := uuid.NewRandom()
			if err != nil {
				mess, _ := hS.ErrHandler.Handle(w, LibErrorUUID, LibErrorUUIDMessage, nil)
				hS.Logger.Print(mess)
				return
			}
			s.ID = sUuid.String()
		}
	}
	// set default project Image if not specified
	if c.Image == "" {
		c.Image = project.DefaultImage
	}

	c.EntityStatus = utils.StatusInited
	if !clusterExists {
		err = hS.Db.WriteCluster(c)
		if err != nil {
			mess, _ := hS.ErrHandler.Handle(w, DBerror, DBerrorMessage, nil)
			hS.Logger.Print(mess)
			return
		}
	}
	go hS.Gc.StartClusterCreation(c)

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(c)
}

func (hS HttpServer) ClustersGetByName(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	projectIdOrName := params.ByName("projectIdOrName")
	clusterIdOrName := params.ByName("clusterIdOrName")
	hS.Logger.Print("Get /projects/"+projectIdOrName+"/clusters/", clusterIdOrName, " GET")

	project, err := hS.getProject(projectIdOrName)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, DBerror, DBerrorMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	if project.Name == "" {
		hS.Logger.Printf("Project with name '%s' not found", projectIdOrName)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// reading cluster info from database
	cluster, err := hS.getCluster(project.ID, clusterIdOrName)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, DBerror, DBerrorMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	if cluster.Name == "" {
		hS.Logger.Printf("Cluster with name or ID '%s' not found", clusterIdOrName)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	err = enc.Encode(cluster)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, LibErrorStructToJson, LibErrorStructToJsonMessage, nil)
		hS.Logger.Print(mess)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

func (hS HttpServer) ClustersStatusGetByName(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	projectIdOrName := params.ByName("projectIdOrName")
	clusterIdOrName := params.ByName("clusterIdOrName")
	hS.Logger.Print("Get /projects/"+projectIdOrName+"/clusters/", clusterIdOrName, "/status", " GET")

	project, err := hS.getProject(projectIdOrName)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, DBerror, DBerrorMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	if project.Name == "" {
		hS.Logger.Printf("Project with name '%s' not found", projectIdOrName)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// reading cluster info from database
	cluster, err := hS.getCluster(project.ID, clusterIdOrName)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, DBerror, DBerrorMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	if cluster.Name == "" {
		hS.Logger.Printf("Cluster with name or ID '%s' not found", clusterIdOrName)
		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(utils.StatusMissing))
		if err != nil {
			hS.Logger.Print(err)
			w.WriteHeader(http.StatusBadRequest)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(cluster.EntityStatus))
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (hS HttpServer) ClustersUpdate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	projectIdOrName := params.ByName("projectIdOrName")
	clusterIdOrName := params.ByName("clusterIdOrName")
	hS.Logger.Print("Get /projects/"+projectIdOrName+"/clusters/", clusterIdOrName, " PUT")

	project, err := hS.getProject(projectIdOrName)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, DBerror, DBerrorMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	if project.Name == "" {
		hS.Logger.Printf("Project with name '%s' not found", projectIdOrName)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	//check that cluster exists
	hS.Logger.Print("Sending request to db-service to check that cluster exists...")

	// reading cluster info from database
	cluster, err := hS.getCluster(project.ID, clusterIdOrName)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, DBerror, DBerrorMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	if cluster.Name == "" {
		hS.Logger.Printf("Cluster with name or ID '%s' not found", clusterIdOrName)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if cluster.EntityStatus != utils.StatusActive && cluster.EntityStatus != utils.StatusFailed {
		mess, _ := hS.ErrHandler.Handle(w, UserErrorClusterStatus, UserErrorClusterStatusMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	// validate request struct
	var newC proto.Cluster
	err = json.NewDecoder(r.Body).Decode(&newC)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, JSONerrorIncorrect, JSONerrorIncorrectMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	if newC.ID != "" || newC.Name != "" || newC.DisplayName != "" || newC.EntityStatus != "" ||
		newC.HostURL != "" || newC.MasterIP != "" || newC.ProjectID != "" || newC.Image != "" {
		mess, _ := hS.ErrHandler.Handle(w, UserErrorProjectUnmodField, UserErrorProjectUnmodFieldMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	//check correctness of new services
	for _, s := range newC.Services {
		if res, _ := ValidateService(hS, s); !res {
			mess, _ := hS.ErrHandler.Handle(w, UserErrorProjectUnmodField, UserErrorProjectUnmodFieldMessage, nil)
			hS.Logger.Print(mess)
			return
		}
	}

	sTypes, err := hS.Db.ListServicesTypes()
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, DBerror, DBerrorMessage, nil)
		hS.Logger.Print(mess)
		return
	}
	//appending old services which does not exist in new cluster configuration
	var serviceTypesOld = make(map[string]serviceExists)

	for _, st := range sTypes {
		serviceTypesOld[st.Type] = serviceExists{
			exists:  false,
			service: nil,
		}
	}

	for _, s := range cluster.Services {
		serviceTypesOld[s.Type] = serviceExists{
			exists:  true,
			service: s,
		}
	}
	//new nodes must be added for some special services types
	newHost := false

	//number of old services
	oldSN := len(cluster.Services)

	for _, s := range newC.Services {
		if serviceTypesOld[s.Type].exists == false {
			sUuid, err := uuid.NewRandom()
			if err != nil {
				mess, _ := hS.ErrHandler.Handle(w, LibErrorUUID, LibErrorUUIDMessage, nil)
				hS.Logger.Print(mess)
				return
			}
			s.ID = sUuid.String()
			cluster.Services = append(cluster.Services, s)
		}

		st, err := hS.Db.ReadServiceType(s.Type)
		if err != nil {
			hS.Logger.Print(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if st.Class == utils.ClassStorage {
			newHost = true
		}
	}

	//check if new services are added
	if oldSN != len(cluster.Services) {
		retryFlag := true
		startIdx := oldSN

		//first for cycle is used for updating range values with appended services
		for retryFlag {
			for i, s := range cluster.Services[startIdx:] {
				st, err := hS.Db.ReadServiceType(s.Type)
				if err != nil {
					mess, _ := hS.ErrHandler.Handle(w, DBerror, DBerrorMessage, nil)
					hS.Logger.Print(mess)
					return
				}

				if s.Version == "" {
					s.Version = st.DefaultVersion
				}

				//add services from dependencies
				sToAdd, err := hS.AddDependencies(cluster, s)
				if err != nil {
					hS.Logger.Println(err)
					mess, _ := hS.ErrHandler.Handle(w, UserErrorBadServiceVersion, UserErrorBadServiceVersionMessage, nil)
					hS.Logger.Print(mess)
					return
				}
				hS.Logger.Println(sToAdd)

				changesFlag := false
				if sToAdd != nil {
					for _, curS := range sToAdd {
						sUuid, err := uuid.NewRandom()
						if err != nil {
							mess, _ := hS.ErrHandler.Handle(w, LibErrorUUID, LibErrorUUIDMessage, nil)
							hS.Logger.Print(mess)
							return
						}
						curS.ID = sUuid.String()

						cluster.Services = append(cluster.Services, curS)
					}
					changesFlag = true
				}

				if !changesFlag {
					retryFlag = false
				} else {
					//update range values if new services has been added and start new iteration from the next value
					startIdx = i + 1
					break
				}

			}
		}
	}

	if newC.Description != "" {
		cluster.Description = newC.Description
	}

	cluster.EntityStatus = utils.StatusInited
	if newC.NHosts != 0 || newHost {
		go hS.Gc.StartClusterCreation(cluster)
	} else {
		go hS.Gc.StartClusterModification(cluster)
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(cluster)
}

func (hS HttpServer) ClustersDelete(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	projectIdOrName := params.ByName("projectIdOrName")
	clusterIdOrName := params.ByName("clusterIdOrName")
	hS.Logger.Print("Get /projects/"+projectIdOrName+"/clusters/", clusterIdOrName, " DELETE")

	project, err := hS.getProject(projectIdOrName)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, DBerror, DBerrorMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	if project.Name == "" {
		hS.Logger.Printf("Project with name '%s' not found", projectIdOrName)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// reading cluster info from database
	cluster, err := hS.getCluster(project.ID, clusterIdOrName)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, DBerror, DBerrorMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	if cluster.Name == "" {
		hS.Logger.Printf("Cluster with name or ID '%s' not found", clusterIdOrName)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if cluster.EntityStatus != utils.StatusActive && cluster.EntityStatus != utils.StatusFailed {
		mess, _ := hS.ErrHandler.Handle(w, UserErrorClusterStatus, UserErrorClusterStatusMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	cluster.EntityStatus = utils.StatusStopping

	go hS.Gc.StartClusterDestroying(cluster)
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(cluster)
}
