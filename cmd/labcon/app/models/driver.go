package models

import "github.com/ktnyt/labcon/driver"

type DriverModel struct {
	Name   string `msgpack:"-"`
	Token  string
	State  interface{}
	Status driver.Status
	Op     *driver.Op `msgpack:",omitempty"`
}

func NewDriver(name, token string, state interface{}) DriverModel {
	return DriverModel{
		Name:   name,
		Token:  token,
		State:  state,
		Status: driver.Idle,
		Op:     nil,
	}
}
