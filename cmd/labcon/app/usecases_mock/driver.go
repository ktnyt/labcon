// Code generated by MockGen. DO NOT EDIT.
// Source: cmd/labcon/app/usecases/driver_iface.go

// Package usecases_mock is a generated GoMock package.
package usecases_mock

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	models "github.com/ktnyt/labcon/cmd/labcon/app/models"
)

// MockDriverUsecase is a mock of DriverUsecase interface.
type MockDriverUsecase struct {
	ctrl     *gomock.Controller
	recorder *MockDriverUsecaseMockRecorder
}

// MockDriverUsecaseMockRecorder is the mock recorder for MockDriverUsecase.
type MockDriverUsecaseMockRecorder struct {
	mock *MockDriverUsecase
}

// NewMockDriverUsecase creates a new mock instance.
func NewMockDriverUsecase(ctrl *gomock.Controller) *MockDriverUsecase {
	mock := &MockDriverUsecase{ctrl: ctrl}
	mock.recorder = &MockDriverUsecaseMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDriverUsecase) EXPECT() *MockDriverUsecaseMockRecorder {
	return m.recorder
}

// Authorize mocks base method.
func (m *MockDriverUsecase) Authorize(name, token string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Authorize", name, token)
	ret0, _ := ret[0].(error)
	return ret0
}

// Authorize indicates an expected call of Authorize.
func (mr *MockDriverUsecaseMockRecorder) Authorize(name, token interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Authorize", reflect.TypeOf((*MockDriverUsecase)(nil).Authorize), name, token)
}

// GetOp mocks base method.
func (m *MockDriverUsecase) GetOp(name string) (*models.DriverOp, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOp", name)
	ret0, _ := ret[0].(*models.DriverOp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOp indicates an expected call of GetOp.
func (mr *MockDriverUsecaseMockRecorder) GetOp(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOp", reflect.TypeOf((*MockDriverUsecase)(nil).GetOp), name)
}

// GetState mocks base method.
func (m *MockDriverUsecase) GetState(name string) (interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetState", name)
	ret0, _ := ret[0].(interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetState indicates an expected call of GetState.
func (mr *MockDriverUsecaseMockRecorder) GetState(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetState", reflect.TypeOf((*MockDriverUsecase)(nil).GetState), name)
}

// GetStatus mocks base method.
func (m *MockDriverUsecase) GetStatus(name string) (models.DriverStatus, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStatus", name)
	ret0, _ := ret[0].(models.DriverStatus)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetStatus indicates an expected call of GetStatus.
func (mr *MockDriverUsecaseMockRecorder) GetStatus(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStatus", reflect.TypeOf((*MockDriverUsecase)(nil).GetStatus), name)
}

// Register mocks base method.
func (m *MockDriverUsecase) Register(name string, state interface{}) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Register", name, state)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Register indicates an expected call of Register.
func (mr *MockDriverUsecaseMockRecorder) Register(name, state interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Register", reflect.TypeOf((*MockDriverUsecase)(nil).Register), name, state)
}

// SetOp mocks base method.
func (m *MockDriverUsecase) SetOp(name string, op models.DriverOp) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetOp", name, op)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetOp indicates an expected call of SetOp.
func (mr *MockDriverUsecaseMockRecorder) SetOp(name, op interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetOp", reflect.TypeOf((*MockDriverUsecase)(nil).SetOp), name, op)
}

// SetState mocks base method.
func (m *MockDriverUsecase) SetState(name string, state interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetState", name, state)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetState indicates an expected call of SetState.
func (mr *MockDriverUsecaseMockRecorder) SetState(name, state interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetState", reflect.TypeOf((*MockDriverUsecase)(nil).SetState), name, state)
}

// SetStatus mocks base method.
func (m *MockDriverUsecase) SetStatus(name string, status models.DriverStatus) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetStatus", name, status)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetStatus indicates an expected call of SetStatus.
func (mr *MockDriverUsecaseMockRecorder) SetStatus(name, status interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetStatus", reflect.TypeOf((*MockDriverUsecase)(nil).SetStatus), name, status)
}
