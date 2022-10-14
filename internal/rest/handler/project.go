package handler

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/rest/handler/validate"
	response "github.com/ispras/michman/internal/rest/response"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

// ProjectsGetList processes a request to get a list of all projects in database
func (hS HttpServer) ProjectsGetList(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	request := "GET /projects"
	hS.Logger.Info(request)

	projects, err := hS.Db.ReadProjectsList()
	if err != nil {
		hS.Logger.Warn("Request ", request, "failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	response.Ok(w, projects, request)
}

// ProjectCreate processes a request to create a project struct in database
func (hS HttpServer) ProjectCreate(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	request := "POST /projects"
	hS.Logger.Info(request)

	var project protobuf.Project
	err := json.NewDecoder(r.Body).Decode(&project)
	if err != nil {
		err = ErrJsonIncorrect
		hS.Logger.Warn("Request ", request, "failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}
	if err := validate.ProjectCreate(hS.Db, &project); err != nil {
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}
	// Generate UUID for project ID
	pUuid, uErr := uuid.NewRandom()
	if uErr != nil {
		err := ErrUuidLibError
		hS.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}
	project.ID = pUuid.String()
	project.Name = project.DisplayName
	// Write new project
	err = hS.Db.WriteProject(&project)
	if err != nil {
		hS.Logger.Warn("Request ", request, "failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusCreated)
	response.Created(w, project, request)
}

// ProjectGet processes a request to get a project struct by id or name from database
func (hS HttpServer) ProjectGet(w http.ResponseWriter, _ *http.Request, params httprouter.Params) {
	projectIdOrName := params.ByName("projectIdOrName")
	request := "GET /projects/" + projectIdOrName
	hS.Logger.Info(request)

	project, err := hS.Db.ReadProject(projectIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, "failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	response.Ok(w, project, request)
}

// ProjectUpdate processes a request to update a project struct in database
func (hS HttpServer) ProjectUpdate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	projectIdOrName := params.ByName("projectIdOrName")
	request := "PUT /projects/" + projectIdOrName
	hS.Logger.Info(request)

	oldProj, err := hS.Db.ReadProject(projectIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, "failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	resProj := oldProj

	var newProj protobuf.Project
	err = json.NewDecoder(r.Body).Decode(&newProj)
	if err != nil {
		err = ErrJsonIncorrect
		hS.Logger.Warn("Request ", request, "failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	hS.Logger.Info("Validating updated values of the project fields...")
	err = validate.ProjectUpdate(hS.Db, &newProj)
	if err != nil {
		hS.Logger.Warn("Request ", request, "failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	if newProj.GroupID != "" {
		resProj.GroupID = newProj.GroupID
	}
	if newProj.Description != "" {
		resProj.Description = newProj.Description
	}
	if newProj.DisplayName != "" {
		resProj.DisplayName = newProj.DisplayName
	}
	if newProj.DefaultImage != "" {
		resProj.DefaultImage = newProj.DefaultImage
	}
	if newProj.DefaultMasterFlavor != "" {
		resProj.DefaultMasterFlavor = newProj.DefaultMasterFlavor
	}
	if newProj.DefaultSlavesFlavor != "" {
		resProj.DefaultSlavesFlavor = newProj.DefaultSlavesFlavor
	}
	if newProj.DefaultStorageFlavor != "" {
		resProj.DefaultStorageFlavor = newProj.DefaultStorageFlavor
	}
	if newProj.DefaultMonitoringFlavor != "" {
		resProj.DefaultMonitoringFlavor = newProj.DefaultMonitoringFlavor
	}

	err = hS.Db.UpdateProject(resProj)
	if err != nil {
		hS.Logger.Warn("Request ", request, "failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	response.Ok(w, resProj, request)
}

// ProjectDelete processes a request to delete a project struct from database
func (hS HttpServer) ProjectDelete(w http.ResponseWriter, _ *http.Request, params httprouter.Params) {
	projectIdOrName := params.ByName("projectIdOrName")
	request := "DELETE /projects/" + projectIdOrName
	hS.Logger.Info(request)

	project, err := hS.Db.ReadProject(projectIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, "failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	err = validate.ProjectDelete(hS.Db, project)
	if err != nil {
		hS.Logger.Warn("Request ", request, "failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	err = hS.Db.DeleteProject(project.ID)
	if err != nil {
		hS.Logger.Warn("Request ", request, "failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusNoContent)
	response.NoContent(w)
}
