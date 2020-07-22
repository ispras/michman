package handlers

import (
	//"bytes"
	//"encoding/json"
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
		request, _ := http.NewRequest("PUT", "/images/"+imageName, nil)
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
		request, _ := http.NewRequest("PUT", "/images/"+imageName, nil)
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