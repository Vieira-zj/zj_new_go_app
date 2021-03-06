// Code generated by MockGen. DO NOT EDIT.
// Source: demo.hello/tests/unit.test (interfaces: IFoo)

// Package unittest is a generated GoMock package.
package unittest

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockIFoo is a mock of IFoo interface
type MockIFoo struct {
	ctrl     *gomock.Controller
	recorder *MockIFooMockRecorder
}

// MockIFooMockRecorder is the mock recorder for MockIFoo
type MockIFooMockRecorder struct {
	mock *MockIFoo
}

// NewMockIFoo creates a new mock instance
func NewMockIFoo(ctrl *gomock.Controller) *MockIFoo {
	mock := &MockIFoo{ctrl: ctrl}
	mock.recorder = &MockIFooMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockIFoo) EXPECT() *MockIFooMockRecorder {
	return m.recorder
}

// Foo mocks base method
func (m *MockIFoo) Foo(arg0 context.Context, arg1 int) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Foo", arg0, arg1)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Foo indicates an expected call of Foo
func (mr *MockIFooMockRecorder) Foo(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Foo", reflect.TypeOf((*MockIFoo)(nil).Foo), arg0, arg1)
}
