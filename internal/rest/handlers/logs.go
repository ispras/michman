package handlers

import (
	cluster_logger "github.com/ispras/michman/internal/logger"
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

func (hS HttpServer) ServeAnsibleServiceLog(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	request := "logs/launcher GET"
	hS.Logger.Info("Get " + request)

	path := MakeLogFilePath(utils.LauncherLogFileName, hS.Config.LogsFilePath)

	if exist, err := CheckFileExists(path); !exist || err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	http.ServeFile(w, r, path)
}

func (hS HttpServer) ServeHttpServerLog(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	request := "logs/http_server GET"
	hS.Logger.Info("Get " + request)

	path := MakeLogFilePath(utils.HttpLogFileName, hS.Config.LogsFilePath)

	if exist, err := CheckFileExists(path); !exist || err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	http.ServeFile(w, r, path)
}

func (hS HttpServer) ServeClusterLog(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	clusterIdOrName := params.ByName("clusterIdOrName")
	projectIdOrName := params.ByName("projectIdOrName")
	request := "logs/project/" + projectIdOrName + "/clusters/" + clusterIdOrName + " GET"
	hS.Logger.Info("Get " + request)

	project, err := hS.Db.ReadProject(projectIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}
	if project.Name == "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrProjectNotFound.Error())
		ResponseBadRequest(w, ErrProjectNotFound)
		return
	}

	// reading cluster info from database
	cluster, err := hS.Db.ReadCluster(project.ID, clusterIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}
	if cluster.Name == "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrClusterNotFound.Error())
		ResponseBadRequest(w, ErrClusterNotFound)
		return
	}

	queryValues := r.URL.Query()
	action := utils.ActionCreate

	if tmpAction := queryValues.Get(respActionKey); tmpAction != "" {
		if tmpAction == utils.ActionCreate || tmpAction == utils.ActionDelete || tmpAction == utils.ActionUpdate {
			action = tmpAction
		} else {
			hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrLogsBadActionParam.Error())
			ResponseBadRequest(w, ErrLogsBadActionParam)
			return
		}
	}

	//initialize cluster logger
	cLogger, err := cluster_logger.MakeNewClusterLogger(hS.Config, cluster.ID, action)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
		return
	}

	clusterLogs, err := cLogger.ReadClusterLogs()
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		ResponseInternalError(w, err)
	}
	resp := clusterLog{ClusterIdOrName: clusterIdOrName, Action: action, ClusterLogs: clusterLogs}
	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	ResponseOK(w, resp, request)
}
