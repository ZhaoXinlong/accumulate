// Code generated by MockGen. DO NOT EDIT.
// Source: abci.go

// Package mock_abci is a generated GoMock package.
package mock_abci

import (
	"github.com/AccumulateNetwork/accumulated/types"
	api "github.com/AccumulateNetwork/accumulated/types/api/query"
	"github.com/AccumulateNetwork/accumulated/types/state"
	reflect "reflect"

	abci "github.com/AccumulateNetwork/accumulated/internal/abci"
	protocol "github.com/AccumulateNetwork/accumulated/protocol"
	transactions "github.com/AccumulateNetwork/accumulated/types/api/transactions"
	gomock "github.com/golang/mock/gomock"
)

// MockChain is a mock of Chain interface.
type MockChain struct {
	ctrl     *gomock.Controller
	recorder *MockChainMockRecorder
}

// MockChainMockRecorder is the mock recorder for MockChain.
type MockChainMockRecorder struct {
	mock *MockChain
}

// NewMockChain creates a new mock instance.
func NewMockChain(ctrl *gomock.Controller) *MockChain {
	mock := &MockChain{ctrl: ctrl}
	mock.recorder = &MockChainMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockChain) EXPECT() *MockChainMockRecorder {
	return m.recorder
}

// BeginBlock mocks base method.
func (m *MockChain) BeginBlock(arg0 abci.BeginBlockRequest) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "BeginBlock", arg0)
}

// BeginBlock indicates an expected call of BeginBlock.
func (mr *MockChainMockRecorder) BeginBlock(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BeginBlock", reflect.TypeOf((*MockChain)(nil).BeginBlock), arg0)
}

// CheckTx mocks base method.
func (m *MockChain) CheckTx(arg0 *transactions.GenTransaction) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckTx", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// CheckTx indicates an expected call of CheckTx.
func (mr *MockChainMockRecorder) CheckTx(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckTx", reflect.TypeOf((*MockChain)(nil).CheckTx), arg0)
}

// Commit mocks base method.
func (m *MockChain) Commit() ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Commit")
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Commit indicates an expected call of Commit.
func (mr *MockChainMockRecorder) Commit() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Commit", reflect.TypeOf((*MockChain)(nil).Commit))
}

// DeliverTx mocks base method.
func (m *MockChain) DeliverTx(arg0 *transactions.GenTransaction) (*protocol.TxResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeliverTx", arg0)
	ret0, _ := ret[0].(*protocol.TxResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeliverTx indicates an expected call of DeliverTx.
func (mr *MockChainMockRecorder) DeliverTx(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeliverTx", reflect.TypeOf((*MockChain)(nil).DeliverTx), arg0)
}

// EndBlock mocks base method.
func (m *MockChain) EndBlock(arg0 abci.EndBlockRequest) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "EndBlock", arg0)
}

// EndBlock indicates an expected call of EndBlock.
func (mr *MockChainMockRecorder) EndBlock(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EndBlock", reflect.TypeOf((*MockChain)(nil).EndBlock), arg0)
}

// Query mocks base method.
func (m *MockChain) Query(arg0 *api.Query) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Query", arg0)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Query indicates an expected call of Query.
func (mr *MockChainMockRecorder) Query(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Query", reflect.TypeOf((*MockChain)(nil).Query), arg0)
}

// MockState is a mock of State interface.
type MockState struct {
	ctrl     *gomock.Controller
	recorder *MockStateMockRecorder
}

// MockStateMockRecorder is the mock recorder for MockState.
type MockStateMockRecorder struct {
	mock *MockState
}

// NewMockState creates a new mock instance.
func NewMockState(ctrl *gomock.Controller) *MockState {
	mock := &MockState{ctrl: ctrl}
	mock.recorder = &MockStateMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockState) EXPECT() *MockStateMockRecorder {
	return m.recorder
}

// BlockIndex mocks base method.
func (m *MockState) BlockIndex() int64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BlockIndex")
	ret0, _ := ret[0].(int64)
	return ret0
}

// BlockIndex indicates an expected call of BlockIndex.
func (mr *MockStateMockRecorder) BlockIndex() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BlockIndex", reflect.TypeOf((*MockState)(nil).BlockIndex))
}

// AddStateEntry used for genesis
func (m *MockState) AddStateEntry(chainId *types.Bytes32, txHash *types.Bytes32, object *state.Object) {
	_ = chainId
	_ = txHash
	_ = object
}

// EnsureRootHash mocks base method.
func (m *MockState) EnsureRootHash() []byte {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnsureRootHash")
	ret0, _ := ret[0].([]byte)
	return ret0
}

// EnsureRootHash indicates an expected call of EnsureRootHash.
func (mr *MockStateMockRecorder) EnsureRootHash() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnsureRootHash", reflect.TypeOf((*MockState)(nil).EnsureRootHash))
}

// RootHash mocks base method.
func (m *MockState) RootHash() []byte {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RootHash")
	ret0, _ := ret[0].([]byte)
	return ret0
}

// RootHash indicates an expected call of RootHash.
func (mr *MockStateMockRecorder) RootHash() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RootHash", reflect.TypeOf((*MockState)(nil).RootHash))
}
