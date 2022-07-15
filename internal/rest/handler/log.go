package handler

import (
	clusterlogger "github.com/ispras/michman/internal/logger"
	"github.com/ispras/michman/internal/rest/handler/check"
	"github.com/ispras/michman/internal/rest/handler/helpfunc"
	"github.com/ispras/michman/internal/rest/handler/response"
	"github.com/ispras/michman/internal/utils"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

const (
	respActionKey = "action"
)

type clusterLog struct {
	ClusterIdOrName string `json:"cluster_id"`
	Action          string `json:"action"`
	ClusterLogs     string `json:"cluster_logs"`
}

// ServeAnsibleServiceLog processes the request to get the launcher.log file on the path specified in the configuration file
func (hS HttpServer) ServeAnsibleServiceLog(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	request := "GET logs/launcher"
	hS.Logger.Info(request)

	path := helpfunc.MakeLogFilePath(utils.LauncherLogFileName, hS.Config.LogsFilePath)

	if exist, err := check.FileExists(path); !exist || err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		response.InternalError(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	http.ServeFile(w, r, path)
}

// ServeHttpServerLog processes the request to get the http.log file on the path specified in the configuration file
func (hS HttpServer) ServeHttpServerLog(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	request := "GET logs/http_server"
	hS.Logger.Info(request)

	path := helpfunc.MakeLogFilePath(utils.HttpLogFileName, hS.Config.LogsFilePath)

	if exist, err := check.FileExists(path); !exist || err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		response.InternalError(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	http.ServeFile(w, r, path)
}

// ServeClusterLog processes the request to get cluster logs on the path specified in the configuration file,
// depending on the action specified in the request(create, delete, update).
// by default create action
func (hS HttpServer) ServeClusterLog(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	clusterIdOrName := params.ByName("clusterIdOrName")
	projectIdOrName := params.ByName("projectIdOrName")
	request := "GET logs/project/" + projectIdOrName + "/clusters/" + clusterIdOrName
	hS.Logger.Info(request)

	// reading project info from database
	project, err := hS.Db.ReadProject(projectIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		response.InternalError(w, err)
		return
	}
	if project.Name == "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrProjectNotFound.Error())
		response.BadRequest(w, ErrProjectNotFound)
		return
	}

	// reading cluster info from database
	cluster, err := hS.Db.ReadCluster(project.ID, clusterIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		response.InternalError(w, err)
		return
	}
	if cluster.Name == "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrClusterNotFound.Error())
		response.BadRequest(w, ErrClusterNotFound)
		return
	}

	queryValues := r.URL.Query()
	action := utils.ActionCreate

	// checking the action field from the request for compliance with: create, delete, update
	if tmpAction := queryValues.Get(respActionKey); tmpAction != "" {
		if tmpAction == utils.ActionCreate || tmpAction == utils.ActionDelete || tmpAction == utils.ActionUpdate {
			action = tmpAction
		} else {
			hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrLogsBadActionParam.Error())
			response.BadRequest(w, ErrLogsBadActionParam)
			return
		}
	}

	// initialize cluster logger
	cLogger, err := clusterlogger.MakeNewClusterLogger(hS.Config, cluster.ID, action)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		response.InternalError(w, err)
		return
	}

	// read cluster logs from file or logstash (depending on 'logs_output' in configuration file)
	clusterLogs, err := cLogger.ReadClusterLogs()
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		response.InternalError(w, err)
	}

	resp := clusterLog{ClusterIdOrName: clusterIdOrName, Action: action, ClusterLogs: clusterLogs}
	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	response.Ok(w, resp, request)
}
