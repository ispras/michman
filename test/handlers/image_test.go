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

func TestImagesGetList(t *testing.T) { //ok
	request, _ := http.NewRequest("GET", "/images", nil)
	response := httptest.NewRecorder()

	mockCtrl := gomock.NewController(t)

	l := log.New(os.Stdout, "TestImagesGetList: ", log.Ldate|log.Ltime)
	mockClient := mocks.NewMockGrpcClient(mockCtrl)
	mockDatabase := mocks.NewMockDatabase(mockCtrl)
	RespHandler := handlers.HttpResponseHandler{}
	mockDatabase.EXPECT().ListImages().Return([]protobuf.Image{}, nil)

	hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, RespHandler: RespHandler}
	hS.ImagesGetList(response, request, httprouter.Params{})

	if response.Code != http.StatusOK {
		t.Fatalf("Expected status code %v, but received: %v", "200", response.Code)
	}
}

func TestImageGet(t *testing.T) { //ok
	l := log.New(os.Stdout, "TestImageGet: ", log.Ldate|log.Ltime)
	mockCtrl := gomock.NewController(t)
	mockClient := mocks.NewMockGrpcClient(mockCtrl)
	mockDatabase := mocks.NewMockDatabase(mockCtrl)
	RespHandler := handlers.HttpResponseHandler{}
	ImageName := "testImageName"

	request, _ := http.NewRequest("GET", "/images/"+ImageName, nil)

	response := httptest.NewRecorder()

	mockDatabase.EXPECT().ReadImage(gomock.Any()).Return(&protobuf.Image{}, nil)
	hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, RespHandler: RespHandler}

	hS.ImageGet(response, request, httprouter.Params{{Key: "projectIdOrName", Value: ImageName}})

	if response.Code != http.StatusOK {
		t.Fatalf("Expected status code %v, but received: %v", "200", response.Code)
	}
}

func TestImageDelete(t *testing.T) {
	l := log.New(os.Stdout, "TestImageDelete: ", log.Ldate|log.Ltime)

	t.Run("Image isn't used", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		mockClient := mocks.NewMockGrpcClient(mockCtrl)
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		imageName := "testImageName"
		RespHandler := handlers.HttpResponseHandler{}
		request, _ := http.NewRequest("DELETE", "/images/"+imageName, nil)
		response := httptest.NewRecorder()

		mockDatabase.EXPECT().ListClusters().Return([]protobuf.Cluster{}, nil)
		mockDatabase.EXPECT().ListProjects().Return([]protobuf.Project{}, nil)
		mockDatabase.EXPECT().DeleteImage(gomock.Any()).Return(nil)

		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, RespHandler: RespHandler}
		hS.ImageDelete(response, request, httprouter.Params{{Key: "imageName", Value: imageName}})

		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusOK, response.Code)
		}
	})

	t.Run("Image is used", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		mockClient := mocks.NewMockGrpcClient(mockCtrl)
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		imageName := "testImageName"
		RespHandler := handlers.HttpResponseHandler{}
		request, _ := http.NewRequest("DELETE", "/images/"+imageName, nil)
		response := httptest.NewRecorder()

		var existedCluster = []protobuf.Cluster{protobuf.Cluster{Name: "Name2",
			ID: "some_ID_123", ProjectID: "test-TEST-UUID-123", Image: imageName}}

		var existedProject = []protobuf.Project{protobuf.Project{Name: "Name1",
			ID: "some_ID_124", DefaultImage: imageName}}

		mockDatabase.EXPECT().ListClusters().Return(existedCluster, nil)
		mockDatabase.EXPECT().ListProjects().Return(existedProject, nil)

		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, RespHandler: RespHandler}
		hS.ImageDelete(response, request, httprouter.Params{{Key: "imageName", Value: imageName}})

		if response.Code != http.StatusBadRequest {
			t.Fatalf("Expected status code %v, but received: %v", http.StatusBadRequest, response.Code)
		}
	})
}

func TestImagePost(t *testing.T) { //переделать
	var image1 = protobuf.Image{
		ID:           "",
		Name:         "testImageName",
		AnsibleUser:  "ubuntu",
		CloudImageID: "456",
	}

	var image2 = protobuf.Image{
		ID:           "123",
		Name:         "testImageName",
		AnsibleUser:  "ubuntu",
		CloudImageID: "456",
	}

	l := log.New(os.Stdout, "TestImagePost: ", log.Ldate|log.Ltime)

	t.Run("Valid JSON, image isn't ok", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		mockClient := mocks.NewMockGrpcClient(mockCtrl)
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		imageName := "testImageName"

		RespHandler := handlers.HttpResponseHandler{}
		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, RespHandler: RespHandler}
		testBody, _ := json.Marshal(image2)
		request, _ := http.NewRequest("POST", "/images", bytes.NewReader(testBody))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		mockDatabase.EXPECT().ReadImage(imageName).Return(&image2, nil)
		hS.ImageCreate(response, request, httprouter.Params{})

		if response.Code != http.StatusBadRequest {
			t.Fatalf("Expected status code %v, but received: %v", "400", response.Code)
		}
	})

	t.Run("Valid JSON, image is ok", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		mockClient := mocks.NewMockGrpcClient(mockCtrl)
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		imageName := "testImageName"

		RespHandler := handlers.HttpResponseHandler{}
		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, RespHandler: RespHandler}
		testBody, _ := json.Marshal(image1)
		request, _ := http.NewRequest("POST", "/images", bytes.NewReader(testBody))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		mockDatabase.EXPECT().ReadImage(imageName).Return(&image1, nil)
		mockDatabase.EXPECT().WriteImage(gomock.Any()).Return(nil)

		hS.ImageCreate(response, request, httprouter.Params{})

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

	t.Run("Invalid JSON", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		mockClient := mocks.NewMockGrpcClient(mockCtrl)
		mockDatabase := mocks.NewMockDatabase(mockCtrl)

		RespHandler := handlers.HttpResponseHandler{}
		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, RespHandler: RespHandler}
		testBody := []byte(`this is invalid json`)
		request, _ := http.NewRequest("POST", "/images", bytes.NewBuffer(testBody))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		hS.ImageCreate(response, request, httprouter.Params{})

		if response.Code != http.StatusBadRequest {
			t.Fatalf("Expected status code %v, but received: %v", "400", response.Code)
		}
	})
}

func TestValidateImage(t *testing.T) { //ok
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
	l := log.New(os.Stdout, "TestValidateImage: ", log.Ldate|log.Ltime)
	RespHandler := handlers.HttpResponseHandler{}
	hS := handlers.HttpServer{Logger: l, RespHandler: RespHandler}

	t.Run("Valid Image", func(t *testing.T) {
		check, _ := handlers.validateImage(hS, &imageVal1)
		if check != true {
			t.Fatalf("Expected status code %v, but received: %v", true, check)
		}
	})
	t.Run("ID not found", func(t *testing.T) {
		check, _ := handlers.validateImage(hS, &imageVal2)
		if check != false {
			t.Fatalf("Expected status code %v, but received: %v", false, check)
		}
	})
	t.Run("Name not found", func(t *testing.T) {
		check, _ := handlers.validateImage(hS, &imageVal3)
		if check != false {
			t.Fatalf("Expected status code %v, but received: %v", false, check)
		}
	})
	t.Run("AnsibleUser not found", func(t *testing.T) {
		check, _ := handlers.validateImage(hS, &imageVal4)
		if check != false {
			t.Fatalf("Expected status code %v, but received: %v", false, check)
		}
	})
	t.Run("CloudImageID not found", func(t *testing.T) {
		check, _ := handlers.validateImage(hS, &imageVal5)
		if check != false {
			t.Fatalf("Expected status code %v, but received: %v", false, check)
		}
	})
}

func TestIsImageUsed(t *testing.T) { //ok
	l := log.New(os.Stdout, "TestIsImageUsed: ", log.Ldate|log.Ltime)

	t.Run("Clusters exist, Projects not exist", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		mockClient := mocks.NewMockGrpcClient(mockCtrl)
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		imageName := "testImageName"
		RespHandler := handlers.HttpResponseHandler{}
		var existedCluster = []protobuf.Cluster{protobuf.Cluster{Image: imageName}}
		mockDatabase.EXPECT().ListClusters().Return(existedCluster, nil)
		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, RespHandler: RespHandler}
		check := handlers.isImageUsed(hS, imageName)
		if check != true {
			t.Fatalf("Expected status code %v, but received: %v", true, check)
		}
	})

	t.Run("Clusters not exist, Projects exist", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		mockClient := mocks.NewMockGrpcClient(mockCtrl)
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		imageName := "testImageName"
		RespHandler := handlers.HttpResponseHandler{}
		var existedProject = []protobuf.Project{protobuf.Project{DefaultImage: imageName}}
		mockDatabase.EXPECT().ListClusters().Return([]protobuf.Cluster{}, nil)
		mockDatabase.EXPECT().ListProjects().Return(existedProject, nil)
		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, RespHandler: RespHandler}
		check := handlers.isImageUsed(hS, imageName)
		if check != true {
			t.Fatalf("Expected status code %v, but received: %v", true, check)
		}
	})

	t.Run("Clusters exist, Projects exist", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		mockClient := mocks.NewMockGrpcClient(mockCtrl)
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		imageName := "testImageName"
		RespHandler := handlers.HttpResponseHandler{}
		mockDatabase.EXPECT().ListClusters().Return([]protobuf.Cluster{}, nil)
		mockDatabase.EXPECT().ListProjects().Return([]protobuf.Project{}, nil)
		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, RespHandler: RespHandler}
		check := handlers.isImageUsed(hS, imageName)
		if check != false {
			t.Fatalf("Expected status code %v, but received: %v", false, check)
		}
	})
}

func TestImagePut(t *testing.T) { // доделать
	var imagePut = protobuf.Image{
		ID:           "",
		Name:         "testImageName",
		AnsibleUser:  "ubuntu",
		CloudImageID: "456",
	}

	var image = protobuf.Image{
		ID:           "123",
		Name:         "testImageName",
		AnsibleUser:  "ubuntu",
		CloudImageID: "456",
	}

	l := log.New(os.Stdout, "TestImagePut: ", log.Ldate|log.Ltime)

	t.Run("Valid JSON, image has no clusters and no projects, new image is ok", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		mockClient := mocks.NewMockGrpcClient(mockCtrl)
		mockDatabase := mocks.NewMockDatabase(mockCtrl)

		RespHandler := handlers.HttpResponseHandler{}
		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, RespHandler: RespHandler}

		testBody, _ := json.Marshal(imagePut)

		request, _ := http.NewRequest("PUT", "/images", bytes.NewReader(testBody))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		mockDatabase.EXPECT().ReadImage(gomock.Any()).Return(&imagePut, nil)
		mockDatabase.EXPECT().ListClusters().Return([]protobuf.Cluster{}, nil)
		mockDatabase.EXPECT().ListProjects().Return([]protobuf.Project{}, nil)
		mockDatabase.EXPECT().ReadImage(gomock.Any()).Return(&imagePut, nil)
		mockDatabase.EXPECT().UpdateImage(gomock.Any(), &imagePut).Return(nil)

		hS.ImageUpdate(response, request, httprouter.Params{})

		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %v, but received: %v", "200", response.Code)
		}
	})

	t.Run("Valid JSON, image has no clusters and no projects, new image isn't ok", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		mockClient := mocks.NewMockGrpcClient(mockCtrl)
		mockDatabase := mocks.NewMockDatabase(mockCtrl)

		RespHandler := handlers.HttpResponseHandler{}
		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, RespHandler: RespHandler}

		testBody, _ := json.Marshal(imagePut)

		request, _ := http.NewRequest("PUT", "/images", bytes.NewReader(testBody))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		mockDatabase.EXPECT().ReadImage(gomock.Any()).Return(&imagePut, nil)
		mockDatabase.EXPECT().ListClusters().Return([]protobuf.Cluster{}, nil)
		mockDatabase.EXPECT().ListProjects().Return([]protobuf.Project{}, nil)
		mockDatabase.EXPECT().ReadImage(gomock.Any()).Return(&image, nil)

		hS.ImageUpdate(response, request, httprouter.Params{})

		if response.Code != http.StatusBadRequest {
			t.Fatalf("Expected status code %v, but received: %v", "400", response.Code)
		}
	})

	t.Run("Valid JSON, image has clusters or projects", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		mockClient := mocks.NewMockGrpcClient(mockCtrl)
		mockDatabase := mocks.NewMockDatabase(mockCtrl)
		imageName := "testImageName"

		RespHandler := handlers.HttpResponseHandler{}
		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, RespHandler: RespHandler}

		testBody, _ := json.Marshal(imagePut)

		request, _ := http.NewRequest("PUT", "/images", bytes.NewReader(testBody))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		var existedCluster = []protobuf.Cluster{protobuf.Cluster{Name: "Name2",
			ID: "some_ID_123", ProjectID: "test-TEST-UUID-123", Image: imageName}}

		var existedProject = []protobuf.Project{protobuf.Project{Name: "Name1",
			ID: "some_ID_124", DefaultImage: imageName}}

		mockDatabase.EXPECT().ReadImage(gomock.Any()).Return(&imagePut, nil)
		mockDatabase.EXPECT().ListClusters().Return(existedCluster, nil)
		mockDatabase.EXPECT().ListProjects().Return(existedProject, nil)

		hS.ImageUpdate(response, request, httprouter.Params{})

		if response.Code != http.StatusBadRequest {
			t.Fatalf("Expected status code %v, but received: %v", "400", response.Code)
		}

	})

	t.Run("Invalid JSON", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		mockClient := mocks.NewMockGrpcClient(mockCtrl)
		mockDatabase := mocks.NewMockDatabase(mockCtrl)

		RespHandler := handlers.HttpResponseHandler{}
		hS := handlers.HttpServer{Gc: mockClient, Logger: l, Db: mockDatabase, RespHandler: RespHandler}

		testBody := []byte(`this is invalid json`)

		request, _ := http.NewRequest("PUT", "/images", bytes.NewBuffer(testBody))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		mockDatabase.EXPECT().ReadImage(gomock.Any()).Return(&imagePut, nil)

		hS.ImageUpdate(response, request, httprouter.Params{})

		if response.Code != http.StatusBadRequest {
			t.Fatalf("Expected status code %v, but received: %v", "400", response.Code)
		}
	})

}
