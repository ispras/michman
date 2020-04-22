package handlers

import (
	"encoding/json"
	"github.com/google/uuid"
	proto "github.com/ispras/michman/src/protobuf"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"regexp"
)

func ValidateProject(project *proto.Project) bool {
	validName := regexp.MustCompile(`^[A-Za-z][A-Za-z0-9-]+$`).MatchString

	if !validName(project.DisplayName) {
		return false
	}
	return true
}

func (hS HttpServer) ProjectsGetList(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	hS.Logger.Print("Get /projects GET")

	hS.Logger.Print("Reading projects information from db...")
	projects, err := hS.Db.ListProjects()
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, DBerror, DBerrorMessage, err)
		hS.Logger.Print(mess)
		return
	}

	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	err = enc.Encode(projects)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, LibErrorStructToJson, LibErrorStructToJsonMessage, err)
		hS.Logger.Print(mess)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

func (hS HttpServer) getProject(idORname string) (*proto.Project, error) {
	is_uuid := true
	_, err := uuid.Parse(idORname)
	if err != nil {
		is_uuid = false
	}

	var project *proto.Project

	if is_uuid {
		project, err = hS.Db.ReadProject(idORname)
	} else {
		project, err = hS.Db.ReadProjectByName(idORname)
	}

	return project, err
}

func (hS HttpServer) ProjectCreate(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	hS.Logger.Print("Get /projects POST")

	var p proto.Project
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, JSONerrorIncorrect, JSONerrorIncorrectMessage, err)
		hS.Logger.Print(mess)
		return
	}

	if p.DisplayName == "" {
		mess, _ := hS.ErrHandler.Handle(w, JSONerrorMissField, JSONerrorMissFieldMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	if !ValidateProject(&p) {
		mess, _ := hS.ErrHandler.Handle(w, JSONerrorIncorrectField, JSONerrorIncorrectFieldMessage, nil)
		hS.Logger.Print(mess)
		return
	}
	p.Name = p.DisplayName
	//check, that project with such name doesn't exist
	dbRes, err := hS.Db.ReadProjectByName(p.Name)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, DBerror, DBerrorMessage, err)
		hS.Logger.Print(mess)
		return
	}

	if dbRes.Name != "" {
		hS.Logger.Print("Project with this name exists")
		w.WriteHeader(http.StatusBadRequest)
		enc := json.NewEncoder(w)
		w.Header().Set("Content-Type", "application/json")
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
		mess, _ := hS.ErrHandler.Handle(w, LibErrorUUID, LibErrorUUIDMessage, err)
		hS.Logger.Print(mess)
		return
	}
	p.ID = pUuid.String()

	err = hS.Db.WriteProject(&p)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, DBerror, DBerrorMessage, err)
		hS.Logger.Print(mess)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(p)
}

func (hS HttpServer) ProjectGetByName(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	projectIdOrName := params.ByName("projectIdOrName")
	hS.Logger.Print("Get /projects/", projectIdOrName, " GET")

	hS.Logger.Print("Reading project information from db...")

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

	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	err = enc.Encode(project)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, DBerror, DBerrorMessage, err)
		hS.Logger.Print(mess)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

func (hS HttpServer) ProjectUpdate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	projectIdOrName := params.ByName("projectIdOrName")
	hS.Logger.Print("Get /projects/", projectIdOrName, " PUT")

	project, err := hS.getProject(projectIdOrName)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, DBerror, DBerrorMessage, err)
		hS.Logger.Print(mess)
		return
	}

	if project.Name == "" {
		hS.Logger.Printf("Project with name or id '%s' not found", projectIdOrName)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var p proto.Project
	err = json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, JSONerrorIncorrect, JSONerrorIncorrectFieldMessage, err)
		hS.Logger.Print(mess)
		return
	}

	if p.Name != "" || p.ID != "" || p.GroupID != 0 || p.DisplayName != "" {
		mess, _ := hS.ErrHandler.Handle(w, UserErrorProjectUnmodField, UserErrorProjectUnmodFieldMessage, err)
		hS.Logger.Print(mess)
		return
	}

	if p.Description != "" {
		project.Description = p.Description
	}

	err = hS.Db.UpdateProject(project)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, DBerror, DBerrorMessage, err)
		hS.Logger.Print(mess)
		return
	}

	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	err = enc.Encode(project)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, LibErrorStructToJson, LibErrorStructToJsonMessage, err)
		hS.Logger.Print(mess)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

func (hS HttpServer) ProjectDelete(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	projectIdOrName := params.ByName("projectIdOrName")
	hS.Logger.Print("Get /projects/", projectIdOrName, " DELETE")

	//reading project info from database
	hS.Logger.Print("Reading project information from db...")
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

	clusters, err := hS.Db.ReadProjectClusters(project.ID)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, DBerror, DBerrorMessage, err)
		hS.Logger.Print(mess)
		return
	}

	if len(clusters) > 0 {
		mess, _ := hS.ErrHandler.Handle(w, UserErrorProjectWithClustersDel, UserErrorProjectWithClustersDelMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	err = hS.Db.DeleteProject(project.ID)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, DBerror, DBerrorMessage, nil)
		hS.Logger.Print(mess)
		return
	}

	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	err = enc.Encode(project)
	if err != nil {
		mess, _ := hS.ErrHandler.Handle(w, LibErrorStructToJson, LibErrorStructToJsonMessage, nil)
		hS.Logger.Print(mess)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}
