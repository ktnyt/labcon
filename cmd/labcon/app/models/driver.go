package models

type DriverStatus string

const (
	DriverIdle  DriverStatus = "idle"
	DriverBusy  DriverStatus = "busy"
	DriverLost  DriverStatus = "lost"
	DriverError DriverStatus = "error"
)

type DriverOp struct {
	Name string      `json:"name" validate:"required"`
	Arg  interface{} `json:"arg,omitempty"`
}

type DriverModel struct {
	Name   string `msgpack:"-"`
	Token  string
	State  interface{}
	Status DriverStatus
	Op     *DriverOp `msgpack:",omitempty"`
}

func NewDriver(name, token string, state interface{}) DriverModel {
	return DriverModel{
		Name:   name,
		Token:  token,
		State:  state,
		Status: DriverIdle,
		Op:     nil,
	}
}
