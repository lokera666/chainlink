// Code generated by mockery v2.50.0. DO NOT EDIT.

package mocks

import (
	legacyevm "github.com/smartcontractkit/chainlink/v2/core/chains/legacyevm"
	mock "github.com/stretchr/testify/mock"

	types "github.com/smartcontractkit/chainlink-integrations/evm/types"
)

// LegacyChainContainer is an autogenerated mock type for the LegacyChainContainer type
type LegacyChainContainer struct {
	mock.Mock
}

type LegacyChainContainer_Expecter struct {
	mock *mock.Mock
}

func (_m *LegacyChainContainer) EXPECT() *LegacyChainContainer_Expecter {
	return &LegacyChainContainer_Expecter{mock: &_m.Mock}
}

// ChainNodeConfigs provides a mock function with no fields
func (_m *LegacyChainContainer) ChainNodeConfigs() types.Configs {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for ChainNodeConfigs")
	}

	var r0 types.Configs
	if rf, ok := ret.Get(0).(func() types.Configs); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(types.Configs)
		}
	}

	return r0
}

// LegacyChainContainer_ChainNodeConfigs_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ChainNodeConfigs'
type LegacyChainContainer_ChainNodeConfigs_Call struct {
	*mock.Call
}

// ChainNodeConfigs is a helper method to define mock.On call
func (_e *LegacyChainContainer_Expecter) ChainNodeConfigs() *LegacyChainContainer_ChainNodeConfigs_Call {
	return &LegacyChainContainer_ChainNodeConfigs_Call{Call: _e.mock.On("ChainNodeConfigs")}
}

func (_c *LegacyChainContainer_ChainNodeConfigs_Call) Run(run func()) *LegacyChainContainer_ChainNodeConfigs_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *LegacyChainContainer_ChainNodeConfigs_Call) Return(_a0 types.Configs) *LegacyChainContainer_ChainNodeConfigs_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *LegacyChainContainer_ChainNodeConfigs_Call) RunAndReturn(run func() types.Configs) *LegacyChainContainer_ChainNodeConfigs_Call {
	_c.Call.Return(run)
	return _c
}

// Get provides a mock function with given fields: id
func (_m *LegacyChainContainer) Get(id string) (legacyevm.Chain, error) {
	ret := _m.Called(id)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 legacyevm.Chain
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (legacyevm.Chain, error)); ok {
		return rf(id)
	}
	if rf, ok := ret.Get(0).(func(string) legacyevm.Chain); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(legacyevm.Chain)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LegacyChainContainer_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type LegacyChainContainer_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - id string
func (_e *LegacyChainContainer_Expecter) Get(id interface{}) *LegacyChainContainer_Get_Call {
	return &LegacyChainContainer_Get_Call{Call: _e.mock.On("Get", id)}
}

func (_c *LegacyChainContainer_Get_Call) Run(run func(id string)) *LegacyChainContainer_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *LegacyChainContainer_Get_Call) Return(_a0 legacyevm.Chain, _a1 error) *LegacyChainContainer_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *LegacyChainContainer_Get_Call) RunAndReturn(run func(string) (legacyevm.Chain, error)) *LegacyChainContainer_Get_Call {
	_c.Call.Return(run)
	return _c
}

// Len provides a mock function with no fields
func (_m *LegacyChainContainer) Len() int {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Len")
	}

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// LegacyChainContainer_Len_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Len'
type LegacyChainContainer_Len_Call struct {
	*mock.Call
}

// Len is a helper method to define mock.On call
func (_e *LegacyChainContainer_Expecter) Len() *LegacyChainContainer_Len_Call {
	return &LegacyChainContainer_Len_Call{Call: _e.mock.On("Len")}
}

func (_c *LegacyChainContainer_Len_Call) Run(run func()) *LegacyChainContainer_Len_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *LegacyChainContainer_Len_Call) Return(_a0 int) *LegacyChainContainer_Len_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *LegacyChainContainer_Len_Call) RunAndReturn(run func() int) *LegacyChainContainer_Len_Call {
	_c.Call.Return(run)
	return _c
}

// List provides a mock function with given fields: ids
func (_m *LegacyChainContainer) List(ids ...string) ([]legacyevm.Chain, error) {
	_va := make([]interface{}, len(ids))
	for _i := range ids {
		_va[_i] = ids[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for List")
	}

	var r0 []legacyevm.Chain
	var r1 error
	if rf, ok := ret.Get(0).(func(...string) ([]legacyevm.Chain, error)); ok {
		return rf(ids...)
	}
	if rf, ok := ret.Get(0).(func(...string) []legacyevm.Chain); ok {
		r0 = rf(ids...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]legacyevm.Chain)
		}
	}

	if rf, ok := ret.Get(1).(func(...string) error); ok {
		r1 = rf(ids...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LegacyChainContainer_List_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'List'
type LegacyChainContainer_List_Call struct {
	*mock.Call
}

// List is a helper method to define mock.On call
//   - ids ...string
func (_e *LegacyChainContainer_Expecter) List(ids ...interface{}) *LegacyChainContainer_List_Call {
	return &LegacyChainContainer_List_Call{Call: _e.mock.On("List",
		append([]interface{}{}, ids...)...)}
}

func (_c *LegacyChainContainer_List_Call) Run(run func(ids ...string)) *LegacyChainContainer_List_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]string, len(args)-0)
		for i, a := range args[0:] {
			if a != nil {
				variadicArgs[i] = a.(string)
			}
		}
		run(variadicArgs...)
	})
	return _c
}

func (_c *LegacyChainContainer_List_Call) Return(_a0 []legacyevm.Chain, _a1 error) *LegacyChainContainer_List_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *LegacyChainContainer_List_Call) RunAndReturn(run func(...string) ([]legacyevm.Chain, error)) *LegacyChainContainer_List_Call {
	_c.Call.Return(run)
	return _c
}

// Slice provides a mock function with no fields
func (_m *LegacyChainContainer) Slice() []legacyevm.Chain {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Slice")
	}

	var r0 []legacyevm.Chain
	if rf, ok := ret.Get(0).(func() []legacyevm.Chain); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]legacyevm.Chain)
		}
	}

	return r0
}

// LegacyChainContainer_Slice_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Slice'
type LegacyChainContainer_Slice_Call struct {
	*mock.Call
}

// Slice is a helper method to define mock.On call
func (_e *LegacyChainContainer_Expecter) Slice() *LegacyChainContainer_Slice_Call {
	return &LegacyChainContainer_Slice_Call{Call: _e.mock.On("Slice")}
}

func (_c *LegacyChainContainer_Slice_Call) Run(run func()) *LegacyChainContainer_Slice_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *LegacyChainContainer_Slice_Call) Return(_a0 []legacyevm.Chain) *LegacyChainContainer_Slice_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *LegacyChainContainer_Slice_Call) RunAndReturn(run func() []legacyevm.Chain) *LegacyChainContainer_Slice_Call {
	_c.Call.Return(run)
	return _c
}

// NewLegacyChainContainer creates a new instance of LegacyChainContainer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewLegacyChainContainer(t interface {
	mock.TestingT
	Cleanup(func())
}) *LegacyChainContainer {
	mock := &LegacyChainContainer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
