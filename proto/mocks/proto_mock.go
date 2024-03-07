// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/mikekulinski/zookeeper/proto (interfaces: Zookeeper_MessageClient,ZookeeperClient)
//
// Generated by this command:
//
//	mockgen -destination mocks/proto_mock.go github.com/mikekulinski/zookeeper/proto Zookeeper_MessageClient,ZookeeperClient
//

// Package mock_proto is a generated GoMock package.
package mock_proto

import (
	context "context"
	reflect "reflect"

	zookeeper "github.com/mikekulinski/zookeeper/proto"
	gomock "go.uber.org/mock/gomock"
	grpc "google.golang.org/grpc"
	metadata "google.golang.org/grpc/metadata"
)

// MockZookeeper_MessageClient is a mock of Zookeeper_MessageClient interface.
type MockZookeeper_MessageClient struct {
	ctrl     *gomock.Controller
	recorder *MockZookeeper_MessageClientMockRecorder
}

// MockZookeeper_MessageClientMockRecorder is the mock recorder for MockZookeeper_MessageClient.
type MockZookeeper_MessageClientMockRecorder struct {
	mock *MockZookeeper_MessageClient
}

// NewMockZookeeper_MessageClient creates a new mock instance.
func NewMockZookeeper_MessageClient(ctrl *gomock.Controller) *MockZookeeper_MessageClient {
	mock := &MockZookeeper_MessageClient{ctrl: ctrl}
	mock.recorder = &MockZookeeper_MessageClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockZookeeper_MessageClient) EXPECT() *MockZookeeper_MessageClientMockRecorder {
	return m.recorder
}

// CloseSend mocks base method.
func (m *MockZookeeper_MessageClient) CloseSend() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CloseSend")
	ret0, _ := ret[0].(error)
	return ret0
}

// CloseSend indicates an expected call of CloseSend.
func (mr *MockZookeeper_MessageClientMockRecorder) CloseSend() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CloseSend", reflect.TypeOf((*MockZookeeper_MessageClient)(nil).CloseSend))
}

// Context mocks base method.
func (m *MockZookeeper_MessageClient) Context() context.Context {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Context")
	ret0, _ := ret[0].(context.Context)
	return ret0
}

// Context indicates an expected call of Context.
func (mr *MockZookeeper_MessageClientMockRecorder) Context() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Context", reflect.TypeOf((*MockZookeeper_MessageClient)(nil).Context))
}

// Header mocks base method.
func (m *MockZookeeper_MessageClient) Header() (metadata.MD, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Header")
	ret0, _ := ret[0].(metadata.MD)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Header indicates an expected call of Header.
func (mr *MockZookeeper_MessageClientMockRecorder) Header() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Header", reflect.TypeOf((*MockZookeeper_MessageClient)(nil).Header))
}

// Recv mocks base method.
func (m *MockZookeeper_MessageClient) Recv() (*zookeeper.ZookeeperResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Recv")
	ret0, _ := ret[0].(*zookeeper.ZookeeperResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Recv indicates an expected call of Recv.
func (mr *MockZookeeper_MessageClientMockRecorder) Recv() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Recv", reflect.TypeOf((*MockZookeeper_MessageClient)(nil).Recv))
}

// RecvMsg mocks base method.
func (m *MockZookeeper_MessageClient) RecvMsg(arg0 any) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RecvMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecvMsg indicates an expected call of RecvMsg.
func (mr *MockZookeeper_MessageClientMockRecorder) RecvMsg(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecvMsg", reflect.TypeOf((*MockZookeeper_MessageClient)(nil).RecvMsg), arg0)
}

// Send mocks base method.
func (m *MockZookeeper_MessageClient) Send(arg0 *zookeeper.ZookeeperRequest) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send.
func (mr *MockZookeeper_MessageClientMockRecorder) Send(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockZookeeper_MessageClient)(nil).Send), arg0)
}

// SendMsg mocks base method.
func (m *MockZookeeper_MessageClient) SendMsg(arg0 any) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendMsg indicates an expected call of SendMsg.
func (mr *MockZookeeper_MessageClientMockRecorder) SendMsg(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMsg", reflect.TypeOf((*MockZookeeper_MessageClient)(nil).SendMsg), arg0)
}

// Trailer mocks base method.
func (m *MockZookeeper_MessageClient) Trailer() metadata.MD {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Trailer")
	ret0, _ := ret[0].(metadata.MD)
	return ret0
}

// Trailer indicates an expected call of Trailer.
func (mr *MockZookeeper_MessageClientMockRecorder) Trailer() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Trailer", reflect.TypeOf((*MockZookeeper_MessageClient)(nil).Trailer))
}

// MockZookeeperClient is a mock of ZookeeperClient interface.
type MockZookeeperClient struct {
	ctrl     *gomock.Controller
	recorder *MockZookeeperClientMockRecorder
}

// MockZookeeperClientMockRecorder is the mock recorder for MockZookeeperClient.
type MockZookeeperClientMockRecorder struct {
	mock *MockZookeeperClient
}

// NewMockZookeeperClient creates a new mock instance.
func NewMockZookeeperClient(ctrl *gomock.Controller) *MockZookeeperClient {
	mock := &MockZookeeperClient{ctrl: ctrl}
	mock.recorder = &MockZookeeperClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockZookeeperClient) EXPECT() *MockZookeeperClientMockRecorder {
	return m.recorder
}

// Message mocks base method.
func (m *MockZookeeperClient) Message(arg0 context.Context, arg1 ...grpc.CallOption) (zookeeper.Zookeeper_MessageClient, error) {
	m.ctrl.T.Helper()
	varargs := []any{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Message", varargs...)
	ret0, _ := ret[0].(zookeeper.Zookeeper_MessageClient)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Message indicates an expected call of Message.
func (mr *MockZookeeperClientMockRecorder) Message(arg0 any, arg1 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Message", reflect.TypeOf((*MockZookeeperClient)(nil).Message), varargs...)
}
