package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/ispras/michman/internal/mocks"
	protobuf "github.com/ispras/michman/protobuf"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var serviceType = "test-service-type"
var svId = "60c18874-f41d-4f7f-a45d-8503abd53e1c"

var testServiceConfig = protobuf.ServiceConfig{
	ParameterName: "test",
	Type:          "string",
	DefaultValue:  "t",
	Required:      false,
	Description:   "test param",
}

var testServiceVersion = protobuf.ServiceVersion{
	ID:          "60c18874-f41d-4f7f-a45d-8503abd53e1c",
	Version:     "testVersion",
	Description: "test",
	//Configs:              []*protobuf.ServiceConfig{&testServiceConfig},
}

var testServiceType = protobuf.ServiceType{
	ID:             "60c18874-f41d-4f7f-a45d-8503abd53e1c",
	Type:           serviceType,
	Description:    "test",
	DefaultVersion: "testVersion",
	Versions:       []*protobuf.ServiceVersion{&testServiceVersion},
	Class:          "storage",
}

func TestIsValidType(t *testing.T) {
	t.Run("Return True", func(t *testing.T) {
		check := IsValidType("int")
		if check != true {
			t.Fatalf("ERROR: return not true")
		}
	})

	t.Run("Return False", func(t *testing.T) {
		check := IsValidType("wrongString")
		if check != false {
			t.Fatalf("ERROR: return not false")
		}
	})
}

func TestCheckVersionUnique(t *testing.T) {
	var testStVersions = []*protobuf.ServiceVersion{
		&protobuf.ServiceVersion{Version: "testVersion_1"},
		&protobuf.ServiceVersion{Version: "testVersion_2"},
		&protobuf.ServiceVersion{Version: "testVersion_3"},
	}
	t.Run("Return True", func(t *testing.T) {
		var testNewVersion = protobuf.ServiceVersion{
			Version: "testVersion_unique",
		}
		check := checkVersionUnique(testStVersions, testNewVersion)
		if check != true {
			t.Fatalf("ERROR: return not true")
		}
	})
	t.Run("Return False", func(t *testing.T) {
		var testNewVersion = protobuf.ServiceVersion{
			Version: "testVersion_2",
		}
		check := checkVersionUnique(testStVersions, testNewVersion)
		if check != false {
			t.Fatalf("ERROR: return not false")
		}
	})
}

func TestCheckDefaultVersion(t *testing.T) {
	var testStVersions = []*protobuf.ServiceVersion{
		&protobuf.ServiceVersion{Version: "testVersion_1"},
		&protobuf.ServiceVersion{Version: "testVersion_2"},
		&protobuf.ServiceVersion{Version: "testVersion_3"},
	}
	t.Run("Return True", func(t *testing.T) {
		var testDefaultVersion string = "testVersion_2"
		check := checkDefaultVersion(testStVersions, testDefaultVersion)
		if check != true {
			t.Fatalf("ERROR: return not true")
		}
	})
	t.Run("Return False", func(t *testing.T) {
		var testDefaultVersion string = "testBadVersion"
		check := checkDefaultVersion(testStVersions, testDefaultVersion)
		if check != false {
			t.Fatalf("ERROR: return not false")
		}
	})
}

func TestCheckConfigs(t *testing.T) {
	l := log.New(os.Stdout, "TestCheckConfigs: ", log.Ldate|log.Ltime)
	errHandler := HttpErrorHandler{}
	t.Run("IsValidType returns false", func(t *testing.T) {
		var testVConfigs = []*protobuf.ServiceConfig{
			&protobuf.ServiceConfig{ParameterName: "Name1", Type: "int"},
			&protobuf.ServiceConfig{ParameterName: "Name2", Type: "wrongType"},
			&protobuf.ServiceConfig{ParameterName: "Name3", Type: "bool"},
		}
		hS := HttpServer{Logger: l, ErrHandler: errHandler}
		check, _ := hS.checkConfigs(testVConfigs)
		if check != false {
			t.Fatalf("ERROR: return is not false")
		}
	})

	t.Run("Param name is nil", func(t *testing.T) {
		var testVConfigs = []*protobuf.ServiceConfig{
			&protobuf.ServiceConfig{ParameterName: "", Type: "int"},
			&protobuf.ServiceConfig{ParameterName: "", Type: "float"},
			&protobuf.ServiceConfig{ParameterName: "", Type: "bool"},
		}
		hS := HttpServer{Logger: l, ErrHandler: errHandler}
		check, _ := hS.checkConfigs(testVConfigs)
		if check != false {
			t.Fatalf("ERROR: param name is not nil")
		}
	})

	t.Run("param name is not unique", func(t *testing.T) {
		var testVConfigs = []*protobuf.ServiceConfig{
			&protobuf.ServiceConfig{ParameterName: "Name1", Type: "int"},
			&protobuf.ServiceConfig{ParameterName: "Name2", Type: "float"},
			&protobuf.ServiceConfig{ParameterName: "Name1", Type: "bool"},
		}
		hS := HttpServer{Logger: l, ErrHandler: errHandler}
		check, _ := hS.checkConfigs(testVConfigs)
		if check != false {
			t.Fatalf("ERROR: param name is not unique")
		}
	})

	t.Run("param name is unique", func(t *testing.T) {
		var testVConfigs = []*protobuf.ServiceConfig{
			&protobuf.ServiceConfig{ParameterName: "Name1", Type: "int"},
			&protobuf.ServiceConfig{ParameterName: "Name2", Type: "float"},
			&protobuf.ServiceConfig{ParameterName: "Name3", Type: "bool"},
		}
		hS := HttpServer{Logger: l, ErrHandler: errHandler}
		check, _ := hS.checkConfigs(testVConfigs)
		if check != true {
			t.Fatalf("ERROR: param name is unique")
		}
	})
}

func TestCheckDependency(t *testing.T) {
	l := log.New(os.Stdout, "TestCheckConfigs: ", log.Ldate|log.Ltime)
	mockCtrl := gomock.NewController(t)
	mockClient := mocks.NewMockGrpcClient(mockCtrl)
	mockDatabase := mocks.NewMockDatabase(mockCtrl)
	errHandler := HttpErrorHandler{}

	t.Run("ReadServiceType() error", func(t *testing.T) {
		var testService = protobuf.ServiceDependency{
			ServiceType: "testType",
		}
		mockDatabase.EXPECT().ReadServiceType(testService.ServiceType).Return(nil, errors.New("ReadServiceType() returns this error"))
		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		check, _ := hS.checkDependency(&testService)
		if check != false {
			t.Fatalf("ERROR: ReadServiceType() returns error")
		}
	})

	t.Run("Type is nil", func(t *testing.T) {
		var testService = protobuf.ServiceDependency{
			ServiceType: "testType",
		}
		var retStruct = protobuf.ServiceType{
			Type: "",
		}
		mockDatabase.EXPECT().ReadServiceType(testService.ServiceType).Return(&retStruct, nil)
		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		check, _ := hS.checkDependency(&testService)
		if check != false {
			t.Fatalf("ERROR: Type is not nil")
		}
	})

	t.Run("ServiceVersions nil", func(t *testing.T) {
		var testService = protobuf.ServiceDependency{
			ServiceType:     "testType",
			ServiceVersions: nil,
		}
		var retStruct = protobuf.ServiceType{
			Type: "testType",
		}
		mockDatabase.EXPECT().ReadServiceType(testService.ServiceType).Return(&retStruct, nil)
		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		check, _ := hS.checkDependency(&testService)
		if check != false {
			t.Fatalf("ERROR: ServiceVersions not nil")
		}
	})

	t.Run("DefaultServiceVersion nil", func(t *testing.T) {
		var testService = protobuf.ServiceDependency{
			ServiceType:           "testType",
			ServiceVersions:       []string{"v_1", "v_2"},
			DefaultServiceVersion: "",
		}
		var retStruct = protobuf.ServiceType{
			Type: "testType",
		}
		mockDatabase.EXPECT().ReadServiceType(testService.ServiceType).Return(&retStruct, nil)
		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		check, _ := hS.checkDependency(&testService)
		if check != false {
			t.Fatalf("ERROR: DefaultServiceVersion not nil")
		}
	})

	t.Run("Service version in dependency doesn't exist", func(t *testing.T) {
		var testService = protobuf.ServiceDependency{
			ServiceType:           "testType",
			ServiceVersions:       []string{"v_1", "v_2"},
			DefaultServiceVersion: "v_3",
		}
		var stVersions = []*protobuf.ServiceVersion{
			&protobuf.ServiceVersion{Version: "v_4"},
			&protobuf.ServiceVersion{Version: "v_5"},
			&protobuf.ServiceVersion{Version: "v_6"},
		}
		var retStruct = protobuf.ServiceType{
			Type:     "testType",
			Versions: stVersions,
		}
		mockDatabase.EXPECT().ReadServiceType(testService.ServiceType).Return(&retStruct, nil)
		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		check, _ := hS.checkDependency(&testService)
		if check != false {
			t.Fatalf("ERROR: Service version in dependency doesn't exist")
		}
	})

	t.Run("Service version in dependency exists, DefaultServiceVersion not", func(t *testing.T) {
		var testService = protobuf.ServiceDependency{
			ServiceType:           "testType",
			ServiceVersions:       []string{"v_1", "v_2", "v_3"},
			DefaultServiceVersion: "v_4",
		}
		var stVersions = []*protobuf.ServiceVersion{
			&protobuf.ServiceVersion{Version: "v_1"},
			&protobuf.ServiceVersion{Version: "v_2"},
			&protobuf.ServiceVersion{Version: "v_3"},
		}
		var retStruct = protobuf.ServiceType{
			Type:     "testType",
			Versions: stVersions,
		}
		mockDatabase.EXPECT().ReadServiceType(testService.ServiceType).Return(&retStruct, nil)
		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		check, _ := hS.checkDependency(&testService)
		if check != false {
			t.Fatalf("ERROR: Service version in dependency exists, DefaultServiceVersion not")
		}
	})

	t.Run("Service version in dependency exists, DefaultServiceVersion exists", func(t *testing.T) {
		var testService = protobuf.ServiceDependency{
			ServiceType:           "testType",
			ServiceVersions:       []string{"v_1", "v_2", "v_3"},
			DefaultServiceVersion: "v_2",
		}
		var stVersions = []*protobuf.ServiceVersion{
			&protobuf.ServiceVersion{Version: "v_1"},
			&protobuf.ServiceVersion{Version: "v_2"},
			&protobuf.ServiceVersion{Version: "v_3"},
		}
		var retStruct = protobuf.ServiceType{
			Type:     "testType",
			Versions: stVersions,
		}
		mockDatabase.EXPECT().ReadServiceType(testService.ServiceType).Return(&retStruct, nil)
		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		check, _ := hS.checkDependency(&testService)
		if check != true {
			t.Fatalf("ERROR: Service version in dependency exists, DefaultServiceVersion exists")
		}
	})
}

func TestCheckClass(t *testing.T) {
	t.Run("Return true", func(t *testing.T) {
		var testService = protobuf.ServiceType{
			Class: "master-slave",
		}
		check := checkClass(&testService)
		if check != true {
			t.Fatalf("ERROR: not true")
		}
	})

	t.Run("Return false", func(t *testing.T) {
		var testService = protobuf.ServiceType{
			Class: "badClass",
		}
		check := checkClass(&testService)
		if check != false {
			t.Fatalf("ERROR: not false")
		}
	})
}

func TestCheckPort(t *testing.T) {
	t.Run("Return true", func(t *testing.T) {
		check := checkPort(20)
		if check != true {
			t.Fatalf("ERROR: not true")
		}
	})

	t.Run("Return false", func(t *testing.T) {
		check := checkPort(-20)
		if check != false {
			t.Fatalf("ERROR: not false")
		}
	})
}

func TestConfigsGetServices(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	l := log.New(os.Stdout, "TestConfigsGetServices: ", log.Ldate|log.Ltime)
	mockClient := mocks.NewMockGrpcClient(mockCtrl)
	errHandler := HttpErrorHandler{}

	t.Run("List of services types", func(t *testing.T) {
		mockDatabase := mocks.NewMockDatabase(mockCtrl)

		request, _ := http.NewRequest("GET", "/configs", nil)
		response := httptest.NewRecorder()

		testServiceType1 := testServiceType
		testServiceType2 := protobuf.ServiceType{
			ID:          "61c18874-f41d-4f7f-a45d-8503abd53e1c",
			Type:        "test-service-type-2",
			Description: "test",
		}

		mockDatabase.EXPECT().ListServicesTypes().Return([]protobuf.ServiceType{testServiceType1, testServiceType2}, nil)

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		hS.ConfigsGetServices(response, request, httprouter.Params{})

		var sTypes []protobuf.ServiceType
		err := json.NewDecoder(response.Body).Decode(&sTypes)
		if err != nil {
			t.Fatalf("Got invalid JSON")
		}

		if len(sTypes) != 2 {
			t.Fatalf("Got wrong count of service versions")
		}

		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusOK, response.Code)
		}
	})
}

func TestConfigsCreateService(t *testing.T) {
	l := log.New(os.Stdout, "TestConfigsCreateService: ", log.Ldate|log.Ltime)
	mockCtrl := gomock.NewController(t)
	mockClient := mocks.NewMockGrpcClient(mockCtrl)
	mockDatabase := mocks.NewMockDatabase(mockCtrl)
	errHandler := HttpErrorHandler{}

	hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}

	t.Run("New service type with valid JSON", func(t *testing.T) {
		testBody, _ := json.Marshal(testServiceType)
		request, _ := http.NewRequest("POST", "/configs", bytes.NewReader(testBody))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		mockDatabase.EXPECT().ReadServiceType(serviceType).Return(&protobuf.ServiceType{}, nil)
		mockDatabase.EXPECT().WriteServiceType(gomock.Any()).Return(nil)

		hS.ConfigsCreateService(response, request, httprouter.Params{})

		var st protobuf.ServiceType
		err := json.NewDecoder(response.Body).Decode(&st)
		if err != nil {
			log.Print(st)
			t.Fatalf("Get invalid JSON")
		}

		if st.ID == "" {
			t.Fatalf("Service type ID wasn't created")
		}

		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusOK, response.Code)
		}
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		testBody := []byte(`this is invalid json`)
		request, _ := http.NewRequest("POST", "/configs", bytes.NewBuffer(testBody))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		hS.ConfigsCreateService(response, request, httprouter.Params{})

		if response.Code != http.StatusBadRequest {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusBadRequest, response.Code)
		}
	})
}

func TestConfigsGetService(t *testing.T) {
	l := log.New(os.Stdout, "TestConfigsGetService: ", log.Ldate|log.Ltime)
	mockCtrl := gomock.NewController(t)
	mockClient := mocks.NewMockGrpcClient(mockCtrl)
	mockDatabase := mocks.NewMockDatabase(mockCtrl)
	errHandler := HttpErrorHandler{}

	t.Run("Existed service type", func(t *testing.T) {
		request, _ := http.NewRequest("GET", "/configs/"+serviceType, nil)
		response := httptest.NewRecorder()
		mockDatabase.EXPECT().ReadServiceType(serviceType).Return(&testServiceType, nil)
		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}

		hS.ConfigsGetService(response, request, httprouter.Params{{Key: "serviceType", Value: serviceType}})

		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %v, but received: %v", "200", response.Code)
		}
	})

	t.Run("Not existed service type", func(t *testing.T) {
		request, _ := http.NewRequest("GET", "/configs/"+serviceType, nil)
		response := httptest.NewRecorder()
		mockDatabase.EXPECT().ReadServiceType(serviceType).Return(&protobuf.ServiceType{}, nil)
		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}

		hS.ConfigsGetService(response, request, httprouter.Params{{Key: "serviceType", Value: serviceType}})

		if response.Code != http.StatusNoContent {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusNoContent, response.Code)
		}
	})
}

func TestConfigsDeleteService(t *testing.T) {
	l := log.New(os.Stdout, "TestConfigsDeleteService: ", log.Ldate|log.Ltime)
	mockCtrl := gomock.NewController(t)
	mockClient := mocks.NewMockGrpcClient(mockCtrl)
	mockDatabase := mocks.NewMockDatabase(mockCtrl)
	errHandler := HttpErrorHandler{}

	t.Run("Existed service type", func(t *testing.T) {
		request, _ := http.NewRequest("DELETE", "/configs/"+serviceType, nil)
		response := httptest.NewRecorder()

		mockDatabase.EXPECT().ReadServiceType(serviceType).Return(&testServiceType, nil)

		testServiceType1 := testServiceType
		testServiceType2 := protobuf.ServiceType{
			ID:          "61c18874-f41d-4f7f-a45d-8503abd53e1c",
			Type:        "test-service-type-2",
			Description: "test",
		}

		mockDatabase.EXPECT().ListServicesTypes().Return([]protobuf.ServiceType{testServiceType1, testServiceType2}, nil)

		mockDatabase.EXPECT().DeleteServiceType(serviceType).Return(nil)

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		hS.ConfigsDeleteService(response, request, httprouter.Params{{Key: "serviceType", Value: serviceType}})

		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusOK, response.Code)
		}
	})

	t.Run("Delete not existed service type", func(t *testing.T) {
		request, _ := http.NewRequest("DELETE", "/configs/"+serviceType, nil)
		response := httptest.NewRecorder()

		mockDatabase.EXPECT().ReadServiceType(serviceType).Return(&protobuf.ServiceType{}, nil)
		//mockDatabase.EXPECT().DeleteServiceType(serviceType).Return(nil)

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		hS.ConfigsDeleteService(response, request, httprouter.Params{{Key: "serviceType", Value: serviceType}})

		if response.Code != http.StatusNoContent {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusNoContent, response.Code)
		}
	})
}

func TestConfigsUpdateService(t *testing.T) {
	l := log.New(os.Stdout, "TestConfigsUpdateService: ", log.Ldate|log.Ltime)
	mockCtrl := gomock.NewController(t)
	mockClient := mocks.NewMockGrpcClient(mockCtrl)
	mockDatabase := mocks.NewMockDatabase(mockCtrl)
	errHandler := HttpErrorHandler{}

	updateBody := protobuf.ServiceType{
		Description:    "updated test",
		DefaultVersion: "testVersion",
	}
	testBody, _ := json.Marshal(updateBody)

	t.Run("Update existed service type with correct body", func(t *testing.T) {
		request, _ := http.NewRequest("PUT", "/configs/"+serviceType, bytes.NewReader(testBody))
		response := httptest.NewRecorder()

		mockDatabase.EXPECT().ReadServiceType(serviceType).Return(&testServiceType, nil)
		mockDatabase.EXPECT().WriteServiceType(gomock.Any()).Return(nil)

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		hS.ConfigsUpdateService(response, request, httprouter.Params{{Key: "serviceType", Value: serviceType}})

		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusOK, response.Code)
		}
	})

	t.Run("Update not existed service type", func(t *testing.T) {
		request, _ := http.NewRequest("PUT", "/configs/"+serviceType, bytes.NewReader(testBody))
		response := httptest.NewRecorder()

		mockDatabase.EXPECT().ReadServiceType(serviceType).Return(&protobuf.ServiceType{}, nil)
		//mockDatabase.EXPECT().UpdateServiceType(gomock.Any()).Return(nil)

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		hS.ConfigsUpdateService(response, request, httprouter.Params{{Key: "serviceType", Value: serviceType}})

		if response.Code != http.StatusNoContent {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusNoContent, response.Code)
		}
	})
}

func TestConfigsCreateVersion(t *testing.T) {
	l := log.New(os.Stdout, "TestConfigsCreateVersion: ", log.Ldate|log.Ltime)
	mockCtrl := gomock.NewController(t)
	mockClient := mocks.NewMockGrpcClient(mockCtrl)
	mockDatabase := mocks.NewMockDatabase(mockCtrl)
	errHandler := HttpErrorHandler{}

	hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}

	t.Run("New service version with valid JSON", func(t *testing.T) {
		testBody, _ := json.Marshal(testServiceVersion)
		request, _ := http.NewRequest("POST", "/configs/"+serviceType+"/versions", bytes.NewReader(testBody))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		st := protobuf.ServiceType{
			ID:          testServiceType.ID,
			Type:        testServiceType.Type,
			Description: testServiceType.Description,
			Versions:    []*protobuf.ServiceVersion{},
		}

		mockDatabase.EXPECT().ReadServiceType(serviceType).Return(&st, nil)
		mockDatabase.EXPECT().UpdateServiceType(gomock.Any()).Return(nil)

		hS.ConfigsCreateVersion(response, request, httprouter.Params{{Key: "serviceType", Value: serviceType}})

		var sv protobuf.ServiceVersion
		err := json.NewDecoder(response.Body).Decode(&sv)
		if err != nil {
			t.Fatalf("Get invalid JSON")
		}

		if sv.ID == "" {
			t.Fatalf("Service type ID wasn't created")
		}

		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusOK, response.Code)
		}
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		testBody := []byte(`this is invalid json`)
		request, _ := http.NewRequest("POST", "/configs/"+serviceType+"/versions", bytes.NewReader(testBody))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		mockDatabase.EXPECT().ReadServiceType(serviceType).Return(&testServiceType, nil)

		hS.ConfigsCreateVersion(response, request, httprouter.Params{{Key: "serviceType", Value: serviceType}})

		if response.Code != http.StatusBadRequest {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusBadRequest, response.Code)
		}
	})
}

func TestConfigsGetVersion(t *testing.T) {
	l := log.New(os.Stdout, "TestConfigsGetVersion: ", log.Ldate|log.Ltime)
	mockCtrl := gomock.NewController(t)
	mockClient := mocks.NewMockGrpcClient(mockCtrl)
	mockDatabase := mocks.NewMockDatabase(mockCtrl)
	errHandler := HttpErrorHandler{}

	t.Run("Existed service version", func(t *testing.T) {
		request, _ := http.NewRequest("GET", "/configs/"+serviceType+"/versions/"+svId, nil)
		response := httptest.NewRecorder()
		mockDatabase.EXPECT().ReadServiceVersion(serviceType, svId).Return(&testServiceVersion, nil)
		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}

		hS.ConfigsGetVersion(response, request, httprouter.Params{{Key: "serviceType", Value: serviceType},
			{Key: "versionId", Value: svId}})

		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %v, but received: %v", "200", response.Code)
		}
	})

	t.Run("Not existed service version", func(t *testing.T) {
		request, _ := http.NewRequest("GET", "/configs/"+serviceType+"/versions/"+svId, nil)
		response := httptest.NewRecorder()
		mockDatabase.EXPECT().ReadServiceVersion(serviceType, svId).Return(&protobuf.ServiceVersion{}, nil)
		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}

		hS.ConfigsGetVersion(response, request, httprouter.Params{{Key: "serviceType", Value: serviceType},
			{Key: "versionId", Value: svId}})

		if response.Code != http.StatusNoContent {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusNoContent, response.Code)
		}
	})
}

func TestConfigsGetVersions(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	l := log.New(os.Stdout, "TestConfigsGetVersions: ", log.Ldate|log.Ltime)
	mockClient := mocks.NewMockGrpcClient(mockCtrl)
	errHandler := HttpErrorHandler{}

	t.Run("List of service versions", func(t *testing.T) {
		mockDatabase := mocks.NewMockDatabase(mockCtrl)

		request, _ := http.NewRequest("GET", "/configs/"+serviceType+"/versions", nil)
		response := httptest.NewRecorder()

		testServiceVersion1 := testServiceVersion
		testServiceVersion2 := protobuf.ServiceVersion{
			ID:          "61c18874-f41d-4f7f-a45d-8503abd53e1s",
			Version:     "testVersion2",
			Description: "test2",
		}

		curSt := protobuf.ServiceType{
			ID:             testServiceType.ID,
			Type:           testServiceType.Type,
			Description:    testServiceType.Description,
			Versions:       []*protobuf.ServiceVersion{&testServiceVersion1, &testServiceVersion2},
			DefaultVersion: testServiceType.DefaultVersion,
		}
		mockDatabase.EXPECT().ReadServiceType(serviceType).Return(&curSt, nil)

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		hS.ConfigsGetVersions(response, request, httprouter.Params{{Key: "serviceType", Value: serviceType}})

		var versions []protobuf.ServiceVersion
		err := json.NewDecoder(response.Body).Decode(&versions)
		if err != nil {
			t.Fatalf("Got invalid JSON")
		}

		if len(versions) != 2 {
			t.Fatalf("Got wrong count of service versions")
		}

		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusOK, response.Code)
		}
	})
}

func TestConfigsUpdateVersion(t *testing.T) {
	l := log.New(os.Stdout, "TestConfigsUpdateVersion: ", log.Ldate|log.Ltime)
	mockCtrl := gomock.NewController(t)
	mockClient := mocks.NewMockGrpcClient(mockCtrl)
	mockDatabase := mocks.NewMockDatabase(mockCtrl)
	errHandler := HttpErrorHandler{}

	updateBody := protobuf.ServiceVersion{
		Description: "updated test",
	}
	testBody, _ := json.Marshal(updateBody)

	t.Run("Update existed service type with correct body", func(t *testing.T) {
		request, _ := http.NewRequest("PUT", "/configs/"+serviceType+"/versions/"+svId, bytes.NewReader(testBody))
		response := httptest.NewRecorder()

		updatedSV := protobuf.ServiceVersion{
			ID:          testServiceVersion.ID,
			Description: updateBody.Description,
			Version:     svId,
		}

		updatedST := protobuf.ServiceType{
			ID:             testServiceType.ID,
			Type:           testServiceType.Type,
			Description:    testServiceType.Description,
			DefaultVersion: testServiceType.DefaultVersion,
			Versions:       []*protobuf.ServiceVersion{&updatedSV},
		}

		mockDatabase.EXPECT().ReadServiceType(serviceType).Return(&updatedST, nil)
		mockDatabase.EXPECT().UpdateServiceType(&updatedST).Return(nil)

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		hS.ConfigsUpdateVersion(response, request, httprouter.Params{{Key: "serviceType", Value: serviceType},
			{Key: "versionId", Value: svId}})

		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusOK, response.Code)
		}
	})

	t.Run("Update not existed service type", func(t *testing.T) {
		request, _ := http.NewRequest("PUT", "/configs/"+serviceType+"/versions/"+svId, bytes.NewReader(testBody))
		response := httptest.NewRecorder()

		mockDatabase.EXPECT().ReadServiceType(serviceType).Return(&protobuf.ServiceType{}, nil)
		//mockDatabase.EXPECT().UpdateServiceType(gomock.Any()).Return(nil)

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		hS.ConfigsUpdateVersion(response, request, httprouter.Params{{Key: "serviceType", Value: serviceType},
			{Key: "versionId", Value: svId}})

		if response.Code != http.StatusNoContent {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusNoContent, response.Code)
		}
	})
	t.Run("Update not existed service version", func(t *testing.T) {
		request, _ := http.NewRequest("PUT", "/configs/"+serviceType+"/versions/"+svId, bytes.NewReader(testBody))
		response := httptest.NewRecorder()

		updatedST := protobuf.ServiceType{
			ID:             testServiceType.ID,
			Type:           testServiceType.Type,
			Description:    testServiceType.Description,
			DefaultVersion: testServiceType.DefaultVersion,
			Versions:       []*protobuf.ServiceVersion{},
		}
		mockDatabase.EXPECT().ReadServiceType(serviceType).Return(&updatedST, nil)
		//mockDatabase.EXPECT().UpdateServiceType(gomock.Any()).Return(nil)

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		hS.ConfigsUpdateVersion(response, request, httprouter.Params{{Key: "serviceType", Value: serviceType},
			{Key: "versionId", Value: svId}})

		if response.Code != http.StatusNoContent {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusNoContent, response.Code)
		}
	})
}

func TestConfigsDeleteVersion(t *testing.T) {
	l := log.New(os.Stdout, "TestConfigsDeleteVersion: ", log.Ldate|log.Ltime)
	mockCtrl := gomock.NewController(t)
	mockClient := mocks.NewMockGrpcClient(mockCtrl)
	mockDatabase := mocks.NewMockDatabase(mockCtrl)
	errHandler := HttpErrorHandler{}

	t.Run("Existed service version", func(t *testing.T) {
		request, _ := http.NewRequest("DELETE", "/configs/"+serviceType+"/versions/"+svId, nil)
		response := httptest.NewRecorder()

		mockDatabase.EXPECT().ReadServiceVersion(serviceType, svId).Return(&testServiceVersion, nil)

		testServiceType1 := testServiceType
		testServiceType2 := protobuf.ServiceType{
			ID:          "61c18874-f41d-4f7f-a45d-8503abd53e1c",
			Type:        "test-service-type-2",
			Description: "test",
		}

		mockDatabase.EXPECT().ListServicesTypes().Return([]protobuf.ServiceType{testServiceType1, testServiceType2}, nil)

		mockDatabase.EXPECT().DeleteServiceVersion(serviceType, svId).Return(&testServiceVersion, nil)

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		hS.ConfigsDeleteVersion(response, request, httprouter.Params{{Key: "serviceType", Value: serviceType},
			{Key: "versionId", Value: svId}})

		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusOK, response.Code)
		}
	})

	t.Run("Delete not existed service version", func(t *testing.T) {
		request, _ := http.NewRequest("DELETE", "/configs/"+serviceType+"/versions/"+svId, nil)
		response := httptest.NewRecorder()

		mockDatabase.EXPECT().ReadServiceVersion(serviceType, svId).Return(&protobuf.ServiceVersion{}, nil)

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		hS.ConfigsDeleteVersion(response, request, httprouter.Params{{Key: "serviceType", Value: serviceType},
			{Key: "versionId", Value: svId}})

		if response.Code != http.StatusNoContent {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusNoContent, response.Code)
		}
	})
}

func TestConfigsCreateConfigParam(t *testing.T) {
	l := log.New(os.Stdout, "TestConfigsCreateConfigParam: ", log.Ldate|log.Ltime)
	mockCtrl := gomock.NewController(t)
	mockClient := mocks.NewMockGrpcClient(mockCtrl)
	mockDatabase := mocks.NewMockDatabase(mockCtrl)
	errHandler := HttpErrorHandler{}

	hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}

	t.Run("New service config param with valid JSON", func(t *testing.T) {
		testBody, _ := json.Marshal(testServiceConfig)
		request, _ := http.NewRequest("POST", "/configs/"+serviceType+"/versions/"+svId, bytes.NewReader(testBody))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		curSV := protobuf.ServiceVersion{
			ID:          testServiceVersion.ID,
			Description: testServiceVersion.Description,
			Version:     svId,
			Configs:     []*protobuf.ServiceConfig{},
		}

		curST := protobuf.ServiceType{
			ID:             testServiceType.ID,
			Type:           testServiceType.Type,
			Description:    testServiceType.Description,
			DefaultVersion: testServiceType.DefaultVersion,
			Versions:       []*protobuf.ServiceVersion{&curSV},
		}

		mockDatabase.EXPECT().ReadServiceType(serviceType).Return(&curST, nil)
		mockDatabase.EXPECT().UpdateServiceType(gomock.Any()).Return(nil)

		hS.ConfigsCreateConfigParam(response, request, httprouter.Params{{Key: "serviceType", Value: serviceType},
			{Key: "versionId", Value: svId}})

		var sv protobuf.ServiceVersion
		err := json.NewDecoder(response.Body).Decode(&sv)
		if err != nil {
			t.Fatalf("Get invalid JSON")
		}

		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusOK, response.Code)
		}
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		testBody := []byte(`this is invalid json`)
		request, _ := http.NewRequest("POST", "/configs/"+serviceType+"/versions/"+svId, bytes.NewReader(testBody))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		mockDatabase.EXPECT().ReadServiceType(serviceType).Return(&testServiceType, nil)
		mockDatabase.EXPECT().UpdateServiceType(gomock.Any()).Return(nil)

		hS.ConfigsCreateConfigParam(response, request, httprouter.Params{{Key: "serviceType", Value: serviceType},
			{Key: "versionId", Value: svId}})

		if response.Code != http.StatusBadRequest {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusBadRequest, response.Code)
		}
	})
	t.Run("Not existed service type", func(t *testing.T) {
		testBody, _ := json.Marshal(testServiceConfig)
		request, _ := http.NewRequest("POST", "/configs/"+serviceType+"/versions/"+svId, bytes.NewReader(testBody))
		response := httptest.NewRecorder()

		mockDatabase.EXPECT().ReadServiceType(serviceType).Return(&protobuf.ServiceType{}, nil)
		//mockDatabase.EXPECT().UpdateServiceType(gomock.Any()).Return(nil)

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		hS.ConfigsCreateConfigParam(response, request, httprouter.Params{{Key: "serviceType", Value: serviceType},
			{Key: "versionId", Value: svId}})

		//TODO: fix this test
		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusNoContent, response.Code)
		}
	})
	t.Run("Not existed service version", func(t *testing.T) {
		testBody, _ := json.Marshal(testServiceConfig)
		request, _ := http.NewRequest("POST", "/configs/"+serviceType+"/versions/"+svId, bytes.NewReader(testBody))
		response := httptest.NewRecorder()
		//
		//curST := protobuf.ServiceType{
		//	ID:             testServiceType.ID,
		//	Type:           testServiceType.Type,
		//	Description:    testServiceType.Description,
		//	DefaultVersion: testServiceType.DefaultVersion,
		//	Versions:       []*protobuf.ServiceVersion{},
		//}
		//
		//mockDatabase.EXPECT().ReadServiceType(serviceType).Return(&curST, nil)
		//mockDatabase.EXPECT().ReadServiceType(serviceType).Return(&curST, nil)
		////mockDatabase.EXPECT().UpdateServiceType(gomock.Any()).Return(nil)

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		hS.ConfigsCreateConfigParam(response, request, httprouter.Params{{Key: "serviceType", Value: serviceType},
			{Key: "versionId", Value: svId}})

		if response.Code != http.StatusNoContent {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusNoContent, response.Code)
		}
	})
}
