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
	mocks "github.com/ispras/michman/mocks"
	protobuf "github.com/ispras/michman/protobuf"
	"github.com/julienschmidt/httprouter"
)

var image = protobuf.Image{
	ID:           "",
	Name:         "testImageName",
	AnsibleUser:  "ubuntu",
	CloudImageID: "456",
}

func TestImagesGetList(t *testing.T) {
	request, _ := http.NewRequest("GET", "/images", nil)
	response := httptest.NewRecorder()

	mockCtrl := gomock.NewController(t)

	l := log.New(os.Stdout, "TestImagesGetList: ", log.Ldate|log.Ltime)
	mockClient := mocks.NewMockGrpcClient(mockCtrl)
	mockDatabase := mocks.NewMockDatabase(mockCtrl)
	errHandler := HttpErrorHandler{}
	mockDatabase.EXPECT().ListImages().Return([]protobuf.Image{}, nil)

	hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
	hS.ImagesGetList(response, request, httprouter.Params{})

	if response.Code != http.StatusOK {
		t.Fatalf("Expected status code %v, but received: %v", "200", response.Code)
	}
} 

func TestImageGet(t *testing.T) {
	l := log.New(os.Stdout, "TestImageGet: ", log.Ldate|log.Ltime)
	mockCtrl := gomock.NewController(t)
	mockClient := mocks.NewMockGrpcClient(mockCtrl)
	mockDatabase := mocks.NewMockDatabase(mockCtrl)
	errHandler := HttpErrorHandler{}
	ImageName := "testImageName"

	request, _ := http.NewRequest("GET", "/images/"+ImageName, nil)
	
	response := httptest.NewRecorder()

	mockDatabase.EXPECT().ReadImage(gomock.Any()).Return(&protobuf.Image{}, nil)
	hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}

	hS.ImageGet(response, request, httprouter.Params{{Key: "projectIdOrName", Value: ImageName}})

	if response.Code != http.StatusOK {
		t.Fatalf("Expected status code %v, but received: %v", "200", response.Code)
	}
}

func TestImageDelete(t *testing.T) {
	l := log.New(os.Stdout, "TestImageDelete: ", log.Ldate|log.Ltime)
	mockCtrl := gomock.NewController(t)
	mockClient := mocks.NewMockGrpcClient(mockCtrl)
	mockDatabase := mocks.NewMockDatabase(mockCtrl)
	imageName := "testImageName"
	errHandler := HttpErrorHandler{}

	t.Run ("Image isn't used", func(t *testing.T){
		request, _ := http.NewRequest("DELETE", "/images/"+imageName, nil)
		response := httptest.NewRecorder()

		mockDatabase.EXPECT().ListClusters().Return([]protobuf.Cluster{}, nil)
		mockDatabase.EXPECT().ListProjects().Return([]protobuf.Project{}, nil)
		mockDatabase.EXPECT().DeleteImage(gomock.Any()).Return(nil)

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		hS.ImageDelete(response, request, httprouter.Params{{Key: "imageName", Value: imageName}})

		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusOK, response.Code)
		}
	})


	t.Run ("Image is used", func(t *testing.T){
		request, _ := http.NewRequest("DELETE", "/images/"+imageName, nil)
		response := httptest.NewRecorder()

		var existedCluster = []protobuf.Cluster{protobuf.Cluster{Name: "Name2",
			ID: "some_ID_123", ProjectID: "test-TEST-UUID-123", Image:imageName }}

		var existedProject = []protobuf.Project{protobuf.Project{Name: "Name1",
			ID: "some_ID_124", DefaultImage: imageName}}

		mockDatabase.EXPECT().ListClusters().Return(existedCluster, nil)
		mockDatabase.EXPECT().ListProjects().Return(existedProject, nil)

		hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}
		hS.ImageDelete(response, request, httprouter.Params{{Key: "imageName", Value: imageName}})

		if response.Code != http.StatusBadRequest {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusBadRequest, response.Code)
		}
	})
}

func TestImagePost(t *testing.T) {
	l := log.New(os.Stdout, "TestImagePost: ", log.Ldate|log.Ltime)
	mockCtrl := gomock.NewController(t)
	mockClient := mocks.NewMockGrpcClient(mockCtrl)
	mockDatabase := mocks.NewMockDatabase(mockCtrl)
	imageName := "testImageName"

	errHandler := HttpErrorHandler{}
	hS := HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, ErrHandler: errHandler}

	t.Run("Valid JSON", func(t *testing.T) {
		testBody, _ := json.Marshal(image)
		request, _ := http.NewRequest("POST", "/images", bytes.NewReader(testBody))
		request.Header.Set("Content-Type", "application/json") // непонятно
		response := httptest.NewRecorder()

		mockDatabase.EXPECT().ReadImage(imageName).Return(&image, nil)
		mockDatabase.EXPECT().WriteImage(gomock.Any()).Return(nil)

		hS.ImagesPost(response, request, httprouter.Params{})

		var im protobuf.Image
		err := json.NewDecoder(response.Body).Decode(&im)
		if err != nil {
			t.Fatalf("Get invalid JSON")
		}

		if im.ID == "" {
			t.Fatalf("Image ID wasn't created")
		}

		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %v, but received: %v", "200", response.Code)
		}
	})

	t.Run ("Invalid JSON", func(t *testing.T){
		testBody := []byte(`this is invalid json`)
		request, _ := http.NewRequest("POST", "/images", bytes.NewBuffer(testBody))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		hS.ImagesPost(response, request, httprouter.Params{})

		if response.Code != http.StatusBadRequest {
			t.Fatalf("Expected status code %v, but received: %v", "400", response.Code)
		}
	})
}

var imageVal1 = protobuf.Image{
	ID:           "",
	Name:         "testImageName",
	AnsibleUser:  "ubuntu",
	CloudImageID: "456",
}

var imageVal2 = protobuf.Image{
	ID:           "123",
	Name:         "testImageName",
	AnsibleUser:  "ubuntu",
	CloudImageID: "456",
}


var imageVal3 = protobuf.Image{
	ID:           "",
	Name:         "",
	AnsibleUser:  "ubuntu",
	CloudImageID: "456",
}


var imageVal4 = protobuf.Image{
	ID:           "",
	Name:         "testImageName",
	AnsibleUser:  "",
	CloudImageID: "456",
}


var imageVal5 = protobuf.Image{
	ID:           "",
	Name:         "testImageName",
	AnsibleUser:  "ubuntu",
	CloudImageID: "",
}

func TestValidateImage(t *testing.T) {
	l := log.New(os.Stdout, "TestValidateImage: ", log.Ldate|log.Ltime)
	errHandler := HttpErrorHandler{}
	hS := HttpServer{Logger: l, ErrHandler: errHandler}

	t.Run ("Valid Image", func(t *testing.T){
		check, _ := validateImage(hS, &imageVal1)
		if check != true {
			t.Fatalf("Expected status code %v, but received: %v", true, check)
		}
	})
	t.Run ("ID not found", func(t *testing.T){
		check, _ := validateImage(hS, &imageVal2)
		if check != false {
			t.Fatalf("Expected status code %v, but received: %v", false, check)
		}
	})
	t.Run ("Name not found", func(t *testing.T){
		check, _ := validateImage(hS, &imageVal3)
		if check != false {
			t.Fatalf("Expected status code %v, but received: %v", false, check)
		}
	})
	t.Run ("AnsibleUser not found", func(t *testing.T){
		check, _ := validateImage(hS, &imageVal4)
		if check != false {
			t.Fatalf("Expected status code %v, but received: %v", false, check)
		}
	})
	t.Run ("CloudImageID not found", func(t *testing.T){
		check, _ := validateImage(hS, &imageVal5)
		if check != false {
			t.Fatalf("Expected status code %v, but received: %v", false, check)
		}
	})

}
