// Code generated by MockGen. DO NOT EDIT.
// Source: demo.apps/go.test/mock (interfaces: Foo)

// Package mocktest is a generated GoMock package.
package mocktest

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockFoo is a mock of Foo interface.
type MockFoo struct {
	ctrl     *gomock.Controller
	recorder *MockFooMockRecorder
}

// MockFooMockRecorder is the mock recorder for MockFoo.
type MockFooMockRecorder struct {
	mock *MockFoo
}

// NewMockFoo creates a new mock instance.
func NewMockFoo(ctrl *gomock.Controller) *MockFoo {
	mock := &MockFoo{ctrl: ctrl}
	mock.recorder = &MockFooMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockFoo) EXPECT() *MockFooMockRecorder {
	return m.recorder
}

// Bar mocks base method.
func (m *MockFoo) Bar(arg0 int) int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Bar", arg0)
	ret0, _ := ret[0].(int)
	return ret0
}

// Bar indicates an expected call of Bar.
func (mr *MockFooMockRecorder) Bar(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Bar", reflect.TypeOf((*MockFoo)(nil).Bar), arg0)
}
