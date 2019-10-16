package handlers

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	//"gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/database"
	protobuf "gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/protobuf"
	//"gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/utils"
	"net/http"
	//"regexp"
)

func (hS HttpServer) ProjectsGetList(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	hS.Logger.Print("Get /projects GET")
	//reading cluster info from database
	hS.Logger.Print("Reading projects information from db...")

	projects, err := hS.Db.ListProjects()
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	err = enc.Encode(projects)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (hS HttpServer) ProjectCreate(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	hS.Logger.Print("Get /projects POST")
	var p protobuf.Project
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		hS.Logger.Print("ERROR:")
		hS.Logger.Print(err)
		hS.Logger.Print(r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// validate struct
	/*
		if !ValidateCluster(&c) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	*/

	//check, that cluster with such name doesn't exist
	dbRes, err := hS.Db.ReadProject(p.Name)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if dbRes.Name != "" {
		hS.Logger.Print("Project with this name exists")
		w.WriteHeader(http.StatusBadRequest)
		enc := json.NewEncoder(w)
		err := enc.Encode("Project with this name exists")
		if err != nil {
			hS.Logger.Print(err)
			w.WriteHeader(http.StatusBadRequest)
		}
		return
	}

	// generating UUID for new project
	pUuid, err := uuid.NewRandom()
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
	}
	p.ID = pUuid.String()

	err = hS.Db.WriteProject(&p)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(p)
}

func (hS HttpServer) ProjectGetByName(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	projectName := params.ByName("projectName")
	hS.Logger.Print("Get /projects/", projectName, " GET")

	//reading cluster info from database
	hS.Logger.Print("Reading project information from db...")
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

	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	err = enc.Encode(project)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (hS HttpServer) ProjectUpdate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	projectName := params.ByName("projectName")
	hS.Logger.Print("Get /projects/", projectName, " PUT")

	//reading project info from database
	hS.Logger.Print("Reading project information from db...")
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

	var p protobuf.Project
	err = json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		hS.Logger.Print("ERROR:")
		hS.Logger.Print(err)
		hS.Logger.Print(r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if p.Name != "" || p.ID != "" || p.GroupID != 0 {
		w.WriteHeader(http.StatusBadRequest)
		enc := json.NewEncoder(w)
		err = enc.Encode("This fields cannot be updated")
		if err != nil {
			hS.Logger.Print(err)
			w.WriteHeader(http.StatusBadRequest)
		}
		return
	}

	if p.Description != "" {
		project.Description = p.Description
	}

	if p.DisplayName != "" {
		project.DisplayName = p.DisplayName
	}

	err = hS.Db.UpdateProject(project)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	err = enc.Encode(project)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (hS HttpServer) ProjectDelete(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	projectName := params.ByName("projectName")
	hS.Logger.Print("Get /projects/", projectName, " DELETE")

	//reading project info from database
	hS.Logger.Print("Reading project information from db...")
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

	if len(clusters) > 0 {
		w.WriteHeader(http.StatusBadRequest)
		enc := json.NewEncoder(w)
		err = enc.Encode("Project has already had clusters. Delete them first")
		if err != nil {
			hS.Logger.Print(err)
			w.WriteHeader(http.StatusBadRequest)
		}
		return
	}

	err = hS.Db.DeleteProject(projectName)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	err = enc.Encode(project)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
