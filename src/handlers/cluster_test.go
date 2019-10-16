package handlers

import (
	"bytes"
	"encoding/json"
	//"gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/utils"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/julienschmidt/httprouter"
	mocks "gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/mocks"
	protobuf "gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/protobuf"
	"gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/utils"
)

func TestClustersGet(t *testing.T) {
	l := log.New(os.Stdout, "TestClustersGet: ", log.Ldate|log.Ltime)
	projectName := "testProjectName"
	request, _ := http.NewRequest("GET", "/projects/"+projectName+"/clusters", nil)
	response := httptest.NewRecorder()

	mockCtrl := gomock.NewController(t)
	mockClient := mocks.NewMockGrpcClient(mockCtrl)
	mockDatabase := mocks.NewMockDatabase(mockCtrl)

	projectTestID := "someID123"

	mockDatabase.EXPECT().ReadProject(projectName).Return(&protobuf.Project{Name: projectName, ID: projectTestID}, nil)
	mockDatabase.EXPECT().ReadProjectClusters(projectTestID).Return([]protobuf.Cluster{}, nil)

	hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase}
	hS.ClustersGet(response, request, httprouter.Params{{Key: "projectName", Value: projectName}})

	if response.Code != http.StatusOK {
		t.Fatalf("Expected status code %v, but received: %v", "200", response.Code)
	}
}

func TestClusterCreate(t *testing.T) {
	l := log.New(os.Stdout, "TestClusterCreate: ", log.Ldate|log.Ltime)
	projectName := "testProjectName"

	mockCtrl := gomock.NewController(t)
	mockClient := mocks.NewMockGrpcClient(mockCtrl)
	mockDatabase := mocks.NewMockDatabase(mockCtrl)

	testClusterName := "spark-test"
	testCluster := []byte(`{
		"Name":"` + testClusterName + `",
		"EntityStatus": "some-status",
		"services":[
		{
			"Name":"spark-test",
			"Type":"spark",
			"Config":{
				"hadoop-version":"2.6"
			},
			"Version":"2.1.0"
		}],
		"NHosts":1
	}`)

	testInvalidCluster := []byte(`{
		"Name":"` + testClusterName + `",
		"EntityStatus": "some-status",
		"InvalidField":35
	}`)

	testInvalidJSON := []byte(`invalid json`)

	t.Run("Project didn't exist", func(t *testing.T) {
		request, _ := http.NewRequest("POST", "/projects/"+projectName+"/clusters", bytes.NewBuffer(testCluster))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		mockDatabase.EXPECT().ReadProject(projectName).Return(&protobuf.Project{}, nil)

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase}
		hS.ClusterCreate(response, request, httprouter.Params{{Key: "projectName", Value: projectName}})

		if response.Code != http.StatusNoContent {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusNoContent, response.Code)
		}
	})

	t.Run("Invalid cluster", func(t *testing.T) {
		request, _ := http.NewRequest("POST", "/projects/"+projectName+"/clusters", bytes.NewBuffer(testInvalidCluster))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		mockDatabase.EXPECT().ReadProject(projectName).Return(&protobuf.Project{ID: "test-TEST-UUID-123", Name: "NotEmptyName"}, nil)

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase}
		hS.ClusterCreate(response, request, httprouter.Params{{Key: "projectName", Value: projectName}})

		if response.Code != http.StatusBadRequest {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusBadRequest, response.Code)
		}
	})

	t.Run("Cluster didn't exist, valid JSON", func(t *testing.T) {
		request, _ := http.NewRequest("POST", "/projects/"+projectName+"/clusters", bytes.NewBuffer(testCluster))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		mockDatabase.EXPECT().ReadProject(projectName).Return(&protobuf.Project{ID: "test-TEST-UUID-123", Name: "NotEmptyName"}, nil)
		mockDatabase.EXPECT().ReadProjectClusters(gomock.Any()).Return([]protobuf.Cluster{}, nil)
		mockDatabase.EXPECT().WriteCluster(gomock.Any()).Return(nil)
		mockClient.EXPECT().StartClusterCreation(gomock.Any())

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase}
		hS.ClusterCreate(response, request, httprouter.Params{{Key: "projectName", Value: projectName}})

		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %v, but received: %v", "200", response.Code)
		}

		var c protobuf.Cluster
		err := json.NewDecoder(response.Body).Decode(&c)
		if err != nil {
			t.Fatalf("Get invalid JSON")
		}

		if c.ID == "" || c.ProjectID == "" || c.EntityStatus != utils.StatusInited {
			t.Fatalf("Cluster wasn't inited correct\n")
		}
	})

	t.Run("Cluster exists, but failed. Valid JSON", func(t *testing.T) {
		request, _ := http.NewRequest("POST", "/projects/"+projectName+"/clusters", bytes.NewBuffer(testCluster))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		var existedCluster = []protobuf.Cluster{protobuf.Cluster{Name: testClusterName, EntityStatus: utils.StatusFailed,
			ID: "some_ID_123", ProjectID: "test-TEST-UUID-123"}}

		mockDatabase.EXPECT().ReadProject(projectName).Return(&protobuf.Project{ID: "test-TEST-UUID-123", Name: "NotEmptyName"}, nil)
		mockDatabase.EXPECT().ReadProjectClusters(gomock.Any()).Return(existedCluster, nil)
		mockDatabase.EXPECT().WriteCluster(gomock.Any()).Return(nil)
		mockClient.EXPECT().StartClusterCreation(gomock.Any())

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase}
		hS.ClusterCreate(response, request, httprouter.Params{{Key: "projectName", Value: projectName}})

		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %v, but received: %v", "200", response.Code)
		}

		var c protobuf.Cluster
		err := json.NewDecoder(response.Body).Decode(&c)
		if err != nil {
			t.Fatalf("Get invalid JSON")
		}

		if c.ID == "" || c.ProjectID == "" || c.EntityStatus != utils.StatusInited {
			t.Fatalf("Cluster wasn't inited correct\n")
		}
	})

	t.Run("Cluster exists. Valid JSON", func(t *testing.T) {
		request, _ := http.NewRequest("POST", "/projects/"+projectName+"/clusters", bytes.NewBuffer(testCluster))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		var existedCluster = []protobuf.Cluster{protobuf.Cluster{Name: testClusterName, EntityStatus: utils.StatusCreated,
			ID: "some_ID_123", ProjectID: "test-TEST-UUID-123"}}

		mockDatabase.EXPECT().ReadProject(projectName).Return(&protobuf.Project{ID: "test-TEST-UUID-123", Name: "NotEmptyName"}, nil)
		mockDatabase.EXPECT().ReadProjectClusters(gomock.Any()).Return(existedCluster, nil)

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase}
		hS.ClusterCreate(response, request, httprouter.Params{{Key: "projectName", Value: projectName}})

		if response.Code != http.StatusBadRequest {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusBadRequest, response.Code)
		}
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		request, _ := http.NewRequest("POST", "/projects/"+projectName+"/clusters", bytes.NewBuffer(testInvalidJSON))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		mockDatabase.EXPECT().ReadProject(projectName).Return(&protobuf.Project{ID: "test-TEST-UUID-123", Name: "NotEmptyName"}, nil)

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase}
		hS.ClusterCreate(response, request, httprouter.Params{{Key: "projectName", Value: projectName}})

		if response.Code != http.StatusBadRequest {
			t.Fatalf("Expected status code %v, but received: %v", "400", response.Code)
		}
	})
}

func TestClustersGetByName(t *testing.T) {
	l := log.New(os.Stdout, "TestClustersGetByName: ", log.Ldate|log.Ltime)
	projectName := "testProjectName"
	clusterName := "testClusterName"

	mockCtrl := gomock.NewController(t)
	mockClient := mocks.NewMockGrpcClient(mockCtrl)
	mockDatabase := mocks.NewMockDatabase(mockCtrl)

	t.Run("Project didn't exist", func(t *testing.T) {
		request, _ := http.NewRequest("GET", "/projects/"+projectName+"/clusters"+clusterName, nil)
		response := httptest.NewRecorder()

		mockDatabase.EXPECT().ReadProject(projectName).Return(&protobuf.Project{}, nil)

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase}
		hS.ClustersGetByName(response, request, httprouter.Params{{Key: "projectName", Value: projectName},
			{Key: "clusterName", Value: clusterName}})

		if response.Code != http.StatusNoContent {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusNoContent, response.Code)
		}
	})

	t.Run("OK case", func(t *testing.T) {
		request, _ := http.NewRequest("GET", "/projects/"+projectName+"/clusters"+clusterName, nil)
		response := httptest.NewRecorder()

		projectTestID := "someID123"
		var testProjectClusters = []protobuf.Cluster{protobuf.Cluster{Name: clusterName}}

		mockDatabase.EXPECT().ReadProject(projectName).Return(&protobuf.Project{Name: projectName, ID: projectTestID}, nil)
		mockDatabase.EXPECT().ReadProjectClusters(projectTestID).Return(testProjectClusters, nil)

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase}
		hS.ClustersGetByName(response, request, httprouter.Params{{Key: "projectName", Value: projectName},
			{Key: "clusterName", Value: clusterName}})

		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %v, but received: %v", "200", response.Code)
		}
	})

	t.Run("Cluster didn't exist", func(t *testing.T) {
		request, _ := http.NewRequest("GET", "/projects/"+projectName+"/clusters"+clusterName, nil)
		response := httptest.NewRecorder()

		projectTestID := "someID123"
		anotherClusterName := "somethingElse"
		var testProjectClusters = []protobuf.Cluster{protobuf.Cluster{Name: anotherClusterName}}

		mockDatabase.EXPECT().ReadProject(projectName).Return(&protobuf.Project{Name: projectName, ID: projectTestID}, nil)
		mockDatabase.EXPECT().ReadProjectClusters(projectTestID).Return(testProjectClusters, nil)

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase}
		hS.ClustersGetByName(response, request, httprouter.Params{{Key: "projectName", Value: projectName},
			{Key: "clusterName", Value: clusterName}})

		if response.Code != http.StatusNoContent {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusNoContent, response.Code)
		}
	})
}
