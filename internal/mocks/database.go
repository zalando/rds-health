// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/zalando/rds-health/internal/database (interfaces: Provider)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	rds "github.com/aws/aws-sdk-go-v2/service/rds"
	gomock "go.uber.org/mock/gomock"
)

// Database is a mock of Provider interface.
type Database struct {
	ctrl     *gomock.Controller
	recorder *DatabaseMockRecorder
}

// DatabaseMockRecorder is the mock recorder for Database.
type DatabaseMockRecorder struct {
	mock *Database
}

// NewDatabase creates a new mock instance.
func NewDatabase(ctrl *gomock.Controller) *Database {
	mock := &Database{ctrl: ctrl}
	mock.recorder = &DatabaseMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *Database) EXPECT() *DatabaseMockRecorder {
	return m.recorder
}

// DescribeDBInstances mocks base method.
func (m *Database) DescribeDBInstances(arg0 context.Context, arg1 *rds.DescribeDBInstancesInput, arg2 ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DescribeDBInstances", varargs...)
	ret0, _ := ret[0].(*rds.DescribeDBInstancesOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DescribeDBInstances indicates an expected call of DescribeDBInstances.
func (mr *DatabaseMockRecorder) DescribeDBInstances(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DescribeDBInstances", reflect.TypeOf((*Database)(nil).DescribeDBInstances), varargs...)
}
