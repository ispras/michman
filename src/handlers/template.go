package handlers

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	protobuf "gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/protobuf"
	"gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/utils"
	"net/http"
)

func (hS HttpServer) TemplateCreate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	hS.Logger.Print("Get /templates or /project/templates POST")

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
	projectID := params.ByName("projectID")
	if projectID == "" {
		// get common templates
		projectID = utils.CommonProjectID
	} else {
		///check that projectID in url == projectID in json
		if projectID != t.ProjectID {
			hS.Logger.Print("projectID in url != projectID in json")
			w.WriteHeader(http.StatusBadRequest)
			enc := json.NewEncoder(w)
			err := enc.Encode("projectID in url != projectID in json")
			if err != nil {
				hS.Logger.Print(err)
				w.WriteHeader(http.StatusBadRequest)
			}
			return
		}
	}

	//check, that template with such ID doesn't exist
	dbRes, err := hS.Db.ReadTemplate(projectID, t.ID)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if dbRes.Name != "" {
		hS.Logger.Print("Template with this ID already exists")
		w.WriteHeader(http.StatusBadRequest)
		enc := json.NewEncoder(w)
		err := enc.Encode("Template with this ID already exists")
		if err != nil {
			hS.Logger.Print(err)
			w.WriteHeader(http.StatusBadRequest)
		}
		return
	}

	// set services uuid to empty string (UUID will be generated when cluster creates) //TODO: check this part
	for _, s := range t.Services {
		sUuid := ""
		s.ID = sUuid
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
	hS.Logger.Print("Get /templates/templateID or /projects/projectID/templates PUT")
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
	//check that ID in URL == ID in json
	if templateID != t.ID {
		hS.Logger.Print("ID in URL != ID in json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// checking do we get certain projectID or not
	projectID := params.ByName("projectID")
	if projectID == "" {
		// get common templates
		projectID = utils.CommonProjectID
	} else {
		///check that projectID in url == projectID in json
		if projectID != t.ProjectID {
			hS.Logger.Print("projectID in url != projectID in json")
			w.WriteHeader(http.StatusBadRequest)
			enc := json.NewEncoder(w)
			err := enc.Encode("projectID in url != projectID in json")
			if err != nil {
				hS.Logger.Print(err)
				w.WriteHeader(http.StatusBadRequest)
			}
			return
		}
	}

	//check, that template with such ID exists
	dbRes, err := hS.Db.ReadTemplate(projectID, t.ID)
	if err != nil {
		hS.Logger.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if dbRes.ID == "" {
		hS.Logger.Print("Template with this ID doesnt exist")
		w.WriteHeader(http.StatusBadRequest)
		enc := json.NewEncoder(w)
		err := enc.Encode("Template with this ID doesnt exists")
		if err != nil {
			hS.Logger.Print(err)
			w.WriteHeader(http.StatusBadRequest)
		}
		return
	}

	// set services uuid to empty string (UUID will be generated when cluster creates) //TODO: check this part
	for _, s := range t.Services {
		sUuid := ""
		s.ID = sUuid
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

func (hS HttpServer) TemplateDelete(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	templateID := params.ByName("templateID")
	hS.Logger.Print("Get /templates/", templateID, "or /projects/projectID/templates/templateID DELETE")

	//check that template exists
	hS.Logger.Print("Sending request to db-service to check that template exists...")

	// checking do we get certain projectID or not
	projectID := params.ByName("projectID")
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
	hS.Logger.Print("Get /templates or /projects/projectID/templates GET")

	// checking do we get certain projectID or not
	projectID := params.ByName("projectID")
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
}

func (hS HttpServer) TemplateGet(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// checking do we get certain projectID or not
	projectID := params.ByName("projectID")
	if projectID == "" {
		// get common templates
		projectID = utils.CommonProjectID
	}
	templateID := params.ByName("templateID")
	hS.Logger.Print("Get /templates/ or /projects/projectID/templates", templateID, " GET")

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
}
