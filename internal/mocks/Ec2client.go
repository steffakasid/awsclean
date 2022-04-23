// Code generated by mockery v2.10.6. DO NOT EDIT.

package mocks

import (
	context "context"

	ec2 "github.com/aws/aws-sdk-go-v2/service/ec2"

	mock "github.com/stretchr/testify/mock"
)

// Ec2client is an autogenerated mock type for the Ec2client type
type Ec2client struct {
	mock.Mock
}

type Ec2client_Expecter struct {
	mock *mock.Mock
}

func (_m *Ec2client) EXPECT() *Ec2client_Expecter {
	return &Ec2client_Expecter{mock: &_m.Mock}
}

// DeregisterImage provides a mock function with given fields: ctx, params, optFns
func (_m *Ec2client) DeregisterImage(ctx context.Context, params *ec2.DeregisterImageInput, optFns ...func(*ec2.Options)) (*ec2.DeregisterImageOutput, error) {
	_va := make([]interface{}, len(optFns))
	for _i := range optFns {
		_va[_i] = optFns[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, params)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *ec2.DeregisterImageOutput
	if rf, ok := ret.Get(0).(func(context.Context, *ec2.DeregisterImageInput, ...func(*ec2.Options)) *ec2.DeregisterImageOutput); ok {
		r0 = rf(ctx, params, optFns...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*ec2.DeregisterImageOutput)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *ec2.DeregisterImageInput, ...func(*ec2.Options)) error); ok {
		r1 = rf(ctx, params, optFns...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Ec2client_DeregisterImage_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeregisterImage'
type Ec2client_DeregisterImage_Call struct {
	*mock.Call
}

// DeregisterImage is a helper method to define mock.On call
//  - ctx context.Context
//  - params *ec2.DeregisterImageInput
//  - optFns ...func(*ec2.Options)
func (_e *Ec2client_Expecter) DeregisterImage(ctx interface{}, params interface{}, optFns ...interface{}) *Ec2client_DeregisterImage_Call {
	return &Ec2client_DeregisterImage_Call{Call: _e.mock.On("DeregisterImage",
		append([]interface{}{ctx, params}, optFns...)...)}
}

func (_c *Ec2client_DeregisterImage_Call) Run(run func(ctx context.Context, params *ec2.DeregisterImageInput, optFns ...func(*ec2.Options))) *Ec2client_DeregisterImage_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]func(*ec2.Options), len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(func(*ec2.Options))
			}
		}
		run(args[0].(context.Context), args[1].(*ec2.DeregisterImageInput), variadicArgs...)
	})
	return _c
}

func (_c *Ec2client_DeregisterImage_Call) Return(_a0 *ec2.DeregisterImageOutput, _a1 error) *Ec2client_DeregisterImage_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// DescribeImages provides a mock function with given fields: ctx, params, optFns
func (_m *Ec2client) DescribeImages(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error) {
	_va := make([]interface{}, len(optFns))
	for _i := range optFns {
		_va[_i] = optFns[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, params)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *ec2.DescribeImagesOutput
	if rf, ok := ret.Get(0).(func(context.Context, *ec2.DescribeImagesInput, ...func(*ec2.Options)) *ec2.DescribeImagesOutput); ok {
		r0 = rf(ctx, params, optFns...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*ec2.DescribeImagesOutput)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *ec2.DescribeImagesInput, ...func(*ec2.Options)) error); ok {
		r1 = rf(ctx, params, optFns...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Ec2client_DescribeImages_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DescribeImages'
type Ec2client_DescribeImages_Call struct {
	*mock.Call
}

// DescribeImages is a helper method to define mock.On call
//  - ctx context.Context
//  - params *ec2.DescribeImagesInput
//  - optFns ...func(*ec2.Options)
func (_e *Ec2client_Expecter) DescribeImages(ctx interface{}, params interface{}, optFns ...interface{}) *Ec2client_DescribeImages_Call {
	return &Ec2client_DescribeImages_Call{Call: _e.mock.On("DescribeImages",
		append([]interface{}{ctx, params}, optFns...)...)}
}

func (_c *Ec2client_DescribeImages_Call) Run(run func(ctx context.Context, params *ec2.DescribeImagesInput, optFns ...func(*ec2.Options))) *Ec2client_DescribeImages_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]func(*ec2.Options), len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(func(*ec2.Options))
			}
		}
		run(args[0].(context.Context), args[1].(*ec2.DescribeImagesInput), variadicArgs...)
	})
	return _c
}

func (_c *Ec2client_DescribeImages_Call) Return(_a0 *ec2.DescribeImagesOutput, _a1 error) *Ec2client_DescribeImages_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// DescribeInstances provides a mock function with given fields: ctx, params, optFns
func (_m *Ec2client) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	_va := make([]interface{}, len(optFns))
	for _i := range optFns {
		_va[_i] = optFns[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, params)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *ec2.DescribeInstancesOutput
	if rf, ok := ret.Get(0).(func(context.Context, *ec2.DescribeInstancesInput, ...func(*ec2.Options)) *ec2.DescribeInstancesOutput); ok {
		r0 = rf(ctx, params, optFns...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*ec2.DescribeInstancesOutput)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *ec2.DescribeInstancesInput, ...func(*ec2.Options)) error); ok {
		r1 = rf(ctx, params, optFns...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Ec2client_DescribeInstances_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DescribeInstances'
type Ec2client_DescribeInstances_Call struct {
	*mock.Call
}

// DescribeInstances is a helper method to define mock.On call
//  - ctx context.Context
//  - params *ec2.DescribeInstancesInput
//  - optFns ...func(*ec2.Options)
func (_e *Ec2client_Expecter) DescribeInstances(ctx interface{}, params interface{}, optFns ...interface{}) *Ec2client_DescribeInstances_Call {
	return &Ec2client_DescribeInstances_Call{Call: _e.mock.On("DescribeInstances",
		append([]interface{}{ctx, params}, optFns...)...)}
}

func (_c *Ec2client_DescribeInstances_Call) Run(run func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options))) *Ec2client_DescribeInstances_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]func(*ec2.Options), len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(func(*ec2.Options))
			}
		}
		run(args[0].(context.Context), args[1].(*ec2.DescribeInstancesInput), variadicArgs...)
	})
	return _c
}

func (_c *Ec2client_DescribeInstances_Call) Return(_a0 *ec2.DescribeInstancesOutput, _a1 error) *Ec2client_DescribeInstances_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// DescribeLaunchTemplateVersions provides a mock function with given fields: ctx, params, optFns
func (_m *Ec2client) DescribeLaunchTemplateVersions(ctx context.Context, params *ec2.DescribeLaunchTemplateVersionsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeLaunchTemplateVersionsOutput, error) {
	_va := make([]interface{}, len(optFns))
	for _i := range optFns {
		_va[_i] = optFns[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, params)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *ec2.DescribeLaunchTemplateVersionsOutput
	if rf, ok := ret.Get(0).(func(context.Context, *ec2.DescribeLaunchTemplateVersionsInput, ...func(*ec2.Options)) *ec2.DescribeLaunchTemplateVersionsOutput); ok {
		r0 = rf(ctx, params, optFns...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*ec2.DescribeLaunchTemplateVersionsOutput)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *ec2.DescribeLaunchTemplateVersionsInput, ...func(*ec2.Options)) error); ok {
		r1 = rf(ctx, params, optFns...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Ec2client_DescribeLaunchTemplateVersions_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DescribeLaunchTemplateVersions'
type Ec2client_DescribeLaunchTemplateVersions_Call struct {
	*mock.Call
}

// DescribeLaunchTemplateVersions is a helper method to define mock.On call
//  - ctx context.Context
//  - params *ec2.DescribeLaunchTemplateVersionsInput
//  - optFns ...func(*ec2.Options)
func (_e *Ec2client_Expecter) DescribeLaunchTemplateVersions(ctx interface{}, params interface{}, optFns ...interface{}) *Ec2client_DescribeLaunchTemplateVersions_Call {
	return &Ec2client_DescribeLaunchTemplateVersions_Call{Call: _e.mock.On("DescribeLaunchTemplateVersions",
		append([]interface{}{ctx, params}, optFns...)...)}
}

func (_c *Ec2client_DescribeLaunchTemplateVersions_Call) Run(run func(ctx context.Context, params *ec2.DescribeLaunchTemplateVersionsInput, optFns ...func(*ec2.Options))) *Ec2client_DescribeLaunchTemplateVersions_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]func(*ec2.Options), len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(func(*ec2.Options))
			}
		}
		run(args[0].(context.Context), args[1].(*ec2.DescribeLaunchTemplateVersionsInput), variadicArgs...)
	})
	return _c
}

func (_c *Ec2client_DescribeLaunchTemplateVersions_Call) Return(_a0 *ec2.DescribeLaunchTemplateVersionsOutput, _a1 error) *Ec2client_DescribeLaunchTemplateVersions_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}
