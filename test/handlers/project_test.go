package handlers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	gomock "github.com/golang/mock/gomock"
	mocks "github.com/ispras/michman/internal/mock"
	protobuf "github.com/ispras/michman/internal/protobuf"
	"github.com/julienschmidt/httprouter"
)

var project = protobuf.Project{
	Name:         "test-project",
	DisplayName:  "test-project",
	Description:  "some description",
	DefaultImage: "ubuntu",
}

var testImage = protobuf.Image{
	ID:           "e2246d19-1221-416e-8c49-ad6dac00000a",
	Name:         "ubuntu",
	AnsibleUser:  "ubuntu",
	CloudImageID: "e2246d19-1221-416e-8c49-ad6dac00000a",
}

func TestProjectValidate(t *testing.T) {

	t.Run("Bad name for project", func(t *testing.T) {
		var projectBadName *protobuf.Project = &protobuf.Project{
			Name:         "#test-project#",
			DisplayName:  "#test-project#",
			Description:  "some description",
			DefaultImage: "ubuntu",
		}

		check := handlers.ValidateProject(projectBadName)

		if check != false {
			t.Fatalf("Expected status code %v, but received: %v", false, check)
		}
	})

	t.Run("Name for Project is Ok", func(t *testing.T) {
		var projectOkName *protobuf.Project = &protobuf.Project{
			Name:         "test-project",
			DisplayName:  "test-project",
			Description:  "some description",
			DefaultImage: "ubuntu",
		}

		check := handlers.ValidateProject(projectOkName)

		if check != true {
			t.Fatalf("Expected status code %v, but received: %v", true, check)
		}
	})

}
func TestProjectsGetList(t *testing.T) {
	request, _ := http.NewRequest("GET", "/projects", nil)
	response := httptest.NewRecorder()

	mockCtrl := gomock.NewController(t)

	l := log.New(os.Stdout, "TestProjectsGetList: ", log.Ldate|log.Ltime)
	mockClient := mocks.NewMockGrpcClient(mockCtrl)
	mockDatabase := mocks.NewMockDatabase(mockCtrl)
	RespHandler := handlers.HttpResponseHandler{}
	mockDatabase.EXPECT().ListProjects().Return([]protobuf.Project{}, nil)

	hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, RespHandler: RespHandler}
	hS.ProjectsGetList(response, request, httprouter.Params{})

	if response.Code != http.StatusOK {
		t.Fatalf("Expected status code %v, but received: %v", "200", response.Code)
	}
}

func TestProjectsCreate(t *testing.T) {
	l := log.New(os.Stdout, "TestProjectsCreate: ", log.Ldate|log.Ltime)
	mockCtrl := gomock.NewController(t)
	mockClient := mocks.NewMockGrpcClient(mockCtrl)
	mockDatabase := mocks.NewMockDatabase(mockCtrl)
	RespHandler := handlers.HttpResponseHandler{}

	hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, RespHandler: RespHandler}

	t.Run("Valid JSON", func(t *testing.T) {
		testBody, _ := json.Marshal(project)
		request, _ := http.NewRequest("POST", "/projects", bytes.NewReader(testBody))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		mockDatabase.EXPECT().ReadProjectByName(project.DisplayName).Return(&protobuf.Project{}, nil)
		mockDatabase.EXPECT().ReadImage(project.DefaultImage).Return(&testImage, nil)
		mockDatabase.EXPECT().WriteProject(gomock.Any()).Return(nil)

		hS.ProjectCreate(response, request, httprouter.Params{})

		var p protobuf.Project
		err := json.NewDecoder(response.Body).Decode(&p)
		if err != nil {
			t.Fatalf("Get invalid JSON")
		}

		if p.ID == "" {
			t.Fatalf("Project ID wasn't created")
		}

		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %v, but received: %v", "200", response.Code)
		}

	})

	t.Run("Invalid JSON", func(t *testing.T) {
		testBody := []byte(`this is invalid json`)
		request, _ := http.NewRequest("POST", "/projects", bytes.NewBuffer(testBody))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		//mockDatabase.EXPECT().ReadProjectByName("test-project").Return(&protobuf.Project{}, nil)
		hS.ProjectCreate(response, request, httprouter.Params{})

		if response.Code != http.StatusBadRequest {
			t.Fatalf("Expected status code %v, but received: %v", "400", response.Code)
		}
	})
}

func TestProjectGetByName(t *testing.T) {
	l := log.New(os.Stdout, "TestProjectGetByName: ", log.Ldate|log.Ltime)
	mockCtrl := gomock.NewController(t)
	mockClient := mocks.NewMockGrpcClient(mockCtrl)
	mockDatabase := mocks.NewMockDatabase(mockCtrl)
	RespHandler := handlers.HttpResponseHandler{}
	projectName := "testProjectName"

	t.Run("Existed project", func(t *testing.T) {
		request, _ := http.NewRequest("GET", "/projects/"+projectName, nil)
		response := httptest.NewRecorder()
		mockDatabase.EXPECT().ReadProjectByName(projectName).Return(&protobuf.Project{Name: "NotEmptyName"}, nil)
		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, RespHandler: RespHandler}

		hS.ProjectGet(response, request, httprouter.Params{{Key: "projectIdOrName", Value: projectName}})

		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %v, but received: %v", "200", response.Code)
		}
	})

	t.Run("Not existed project", func(t *testing.T) {
		request, _ := http.NewRequest("GET", "/projects/"+projectName, nil)
		response := httptest.NewRecorder()
		mockDatabase.EXPECT().ReadProjectByName(projectName).Return(&protobuf.Project{}, nil)
		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, RespHandler: RespHandler}

		hS.ProjectGet(response, request, httprouter.Params{{Key: "projectIdOrName", Value: projectName}})

		if response.Code != http.StatusNoContent {
			t.Fatalf("Expected status code %v, but received: %v", "200", response.Code)
		}
	})
}

func TestProjectUpdate(t *testing.T) {
	l := log.New(os.Stdout, "TestProjectUpdate: ", log.Ldate|log.Ltime)
	mockCtrl := gomock.NewController(t)
	mockClient := mocks.NewMockGrpcClient(mockCtrl)
	mockDatabase := mocks.NewMockDatabase(mockCtrl)
	RespHandler := handlers.HttpResponseHandler{}
	projectName := "testProjectName"

	correctBudy := []byte(`{
		"Description": "some description"
	}`)

	incorrectBudy := []byte(`{
		"Name": "test-project",
		"DisplayName": "test-project-display",
		"GroupId": 1,
		"Description": "some description"
	}`)

	invalidJSON := []byte(`invalid json`)

	t.Run("Existed project, incorrect update fields", func(t *testing.T) {
		request, _ := http.NewRequest("PUT", "/projects/"+projectName, bytes.NewBuffer(incorrectBudy))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		mockDatabase.EXPECT().ReadProjectByName(projectName).Return(&protobuf.Project{Name: "NotEmptyName"}, nil)

		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, RespHandler: RespHandler}
		hS.ProjectUpdate(response, request, httprouter.Params{{Key: "projectIdOrName", Value: projectName}})

		if response.Code != http.StatusBadRequest {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusBadRequest, response.Code)
		}
	})

	t.Run("Existed project, correct fields", func(t *testing.T) {
		request, _ := http.NewRequest("PUT", "/projects/"+projectName, bytes.NewBuffer(correctBudy))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		mockDatabase.EXPECT().ReadProjectByName(projectName).Return(&protobuf.Project{Name: "NotEmptyName"}, nil)
		mockDatabase.EXPECT().UpdateProject(gomock.Any()).Return(nil)

		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, RespHandler: RespHandler}
		hS.ProjectUpdate(response, request, httprouter.Params{{Key: "projectIdOrName", Value: projectName}})

		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusOK, response.Code)
		}
	})

	t.Run("Incorrect JSON", func(t *testing.T) {
		request, _ := http.NewRequest("PUT", "/projects/"+projectName, bytes.NewBuffer(invalidJSON))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		mockDatabase.EXPECT().ReadProjectByName(projectName).Return(&protobuf.Project{Name: "NotEmptyName"}, nil)

		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, RespHandler: RespHandler}
		hS.ProjectUpdate(response, request, httprouter.Params{{Key: "projectIdOrName", Value: projectName}})

		if response.Code != http.StatusBadRequest {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusBadRequest, response.Code)
		}
	})

	t.Run("Project didn't exist", func(t *testing.T) {
		request, _ := http.NewRequest("PUT", "/projects/"+projectName, bytes.NewBuffer(correctBudy))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		mockDatabase.EXPECT().ReadProjectByName(projectName).Return(&protobuf.Project{}, nil)

		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, RespHandler: RespHandler}
		hS.ProjectUpdate(response, request, httprouter.Params{{Key: "projectIdOrName", Value: projectName}})

		if response.Code != http.StatusNoContent {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusNoContent, response.Code)
		}
	})
}

func TestProjectDelete(t *testing.T) {
	l := log.New(os.Stdout, "TestProjectUpdate: ", log.Ldate|log.Ltime)
	mockCtrl := gomock.NewController(t)
	mockClient := mocks.NewMockGrpcClient(mockCtrl)
	mockDatabase := mocks.NewMockDatabase(mockCtrl)
	RespHandler := handlers.HttpResponseHandler{}
	projectName := "testProjectName"
	projectID := "someID123"

	t.Run("Project has clusters", func(t *testing.T) {
		request, _ := http.NewRequest("PUT", "/projects/"+projectName, nil)
		response := httptest.NewRecorder()

		var existedCluster = []protobuf.Cluster{protobuf.Cluster{Name: "Some-name",
			ID: "some_ID_123", ProjectID: "test-TEST-UUID-123"}}

		mockDatabase.EXPECT().ReadProjectByName(projectName).Return(&protobuf.Project{Name: projectName, ID: projectID}, nil)
		mockDatabase.EXPECT().ReadProjectClusters(projectID).Return(existedCluster, nil)

		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, RespHandler: RespHandler}
		hS.ProjectDelete(response, request, httprouter.Params{{Key: "projectIdOrName", Value: projectName}})

		if response.Code != http.StatusBadRequest {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusBadRequest, response.Code)
		}
	})

	t.Run("Project has no clusters", func(t *testing.T) {
		request, _ := http.NewRequest("PUT", "/projects/"+projectName, nil)
		response := httptest.NewRecorder()

		mockDatabase.EXPECT().ReadProjectByName(projectName).Return(&protobuf.Project{Name: projectName, ID: projectID}, nil)
		mockDatabase.EXPECT().ReadProjectClusters(projectID).Return([]protobuf.Cluster{}, nil)
		mockDatabase.EXPECT().DeleteProject(projectID).Return(nil)

		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, RespHandler: RespHandler}
		hS.ProjectDelete(response, request, httprouter.Params{{Key: "projectIdOrName", Value: projectName}})

		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusOK, response.Code)
		}
	})

	t.Run("Project didn't exist", func(t *testing.T) {
		request, _ := http.NewRequest("DELETE", "/projects/"+projectName, nil)
		response := httptest.NewRecorder()

		mockDatabase.EXPECT().ReadProjectByName(projectName).Return(&protobuf.Project{}, nil)

		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, RespHandler: RespHandler}
		hS.ProjectDelete(response, request, httprouter.Params{{Key: "projectIdOrName", Value: projectName}})

		if response.Code != http.StatusNoContent {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusNoContent, response.Code)
		}
	})
}
