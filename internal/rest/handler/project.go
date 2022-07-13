package handler

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/rest/handler/response"
	"github.com/ispras/michman/internal/rest/handler/validate"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (hS HttpServer) ProjectsGetList(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	request := "/projects GET"
	hS.Logger.Info("Get " + request)

	projects, err := hS.Db.ReadProjectsList()
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		response.InternalError(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	response.Ok(w, projects, request)
}

func (hS HttpServer) ProjectCreate(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	request := "/projects POST"
	hS.Logger.Info("Get " + request)

	var project protobuf.Project
	err := json.NewDecoder(r.Body).Decode(&project)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrJsonIncorrect.Error())
		response.BadRequest(w, ErrJsonIncorrect)
		return
	}

	err, status := validate.ProjectCreate(hS.Db, hS.Logger, &project)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", status, ": ", err.Error())
		switch status {
		case http.StatusBadRequest:
			response.BadRequest(w, err)
			return
		case http.StatusInternalServerError:
			response.InternalError(w, err)
			return
		}
	}

	project.Name = project.DisplayName

	pUuid, err := uuid.NewRandom()
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", ErrUuidLibError.Error())
		response.InternalError(w, ErrUuidLibError)
		return
	}
	project.ID = pUuid.String()

	err = hS.Db.WriteProject(&project)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		response.InternalError(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusCreated)
	response.Created(w, project, request)
}

func (hS HttpServer) ProjectGet(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	projectIdOrName := params.ByName("projectIdOrName")
	request := "/projects/" + projectIdOrName + " GET"
	hS.Logger.Info("Get " + request)

	project, err := hS.Db.ReadProject(projectIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		response.InternalError(w, err)
		return
	}
	if project.Name == "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusNotFound, ": ", ErrProjectNotFound.Error())
		response.NotFound(w, ErrProjectNotFound)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	response.Ok(w, project, request)
}

func (hS HttpServer) ProjectUpdate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	projectIdOrName := params.ByName("projectIdOrName")
	request := "/projects/" + projectIdOrName + " PUT"
	hS.Logger.Info("Get " + request)

	oldProj, err := hS.Db.ReadProject(projectIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		response.InternalError(w, err)
		return
	}
	if oldProj.Name == "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusNotFound, ": ", ErrProjectNotFound.Error())
		response.NotFound(w, ErrProjectNotFound)
		return
	}

	resProj := oldProj

	var newProj protobuf.Project
	err = json.NewDecoder(r.Body).Decode(&newProj)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusBadRequest, ": ", ErrJsonIncorrect.Error())
		response.BadRequest(w, ErrJsonIncorrect)
		return
	}

	err, status := validate.ProjectUpdate(hS.Db, hS.Logger, &newProj)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", status, ": ", err.Error())
		switch status {
		case http.StatusBadRequest:
			response.BadRequest(w, err)
			return
		case http.StatusInternalServerError:
			response.InternalError(w, err)
			return
		}
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
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		response.InternalError(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusOK)
	response.Ok(w, resProj, request)
}

func (hS HttpServer) ProjectDelete(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	projectIdOrName := params.ByName("projectIdOrName")
	request := "/projects/" + projectIdOrName + " DELETE"
	hS.Logger.Info("Get " + request)

	project, err := hS.Db.ReadProject(projectIdOrName)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		response.InternalError(w, err)
		return
	}
	if project.Name == "" {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusNotFound, ": ", ErrProjectNotFound.Error())
		response.NotFound(w, ErrProjectNotFound)
		return
	}

	clusters, err := hS.Db.ReadProjectClusters(project.ID)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		response.InternalError(w, err)
		return
	}
	if len(clusters) > 0 {
		hS.Logger.Warn("Request ", request, " failed  with status ", http.StatusBadRequest, ": ", ErrProjectHasClusters.Error())
		response.BadRequest(w, ErrProjectHasClusters)
		return
	}

	err = hS.Db.DeleteProject(project.ID)
	if err != nil {
		hS.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
		response.InternalError(w, err)
		return
	}

	hS.Logger.Info("Request ", request, " has succeeded with status ", http.StatusNoContent)
	response.NoContent(w)
}
