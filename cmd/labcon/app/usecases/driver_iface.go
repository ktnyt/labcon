package usecases

import "github.com/ktnyt/labcon/driver"

type DriverUsecase interface {
	Register(name string, state interface{}) (string, error)
	Authorize(name string, token string) error
	GetState(name string) (interface{}, error)
	SetState(name string, state interface{}) error
	GetStatus(name string) (driver.Status, error)
	SetStatus(name string, status driver.Status) error
	GetOp(name string) (*driver.Op, error)
	SetOp(name string, op driver.Op) error
}
