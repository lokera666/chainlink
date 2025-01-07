// Code generated by mockery v2.50.0. DO NOT EDIT.

package client

import (
	big "math/big"

	mock "github.com/stretchr/testify/mock"
)

// mockHead is an autogenerated mock type for the Head type
type mockHead struct {
	mock.Mock
}

type mockHead_Expecter struct {
	mock *mock.Mock
}

func (_m *mockHead) EXPECT() *mockHead_Expecter {
	return &mockHead_Expecter{mock: &_m.Mock}
}

// BlockDifficulty provides a mock function with no fields
func (_m *mockHead) BlockDifficulty() *big.Int {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for BlockDifficulty")
	}

	var r0 *big.Int
	if rf, ok := ret.Get(0).(func() *big.Int); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*big.Int)
		}
	}

	return r0
}

// mockHead_BlockDifficulty_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'BlockDifficulty'
type mockHead_BlockDifficulty_Call struct {
	*mock.Call
}

// BlockDifficulty is a helper method to define mock.On call
func (_e *mockHead_Expecter) BlockDifficulty() *mockHead_BlockDifficulty_Call {
	return &mockHead_BlockDifficulty_Call{Call: _e.mock.On("BlockDifficulty")}
}

func (_c *mockHead_BlockDifficulty_Call) Run(run func()) *mockHead_BlockDifficulty_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *mockHead_BlockDifficulty_Call) Return(_a0 *big.Int) *mockHead_BlockDifficulty_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockHead_BlockDifficulty_Call) RunAndReturn(run func() *big.Int) *mockHead_BlockDifficulty_Call {
	_c.Call.Return(run)
	return _c
}

// BlockNumber provides a mock function with no fields
func (_m *mockHead) BlockNumber() int64 {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for BlockNumber")
	}

	var r0 int64
	if rf, ok := ret.Get(0).(func() int64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int64)
	}

	return r0
}

// mockHead_BlockNumber_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'BlockNumber'
type mockHead_BlockNumber_Call struct {
	*mock.Call
}

// BlockNumber is a helper method to define mock.On call
func (_e *mockHead_Expecter) BlockNumber() *mockHead_BlockNumber_Call {
	return &mockHead_BlockNumber_Call{Call: _e.mock.On("BlockNumber")}
}

func (_c *mockHead_BlockNumber_Call) Run(run func()) *mockHead_BlockNumber_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *mockHead_BlockNumber_Call) Return(_a0 int64) *mockHead_BlockNumber_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockHead_BlockNumber_Call) RunAndReturn(run func() int64) *mockHead_BlockNumber_Call {
	_c.Call.Return(run)
	return _c
}

// IsValid provides a mock function with no fields
func (_m *mockHead) IsValid() bool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for IsValid")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// mockHead_IsValid_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IsValid'
type mockHead_IsValid_Call struct {
	*mock.Call
}

// IsValid is a helper method to define mock.On call
func (_e *mockHead_Expecter) IsValid() *mockHead_IsValid_Call {
	return &mockHead_IsValid_Call{Call: _e.mock.On("IsValid")}
}

func (_c *mockHead_IsValid_Call) Run(run func()) *mockHead_IsValid_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *mockHead_IsValid_Call) Return(_a0 bool) *mockHead_IsValid_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockHead_IsValid_Call) RunAndReturn(run func() bool) *mockHead_IsValid_Call {
	_c.Call.Return(run)
	return _c
}

// newMockHead creates a new instance of mockHead. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockHead(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockHead {
	mock := &mockHead{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
