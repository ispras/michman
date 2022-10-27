package handler

import (
	"encoding/json"
	proto "github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/rest/handler/check"
	"github.com/ispras/michman/internal/rest/handler/helpfunc"
	"github.com/ispras/michman/internal/rest/handler/validate"
	response "github.com/ispras/michman/internal/rest/response"
	"github.com/ispras/michman/internal/utils"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

// ClustersGetList processes a request to get a list of all clusters in database
func (hS HttpServer) ClustersGetList(w http.ResponseWriter, _ *http.Request, params httprouter.Params) {
	projectIdOrName := params.ByName("projectIdOrName")
	request := "GET /projects/" + projectIdOrName + "/clusters"
	hS.Logger.Info(request)

	// reading project info from database
	project, err := hS.Db.ReadProject(projectIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	clusters, err := hS.Db.ReadProjectClusters(project.ID)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	response.Ok(w, clusters, request)
}

// ClusterCreate processes a request to create a cluster struct in database
func (hS HttpServer) ClusterCreate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	projectIdOrName := params.ByName("projectIdOrName")
	request := "POST /project/" + projectIdOrName + "/clusters"
	hS.Logger.Info(request)

	// reading project info from database
	project, err := hS.Db.ReadProject(projectIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	var resCluster *proto.Cluster
	err = json.NewDecoder(r.Body).Decode(&resCluster)
	if err != nil {
		err = ErrJsonIncorrect
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	// set fields by defaults if not specified by user
	helpfunc.SetClusterDefaults(resCluster, project)

	hS.Logger.Infof("Validating cluster %s general info...", resCluster.Name)
	err = validate.ClusterCreateGeneral(hS.Db, resCluster)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	// check, that cluster with such name doesn't exist
	clusterExists, oldCluster, retErr := check.ClusterExist(hS.Db, resCluster, project)
	if retErr != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", retErr.Error())
		response.Error(w, retErr)
		return
	}
	// If cluster was failed
	if clusterExists {
		resCluster = oldCluster
	} else {
		// Set ID, ProjectID, Name for new cluster
		err := helpfunc.SetClusterGeneratedFields(resCluster, project)
		if err != nil {
			hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
			response.Error(w, err)
			return
		}
		// Set OwnerID from the request
		resCluster.OwnerID = helpfunc.GetClusterOwnerId(r)

		// add services from user request and from dependencies
		if err := helpfunc.SetServices(hS.Db, resCluster); err != nil {
			hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
			response.Error(w, err)
			return
		}
		// cluster should be validated after addition services from dependencies
		err = validate.ClusterServices(hS.Db, resCluster)
		if err != nil {
			hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
			response.Error(w, err)
			return
		}
	}

	hS.Logger.Info("validate services after adding service dependencies...")
	sErr := validate.ClusterCreateServices(hS.Db, resCluster)
	if sErr != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", sErr.Error())
		response.Error(w, err)
		return
	}
	resCluster.EntityStatus = utils.StatusInited

	if !clusterExists {
		err = hS.Db.WriteCluster(resCluster)
		if err != nil {
			hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
			response.Error(w, err)
			return
		}
	}
	go hS.Gc.StartClusterCreation(resCluster)

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusCreated)
	response.Created(w, resCluster, request)
}

// ClusterGet processes a request to get a cluster struct by id or name from database
func (hS HttpServer) ClusterGet(w http.ResponseWriter, _ *http.Request, params httprouter.Params) {
	projectIdOrName := params.ByName("projectIdOrName")
	clusterIdOrName := params.ByName("clusterIdOrName")
	request := "GET /projects/" + projectIdOrName + "/clusters/" + clusterIdOrName
	hS.Logger.Info(request)

	// reading project info from database
	project, err := hS.Db.ReadProject(projectIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	// reading cluster info from database
	cluster, err := hS.Db.ReadCluster(project.ID, clusterIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	response.Ok(w, cluster, request)
}

// ClusterStatusGet processes a request to get a cluster status message by id or name from database
func (hS HttpServer) ClusterStatusGet(w http.ResponseWriter, _ *http.Request, params httprouter.Params) {
	projectIdOrName := params.ByName("projectIdOrName")
	clusterIdOrName := params.ByName("clusterIdOrName")
	request := "GET /projects/" + projectIdOrName + "/clusters/" + clusterIdOrName + "/status"
	hS.Logger.Info(request)

	// reading project info from database
	project, err := hS.Db.ReadProject(projectIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	// reading cluster info from database
	cluster, err := hS.Db.ReadCluster(project.ID, clusterIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	response.Ok(w, cluster.EntityStatus, request)
}

// ClustersUpdate processes a request to update a cluster struct in database
func (hS HttpServer) ClustersUpdate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	projectIdOrName := params.ByName("projectIdOrName")
	clusterIdOrName := params.ByName("clusterIdOrName")
	request := "PUT /projects/" + projectIdOrName + "/clusters/" + clusterIdOrName
	hS.Logger.Info(request)

	// reading project info from database
	project, err := hS.Db.ReadProject(projectIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	// reading cluster info from database
	oldCluster, err := hS.Db.ReadCluster(project.ID, clusterIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	var newCluster proto.Cluster
	err = json.NewDecoder(r.Body).Decode(&newCluster)
	if err != nil {
		err = ErrJsonIncorrect
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	hS.Logger.Info("Validating updated values of the cluster fields...")
	err = validate.ClusterUpdate(hS.Db, oldCluster, &newCluster)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	resCluster := oldCluster

	// set existed services
	serviceTypesOld, oldServiceNumber, err := helpfunc.SetServiceExistInfo(hS.Db, oldCluster)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	// append new services to the resCluster struct
	newHost, err := helpfunc.AppendNewServices(hS.Db, serviceTypesOld, &newCluster, resCluster)
	if err != nil {
		err = ErrUuidLibError
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	// check if new services are added
	if oldServiceNumber != len(resCluster.Services) {
		// updating range values of appended services
		err = helpfunc.UpdateRangeValuesAppendedServices(hS.Db, oldServiceNumber, resCluster, utils.ActionUpdate)
		if err != nil {
			hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
			response.Error(w, err)
			return
		}
	}

	// cluster should be validated after addition services from dependencies
	err = validate.ClusterServices(hS.Db, resCluster)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	if newCluster.Description != "" {
		resCluster.Description = newCluster.Description
	}
	if newCluster.DisplayName != "" {
		resCluster.DisplayName = newCluster.DisplayName
	}
	if newCluster.Keys != nil {
		for _, key := range newCluster.Keys {
			if !utils.ItemExists(resCluster.Keys, key) {
				resCluster.Keys = append(resCluster.Keys, key)
			}
		}
	}

	resCluster.EntityStatus = utils.StatusInited
	if newCluster.NHosts != 0 || newHost {
		go hS.Gc.StartClusterCreation(resCluster)
	} else {
		go hS.Gc.StartClusterModification(resCluster)
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	response.Ok(w, resCluster, request)
}

// ClustersDelete processes a request to delete a cluster struct from database
func (hS HttpServer) ClustersDelete(w http.ResponseWriter, _ *http.Request, params httprouter.Params) {
	projectIdOrName := params.ByName("projectIdOrName")
	clusterIdOrName := params.ByName("clusterIdOrName")
	request := "DELETE /projects/" + projectIdOrName + "/clusters/" + clusterIdOrName
	hS.Logger.Info(request)

	// reading project info from database
	project, err := hS.Db.ReadProject(projectIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	// reading cluster info from database
	cluster, err := hS.Db.ReadCluster(project.ID, clusterIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	err = validate.ClusterDelete(cluster)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	cluster.EntityStatus = utils.StatusStopping

	go hS.Gc.StartClusterDestroying(cluster)

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	response.Ok(w, cluster, request)
}
