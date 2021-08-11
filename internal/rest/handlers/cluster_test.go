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
	mocks "github.com/ispras/michman/internal/mocks"
	protobuf "github.com/ispras/michman/internal/protobuf"
	"github.com/ispras/michman/utils"
	"github.com/julienschmidt/httprouter"
)

var testClusterName = "spark-test"

var testService = protobuf.Service{
	DisplayName: "test",
	Type:        "spark",
}

var testCluster = protobuf.Cluster{
	DisplayName: testClusterName,
	NHosts:      3,
	Image:       "ubuntu",
	Services:    []*protobuf.Service{&testService},
}

func TestClustersGet(t *testing.T) {
	l := log.New(os.Stdout, "TestClustersGet: ", log.Ldate|log.Ltime)
	projectName := "testProjectName"
	request, _ := http.NewRequest("GET", "/projects/"+projectName+"/clusters", nil)
	response := httptest.NewRecorder()

	mockCtrl := gomock.NewController(t)
	mockClient := mocks.NewMockGrpcClient(mockCtrl)
	mockDatabase := mocks.NewMockDatabase(mockCtrl)
	errHandler := HttpErrorHandler{}

	projectTestID := "someID123"

	mockDatabase.EXPECT().ReadProjectByName(projectName).Return(&protobuf.Project{Name: projectName, ID: projectTestID}, nil)
	mockDatabase.EXPECT().ReadProjectClusters(projectTestID).Return([]protobuf.Cluster{}, nil)

	hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
	hS.ClustersGet(response, request, httprouter.Params{{Key: "projectIdOrName", Value: projectName}})

	if response.Code != http.StatusOK {
		t.Fatalf("Expected status code %v, but received: %v", "200", response.Code)
	}
}

func TestAddDependencies(t *testing.T) {
	l := log.New(os.Stdout, "TestAddDependencies: ", log.Ldate|log.Ltime)

	t.Run("no Dependencies", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		mockClient := mocks.NewMockGrpcClient(mockCtrl)
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		errHandler := HttpErrorHandler{}

		var testClusterOk *protobuf.Cluster = &protobuf.Cluster{
			DisplayName: "test",
			NHosts:      3,
			Image:       "ubuntu",
			Services:    []*protobuf.Service{&testService},
		}

		var testServiceCluster *protobuf.Service = &protobuf.Service{
			DisplayName: "test",
			Type:        "spark",
			Version:     "DefaultVersion",
		}

		var V *protobuf.ServiceVersion = &protobuf.ServiceVersion{Version: "DefaultVersion"}
		mockDatabase.EXPECT().ReadServiceVersionByName(testServiceCluster.Type, testServiceCluster.Version).Return(V, nil)

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}

		servicesList, _ := hS.AddDependencies(testClusterOk, testServiceCluster)
		if servicesList != nil {
			t.Fatalf("Expected servicesList without any parameters")
		}

	})

	t.Run("Add service from dependencies", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		mockClient := mocks.NewMockGrpcClient(mockCtrl)
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		errHandler := HttpErrorHandler{}

		var testClusterOk *protobuf.Cluster = &protobuf.Cluster{
			DisplayName: "test",
			NHosts:      3,
			Image:       "ubuntu",
			Services:    []*protobuf.Service{&testService},
		}

		var testServiceCluster *protobuf.Service = &protobuf.Service{
			DisplayName: "test",
			Type:        "spark",
			Version:     "DefaultVersion",
		}

		var D *protobuf.ServiceDependency = &protobuf.ServiceDependency{ServiceType: "sp"}

		var Dependencies = []*protobuf.ServiceDependency{D}

		var V *protobuf.ServiceVersion = &protobuf.ServiceVersion{Version: "DefaultVersion", Dependencies: Dependencies}

		mockDatabase.EXPECT().ReadServiceVersionByName(testServiceCluster.Type, testServiceCluster.Version).Return(V, nil)

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}

		servicesList, _ := hS.AddDependencies(testClusterOk, testServiceCluster)
		if servicesList == nil {
			t.Fatalf("Expected servicesList with parameters")
		}

	})

	t.Run("error: bad service version from user list", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		mockClient := mocks.NewMockGrpcClient(mockCtrl)
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		errHandler := HttpErrorHandler{}

		var testClusterOk *protobuf.Cluster = &protobuf.Cluster{
			DisplayName: "test",
			NHosts:      3,
			Image:       "ubuntu",
			Services:    []*protobuf.Service{&testService},
		}

		var testServiceCluster *protobuf.Service = &protobuf.Service{
			DisplayName: "test",
			Type:        "spark",
			Version:     "DefaultVersion",
		}

		var D *protobuf.ServiceDependency = &protobuf.ServiceDependency{ServiceType: "spark"}

		var Dependencies = []*protobuf.ServiceDependency{D}

		var V *protobuf.ServiceVersion = &protobuf.ServiceVersion{Version: "DefaultVersion", Dependencies: Dependencies}

		mockDatabase.EXPECT().ReadServiceVersionByName(testServiceCluster.Type, testServiceCluster.Version).Return(V, nil)

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}

		_, err := hS.AddDependencies(testClusterOk, testServiceCluster)
		if err == nil {
			t.Fatalf("Expected error")
		}

	})

}
func TestGetCluster(t *testing.T) {
	l := log.New(os.Stdout, "TestGetCluster: ", log.Ldate|log.Ltime)

	t.Run("Cluster with Name", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		mockClient := mocks.NewMockGrpcClient(mockCtrl)
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		errHandler := HttpErrorHandler{}

		projectID := "some_ID_123"
		IDorName := "testClusterName"

		var existedCluster = protobuf.Cluster{Name: "testClusterName", ProjectID: projectID}

		mockDatabase.EXPECT().ReadClusterByName(projectID, IDorName).Return(&existedCluster, nil)

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		c, _ := hS.getCluster(projectID, IDorName)

		if c.Name == "" {
			t.Fatalf("Expected existing cluster")
		}
	})

	t.Run("Cluster with ID", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		mockClient := mocks.NewMockGrpcClient(mockCtrl)
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		errHandler := HttpErrorHandler{}

		projectID := "some_ID_123"
		IDorName := "e2246d19-1221-416e-8c49-ad6dac00000a"

		var existedCluster = protobuf.Cluster{Name: "testClusterName", ProjectID: projectID}

		mockDatabase.EXPECT().ReadCluster(IDorName).Return(&existedCluster, nil)

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		c, _ := hS.getCluster(projectID, IDorName)

		if c.Name == "" {
			t.Fatalf("Expected existing cluster")
		}
	})

}

func TestClusterCreate(t *testing.T) {
	l := log.New(os.Stdout, "TestClusterCreate: ", log.Ldate|log.Ltime)
	projectName := "testProjectName"
	projectID := "test-TEST-UUID-123"

	mockCtrl := gomock.NewController(t)
	mockClient := mocks.NewMockGrpcClient(mockCtrl)
	mockDatabase := mocks.NewMockDatabase(mockCtrl)
	errHandler := HttpErrorHandler{}

	//testCluster := []byte(`{
	//	"DisplayName":"` + testClusterName + `",
	//	"EntityStatus": "some-status",
	//	"Services":[
	//	{
	//		"Name":"spark-test",
	//		"Type":"spark",
	//		"Config":{
	//			"hadoop-version":"2.6"
	//		},
	//		"Version":"2.1.0"
	//	}],
	//	"NHosts":1
	//}`)

	testInvalidCluster := []byte(`{
		"Name":"` + testClusterName + `",
		"EntityStatus": "some-status",
		"InvalidField":35
	}`)

	testInvalidJSON := []byte(`invalid json`)
	testBody, _ := json.Marshal(testCluster)

	t.Run("Project didn't exist", func(t *testing.T) {
		request, _ := http.NewRequest("POST", "/projects/"+projectName+"/clusters", bytes.NewBuffer(testBody))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		mockDatabase.EXPECT().ReadProjectByName(projectName).Return(&protobuf.Project{}, nil)

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		hS.ClusterCreate(response, request, httprouter.Params{{Key: "projectIdOrName", Value: projectName}})

		if response.Code != http.StatusNoContent {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusNoContent, response.Code)
		}
	})

	t.Run("Invalid cluster", func(t *testing.T) {
		request, _ := http.NewRequest("POST", "/projects/"+projectName+"/clusters", bytes.NewBuffer(testInvalidCluster))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		mockDatabase.EXPECT().ReadProjectByName(projectName).Return(&protobuf.Project{ID: "test-TEST-UUID-123", Name: projectName}, nil)

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		hS.ClusterCreate(response, request, httprouter.Params{{Key: "projectIdOrName", Value: projectName}})

		if response.Code != http.StatusBadRequest {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusBadRequest, response.Code)
		}
	})

	t.Run("Cluster didn't exist, valid JSON", func(t *testing.T) {
		request, _ := http.NewRequest("POST", "/projects/"+projectName+"/clusters", bytes.NewBuffer(testBody))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		mockDatabase.EXPECT().ReadProjectByName(projectName).Return(&protobuf.Project{ID: projectID, Name: projectName}, nil)
		mockDatabase.EXPECT().ReadClusterByName(projectID, testClusterName+"-"+projectName).Return(&protobuf.Cluster{}, nil)

		var testServiceVersion = protobuf.ServiceVersion{
			ID:          "60c18874-f41d-4f7f-a45d-8503abd53e1c",
			Version:     "testVersion",
			Description: "test",
			//Configs:              []*protobuf.ServiceConfig{&testServiceConfig},
		}

		testServiceType1 := protobuf.ServiceType{
			ID:             "60c18874-f41d-4f7f-a45d-8503abd53e1c",
			Type:           "spark",
			Description:    "test",
			DefaultVersion: "testVersion",
			Class:          "storage",
			Versions:       []*protobuf.ServiceVersion{&testServiceVersion},
		}
		testServiceType2 := protobuf.ServiceType{
			ID:          "61c18874-f41d-4f7f-a45d-8503abd53e1c",
			Type:        "test-service-type-2",
			Description: "test",
			Class:       "stand-alone",
		}

		mockDatabase.EXPECT().ListServicesTypes().Return([]protobuf.ServiceType{testServiceType1, testServiceType2}, nil)
		for _, s := range testCluster.Services {
			mockDatabase.EXPECT().ListServicesTypes().Return([]protobuf.ServiceType{testServiceType1, testServiceType2}, nil)
			log.Println(s)
		}
		mockDatabase.EXPECT().ListServicesTypes().Return([]protobuf.ServiceType{testServiceType1, testServiceType2}, nil)

		for _, s := range testCluster.Services {
			mockDatabase.EXPECT().ReadServiceType(s.Type).Return(&testServiceType1, nil)
			mockDatabase.EXPECT().ReadServiceVersionByName(testServiceType1.Type, testServiceVersion.Version).Return(&testServiceVersion, nil)
		}
		mockDatabase.EXPECT().WriteCluster(gomock.Any()).Return(nil)
		mockClient.EXPECT().StartClusterCreation(gomock.Any())

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		hS.ClusterCreate(response, request, httprouter.Params{{Key: "projectIdOrName", Value: projectName}})

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
		request, _ := http.NewRequest("POST", "/projects/"+projectName+"/clusters", bytes.NewBuffer(testBody))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		var existedCluster = protobuf.Cluster{Name: testClusterName, EntityStatus: utils.StatusFailed,
			ID: "some_ID_123", ProjectID: projectID}

		mockDatabase.EXPECT().ReadProjectByName(projectName).Return(&protobuf.Project{ID: projectID, Name: projectName}, nil)
		mockDatabase.EXPECT().ReadClusterByName(projectID, testClusterName+"-"+projectName).Return(&existedCluster, nil)
		mockClient.EXPECT().StartClusterCreation(gomock.Any())

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		hS.ClusterCreate(response, request, httprouter.Params{{Key: "projectIdOrName", Value: projectName}})

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
		request, _ := http.NewRequest("POST", "/projects/"+projectName+"/clusters", bytes.NewBuffer(testBody))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		var existedCluster = protobuf.Cluster{Name: testClusterName, EntityStatus: utils.StatusActive,
			ID: "some_ID_123", ProjectID: projectID}

		mockDatabase.EXPECT().ReadProjectByName(projectName).Return(&protobuf.Project{ID: projectID, Name: projectName}, nil)
		mockDatabase.EXPECT().ReadClusterByName(projectID, testClusterName+"-"+projectName).Return(&existedCluster, nil)

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		hS.ClusterCreate(response, request, httprouter.Params{{Key: "projectIdOrName", Value: projectName}})

		if response.Code != http.StatusBadRequest {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusBadRequest, response.Code)
		}
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		request, _ := http.NewRequest("POST", "/projects/"+projectName+"/clusters", bytes.NewBuffer(testInvalidJSON))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		mockDatabase.EXPECT().ReadProjectByName(projectName).Return(&protobuf.Project{ID: "test-TEST-UUID-123", Name: projectName}, nil)

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		hS.ClusterCreate(response, request, httprouter.Params{{Key: "projectIdOrName", Value: projectName}})

		if response.Code != http.StatusBadRequest {
			t.Fatalf("Expected status code %v, but received: %v", "400", response.Code)
		}
	})
}

func TestValidateCluster(t *testing.T) {
	l := log.New(os.Stdout, "TestValidateClusters: ", log.Ldate|log.Ltime)

	t.Run("Cluster is OK", func(t *testing.T) {

		mockCtrl := gomock.NewController(t)
		mockClient := mocks.NewMockGrpcClient(mockCtrl)
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		errHandler := HttpErrorHandler{}

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}

		var testClusterOk = protobuf.Cluster{
			DisplayName: "test",
			NHosts:      3,
			Image:       "ubuntu",
			Services:    []*protobuf.Service{&testService},
		}

		var V *protobuf.ServiceVersion = &protobuf.ServiceVersion{Version: "DefaultVersion"}

		var Version = []*protobuf.ServiceVersion{V}

		var existedServiceType = []protobuf.ServiceType{protobuf.ServiceType{Type: "spark", DefaultVersion: "DefaultVersion", Versions: Version}}

		mockDatabase.EXPECT().ListServicesTypes().Return(existedServiceType, nil)

		check := ValidateCluster(hS, &testClusterOk)
		if check != true {
			t.Fatalf("Expected status code %v, but received: %v", true, check)
		}
	})

	t.Run("Bad cluster's name", func(t *testing.T) {

		mockCtrl := gomock.NewController(t)
		mockClient := mocks.NewMockGrpcClient(mockCtrl)
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		errHandler := HttpErrorHandler{}

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}

		var testClusterBadName = protobuf.Cluster{
			DisplayName: "#test#",
			NHosts:      3,
			Image:       "ubuntu",
			Services:    []*protobuf.Service{&testService},
		}

		check := ValidateCluster(hS, &testClusterBadName)
		if check != false {
			t.Fatalf("Expected status code %v, but received: %v", false, check)
		}
	})

	t.Run("Bad cluster's hosts", func(t *testing.T) {

		mockCtrl := gomock.NewController(t)
		mockClient := mocks.NewMockGrpcClient(mockCtrl)
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		errHandler := HttpErrorHandler{}

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}

		var testClusterQuantityofHosts = protobuf.Cluster{
			DisplayName: "test",
			NHosts:      0,
			Image:       "ubuntu",
			Services:    []*protobuf.Service{&testService},
		}

		check := ValidateCluster(hS, &testClusterQuantityofHosts)
		if check != false {
			t.Fatalf("Expected status code %v, but received: %v", false, check)
		}

	})

	t.Run("Bad cluster's service", func(t *testing.T) {

		mockCtrl := gomock.NewController(t)
		mockClient := mocks.NewMockGrpcClient(mockCtrl)
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		errHandler := HttpErrorHandler{}

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}

		var testService1 = protobuf.Service{
			DisplayName: "test",
			Type:        "",
		}

		var testClusterService = protobuf.Cluster{
			DisplayName: "test",
			NHosts:      3,
			Image:       "ubuntu",
			Services:    []*protobuf.Service{&testService1},
		}

		check := ValidateCluster(hS, &testClusterService)
		if check != false {
			t.Fatalf("Expected status code %v, but received: %v", false, check)
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
	errHandler := HttpErrorHandler{}

	t.Run("Project didn't exist", func(t *testing.T) {
		request, _ := http.NewRequest("GET", "/projects/"+projectName+"/clusters"+clusterName, nil)
		response := httptest.NewRecorder()

		mockDatabase.EXPECT().ReadProjectByName(projectName).Return(&protobuf.Project{}, nil)

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		hS.ClustersGetByName(response, request, httprouter.Params{{Key: "projectIdOrName", Value: projectName},
			{Key: "clusterName", Value: clusterName}})

		if response.Code != http.StatusNoContent {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusNoContent, response.Code)
		}
	})

	t.Run("OK case", func(t *testing.T) {
		request, _ := http.NewRequest("GET", "/projects/"+projectName+"/clusters"+clusterName, nil)
		response := httptest.NewRecorder()

		projectTestID := "someID123"
		var testProjectClusters = protobuf.Cluster{Name: clusterName}

		mockDatabase.EXPECT().ReadProjectByName(projectName).Return(&protobuf.Project{Name: projectName, ID: projectTestID}, nil)
		mockDatabase.EXPECT().ReadClusterByName(projectTestID, clusterName).Return(&testProjectClusters, nil)

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		hS.ClustersGetByName(response, request, httprouter.Params{{Key: "projectIdOrName", Value: projectName},
			{Key: "clusterIdOrName", Value: clusterName}})

		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %v, but received: %v", "200", response.Code)
		}
	})

	t.Run("Cluster didn't exist", func(t *testing.T) {
		request, _ := http.NewRequest("GET", "/projects/"+projectName+"/clusters"+clusterName, nil)
		response := httptest.NewRecorder()

		projectTestID := "someID123"
		var testProjectClusters = protobuf.Cluster{}

		mockDatabase.EXPECT().ReadProjectByName(projectName).Return(&protobuf.Project{Name: projectName, ID: projectTestID}, nil)
		mockDatabase.EXPECT().ReadClusterByName(projectTestID, clusterName).Return(&testProjectClusters, nil)

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		hS.ClustersGetByName(response, request, httprouter.Params{{Key: "projectIdOrName", Value: projectName},
			{Key: "clusterIdOrName", Value: clusterName}})

		if response.Code != http.StatusNoContent {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusNoContent, response.Code)
		}
	})
}
