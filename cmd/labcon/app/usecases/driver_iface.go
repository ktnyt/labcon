package usecases

import "github.com/ktnyt/labcon/cmd/labcon/app/models"

type DriverUsecase interface {
	Register(name string, state interface{}) (string, error)
	Authorize(name string, token string) error
	GetState(name string) (interface{}, error)
	SetState(name string, state interface{}) error
	GetStatus(name string) (models.DriverStatus, error)
	SetStatus(name string, status models.DriverStatus) error
	GetOp(name string) (*models.DriverOp, error)
	SetOp(name string, op models.DriverOp) error
}
