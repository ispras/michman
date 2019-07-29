package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	protobuf "gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/protobuf"
)

type mockGrpcClient struct{}

func (mockCl mockGrpcClient) GetID(c *protobuf.Cluster) (int32, error) {
	log.Print("MOCK gRPC call for getting ID")

	return 0, nil
}

func (mockCl mockGrpcClient) StartClusterCreation(c *protobuf.Cluster) {
	log.Print("MOCK gRPC call for cluster creating")
}

func TestClustersGet(t *testing.T) {
	request, _ := http.NewRequest("GET", "/", nil)
	response := httptest.NewRecorder()

	hS := httpServer{}

	hS.clustersHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("Expected status code %v, but received: %v", "200", response.Code)
	}
}

func TestClustersPost(t *testing.T) {
	l := log.New(os.Stdout, "TEST_LOGGER: ", log.Ldate|log.Ltime)

	t.Run("Valid JSON", func(t *testing.T) {
		testBody := []byte(`{"Name": "test-cluster"}`)
		request, _ := http.NewRequest("POST", "/", bytes.NewBuffer(testBody))
		request.Header.Set("Content-Type", "application/json")

		response := httptest.NewRecorder()

		mockClient := mockGrpcClient{}
		hS := httpServer{mockClient, l}

		hS.clustersHandler(response, request)

		if response.Code != http.StatusOK {
			t.Fatalf("Expected status code %v, but received: %v", "200", response.Code)
		}

		var c protobuf.Cluster
		err := json.NewDecoder(response.Body).Decode(&c)
		if err != nil {
			t.Fatalf("Get invalid JSON")
		}

		if c.EntityStatus != "INITED" {
			t.Fatalf("Expected EntityStatus %s, but received: %s", "INITED", c.EntityStatus)
		}
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		testBody := []byte(`this is invalid json`)
		request, _ := http.NewRequest("POST", "/", bytes.NewBuffer(testBody))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		mockClient := mockGrpcClient{}
		hS := httpServer{mockClient, l}

		hS.clustersHandler(response, request)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("Expected status code %v, but received: %v", "400", response.Code)
		}
	})
}
