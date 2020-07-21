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