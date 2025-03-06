// Code generated by MockGen. DO NOT EDIT.
// Source: infra.go
//
// Generated by this command:
//
//	mockgen -source=infra.go -package=app -destination=./mock_infra.go
//

// Package app is a generated GoMock package.
package app

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockItemRepository is a mock of ItemRepository interface.
type MockItemRepository struct {
	ctrl     *gomock.Controller
	recorder *MockItemRepositoryMockRecorder
	isgomock struct{}
}

// MockItemRepositoryMockRecorder is the mock recorder for MockItemRepository.
type MockItemRepositoryMockRecorder struct {
	mock *MockItemRepository
}

// NewMockItemRepository creates a new mock instance.
func NewMockItemRepository(ctrl *gomock.Controller) *MockItemRepository {
	mock := &MockItemRepository{ctrl: ctrl}
	mock.recorder = &MockItemRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockItemRepository) EXPECT() *MockItemRepositoryMockRecorder {
	return m.recorder
}

// GetItemID mocks base method.
func (m *MockItemRepository) GetItemID(ctx context.Context, itemID int) (Item, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetItemID", ctx, itemID)
	ret0, _ := ret[0].(Item)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetItemID indicates an expected call of GetItemID.
func (mr *MockItemRepositoryMockRecorder) GetItemID(ctx, itemID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetItemID", reflect.TypeOf((*MockItemRepository)(nil).GetItemID), ctx, itemID)
}

// GetItems mocks base method.
func (m *MockItemRepository) GetItems(ctx context.Context) ([]Item, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetItems", ctx)
	ret0, _ := ret[0].([]Item)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetItems indicates an expected call of GetItems.
func (mr *MockItemRepositoryMockRecorder) GetItems(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetItems", reflect.TypeOf((*MockItemRepository)(nil).GetItems), ctx)
}

// Insert mocks base method.
func (m *MockItemRepository) Insert(ctx context.Context, item *Item) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Insert", ctx, item)
	ret0, _ := ret[0].(error)
	return ret0
}

// Insert indicates an expected call of Insert.
func (mr *MockItemRepositoryMockRecorder) Insert(ctx, item any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Insert", reflect.TypeOf((*MockItemRepository)(nil).Insert), ctx, item)
}
