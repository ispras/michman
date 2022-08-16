package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/golang/mock/gomock"
	protobuf "github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/internal/utils"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var commonTestTemplateRequest = protobuf.Template{
	//ID:          "af916f2b-8c25-4e72-8cb6-27170583128c",
	//ProjectID:   utils.CommonProjectID,
	//Name:        "test1",
	DisplayName: "test1",
	Services:    nil,
	NHosts:      1,
	Description: "description1",
}

var commonTestTemplateResponse = protobuf.Template{
	ID:          "af916f2b-8c25-4e72-8cb6-27170583128c",
	ProjectID:   utils.CommonProjectID,
	Name:        "test1",
	DisplayName: "test1",
	Services:    nil,
	NHosts:      1,
	Description: "description1",
}

var testProject = protobuf.Project{
	ID:          "2f20b979-6eb8-46b2-8f14-984f87d96978]",
	Name:        "test1",
	DisplayName: "test1",
	GroupID:     "0",
	Description: "test-description",
}

var testProjectID = "2f20b979-6eb8-46b2-8f14-984f87d96978"
var testCommonProjectID = utils.CommonProjectID

func TestTemplatesGetList(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	l := log.New(os.Stdout, "TestCommonTemplatesGetList: ", log.Ldate|log.Ltime)
	mockClient := mocks.NewMockGrpcClient(mockCtrl)

	t.Run("Common templates list", func(t *testing.T) {
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		request, _ := http.NewRequest("GET", "/templates", nil)
		response := httptest.NewRecorder()
		testTemplate1 := commonTestTemplateResponse
		testTemplate2 := protobuf.Template{
			ID:          "bf916f2b-8c25-4e72-8cb6-27170583128c",
			ProjectID:   utils.CommonProjectID,
			Name:        "test2",
			DisplayName: "test2",
			Services:    nil,
			NHosts:      2,
			Description: "description2",
		}

		mockDatabase.EXPECT().ListTemplates(testCommonProjectID).Return([]protobuf.Template{testTemplate1,
			testTemplate2}, nil)
		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase}

		hS.TemplatesGetList(response, request, httprouter.Params{})

		var tt []protobuf.Template
		err := json.NewDecoder(response.Body).Decode(&tt)

		if err != nil {
			t.Fatalf("Got invalid JSON")
		}

		if len(tt) != 2 {
			t.Fatalf("Got wrong count of templates")
		}

		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusOK, response.Code)
		}
	})

	t.Run("Templates list for certain project", func(t *testing.T) {
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		request, _ := http.NewRequest("GET", "/projects/"+testProjectID+"/templates", nil)
		response := httptest.NewRecorder()
		testTemplate1 := protobuf.Template{
			ID:          "yf916f2b-8c25-4e72-8cb6-27170583128c",
			ProjectID:   testProjectID,
			Name:        "test2",
			DisplayName: "test2",
			Services:    nil,
			NHosts:      2,
			Description: "description2",
		}
		testTemplate2 := protobuf.Template{
			ID:          "bf916f2b-8c25-4e72-8cb6-27170583128c",
			ProjectID:   testProjectID,
			Name:        "test2",
			DisplayName: "test2",
			Services:    nil,
			NHosts:      2,
			Description: "description2",
		}

		mockDatabase.EXPECT().ListTemplates(testProjectID).Return([]protobuf.Template{testTemplate1,
			testTemplate2}, nil)
		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase}

		hS.TemplatesGetList(response, request, httprouter.Params{{"projectIdOrName", testProjectID}})

		var tt []protobuf.Template
		err := json.NewDecoder(response.Body).Decode(&tt)

		if err != nil {
			t.Fatalf("Got invalid JSON")
		}

		if len(tt) != 2 {
			t.Fatalf("Got wrong count of templates")
		}

		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusOK, response.Code)
		}
	})

}

func TestTemplatesGet(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	l := log.New(os.Stdout, "TestTemplatesGet: ", log.Ldate|log.Ltime)
	mockClient := mocks.NewMockGrpcClient(mockCtrl)

	t.Run("Existing common template", func(t *testing.T) {
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		testTemplateID := "af916f2b-8c25-4e72-8cb6-27170583128c"
		request, _ := http.NewRequest("GET",
			"/templates/"+testTemplateID, nil)
		response := httptest.NewRecorder()

		testTemplate := commonTestTemplateResponse

		mockDatabase.EXPECT().ReadTemplate(utils.CommonProjectID, testTemplateID).Return(&testTemplate, nil)
		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase}

		hS.TemplateGet(response, request, httprouter.Params{{Key: "templateID", Value: testTemplateID}})

		var tmp protobuf.Template
		err := json.NewDecoder(response.Body).Decode(&tmp)

		if err != nil {
			t.Fatalf("Got invalid JSON")
		}

		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusOK, response.Code)
		}
	})

	t.Run("Not existing common template", func(t *testing.T) {
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		testTemplateID := "af916f2b-8c25-4e72-8cb6-27170583128c"
		request, _ := http.NewRequest("GET",
			"/templates/"+testTemplateID, nil)
		response := httptest.NewRecorder()

		mockDatabase.EXPECT().ReadTemplate(utils.CommonProjectID, testTemplateID).Return(&protobuf.Template{}, nil)
		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase}

		hS.TemplateGet(response, request, httprouter.Params{{Key: "templateID", Value: testTemplateID}})

		if response.Code != http.StatusNoContent {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusNoContent, response.Code)
		}
	})

	t.Run("Existing projects template", func(t *testing.T) {
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		testTemplateID := "af916f2b-8c25-4e72-8cb6-27170583128c"
		request, _ := http.NewRequest("GET",
			"/projects/"+testProjectID+"/templates/"+testTemplateID, nil)
		response := httptest.NewRecorder()

		testTemplate := commonTestTemplateResponse
		testTemplate.ProjectID = testProjectID

		mockDatabase.EXPECT().ReadTemplate(testProjectID, testTemplateID).Return(&testTemplate, nil)
		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase}

		hS.TemplateGet(response, request, httprouter.Params{{"projectIdOrName", testProjectID},
			{"templateID", testTemplateID}})

		var tmp protobuf.Template
		err := json.NewDecoder(response.Body).Decode(&tmp)

		if err != nil {
			t.Fatalf("Got invalid JSON")
		}

		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusOK, response.Code)
		}
	})

	t.Run("Not existing projects template", func(t *testing.T) {
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		testTemplateID := "af916f2b-8c25-4e72-8cb6-27170583128c"
		request, _ := http.NewRequest("GET",
			"/projects/"+testProjectID+"/templates/"+testTemplateID, nil)
		response := httptest.NewRecorder()

		mockDatabase.EXPECT().ReadTemplate(testProjectID, testTemplateID).Return(&protobuf.Template{}, nil)
		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase}

		hS.TemplateGet(response, request, httprouter.Params{{"projectIdOrName", testProjectID},
			{"templateID", testTemplateID}})

		if response.Code != http.StatusNoContent {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusNoContent, response.Code)
		}
	})
}

func TestTemplatesCreate(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	l := log.New(os.Stdout, "TestTemplatesCreate: ", log.Ldate|log.Ltime)
	mockClient := mocks.NewMockGrpcClient(mockCtrl)

	t.Run("New common template", func(t *testing.T) {
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		testTemplate := commonTestTemplateRequest
		jsonTemplate, _ := json.Marshal(testTemplate)
		request, _ := http.NewRequest("POST",
			"/templates", bytes.NewReader(jsonTemplate))
		response := httptest.NewRecorder()
		mockDatabase.EXPECT().ReadTemplateByName(testTemplate.DisplayName+"-common").Return(&protobuf.Template{}, nil)
		mockDatabase.EXPECT().WriteTemplate(gomock.Any()).Return(nil)
		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase}

		hS.TemplateCreate(response, request, httprouter.Params{})

		var tmp protobuf.Template
		err := json.NewDecoder(response.Body).Decode(&tmp)

		if err != nil {
			t.Fatalf("Got invalid JSON in responce")
		}

		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusOK, response.Code)
		}
	})

	t.Run("common template with error from db side", func(t *testing.T) {
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		testTemplate := commonTestTemplateRequest
		jsonTemplate, _ := json.Marshal(testTemplate)
		request, _ := http.NewRequest("POST",
			"/templates", bytes.NewReader(jsonTemplate))
		response := httptest.NewRecorder()
		mockDatabase.EXPECT().ReadTemplateByName(testTemplate.DisplayName+"-common").Return(&protobuf.Template{}, nil)
		mockDatabase.EXPECT().WriteTemplate(gomock.Any()).Return(errors.New("error on db side"))
		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase}

		hS.TemplateCreate(response, request, httprouter.Params{})

		if response.Code != http.StatusBadRequest {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusBadRequest, response.Code)
		}
	})

	t.Run("Already have this common template", func(t *testing.T) {
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		testTemplate := commonTestTemplateRequest
		testTemplateResponse := commonTestTemplateResponse
		jsonTemplate, _ := json.Marshal(testTemplate)
		request, _ := http.NewRequest("POST",
			"/templates", bytes.NewReader(jsonTemplate))
		response := httptest.NewRecorder()
		mockDatabase.EXPECT().ReadTemplateByName(testTemplate.DisplayName+"-common").Return(&testTemplateResponse, nil)
		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase}

		hS.TemplateCreate(response, request, httprouter.Params{})

		if response.Code != http.StatusBadRequest {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusBadRequest, response.Code)
		}
	})

	t.Run("New projects template", func(t *testing.T) {
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		testTemplate := commonTestTemplateRequest
		testTemplate.ProjectID = testProjectID
		testTemplateResponse := commonTestTemplateResponse
		testTemplateResponse.ProjectID = testProjectID
		jsonTemplate, _ := json.Marshal(testTemplate)
		request, _ := http.NewRequest("POST",
			"/projects/"+testProjectID+"/templates", bytes.NewReader(jsonTemplate))
		response := httptest.NewRecorder()
		mockDatabase.EXPECT().ReadProject(testTemplate.ProjectID).Return(&testProject, nil)
		mockDatabase.EXPECT().ReadTemplateByName(testTemplate.DisplayName+"-"+testProject.Name).Return(&protobuf.Template{}, nil)
		mockDatabase.EXPECT().WriteTemplate(gomock.Any()).Return(nil)
		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase}

		hS.TemplateCreate(response, request, httprouter.Params{{"projectIdOrName", testProjectID}})

		var tmp protobuf.Template
		err := json.NewDecoder(response.Body).Decode(&tmp)

		if err != nil {
			t.Fatalf("Got invalid JSON in responce")
		}

		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusOK, response.Code)
		}
	})

	t.Run("projects template with error from db side", func(t *testing.T) {
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		testTemplate := commonTestTemplateRequest
		testTemplate.ProjectID = testProjectID
		testTemplateResponse := commonTestTemplateResponse
		testTemplateResponse.ProjectID = testProjectID
		jsonTemplate, _ := json.Marshal(testTemplate)
		request, _ := http.NewRequest("POST",
			"/projects/"+testProjectID+"/templates", bytes.NewReader(jsonTemplate))
		response := httptest.NewRecorder()
		mockDatabase.EXPECT().ReadProject(testTemplate.ProjectID).Return(&testProject, nil)
		mockDatabase.EXPECT().ReadTemplateByName(testTemplate.DisplayName+"-"+testProject.Name).Return(&protobuf.Template{}, nil)
		mockDatabase.EXPECT().WriteTemplate(gomock.Any()).Return(errors.New("error on db side"))
		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase}

		hS.TemplateCreate(response, request, httprouter.Params{{"projectIdOrName", testProjectID}})

		if response.Code != http.StatusBadRequest {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusBadRequest, response.Code)
		}
	})

	t.Run("Already have this projects template", func(t *testing.T) {
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		testTemplate := commonTestTemplateRequest
		testTemplate.ProjectID = testProjectID
		testTemplateResponse := commonTestTemplateResponse
		testTemplateResponse.ProjectID = testProjectID
		jsonTemplate, _ := json.Marshal(testTemplate)
		request, _ := http.NewRequest("POST",
			"/projects/"+testProjectID+"/templates", bytes.NewReader(jsonTemplate))
		response := httptest.NewRecorder()
		mockDatabase.EXPECT().ReadProject(testTemplate.ProjectID).Return(&testProject, nil)
		mockDatabase.EXPECT().ReadTemplateByName(testTemplate.DisplayName+"-"+testProject.Name).Return(&testTemplateResponse, nil)
		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase}

		hS.TemplateCreate(response, request, httprouter.Params{{"projectIdOrName", testProjectID}})

		if response.Code != http.StatusBadRequest {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusBadRequest, response.Code)
		}
	})
}

func TestTemplatesUpdate(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	l := log.New(os.Stdout, "TestTemplatesUpdate: ", log.Ldate|log.Ltime)
	mockClient := mocks.NewMockGrpcClient(mockCtrl)

	t.Run("Success update common template", func(t *testing.T) {
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		testTemplate := commonTestTemplateRequest
		testTemplateResponse := commonTestTemplateResponse
		jsonTemplate, _ := json.Marshal(testTemplate)
		request, _ := http.NewRequest("PUT",
			"/templates/"+testTemplate.ID, bytes.NewReader(jsonTemplate))
		response := httptest.NewRecorder()
		mockDatabase.EXPECT().ReadTemplate(utils.CommonProjectID, testTemplate.ID).Return(&testTemplateResponse, nil)
		mockDatabase.EXPECT().WriteTemplate(&testTemplateResponse).Return(nil)
		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase}

		hS.TemplateUpdate(response, request, httprouter.Params{{Key: "templateID", Value: testTemplate.ID}})

		var tmp protobuf.Template
		err := json.NewDecoder(response.Body).Decode(&tmp)

		if err != nil {
			t.Fatalf("Got invalid JSON in responce")
		}

		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusOK, response.Code)
		}
	})

	t.Run("update common template with error from db side", func(t *testing.T) {
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		testTemplate := commonTestTemplateRequest
		jsonTemplate, _ := json.Marshal(testTemplate)
		request, _ := http.NewRequest("PUT",
			"/templates/"+testTemplate.ID, bytes.NewReader(jsonTemplate))
		response := httptest.NewRecorder()
		mockDatabase.EXPECT().ReadTemplate(utils.CommonProjectID, testTemplate.ID).Return(&testTemplate, nil)
		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase}

		hS.TemplateUpdate(response, request, httprouter.Params{{Key: "templateID", Value: testTemplate.ID}})

		if response.Code != http.StatusBadRequest {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusBadRequest, response.Code)
		}
	})

	t.Run("Dont have common template with such ID", func(t *testing.T) {
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		testTemplate := commonTestTemplateRequest
		jsonTemplate, _ := json.Marshal(testTemplate)
		request, _ := http.NewRequest("POST",
			"/templates", bytes.NewReader(jsonTemplate))
		response := httptest.NewRecorder()
		mockDatabase.EXPECT().ReadTemplate(utils.CommonProjectID, testTemplate.ID).Return(&protobuf.Template{}, nil)
		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase}

		hS.TemplateUpdate(response, request, httprouter.Params{{Key: "templateID", Value: testTemplate.ID}})

		if response.Code != http.StatusBadRequest {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusBadRequest, response.Code)
		}
	})

	t.Run("Success update projects template", func(t *testing.T) {
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		testTemplate := commonTestTemplateRequest
		testTemplateResponse := commonTestTemplateResponse
		testTemplateResponse.ProjectID = testProjectID
		jsonTemplate, _ := json.Marshal(testTemplate)
		request, _ := http.NewRequest("PUT",
			"/projects/"+testProjectID+"/templates/"+testTemplate.ID, bytes.NewReader(jsonTemplate))
		response := httptest.NewRecorder()
		mockDatabase.EXPECT().ReadProject(testTemplateResponse.ProjectID).Return(&testProject, nil)
		mockDatabase.EXPECT().ReadTemplate(testProjectID, testTemplate.ID).Return(&testTemplateResponse, nil)
		mockDatabase.EXPECT().WriteTemplate(&testTemplateResponse).Return(nil)
		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase}

		hS.TemplateUpdate(response, request, httprouter.Params{{"projectIdOrName", testProjectID},
			{"templateID", testTemplate.ID}})

		var tmp protobuf.Template
		err := json.NewDecoder(response.Body).Decode(&tmp)

		if err != nil {
			t.Fatalf("Got invalid JSON in responce")
		}

		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusOK, response.Code)
		}
	})

	t.Run("update projects template with error from db side", func(t *testing.T) {
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		testTemplate := commonTestTemplateRequest
		testTemplate.ProjectID = testProjectID
		testTemplateResponse := commonTestTemplateResponse
		testTemplateResponse.ProjectID = testProjectID
		jsonTemplate, _ := json.Marshal(testTemplate)
		request, _ := http.NewRequest("PUT",
			"/projects/"+testProjectID+"/templates/"+testTemplate.ID, bytes.NewReader(jsonTemplate))
		response := httptest.NewRecorder()
		mockDatabase.EXPECT().ReadProject(testTemplateResponse.ProjectID).Return(&testProject, nil)
		mockDatabase.EXPECT().ReadTemplate(testProjectID, testTemplate.ID).Return(&testTemplate, nil)
		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase}

		hS.TemplateUpdate(response, request, httprouter.Params{{"projectIdOrName", testProjectID},
			{"templateID", testTemplate.ID}})

		if response.Code != http.StatusBadRequest {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusBadRequest, response.Code)
		}
	})

	t.Run("Dont have projects template with such ID", func(t *testing.T) {
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		testTemplate := commonTestTemplateRequest
		testTemplate.ProjectID = testProjectID
		testTemplateResponse := commonTestTemplateResponse
		testTemplateResponse.ProjectID = testProjectID
		jsonTemplate, _ := json.Marshal(testTemplate)
		request, _ := http.NewRequest("POST",
			"/projects/"+testProjectID+"/templates", bytes.NewReader(jsonTemplate))
		response := httptest.NewRecorder()
		mockDatabase.EXPECT().ReadProject(testTemplateResponse.ProjectID).Return(&testProject, nil)
		mockDatabase.EXPECT().ReadTemplate(testProjectID, testTemplate.ID).Return(&protobuf.Template{}, nil)
		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase}

		hS.TemplateUpdate(response, request, httprouter.Params{{"projectIdOrName", testProjectID},
			{"templateID", testTemplate.ID}})

		if response.Code != http.StatusBadRequest {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusBadRequest, response.Code)
		}
	})
}

func TestTemplatesDelete(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	l := log.New(os.Stdout, "TestTemplatesDelete: ", log.Ldate|log.Ltime)
	mockClient := mocks.NewMockGrpcClient(mockCtrl)

	t.Run("Success delete common template", func(t *testing.T) {
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		testTemplate := commonTestTemplateResponse
		request, _ := http.NewRequest("DELETE",
			"/templates/"+testTemplate.ID, nil)
		response := httptest.NewRecorder()
		mockDatabase.EXPECT().ReadTemplate(utils.CommonProjectID, testTemplate.ID).Return(&testTemplate, nil)
		mockDatabase.EXPECT().DeleteTemplate(testTemplate.ID).Return(nil)
		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase}

		hS.TemplateDelete(response, request, httprouter.Params{{Key: "templateID", Value: testTemplate.ID}})

		var tmp protobuf.Template
		err := json.NewDecoder(response.Body).Decode(&tmp)

		if err != nil {
			t.Fatalf("Got invalid JSON in responce")
		}

		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusOK, response.Code)
		}
	})

	t.Run("delete common template with error from db side", func(t *testing.T) {
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		testTemplate := commonTestTemplateResponse
		request, _ := http.NewRequest("DELETE",
			"/templates/"+testTemplate.ID, nil)
		response := httptest.NewRecorder()
		mockDatabase.EXPECT().ReadTemplate(utils.CommonProjectID, testTemplate.ID).Return(&testTemplate, nil)
		mockDatabase.EXPECT().DeleteTemplate(testTemplate.ID).Return(errors.New("some db error"))
		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase}

		hS.TemplateDelete(response, request, httprouter.Params{{Key: "templateID", Value: testTemplate.ID}})

		if response.Code != http.StatusBadRequest {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusBadRequest, response.Code)
		}
	})

	t.Run("Delete not existing common template", func(t *testing.T) {
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		testTemplate := commonTestTemplateResponse
		request, _ := http.NewRequest("DELETE",
			"/templates/"+testTemplate.ID, nil)
		response := httptest.NewRecorder()
		mockDatabase.EXPECT().ReadTemplate(utils.CommonProjectID, testTemplate.ID).Return(&protobuf.Template{}, nil)
		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase}

		hS.TemplateDelete(response, request, httprouter.Params{{Key: "templateID", Value: testTemplate.ID}})

		if response.Code != http.StatusNoContent {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusNoContent, response.Code)
		}
	})

	t.Run("Success delete projects template", func(t *testing.T) {
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		testTemplate := commonTestTemplateResponse
		testTemplate.ProjectID = testProjectID
		request, _ := http.NewRequest("DELETE",
			"/projects/"+testProjectID+"/templates/"+testTemplate.ID, nil)
		response := httptest.NewRecorder()
		mockDatabase.EXPECT().ReadTemplate(testProjectID, testTemplate.ID).Return(&testTemplate, nil)
		mockDatabase.EXPECT().DeleteTemplate(testTemplate.ID).Return(nil)
		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase}

		hS.TemplateDelete(response, request, httprouter.Params{{"projectIdOrName", testProjectID},
			{"templateID", testTemplate.ID}})

		var tmp protobuf.Template
		err := json.NewDecoder(response.Body).Decode(&tmp)

		if err != nil {
			t.Fatalf("Got invalid JSON in responce")
		}

		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusOK, response.Code)
		}
	})

	t.Run("delete projects template with error from db side", func(t *testing.T) {
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		testTemplate := commonTestTemplateResponse
		testTemplate.ProjectID = testProjectID
		request, _ := http.NewRequest("DELETE",
			"/projects/"+testProjectID+"/templates/"+testTemplate.ID, nil)
		response := httptest.NewRecorder()
		mockDatabase.EXPECT().ReadTemplate(testProjectID, testTemplate.ID).Return(&testTemplate, nil)
		mockDatabase.EXPECT().DeleteTemplate(testTemplate.ID).Return(errors.New("some db error"))
		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase}

		hS.TemplateDelete(response, request, httprouter.Params{{"projectIdOrName", testProjectID},
			{"templateID", testTemplate.ID}})

		if response.Code != http.StatusBadRequest {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusBadRequest, response.Code)
		}
	})

	t.Run("Delete not existing projects template", func(t *testing.T) {
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		testTemplate := commonTestTemplateResponse
		testTemplate.ProjectID = testProjectID
		testTemplate.ProjectID = testProjectID
		request, _ := http.NewRequest("DELETE",
			"/projects/"+testProjectID+"/templates/"+testTemplate.ID, nil)
		response := httptest.NewRecorder()
		mockDatabase.EXPECT().ReadTemplate(testProjectID, testTemplate.ID).Return(&protobuf.Template{}, nil)
		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase}

		hS.TemplateDelete(response, request, httprouter.Params{{"projectIdOrName", testProjectID},
			{"templateID", testTemplate.ID}})

		if response.Code != http.StatusNoContent {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusNoContent, response.Code)
		}
	})
}
