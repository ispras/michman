// Code generated by MockGen. DO NOT EDIT.
// Source: gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/handlers (interfaces: GrpcClient)

// Package mocks is a generated GoMock package.
package mocks

import (
	gomock "github.com/golang/mock/gomock"
	protobuf "gitlab.at.ispras.ru/openstack_bigdata_tools/spark-openstack/src/protobuf"
	reflect "reflect"
)

// MockGrpcClient is a mock of GrpcClient interface
type MockGrpcClient struct {
	ctrl     *gomock.Controller
	recorder *MockGrpcClientMockRecorder
}

// MockGrpcClientMockRecorder is the mock recorder for MockGrpcClient
type MockGrpcClientMockRecorder struct {
	mock *MockGrpcClient
}

// NewMockGrpcClient creates a new mock instance
func NewMockGrpcClient(ctrl *gomock.Controller) *MockGrpcClient {
	mock := &MockGrpcClient{ctrl: ctrl}
	mock.recorder = &MockGrpcClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockGrpcClient) EXPECT() *MockGrpcClientMockRecorder {
	return m.recorder
}

// GetID mocks base method
func (m *MockGrpcClient) GetID(arg0 *protobuf.Cluster) (int32, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetID", arg0)
	ret0, _ := ret[0].(int32)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetID indicates an expected call of GetID
func (mr *MockGrpcClientMockRecorder) GetID(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetID", reflect.TypeOf((*MockGrpcClient)(nil).GetID), arg0)
}

// StartClusterCreation mocks base method
func (m *MockGrpcClient) StartClusterCreation(arg0 *protobuf.Cluster) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "StartClusterCreation", arg0)
}

// StartClusterCreation indicates an expected call of StartClusterCreation
func (mr *MockGrpcClientMockRecorder) StartClusterCreation(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartClusterCreation", reflect.TypeOf((*MockGrpcClient)(nil).StartClusterCreation), arg0)
}

// StartClusterDestroying mocks base method
func (m *MockGrpcClient) StartClusterDestroying(arg0 *protobuf.Cluster) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "StartClusterDestroying", arg0)
}

// StartClusterDestroying indicates an expected call of StartClusterDestroying
func (mr *MockGrpcClientMockRecorder) StartClusterDestroying(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartClusterDestroying", reflect.TypeOf((*MockGrpcClient)(nil).StartClusterDestroying), arg0)
}

// StartClusterModification mocks base method
func (m *MockGrpcClient) StartClusterModification(arg0 *protobuf.Cluster) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "StartClusterModification", arg0)
}

// StartClusterModification indicates an expected call of StartClusterModification
func (mr *MockGrpcClientMockRecorder) StartClusterModification(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartClusterModification", reflect.TypeOf((*MockGrpcClient)(nil).StartClusterModification), arg0)
}