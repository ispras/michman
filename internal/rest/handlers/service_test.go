package handlers

import (
	"errors"
	gomock "github.com/golang/mock/gomock"
	mocks "github.com/ispras/michman/internal/mocks"
	protobuf "github.com/ispras/michman/internal/protobuf"
	"log"
	"os"
	"testing"
)

func TestDeleteSpaces(t *testing.T) {
	t.Run("Return True", func(t *testing.T) {
		valStr := "[ 11, 22  , 33, 4]"
		resStr := deleteSpaces(valStr)
		if resStr != "[11,22,33,4]" {
			t.Fatalf("ERROR: Invalid output string format")
		}
	})
	t.Run("Return True", func(t *testing.T) {
		valStr := "[ true, false  ,true, false  ]"
		resStr := deleteSpaces(valStr)
		if resStr != "[true,false,true,false]" {
			t.Fatalf("ERROR: Invalid output string format")
		}
	})
	t.Run("Return True", func(t *testing.T) {
		valStr := "[\"val1\"  ,\"val2\", \"val3\", \"val4\"  ]"
		resStr := deleteSpaces(valStr)
		if resStr != "[\"val1\",\"val2\",\"val3\",\"val4\"]" {
			t.Fatalf("ERROR: Invalid output string format")
		}
	})
}

func TestValidateService(t *testing.T) {

	l := log.New(os.Stdout, "TestValidateService: ", log.Ldate|log.Ltime)

	t.Run("Error service type", func(t *testing.T) {
		var testServiceTypeError = protobuf.Service{
			Type: "",
		}
		errHandler := HttpErrorHandler{}
		hS := HttpServer{Logger: l, ErrHandler: errHandler}
		check, _ := ValidateService(hS, &testServiceTypeError)
		if check != false {
			t.Fatalf("ERROR: service type can't be nil.")
		}
	})

	t.Run("ListServicesTypes error", func(t *testing.T) {
		var testService = protobuf.Service{
			Type: "int",
		}
		var testServiceExpect = []protobuf.ServiceType{}
		mockCtrl := gomock.NewController(t)
		mockClient := mocks.NewMockGrpcClient(mockCtrl)
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		mockDatabase.EXPECT().ListServicesTypes().Return(testServiceExpect, errors.New("ERROR: ListServicesTypes() returns not nil"))
		errHandler := HttpErrorHandler{}
		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		check, err := ValidateService(hS, &testService)
		if check != false {
			t.Fatalf("ERROR: %v", err)
		}
	})

	t.Run("Service type is not supported", func(t *testing.T) {
		var testService = protobuf.Service{
			Type: "test_type",
		}
		var testServiceExpect = []protobuf.ServiceType{}
		mockCtrl := gomock.NewController(t)
		mockClient := mocks.NewMockGrpcClient(mockCtrl)
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		mockDatabase.EXPECT().ListServicesTypes().Return(testServiceExpect, nil)
		errHandler := HttpErrorHandler{}
		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		check, _ := ValidateService(hS, &testService)
		if check != false {
			t.Fatalf("Service type is not supported")
		}
	})

	t.Run("Service type is supported, default version for service type is nil", func(t *testing.T) {
		var testService = protobuf.Service{
			Type:    "test_type",
			Version: "",
		}
		var testServiceExpect = []protobuf.ServiceType{
			protobuf.ServiceType{Type: "bad_type_1"},
			protobuf.ServiceType{Type: "bad_type_2"},
			protobuf.ServiceType{
				Type:           "test_type",
				DefaultVersion: "",
			},
		}
		mockCtrl := gomock.NewController(t)
		mockClient := mocks.NewMockGrpcClient(mockCtrl)
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		mockDatabase.EXPECT().ListServicesTypes().Return(testServiceExpect, nil)
		errHandler := HttpErrorHandler{}
		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		check, err := ValidateService(hS, &testService)
		if check != false {
			t.Fatalf("ERROR: %v", err)
		}
	})

	t.Run("Service type is supported, default version for service type is not supported", func(t *testing.T) {
		var testService = protobuf.Service{
			Type:    "test_type",
			Version: "",
		}
		var testServiceExpect = []protobuf.ServiceType{
			protobuf.ServiceType{Type: "bad_type_1"},
			protobuf.ServiceType{Type: "bad_type_2"},
			protobuf.ServiceType{
				Type:           "test_type",
				DefaultVersion: "TestDefaultVersion",
			},
		}
		mockCtrl := gomock.NewController(t)
		mockClient := mocks.NewMockGrpcClient(mockCtrl)
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		mockDatabase.EXPECT().ListServicesTypes().Return(testServiceExpect, nil)
		errHandler := HttpErrorHandler{}
		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		check, err := ValidateService(hS, &testService)
		if check != false {
			t.Fatalf("ERROR: %v", err)
		}
	})

	t.Run("Service config param name is not supported", func(t *testing.T) {
		testMapService := map[string]string{
			"config_1": "someInformation_1",
			"config_2": "someInformation_2",
			"config_3": "someInformation_3",
		}
		var testService = protobuf.Service{
			Type:    "test_type",
			Version: "",
			Config:  testMapService,
		}

		var testServiceConfig = []*protobuf.ServiceConfig{
			&protobuf.ServiceConfig{ParameterName: "bad_config_1"},
			&protobuf.ServiceConfig{ParameterName: "bad_config_2"},
			&protobuf.ServiceConfig{ParameterName: "config_2"},
		}

		var testServiceVersion = []*protobuf.ServiceVersion{
			&protobuf.ServiceVersion{Version: "test_1"},
			&protobuf.ServiceVersion{
				Version: "TestDefaultVersion",
				Configs: testServiceConfig,
			},
			&protobuf.ServiceVersion{Version: "test_2"},
		}

		var testServiceExpect = []protobuf.ServiceType{
			protobuf.ServiceType{Type: "bad_type_1"},
			protobuf.ServiceType{Type: "bad_type_2"},
			protobuf.ServiceType{
				Type:           "test_type",
				DefaultVersion: "TestDefaultVersion",
				Versions:       testServiceVersion,
			},
		}
		mockCtrl := gomock.NewController(t)
		mockClient := mocks.NewMockGrpcClient(mockCtrl)
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		mockDatabase.EXPECT().ListServicesTypes().Return(testServiceExpect, nil)
		errHandler := HttpErrorHandler{}
		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		check, _ := ValidateService(hS, &testService)
		if check != false {
			t.Fatalf("ERROR: service config param name is not supported.")
		}
	})

	t.Run("Config param is LIST, but value isn't LIST", func(t *testing.T) {
		testMapService := map[string]string{
			"config_1": "[1123]",
			"config_2": "true",
		}
		var testService = protobuf.Service{
			Type:    "test_type",
			Version: "",
			Config:  testMapService,
		}

		var testServiceConfig = []*protobuf.ServiceConfig{
			&protobuf.ServiceConfig{ParameterName: "config_1", Type: "int", IsList: true},
			&protobuf.ServiceConfig{ParameterName: "config_2", Type: "bool", IsList: true},
		}

		var testServiceVersion = []*protobuf.ServiceVersion{
			&protobuf.ServiceVersion{Version: "test_1"},
			&protobuf.ServiceVersion{
				Version: "TestDefaultVersion",
				Configs: testServiceConfig,
			},
			&protobuf.ServiceVersion{Version: "test_2"},
		}

		var testServiceExpect = []protobuf.ServiceType{
			protobuf.ServiceType{Type: "bad_type_1"},
			protobuf.ServiceType{Type: "bad_type_2"},
			protobuf.ServiceType{
				Type:           "test_type",
				DefaultVersion: "TestDefaultVersion",
				Versions:       testServiceVersion,
			},
		}
		mockCtrl := gomock.NewController(t)
		mockClient := mocks.NewMockGrpcClient(mockCtrl)
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		mockDatabase.EXPECT().ListServicesTypes().Return(testServiceExpect, nil)
		errHandler := HttpErrorHandler{}
		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		check, _ := ValidateService(hS, &testService)
		if check != false {
			t.Fatalf("ERROR: incorrect bool list int config.")
		}
	})

	t.Run("Config param isn't LIST, but value is LIST", func(t *testing.T) {
		testMapService := map[string]string{
			"config_1": "[1123, 12, 111]",
			"config_2": "true",
		}
		var testService = protobuf.Service{
			Type:    "test_type",
			Version: "",
			Config:  testMapService,
		}

		var testServiceConfig = []*protobuf.ServiceConfig{
			&protobuf.ServiceConfig{ParameterName: "config_1", Type: "int"},
			&protobuf.ServiceConfig{ParameterName: "config_2", Type: "bool"},
		}

		var testServiceVersion = []*protobuf.ServiceVersion{
			&protobuf.ServiceVersion{Version: "test_1"},
			&protobuf.ServiceVersion{
				Version: "TestDefaultVersion",
				Configs: testServiceConfig,
			},
			&protobuf.ServiceVersion{Version: "test_2"},
		}

		var testServiceExpect = []protobuf.ServiceType{
			protobuf.ServiceType{Type: "bad_type_1"},
			protobuf.ServiceType{Type: "bad_type_2"},
			protobuf.ServiceType{
				Type:           "test_type",
				DefaultVersion: "TestDefaultVersion",
				Versions:       testServiceVersion,
			},
		}
		mockCtrl := gomock.NewController(t)
		mockClient := mocks.NewMockGrpcClient(mockCtrl)
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		mockDatabase.EXPECT().ListServicesTypes().Return(testServiceExpect, nil)
		errHandler := HttpErrorHandler{}
		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		check, _ := ValidateService(hS, &testService)
		if check != false {
			t.Fatalf("ERROR: incorrect int config.")
		}
	})

	t.Run("sc.PossibleValues == nil, type INT fail", func(t *testing.T) {
		testMapService := map[string]string{
			"config_1": "not INT",
			"config_2": "+Inf",
			"config_3": "true",
		}
		var testService = protobuf.Service{
			Type:    "test_type",
			Version: "",
			Config:  testMapService,
		}

		var testServiceConfig = []*protobuf.ServiceConfig{
			&protobuf.ServiceConfig{ParameterName: "config_1", PossibleValues: nil, Type: "int"},
			&protobuf.ServiceConfig{ParameterName: "config_2", PossibleValues: nil, Type: "float"},
			&protobuf.ServiceConfig{ParameterName: "config_3", PossibleValues: nil, Type: "bool"},
		}

		var testServiceVersion = []*protobuf.ServiceVersion{
			&protobuf.ServiceVersion{Version: "test_1"},
			&protobuf.ServiceVersion{
				Version: "TestDefaultVersion",
				Configs: testServiceConfig,
			},
			&protobuf.ServiceVersion{Version: "test_2"},
		}

		var testServiceExpect = []protobuf.ServiceType{
			protobuf.ServiceType{Type: "bad_type_1"},
			protobuf.ServiceType{Type: "bad_type_2"},
			protobuf.ServiceType{
				Type:           "test_type",
				DefaultVersion: "TestDefaultVersion",
				Versions:       testServiceVersion,
			},
		}
		mockCtrl := gomock.NewController(t)
		mockClient := mocks.NewMockGrpcClient(mockCtrl)
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		mockDatabase.EXPECT().ListServicesTypes().Return(testServiceExpect, nil)
		errHandler := HttpErrorHandler{}
		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		check, _ := ValidateService(hS, &testService)
		if check != false {
			t.Fatalf("ERROR: incorrect int config")
		}
	})

	t.Run("sc.PossibleValues == nil, type FLOAT fail", func(t *testing.T) {
		testMapService := map[string]string{
			"config_1": "123456789",
			"config_2": "Not_FLOAT",
			"config_3": "true",
		}
		var testService = protobuf.Service{
			Type:    "test_type",
			Version: "",
			Config:  testMapService,
		}

		var testServiceConfig = []*protobuf.ServiceConfig{
			&protobuf.ServiceConfig{ParameterName: "config_1", PossibleValues: nil, Type: "int"},
			&protobuf.ServiceConfig{ParameterName: "config_2", PossibleValues: nil, Type: "float"},
			&protobuf.ServiceConfig{ParameterName: "config_3", PossibleValues: nil, Type: "bool"},
		}

		var testServiceVersion = []*protobuf.ServiceVersion{
			&protobuf.ServiceVersion{Version: "test_1"},
			&protobuf.ServiceVersion{
				Version: "TestDefaultVersion",
				Configs: testServiceConfig,
			},
			&protobuf.ServiceVersion{Version: "test_2"},
		}

		var testServiceExpect = []protobuf.ServiceType{
			protobuf.ServiceType{Type: "bad_type_1"},
			protobuf.ServiceType{Type: "bad_type_2"},
			protobuf.ServiceType{
				Type:           "test_type",
				DefaultVersion: "TestDefaultVersion",
				Versions:       testServiceVersion,
			},
		}
		mockCtrl := gomock.NewController(t)
		mockClient := mocks.NewMockGrpcClient(mockCtrl)
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		mockDatabase.EXPECT().ListServicesTypes().Return(testServiceExpect, nil)
		errHandler := HttpErrorHandler{}
		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		check, _ := ValidateService(hS, &testService)
		if check != false {
			t.Fatalf("ERROR: incorrect float config")
		}
	})

	t.Run("sc.PossibleValues == nil, type BOOL fail", func(t *testing.T) {
		testMapService := map[string]string{
			"config_1": "123456789",
			"config_2": "+Inf",
			"config_3": "Not bool",
		}
		var testService = protobuf.Service{
			Type:    "test_type",
			Version: "",
			Config:  testMapService,
		}

		var testServiceConfig = []*protobuf.ServiceConfig{
			&protobuf.ServiceConfig{ParameterName: "config_1", PossibleValues: nil, Type: "int"},
			&protobuf.ServiceConfig{ParameterName: "config_2", PossibleValues: nil, Type: "float"},
			&protobuf.ServiceConfig{ParameterName: "config_3", PossibleValues: nil, Type: "bool"},
		}

		var testServiceVersion = []*protobuf.ServiceVersion{
			&protobuf.ServiceVersion{Version: "test_1"},
			&protobuf.ServiceVersion{
				Version: "TestDefaultVersion",
				Configs: testServiceConfig,
			},
			&protobuf.ServiceVersion{Version: "test_2"},
		}

		var testServiceExpect = []protobuf.ServiceType{
			protobuf.ServiceType{Type: "bad_type_1"},
			protobuf.ServiceType{Type: "bad_type_2"},
			protobuf.ServiceType{
				Type:           "test_type",
				DefaultVersion: "TestDefaultVersion",
				Versions:       testServiceVersion,
			},
		}
		mockCtrl := gomock.NewController(t)
		mockClient := mocks.NewMockGrpcClient(mockCtrl)
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		mockDatabase.EXPECT().ListServicesTypes().Return(testServiceExpect, nil)
		errHandler := HttpErrorHandler{}
		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		check, _ := ValidateService(hS, &testService)
		if check != false {
			t.Fatalf("ERROR: incorrect bool config")
		}
	})

	t.Run("sc.PossibleValues == nil, type INT LIST fail", func(t *testing.T) {
		testMapService := map[string]string{
			"config_1": "[not int1, not int2, not int3]",
			"config_2": "+Inf",
			"config_3": "true",
		}
		var testService = protobuf.Service{
			Type:    "test_type",
			Version: "",
			Config:  testMapService,
		}

		var testServiceConfig = []*protobuf.ServiceConfig{
			&protobuf.ServiceConfig{ParameterName: "config_1", PossibleValues: nil, Type: "int", IsList: true},
			&protobuf.ServiceConfig{ParameterName: "config_2", PossibleValues: nil, Type: "float", IsList: false},
			&protobuf.ServiceConfig{ParameterName: "config_3", PossibleValues: nil, Type: "bool", IsList: false},
		}

		var testServiceVersion = []*protobuf.ServiceVersion{
			&protobuf.ServiceVersion{Version: "test_1"},
			&protobuf.ServiceVersion{
				Version: "TestDefaultVersion",
				Configs: testServiceConfig,
			},
			&protobuf.ServiceVersion{Version: "test_2"},
		}

		var testServiceExpect = []protobuf.ServiceType{
			protobuf.ServiceType{Type: "bad_type_1"},
			protobuf.ServiceType{Type: "bad_type_2"},
			protobuf.ServiceType{
				Type:           "test_type",
				DefaultVersion: "TestDefaultVersion",
				Versions:       testServiceVersion,
			},
		}
		mockCtrl := gomock.NewController(t)
		mockClient := mocks.NewMockGrpcClient(mockCtrl)
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		mockDatabase.EXPECT().ListServicesTypes().Return(testServiceExpect, nil)
		errHandler := HttpErrorHandler{}
		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		check, _ := ValidateService(hS, &testService)
		if check != false {
			t.Fatalf("ERROR: incorrect int list config")
		}
	})

	t.Run("sc.PossibleValues == nil, type BOOL LIST fail", func(t *testing.T) {
		testMapService := map[string]string{
			"config_1": "12345",
			"config_2": "+Inf",
			"config_3": "[not bool1, not bool2, not bool3]",
		}
		var testService = protobuf.Service{
			Type:    "test_type",
			Version: "",
			Config:  testMapService,
		}

		var testServiceConfig = []*protobuf.ServiceConfig{
			&protobuf.ServiceConfig{ParameterName: "config_1", PossibleValues: nil, Type: "int", IsList: false},
			&protobuf.ServiceConfig{ParameterName: "config_2", PossibleValues: nil, Type: "float", IsList: false},
			&protobuf.ServiceConfig{ParameterName: "config_3", PossibleValues: nil, Type: "bool", IsList: true},
		}

		var testServiceVersion = []*protobuf.ServiceVersion{
			&protobuf.ServiceVersion{Version: "test_1"},
			&protobuf.ServiceVersion{
				Version: "TestDefaultVersion",
				Configs: testServiceConfig,
			},
			&protobuf.ServiceVersion{Version: "test_2"},
		}

		var testServiceExpect = []protobuf.ServiceType{
			protobuf.ServiceType{Type: "bad_type_1"},
			protobuf.ServiceType{Type: "bad_type_2"},
			protobuf.ServiceType{
				Type:           "test_type",
				DefaultVersion: "TestDefaultVersion",
				Versions:       testServiceVersion,
			},
		}
		mockCtrl := gomock.NewController(t)
		mockClient := mocks.NewMockGrpcClient(mockCtrl)
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		mockDatabase.EXPECT().ListServicesTypes().Return(testServiceExpect, nil)
		errHandler := HttpErrorHandler{}
		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		check, _ := ValidateService(hS, &testService)
		if check != false {
			t.Fatalf("ERROR: incorrect bool list config")
		}
	})

	t.Run("sc.PossibleValues == nil, type FLOAT LIST fail", func(t *testing.T) {
		testMapService := map[string]string{
			"config_1": "12345",
			"config_2": "[not float1, not float2]",
			"config_3": "false",
		}
		var testService = protobuf.Service{
			Type:    "test_type",
			Version: "",
			Config:  testMapService,
		}

		var testServiceConfig = []*protobuf.ServiceConfig{
			&protobuf.ServiceConfig{ParameterName: "config_1", PossibleValues: nil, Type: "int", IsList: false},
			&protobuf.ServiceConfig{ParameterName: "config_2", PossibleValues: nil, Type: "float", IsList: true},
			&protobuf.ServiceConfig{ParameterName: "config_3", PossibleValues: nil, Type: "bool", IsList: false},
		}

		var testServiceVersion = []*protobuf.ServiceVersion{
			&protobuf.ServiceVersion{Version: "test_1"},
			&protobuf.ServiceVersion{
				Version: "TestDefaultVersion",
				Configs: testServiceConfig,
			},
			&protobuf.ServiceVersion{Version: "test_2"},
		}

		var testServiceExpect = []protobuf.ServiceType{
			protobuf.ServiceType{Type: "bad_type_1"},
			protobuf.ServiceType{Type: "bad_type_2"},
			protobuf.ServiceType{
				Type:           "test_type",
				DefaultVersion: "TestDefaultVersion",
				Versions:       testServiceVersion,
			},
		}
		mockCtrl := gomock.NewController(t)
		mockClient := mocks.NewMockGrpcClient(mockCtrl)
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		mockDatabase.EXPECT().ListServicesTypes().Return(testServiceExpect, nil)
		errHandler := HttpErrorHandler{}
		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		check, _ := ValidateService(hS, &testService)
		if check != false {
			t.Fatalf("ERROR: incorrect float list config")
		}
	})

	t.Run("sc.PossibleValues == nil, type STRING LIST fail", func(t *testing.T) {
		testMapService := map[string]string{
			"config_1": "12345",
			"config_2": "[1, 2, 45, 0]",
		}
		var testService = protobuf.Service{
			Type:    "test_type",
			Version: "",
			Config:  testMapService,
		}

		var testServiceConfig = []*protobuf.ServiceConfig{
			&protobuf.ServiceConfig{ParameterName: "config_1", PossibleValues: nil, Type: "int", IsList: false},
			&protobuf.ServiceConfig{ParameterName: "config_2", PossibleValues: nil, Type: "string", IsList: true},
		}

		var testServiceVersion = []*protobuf.ServiceVersion{
			&protobuf.ServiceVersion{Version: "test_1"},
			&protobuf.ServiceVersion{
				Version: "TestDefaultVersion",
				Configs: testServiceConfig,
			},
			&protobuf.ServiceVersion{Version: "test_2"},
		}

		var testServiceExpect = []protobuf.ServiceType{
			protobuf.ServiceType{Type: "bad_type_1"},
			protobuf.ServiceType{Type: "bad_type_2"},
			protobuf.ServiceType{
				Type:           "test_type",
				DefaultVersion: "TestDefaultVersion",
				Versions:       testServiceVersion,
			},
		}
		mockCtrl := gomock.NewController(t)
		mockClient := mocks.NewMockGrpcClient(mockCtrl)
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		mockDatabase.EXPECT().ListServicesTypes().Return(testServiceExpect, nil)
		errHandler := HttpErrorHandler{}
		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		check, _ := ValidateService(hS, &testService)
		if check != false {
			t.Fatalf("ERROR: incorrect string list config")
		}
	})

	t.Run("sc.PossibleValues == nil, type is OK", func(t *testing.T) {
		testMapService := map[string]string{
			"config_1": "123456789",
			"config_2": "+Inf",
			"config_3": "true",
			"config_4": "string",
			"config_5": "[1, 2, 3, 4, 5]",
			"config_6": "[0.0, null, 1.2, 144.665]",
			"config_7": "[true, true, false, true, false]",
			"config_8": "[\"string1\", \"string2\", \"string3\"]",
		}
		var testService = protobuf.Service{
			Type:    "test_type",
			Version: "",
			Config:  testMapService,
		}

		var testServiceConfig = []*protobuf.ServiceConfig{
			&protobuf.ServiceConfig{ParameterName: "config_1", PossibleValues: nil, Type: "int"},
			&protobuf.ServiceConfig{ParameterName: "config_2", PossibleValues: nil, Type: "float"},
			&protobuf.ServiceConfig{ParameterName: "config_3", PossibleValues: nil, Type: "bool"},
			&protobuf.ServiceConfig{ParameterName: "config_4", PossibleValues: nil, Type: "string"},
			&protobuf.ServiceConfig{ParameterName: "config_5", PossibleValues: nil, Type: "int", IsList: true},
			&protobuf.ServiceConfig{ParameterName: "config_6", PossibleValues: nil, Type: "float", IsList: true},
			&protobuf.ServiceConfig{ParameterName: "config_7", PossibleValues: nil, Type: "bool", IsList: true},
			&protobuf.ServiceConfig{ParameterName: "config_8", PossibleValues: nil, Type: "string", IsList: true},
		}

		var testServiceVersion = []*protobuf.ServiceVersion{
			&protobuf.ServiceVersion{Version: "test_1"},
			&protobuf.ServiceVersion{
				Version: "TestDefaultVersion",
				Configs: testServiceConfig,
			},
			&protobuf.ServiceVersion{Version: "test_2"},
		}

		var testServiceExpect = []protobuf.ServiceType{
			protobuf.ServiceType{Type: "bad_type_1"},
			protobuf.ServiceType{Type: "bad_type_2"},
			protobuf.ServiceType{
				Type:           "test_type",
				DefaultVersion: "TestDefaultVersion",
				Versions:       testServiceVersion,
			},
		}
		mockCtrl := gomock.NewController(t)
		mockClient := mocks.NewMockGrpcClient(mockCtrl)
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		mockDatabase.EXPECT().ListServicesTypes().Return(testServiceExpect, nil)
		errHandler := HttpErrorHandler{}
		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		check, _ := ValidateService(hS, &testService)
		if check != true {
			t.Fatalf("ERROR: unexpected error")
		}
	})

	t.Run("sc.PossibleValues != nil, PossibleValues are OK, value isn't list", func(t *testing.T) {
		testMapService := map[string]string{
			"config_1": "123456789",
			"config_2": "+Inf",
			"config_3": "true",
			"config_4": "value1",
		}
		var testService = protobuf.Service{
			Type:    "test_type",
			Version: "",
			Config:  testMapService,
		}

		var testServiceConfig = []*protobuf.ServiceConfig{
			&protobuf.ServiceConfig{ParameterName: "config_1", PossibleValues: []string{"val11", "123456789", "val13"}, Type: "int"},
			&protobuf.ServiceConfig{ParameterName: "config_2", PossibleValues: []string{"+Inf", "val22", "val23"}, Type: "float"},
			&protobuf.ServiceConfig{ParameterName: "config_3", PossibleValues: []string{"val31", "val32", "true"}, Type: "bool"},
			&protobuf.ServiceConfig{ParameterName: "config_4", PossibleValues: []string{"val2", "value1", "val3"}, Type: "string"},
		}

		var testServiceVersion = []*protobuf.ServiceVersion{
			&protobuf.ServiceVersion{Version: "test_1"},
			&protobuf.ServiceVersion{
				Version: "TestDefaultVersion",
				Configs: testServiceConfig,
			},
			&protobuf.ServiceVersion{Version: "test_2"},
		}

		var testServiceExpect = []protobuf.ServiceType{
			protobuf.ServiceType{Type: "bad_type_1"},
			protobuf.ServiceType{Type: "bad_type_2"},
			protobuf.ServiceType{
				Type:           "test_type",
				DefaultVersion: "TestDefaultVersion",
				Versions:       testServiceVersion,
			},
		}
		mockCtrl := gomock.NewController(t)
		mockClient := mocks.NewMockGrpcClient(mockCtrl)
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		mockDatabase.EXPECT().ListServicesTypes().Return(testServiceExpect, nil)
		errHandler := HttpErrorHandler{}
		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		check, _ := ValidateService(hS, &testService)
		if check != true {
			t.Fatalf("ERROR: unexpected error")
		}
	})

	t.Run("sc.PossibleValues != nil, PossibleValues are not OK, value isn't list", func(t *testing.T) {
		testMapService := map[string]string{
			"config_1": "123",
			"config_2": "15.3",
			"config_3": "true",
		}
		var testService = protobuf.Service{
			Type:    "test_type",
			Version: "",
			Config:  testMapService,
		}

		var testServiceConfig = []*protobuf.ServiceConfig{
			&protobuf.ServiceConfig{ParameterName: "config_1", PossibleValues: []string{"123", "456", "789"}, Type: "int"},
			&protobuf.ServiceConfig{ParameterName: "config_2", PossibleValues: []string{"+Inf", "2.0", "0.0"}, Type: "float"},
			&protobuf.ServiceConfig{ParameterName: "config_3", PossibleValues: []string{"val31", "val32", "true"}, Type: "bool"},
		}

		var testServiceVersion = []*protobuf.ServiceVersion{
			&protobuf.ServiceVersion{Version: "test_1"},
			&protobuf.ServiceVersion{
				Version: "TestDefaultVersion",
				Configs: testServiceConfig,
			},
			&protobuf.ServiceVersion{Version: "test_2"},
		}

		var testServiceExpect = []protobuf.ServiceType{
			protobuf.ServiceType{Type: "bad_type_1"},
			protobuf.ServiceType{Type: "bad_type_2"},
			protobuf.ServiceType{
				Type:           "test_type",
				DefaultVersion: "TestDefaultVersion",
				Versions:       testServiceVersion,
			},
		}
		mockCtrl := gomock.NewController(t)
		mockClient := mocks.NewMockGrpcClient(mockCtrl)
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		mockDatabase.EXPECT().ListServicesTypes().Return(testServiceExpect, nil)
		errHandler := HttpErrorHandler{}
		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		check, _ := ValidateService(hS, &testService)
		if check != false {
			t.Fatalf("ERROR: service config param value is not supported")
		}
	})

	t.Run("sc.PossibleValues != nil, PossibleValues are OK, value is list", func(t *testing.T) {
		testMapService := map[string]string{
			"config_1": "[123, 456]",
			"config_2": "[0.0, 2.0]",
			"config_3": "[true, true, false, true]",
			"config_4": "[\"val1\", \"val3\", \"val2\"]",
			"config_5": "3.0485",
		}
		var testService = protobuf.Service{
			Type:    "test_type",
			Version: "",
			Config:  testMapService,
		}

		var testServiceConfig = []*protobuf.ServiceConfig{
			&protobuf.ServiceConfig{ParameterName: "config_1", PossibleValues: []string{"[12,15,3,0]", "[123,456]", "[1,2,3,4,5]"}, Type: "int", IsList: true},
			&protobuf.ServiceConfig{ParameterName: "config_2", PossibleValues: []string{"[2.0,1.0]", "[0.0,2.0]", "[3.4, 1.1]"}, Type: "float", IsList: true},
			&protobuf.ServiceConfig{ParameterName: "config_3", PossibleValues: []string{"[true,true,false,true]", "[false,true]"}, Type: "bool", IsList: true},
			&protobuf.ServiceConfig{ParameterName: "config_4", PossibleValues: []string{"[\"val1\",\"val3\",\"val2\"]", "[\"val4\",\"val5\"]"}, Type: "string", IsList: true},
			&protobuf.ServiceConfig{ParameterName: "config_5", PossibleValues: []string{"+Inf", "2.0", "0.0", "3.0485"}, Type: "float"},
		}

		var testServiceVersion = []*protobuf.ServiceVersion{
			&protobuf.ServiceVersion{Version: "test_1"},
			&protobuf.ServiceVersion{
				Version: "TestDefaultVersion",
				Configs: testServiceConfig,
			},
			&protobuf.ServiceVersion{Version: "test_2"},
		}

		var testServiceExpect = []protobuf.ServiceType{
			protobuf.ServiceType{Type: "bad_type_1"},
			protobuf.ServiceType{Type: "bad_type_2"},
			protobuf.ServiceType{
				Type:           "test_type",
				DefaultVersion: "TestDefaultVersion",
				Versions:       testServiceVersion,
			},
		}
		mockCtrl := gomock.NewController(t)
		mockClient := mocks.NewMockGrpcClient(mockCtrl)
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		mockDatabase.EXPECT().ListServicesTypes().Return(testServiceExpect, nil)
		errHandler := HttpErrorHandler{}
		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		check, _ := ValidateService(hS, &testService)
		if check != true {
			t.Fatalf("ERROR: service config param value is not supported")
		}
	})

	t.Run("sc.PossibleValues != nil, PossibleValues aren't OK, value is list", func(t *testing.T) {
		testMapService := map[string]string{
			"config_1": "[15, 46]",
			"config_2": "[0.0, 2.0]",
			"config_3": "[false, false, true]",
			"config_4": "[\"val1\", \"val32\", \"val2\"]",
		}
		var testService = protobuf.Service{
			Type:    "test_type",
			Version: "",
			Config:  testMapService,
		}

		var testServiceConfig = []*protobuf.ServiceConfig{
			&protobuf.ServiceConfig{ParameterName: "config_1", PossibleValues: []string{"[123,456,789]", "[15,41]"}, Type: "int", IsList: true},
			&protobuf.ServiceConfig{ParameterName: "config_2", PossibleValues: []string{"[+Inf,-Inf]", "[2.0,0.0]", "[3.4,1.012]"}, Type: "float", IsList: true},
			&protobuf.ServiceConfig{ParameterName: "config_3", PossibleValues: []string{"[false,true,false]"}, Type: "bool", IsList: true},
			&protobuf.ServiceConfig{ParameterName: "config_4", PossibleValues: []string{"[\"val1\",\"val2\",\"val3\",\"val4\"]"}, Type: "string", IsList: true},
		}

		var testServiceVersion = []*protobuf.ServiceVersion{
			&protobuf.ServiceVersion{Version: "test_1"},
			&protobuf.ServiceVersion{
				Version: "TestDefaultVersion",
				Configs: testServiceConfig,
			},
			&protobuf.ServiceVersion{Version: "test_2"},
		}

		var testServiceExpect = []protobuf.ServiceType{
			protobuf.ServiceType{Type: "bad_type_1"},
			protobuf.ServiceType{Type: "bad_type_2"},
			protobuf.ServiceType{
				Type:           "test_type",
				DefaultVersion: "TestDefaultVersion",
				Versions:       testServiceVersion,
			},
		}
		mockCtrl := gomock.NewController(t)
		mockClient := mocks.NewMockGrpcClient(mockCtrl)
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		mockDatabase.EXPECT().ListServicesTypes().Return(testServiceExpect, nil)
		errHandler := HttpErrorHandler{}
		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		check, _ := ValidateService(hS, &testService)
		if check != false {
			t.Fatalf("ERROR: service config param value is not supported")
		}
	})
}
