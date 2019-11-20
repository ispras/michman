package handlers

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	protobuf "gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/protobuf"
	"gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/utils"
	"net/http"
)

func (hS HttpServer) TemplateCreate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	hS.Logger.Print("Get /templates or /project/projectIdOrName/templates POST")

	var t protobuf.Template
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		hS.Logger.Print("ERROR:")
		hS.Logger.Print(err)
		hS.Logger.Print(r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// checking do we get certain projectID or not
	projectID := params.ByName("projectIdOrName")
	var projectName string
	if projectID == "" {
		// get common templates
		projectID = utils.CommonProjectID
		projectName = "common"
	} else {
		// try to get project with such ID from DB
		project, err := hS.Db.ReadProject(projectID)
		if err != nil {
			hS.Logger.Print(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if project.ID == "" {
			hS.Logger.Print("no project with such id ", projectID)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		projectName = project.Name
	}

	tUuid, err := uuid.NewRandom()
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	t.ID = tUuid.String()
	t.ProjectID = projectID
	t.Name = t.DisplayName + "-" + projectName

	//check, that template with such Name doesn't exist
	dbTemplate, err := hS.Db.ReadTemplateByName(t.Name)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if dbTemplate.ID != "" {
		hS.Logger.Print("template with this Name already exists")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = hS.Db.WriteTemplate(&t)

	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(t)
}

func (hS HttpServer) TemplateUpdate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	hS.Logger.Print("Get /templates/templateID or /projects/projectIdOrName/templates PUT")
	templateID := params.ByName("templateID")
	var t protobuf.Template
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		hS.Logger.Print("ERROR:")
		hS.Logger.Print(err)
		hS.Logger.Print(r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// check immutable field
	if t.ID != "" || t.ProjectID != "" || t.Name != "" {
		hS.Logger.Print("got immutable fields")
		w.WriteHeader(http.StatusBadRequest)
	}

	// checking do we get certain projectID or not
	projectID := params.ByName("projectIdOrName")
	if projectID == "" {
		// get common templates
		projectID = utils.CommonProjectID
	} else {
		// try to get project with such ID from DB
		project, err := hS.Db.ReadProject(projectID)
		if err != nil {
			hS.Logger.Print(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if project.ID == "" {
			hS.Logger.Print("no project with such id ", projectID)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	//check, that template with such ID exists
	dbTemplate, err := hS.Db.ReadTemplate(projectID, templateID)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if dbTemplate.ID == "" {
		hS.Logger.Print("Template with this ID doesnt exist")
		w.WriteHeader(http.StatusBadRequest)
		enc := json.NewEncoder(w)
		w.Header().Set("Content-Type", "application/json")
		err := enc.Encode("Template with this ID doesnt exists")
		if err != nil {
			hS.Logger.Print(err)
			w.WriteHeader(http.StatusBadRequest)
		}
		return
	}

	//update fields
	dbTemplate.DisplayName = t.DisplayName
	dbTemplate.Services = t.Services
	dbTemplate.NHosts = t.NHosts
	dbTemplate.Description = t.Description

	err = hS.Db.WriteTemplate(dbTemplate)

	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(t)
}

func (hS HttpServer) TemplateDelete(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	templateID := params.ByName("templateID")
	hS.Logger.Print("Get /templates/", templateID, "or /projects/projectIdOrName/templates/templateID DELETE")

	//check that template exists
	hS.Logger.Print("Check that template exists...")

	// checking do we get certain projectID or not
	projectID := params.ByName("projectIdOrName")
	if projectID == "" {
		// get common templates
		projectID = utils.CommonProjectID
	}

	t, err := hS.Db.ReadTemplate(projectID, templateID)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if t.Name == "" {
		hS.Logger.Print("Template not found")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	err = hS.Db.DeleteTemplate(templateID)

	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(t)
}

func (hS HttpServer) TemplatesGetList(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	hS.Logger.Print("Get /templates or /projects/projectIdOrName/templates GET")

	// checking do we get certain projectID or not
	projectID := params.ByName("projectIdOrName")
	if projectID == "" {
		// get common templates
		projectID = utils.CommonProjectID
	}
	//reading cluster info from database
	hS.Logger.Print("Reading templates information from db...")

	templates, err := hS.Db.ListTemplates(projectID)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	err = enc.Encode(templates)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

func (hS HttpServer) TemplateGet(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// checking do we get certain projectID or not
	projectID := params.ByName("projectIdOrName")
	if projectID == "" {
		// get common templates
		projectID = utils.CommonProjectID
	}
	templateID := params.ByName("templateID")
	hS.Logger.Print("Get /templates/ or /projects/projectIdOrName/templates", templateID, " GET")

	//reading template info from database
	hS.Logger.Print("Reading template information from db...")
	template, err := hS.Db.ReadTemplate(projectID, templateID)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if template.Name == "" {
		hS.Logger.Print("Template not found")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	err = enc.Encode(template)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}
