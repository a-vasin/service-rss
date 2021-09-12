// Code generated by MockGen. DO NOT EDIT.
// Source: database.go

// Package database is a generated GoMock package.
package database

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockDatabase is a mock of Database interface.
type MockDatabase struct {
	ctrl     *gomock.Controller
	recorder *MockDatabaseMockRecorder
}

// MockDatabaseMockRecorder is the mock recorder for MockDatabase.
type MockDatabaseMockRecorder struct {
	mock *MockDatabase
}

// NewMockDatabase creates a new mock instance.
func NewMockDatabase(ctrl *gomock.Controller) *MockDatabase {
	mock := &MockDatabase{ctrl: ctrl}
	mock.recorder = &MockDatabaseMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDatabase) EXPECT() *MockDatabaseMockRecorder {
	return m.recorder
}

// CreateRss mocks base method.
func (m *MockDatabase) CreateRss(name string, sources []string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateRss", name, sources)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateRss indicates an expected call of CreateRss.
func (mr *MockDatabaseMockRecorder) CreateRss(name, sources interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateRss", reflect.TypeOf((*MockDatabase)(nil).CreateRss), name, sources)
}

// Shutdown mocks base method.
func (m *MockDatabase) Shutdown() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Shutdown")
	ret0, _ := ret[0].(error)
	return ret0
}

// Shutdown indicates an expected call of Shutdown.
func (mr *MockDatabaseMockRecorder) Shutdown() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Shutdown", reflect.TypeOf((*MockDatabase)(nil).Shutdown))
}
