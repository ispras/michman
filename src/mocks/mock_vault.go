// Code generated by MockGen. DO NOT EDIT.

// Package mocks is a generated GoMock package.
package mocks

import (
	gomock "github.com/golang/mock/gomock"
	api "github.com/hashicorp/vault/api"
	utils "github.com/ispras/michman/src/utils"
	reflect "reflect"
)

// MockSecretStorage is a mock of SecretStorage interface
type MockSecretStorage struct {
	ctrl     *gomock.Controller
	recorder *MockSecretStorageMockRecorder
}

// MockSecretStorageMockRecorder is the mock recorder for MockSecretStorage
type MockSecretStorageMockRecorder struct {
	mock *MockSecretStorage
}

// NewMockSecretStorage creates a new mock instance
func NewMockSecretStorage(ctrl *gomock.Controller) *MockSecretStorage {
	mock := &MockSecretStorage{ctrl: ctrl}
	mock.recorder = &MockSecretStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockSecretStorage) EXPECT() *MockSecretStorageMockRecorder {
	return m.recorder
}

// ConnectVault mocks base method
func (m *MockSecretStorage) ConnectVault() (*api.Client, *utils.Config) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ConnectVault")
	ret0, _ := ret[0].(*api.Client)
	ret1, _ := ret[1].(*utils.Config)
	return ret0, ret1
}

// ConnectVault indicates an expected call of ConnectVault
func (mr *MockSecretStorageMockRecorder) ConnectVault() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ConnectVault", reflect.TypeOf((*MockSecretStorage)(nil).ConnectVault))
}
