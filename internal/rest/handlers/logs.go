package handlers

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"encoding/json"
	"github.com/ispras/michman/internal/utils"
	cluster_logger "github.com/ispras/michman/internal/logger"
)

const (
	respActionKey = "action"
)

type clusterLog struct {
	ClusterId string `json:"cluster_id"`
	Action string `json:"action"`
    ClusterLogs string `json:"cluster_logs"`
}

func (hS HttpServer) ServeAnsibleOutput(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	hS.Logger.Print(r.URL.Path)
	hS.Logger.Print("Request to serve logs/ansible_output.log")
	p := "./logs/ansible_output.log"
	http.ServeFile(w, r, p)
}

func (hS HttpServer) ServeAnsibleServiceLog(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	hS.Logger.Print(r.URL.Path)
	hS.Logger.Print("Request to serve logs/launcher.log")
	p := "./logs/launcher.log"
	http.ServeFile(w, r, p)
}

func (hS HttpServer) ServeHttpServerLog(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	hS.Logger.Print(r.URL.Path)
	hS.Logger.Print("Request to serve logs/http_server.log")
	p := "./logs/http_server.log"
	http.ServeFile(w, r, p)
}

func (hS HttpServer) ServeHttpServerLogstash(w http.ResponseWriter, r *http.Request, params httprouter.Params) (){
	clusterID:= params.ByName("clusterID")
	projectIdOrName := params.ByName("projectIdOrName")
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
	cluster, err := hS.getCluster(project.ID, clusterID)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, DBerror, DBerrorMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	if cluster.Name == "" {
		hS.Logger.Printf("Cluster with name or ID '%s' not found", clusterID)
		w.WriteHeader(http.StatusBadRequest)
		_, err = w.Write([]byte(utils.StatusMissing))
		if err != nil {
			hS.Logger.Print(err)
			w.WriteHeader(http.StatusBadRequest)
		}
		return
	}

	queryValues := r.URL.Query()
	action := utils.ActionCreate

	if a := queryValues.Get(respActionKey); a != "" {
		if a == utils.ActionCreate || a == utils.ActionDelete || a == utils.ActionUpdate {
			action = a
		} else {
			hS.Logger.Print("Error: bad action param. Supported query variables for action parameter are 'create', 'update' and 'delete', 'create' is default.")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	//initialize cluster logger
	cLogger, err := cluster_logger.MakeNewClusterLogger(hS.Config, cluster.ID, action)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	clusterLogs, err := cLogger.ReadClusterLogs()
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
	}
	resp := clusterLog{ClusterId: clusterID, Action: action, ClusterLogs: clusterLogs}
	enc := json.NewEncoder(w)
	err = enc.Encode(resp)
}
