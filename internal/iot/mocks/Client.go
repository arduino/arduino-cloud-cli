// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	iot "github.com/arduino/iot-client-go"
	mock "github.com/stretchr/testify/mock"

	os "os"
)

// Client is an autogenerated mock type for the Client type
type Client struct {
	mock.Mock
}

// CertificateCreate provides a mock function with given fields: id, csr
func (_m *Client) CertificateCreate(id string, csr string) (*iot.ArduinoCompressedv2, error) {
	ret := _m.Called(id, csr)

	var r0 *iot.ArduinoCompressedv2
	if rf, ok := ret.Get(0).(func(string, string) *iot.ArduinoCompressedv2); ok {
		r0 = rf(id, csr)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*iot.ArduinoCompressedv2)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(id, csr)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DashboardCreate provides a mock function with given fields: dashboard
func (_m *Client) DashboardCreate(dashboard *iot.Dashboardv2) (*iot.ArduinoDashboardv2, error) {
	ret := _m.Called(dashboard)

	var r0 *iot.ArduinoDashboardv2
	if rf, ok := ret.Get(0).(func(*iot.Dashboardv2) *iot.ArduinoDashboardv2); ok {
		r0 = rf(dashboard)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*iot.ArduinoDashboardv2)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*iot.Dashboardv2) error); ok {
		r1 = rf(dashboard)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DashboardDelete provides a mock function with given fields: id
func (_m *Client) DashboardDelete(id string) error {
	ret := _m.Called(id)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DashboardList provides a mock function with given fields:
func (_m *Client) DashboardList() ([]iot.ArduinoDashboardv2, error) {
	ret := _m.Called()

	var r0 []iot.ArduinoDashboardv2
	if rf, ok := ret.Get(0).(func() []iot.ArduinoDashboardv2); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]iot.ArduinoDashboardv2)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DashboardShow provides a mock function with given fields: id
func (_m *Client) DashboardShow(id string) (*iot.ArduinoDashboardv2, error) {
	ret := _m.Called(id)

	var r0 *iot.ArduinoDashboardv2
	if rf, ok := ret.Get(0).(func(string) *iot.ArduinoDashboardv2); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*iot.ArduinoDashboardv2)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeviceCreate provides a mock function with given fields: fqbn, name, serial, devType
func (_m *Client) DeviceCreate(fqbn string, name string, serial string, devType string) (*iot.ArduinoDevicev2, error) {
	ret := _m.Called(fqbn, name, serial, devType)

	var r0 *iot.ArduinoDevicev2
	if rf, ok := ret.Get(0).(func(string, string, string, string) *iot.ArduinoDevicev2); ok {
		r0 = rf(fqbn, name, serial, devType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*iot.ArduinoDevicev2)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string, string) error); ok {
		r1 = rf(fqbn, name, serial, devType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeviceDelete provides a mock function with given fields: id
func (_m *Client) DeviceDelete(id string) error {
	ret := _m.Called(id)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeviceList provides a mock function with given fields:
func (_m *Client) DeviceList() ([]iot.ArduinoDevicev2, error) {
	ret := _m.Called()

	var r0 []iot.ArduinoDevicev2
	if rf, ok := ret.Get(0).(func() []iot.ArduinoDevicev2); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]iot.ArduinoDevicev2)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeviceOTA provides a mock function with given fields: id, file, expireMins
func (_m *Client) DeviceOTA(id string, file *os.File, expireMins int) error {
	ret := _m.Called(id, file, expireMins)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, *os.File, int) error); ok {
		r0 = rf(id, file, expireMins)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeviceShow provides a mock function with given fields: id
func (_m *Client) DeviceShow(id string) (*iot.ArduinoDevicev2, error) {
	ret := _m.Called(id)

	var r0 *iot.ArduinoDevicev2
	if rf, ok := ret.Get(0).(func(string) *iot.ArduinoDevicev2); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*iot.ArduinoDevicev2)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ThingCreate provides a mock function with given fields: thing, force
func (_m *Client) ThingCreate(thing *iot.Thing, force bool) (*iot.ArduinoThing, error) {
	ret := _m.Called(thing, force)

	var r0 *iot.ArduinoThing
	if rf, ok := ret.Get(0).(func(*iot.Thing, bool) *iot.ArduinoThing); ok {
		r0 = rf(thing, force)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*iot.ArduinoThing)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*iot.Thing, bool) error); ok {
		r1 = rf(thing, force)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ThingDelete provides a mock function with given fields: id
func (_m *Client) ThingDelete(id string) error {
	ret := _m.Called(id)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ThingList provides a mock function with given fields: ids, device, props
func (_m *Client) ThingList(ids []string, device *string, props bool) ([]iot.ArduinoThing, error) {
	ret := _m.Called(ids, device, props)

	var r0 []iot.ArduinoThing
	if rf, ok := ret.Get(0).(func([]string, *string, bool) []iot.ArduinoThing); ok {
		r0 = rf(ids, device, props)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]iot.ArduinoThing)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]string, *string, bool) error); ok {
		r1 = rf(ids, device, props)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ThingShow provides a mock function with given fields: id
func (_m *Client) ThingShow(id string) (*iot.ArduinoThing, error) {
	ret := _m.Called(id)

	var r0 *iot.ArduinoThing
	if rf, ok := ret.Get(0).(func(string) *iot.ArduinoThing); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*iot.ArduinoThing)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ThingUpdate provides a mock function with given fields: id, thing, force
func (_m *Client) ThingUpdate(id string, thing *iot.Thing, force bool) error {
	ret := _m.Called(id, thing, force)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, *iot.Thing, bool) error); ok {
		r0 = rf(id, thing, force)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
