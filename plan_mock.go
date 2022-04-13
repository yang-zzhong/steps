// Code generated by MockGen. DO NOT EDIT.
// Source: phasemachine.go

// Package pm is a generated GoMock package.
package steps

import (
    reflect "reflect"

    gomock "github.com/golang/mock/gomock"
)

// MockAdapter is a mock of Adapter interface.
type MockAdapter struct {
    ctrl     *gomock.Controller
    recorder *MockAdapterMockRecorder
}

// MockAdapterMockRecorder is the mock recorder for MockAdapter.
type MockAdapterMockRecorder struct {
    mock *MockAdapter
}

// NewMockAdapter creates a new mock instance.
func NewMockAdapter(ctrl *gomock.Controller) *MockAdapter {
    mock := &MockAdapter{ctrl: ctrl}
    mock.recorder = &MockAdapterMockRecorder{mock}
    return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAdapter) EXPECT() *MockAdapterMockRecorder {
    return m.recorder
}

// SaveState mocks base method.
func (m *MockAdapter) SaveState(state *State) error {
    m.ctrl.T.Helper()
    ret := m.ctrl.Call(m, "SaveState", state)
    ret0, _ := ret[0].(error)
    return ret0
}

// SaveState indicates an expected call of SaveState.
func (mr *MockAdapterMockRecorder) SaveState(state interface{}) *gomock.Call {
    mr.mock.ctrl.T.Helper()
    return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveState", reflect.TypeOf((*MockAdapter)(nil).SaveState), state)
}

// State mocks base method.
func (m *MockAdapter) State(state *State) error {
    m.ctrl.T.Helper()
    ret := m.ctrl.Call(m, "State", state)
    ret0, _ := ret[0].(error)
    return ret0
}

// State indicates an expected call of State.
func (mr *MockAdapterMockRecorder) State(state interface{}) *gomock.Call {
    mr.mock.ctrl.T.Helper()
    return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "State", reflect.TypeOf((*MockAdapter)(nil).State), state)
}

// MockTask is a mock of Task interface.
type MockTask struct {
    ctrl     *gomock.Controller
    recorder *MockTaskMockRecorder
}

// MockTaskMockRecorder is the mock recorder for MockTask.
type MockTaskMockRecorder struct {
    mock *MockTask
}

// NewMockTask creates a new mock instance.
func NewMockTask(ctrl *gomock.Controller) *MockTask {
    mock := &MockTask{ctrl: ctrl}
    mock.recorder = &MockTaskMockRecorder{mock}
    return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTask) EXPECT() *MockTaskMockRecorder {
    return m.recorder
}

// Name mocks base method.
func (m *MockTask) Name() string {
    m.ctrl.T.Helper()
    ret := m.ctrl.Call(m, "Name")
    ret0, _ := ret[0].(string)
    return ret0
}

// Name indicates an expected call of Name.
func (mr *MockTaskMockRecorder) Name() *gomock.Call {
    mr.mock.ctrl.T.Helper()
    return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Name", reflect.TypeOf((*MockTask)(nil).Name))
}

// Tasks mocks base method.
func (m *MockTask) Tasks() []Task {
    m.ctrl.T.Helper()
    ret := m.ctrl.Call(m, "Tasks")
    ret0, _ := ret[0].([]Task)
    return ret0
}

// Tasks indicates an expected call of Tasks.
func (mr *MockTaskMockRecorder) Tasks() *gomock.Call {
    mr.mock.ctrl.T.Helper()
    return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Tasks", reflect.TypeOf((*MockTask)(nil).Tasks))
}

