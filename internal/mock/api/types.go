// Code generated by MockGen. DO NOT EDIT.
// Source: types.go

// Package mock_api is a generated GoMock package.
package mock_api

import (
	context "context"
	reflect "reflect"

	api "github.com/AccumulateNetwork/accumulate/internal/api/v2"
	gomock "github.com/golang/mock/gomock"
	bytes "github.com/tendermint/tendermint/libs/bytes"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	types "github.com/tendermint/tendermint/types"
)

// MockQuerier is a mock of Querier interface.
type MockQuerier struct {
	ctrl     *gomock.Controller
	recorder *MockQuerierMockRecorder
}

// MockQuerierMockRecorder is the mock recorder for MockQuerier.
type MockQuerierMockRecorder struct {
	mock *MockQuerier
}

// NewMockQuerier creates a new mock instance.
func NewMockQuerier(ctrl *gomock.Controller) *MockQuerier {
	mock := &MockQuerier{ctrl: ctrl}
	mock.recorder = &MockQuerierMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockQuerier) EXPECT() *MockQuerierMockRecorder {
	return m.recorder
}

// QueryChain mocks base method.
func (m *MockQuerier) QueryChain(id []byte) (*api.QueryResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryChain", id)
	ret0, _ := ret[0].(*api.QueryResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryChain indicates an expected call of QueryChain.
func (mr *MockQuerierMockRecorder) QueryChain(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryChain", reflect.TypeOf((*MockQuerier)(nil).QueryChain), id)
}

// QueryDirectory mocks base method.
func (m *MockQuerier) QueryDirectory(url string) (*api.QueryResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryDirectory", url)
	ret0, _ := ret[0].(*api.QueryResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryDirectory indicates an expected call of QueryDirectory.
func (mr *MockQuerierMockRecorder) QueryDirectory(url interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryDirectory", reflect.TypeOf((*MockQuerier)(nil).QueryDirectory), url)
}

// QueryTx mocks base method.
func (m *MockQuerier) QueryTx(id []byte) (*api.QueryResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryTx", id)
	ret0, _ := ret[0].(*api.QueryResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryTx indicates an expected call of QueryTx.
func (mr *MockQuerierMockRecorder) QueryTx(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryTx", reflect.TypeOf((*MockQuerier)(nil).QueryTx), id)
}

// QueryTxHistory mocks base method.
func (m *MockQuerier) QueryTxHistory(url string, start, count int64) (*api.QueryMultiResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryTxHistory", url, start, count)
	ret0, _ := ret[0].(*api.QueryMultiResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryTxHistory indicates an expected call of QueryTxHistory.
func (mr *MockQuerierMockRecorder) QueryTxHistory(url, start, count interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryTxHistory", reflect.TypeOf((*MockQuerier)(nil).QueryTxHistory), url, start, count)
}

// QueryUrl mocks base method.
func (m *MockQuerier) QueryUrl(url string) (*api.QueryResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryUrl", url)
	ret0, _ := ret[0].(*api.QueryResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryUrl indicates an expected call of QueryUrl.
func (mr *MockQuerierMockRecorder) QueryUrl(url interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryUrl", reflect.TypeOf((*MockQuerier)(nil).QueryUrl), url)
}

// MockABCIQueryClient is a mock of ABCIQueryClient interface.
type MockABCIQueryClient struct {
	ctrl     *gomock.Controller
	recorder *MockABCIQueryClientMockRecorder
}

// MockABCIQueryClientMockRecorder is the mock recorder for MockABCIQueryClient.
type MockABCIQueryClientMockRecorder struct {
	mock *MockABCIQueryClient
}

// NewMockABCIQueryClient creates a new mock instance.
func NewMockABCIQueryClient(ctrl *gomock.Controller) *MockABCIQueryClient {
	mock := &MockABCIQueryClient{ctrl: ctrl}
	mock.recorder = &MockABCIQueryClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockABCIQueryClient) EXPECT() *MockABCIQueryClientMockRecorder {
	return m.recorder
}

// ABCIQuery mocks base method.
func (m *MockABCIQueryClient) ABCIQuery(ctx context.Context, path string, data bytes.HexBytes) (*coretypes.ResultABCIQuery, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ABCIQuery", ctx, path, data)
	ret0, _ := ret[0].(*coretypes.ResultABCIQuery)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ABCIQuery indicates an expected call of ABCIQuery.
func (mr *MockABCIQueryClientMockRecorder) ABCIQuery(ctx, path, data interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ABCIQuery", reflect.TypeOf((*MockABCIQueryClient)(nil).ABCIQuery), ctx, path, data)
}

// MockABCIBroadcastClient is a mock of ABCIBroadcastClient interface.
type MockABCIBroadcastClient struct {
	ctrl     *gomock.Controller
	recorder *MockABCIBroadcastClientMockRecorder
}

// MockABCIBroadcastClientMockRecorder is the mock recorder for MockABCIBroadcastClient.
type MockABCIBroadcastClientMockRecorder struct {
	mock *MockABCIBroadcastClient
}

// NewMockABCIBroadcastClient creates a new mock instance.
func NewMockABCIBroadcastClient(ctrl *gomock.Controller) *MockABCIBroadcastClient {
	mock := &MockABCIBroadcastClient{ctrl: ctrl}
	mock.recorder = &MockABCIBroadcastClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockABCIBroadcastClient) EXPECT() *MockABCIBroadcastClientMockRecorder {
	return m.recorder
}

// BroadcastTxAsync mocks base method.
func (m *MockABCIBroadcastClient) BroadcastTxAsync(arg0 context.Context, arg1 types.Tx) (*coretypes.ResultBroadcastTx, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BroadcastTxAsync", arg0, arg1)
	ret0, _ := ret[0].(*coretypes.ResultBroadcastTx)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BroadcastTxAsync indicates an expected call of BroadcastTxAsync.
func (mr *MockABCIBroadcastClientMockRecorder) BroadcastTxAsync(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BroadcastTxAsync", reflect.TypeOf((*MockABCIBroadcastClient)(nil).BroadcastTxAsync), arg0, arg1)
}

// BroadcastTxSync mocks base method.
func (m *MockABCIBroadcastClient) BroadcastTxSync(arg0 context.Context, arg1 types.Tx) (*coretypes.ResultBroadcastTx, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BroadcastTxSync", arg0, arg1)
	ret0, _ := ret[0].(*coretypes.ResultBroadcastTx)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BroadcastTxSync indicates an expected call of BroadcastTxSync.
func (mr *MockABCIBroadcastClientMockRecorder) BroadcastTxSync(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BroadcastTxSync", reflect.TypeOf((*MockABCIBroadcastClient)(nil).BroadcastTxSync), arg0, arg1)
}

// CheckTx mocks base method.
func (m *MockABCIBroadcastClient) CheckTx(ctx context.Context, tx types.Tx) (*coretypes.ResultCheckTx, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckTx", ctx, tx)
	ret0, _ := ret[0].(*coretypes.ResultCheckTx)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CheckTx indicates an expected call of CheckTx.
func (mr *MockABCIBroadcastClientMockRecorder) CheckTx(ctx, tx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckTx", reflect.TypeOf((*MockABCIBroadcastClient)(nil).CheckTx), ctx, tx)
}
