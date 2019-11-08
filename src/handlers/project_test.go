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
	"github.com/julienschmidt/httprouter"
	mocks "gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/mocks"
	protobuf "gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/protobuf"
)

func TestProjectsGetList(t *testing.T) {
	request, _ := http.NewRequest("GET", "/projects", nil)
	response := httptest.NewRecorder()

	mockCtrl := gomock.NewController(t)

	l := log.New(os.Stdout, "TestProjectsGetList: ", log.Ldate|log.Ltime)
	mockClient := mocks.NewMockGrpcClient(mockCtrl)
	mockDatabase := mocks.NewMockDatabase(mockCtrl)
	errHandler := HttpErrorHandler{}
	mockDatabase.EXPECT().ListProjects().Return([]protobuf.Project{}, nil)

	hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
	hS.ProjectsGetList(response, request, httprouter.Params{})

	if response.Code != http.StatusOK {
		t.Fatalf("Expected status code %v, but received: %v", "200", response.Code)
	}
}

func TestProjectsPost(t *testing.T) {
	l := log.New(os.Stdout, "TestProjectsPost: ", log.Ldate|log.Ltime)
	mockCtrl := gomock.NewController(t)
	mockClient := mocks.NewMockGrpcClient(mockCtrl)
	mockDatabase := mocks.NewMockDatabase(mockCtrl)
	errHandler := HttpErrorHandler{}
	mockDatabase.EXPECT().ReadProjectByName("test-project").Return(&protobuf.Project{}, nil)

	// Because of generating uuid in WriteProject we can't know
	// what it will be
	// So gomock.Any() is used
	mockDatabase.EXPECT().WriteProject(gomock.Any()).Return(nil)

	hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}

	t.Run("Valid JSON", func(t *testing.T) {
		testBody := []byte(`{"Name": "test-project","DisplayName": "test-project","GroupId": 1,"Description": "some description"}`)
		request, _ := http.NewRequest("POST", "/projects", bytes.NewBuffer(testBody))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		hS.ProjectCreate(response, request, httprouter.Params{})

		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %v, but received: %v", "200", response.Code)
		}

		var p protobuf.Project
		err := json.NewDecoder(response.Body).Decode(&p)
		if err != nil {
			t.Fatalf("Get invalid JSON")
		}

		if p.ID == "" {
			t.Fatalf("Project ID wasn't created")
		}

	})

	t.Run("Invalid JSON", func(t *testing.T) {
		testBody := []byte(`this is invalid json`)
		request, _ := http.NewRequest("POST", "/projects", bytes.NewBuffer(testBody))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

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
	errHandler := HttpErrorHandler{}
	projectName := "testProjectName"

	t.Run("Existed project", func(t *testing.T) {
		request, _ := http.NewRequest("GET", "/projects/"+projectName, nil)
		response := httptest.NewRecorder()
		mockDatabase.EXPECT().ReadProjectByName(projectName).Return(&protobuf.Project{Name: "NotEmptyName"}, nil)
		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}

		hS.ProjectGetByName(response, request, httprouter.Params{{Key: "projectIdOrName", Value: projectName}})

		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %v, but received: %v", "200", response.Code)
		}
	})

	t.Run("Not existed project", func(t *testing.T) {
		request, _ := http.NewRequest("GET", "/projects/"+projectName, nil)
		response := httptest.NewRecorder()
		mockDatabase.EXPECT().ReadProjectByName(projectName).Return(&protobuf.Project{}, nil)
		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}

		hS.ProjectGetByName(response, request, httprouter.Params{{Key: "projectIdOrName", Value: projectName}})

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
	errHandler := HttpErrorHandler{}
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

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
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

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
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

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
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

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
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
	errHandler := HttpErrorHandler{}
	projectName := "testProjectName"
	projectID := "someID123"

	t.Run("Project has no clusters", func(t *testing.T) {
		request, _ := http.NewRequest("PUT", "/projects/"+projectName, nil)
		response := httptest.NewRecorder()

		mockDatabase.EXPECT().ReadProjectByName(projectName).Return(&protobuf.Project{Name: projectName, ID: projectID}, nil)
		mockDatabase.EXPECT().ReadProjectClusters(projectID).Return([]protobuf.Cluster{}, nil)
		mockDatabase.EXPECT().DeleteProject(projectID).Return(nil)

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		hS.ProjectDelete(response, request, httprouter.Params{{Key: "projectIdOrName", Value: projectName}})

		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusOK, response.Code)
		}
	})

	t.Run("Project has clusters", func(t *testing.T) {
		request, _ := http.NewRequest("PUT", "/projects/"+projectName, nil)
		response := httptest.NewRecorder()

		var existedCluster = []protobuf.Cluster{protobuf.Cluster{Name: "Some-name",
			ID: "some_ID_123", ProjectID: "test-TEST-UUID-123"}}

		mockDatabase.EXPECT().ReadProjectByName(projectName).Return(&protobuf.Project{Name: projectName, ID: projectID}, nil)
		mockDatabase.EXPECT().ReadProjectClusters(projectID).Return(existedCluster, nil)

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		hS.ProjectDelete(response, request, httprouter.Params{{Key: "projectIdOrName", Value: projectName}})

		if response.Code != http.StatusBadRequest {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusBadRequest, response.Code)
		}
	})

	t.Run("Project didn't exist", func(t *testing.T) {
		request, _ := http.NewRequest("DELETE", "/projects/"+projectName, nil)
		response := httptest.NewRecorder()

		mockDatabase.EXPECT().ReadProjectByName(projectName).Return(&protobuf.Project{}, nil)

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		hS.ProjectDelete(response, request, httprouter.Params{{Key: "projectIdOrName", Value: projectName}})

		if response.Code != http.StatusNoContent {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusNoContent, response.Code)
		}
	})
}
