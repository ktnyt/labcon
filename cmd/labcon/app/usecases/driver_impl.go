package usecases

import (
	"github.com/ktnyt/labcon/cmd/labcon/app/repositories"
	"github.com/ktnyt/labcon/cmd/labcon/lib"
	"github.com/ktnyt/labcon/driver"
)

type DriverUsecaseImpl struct {
	repository repositories.DriverRepository
	generate   func() string
}

func NewDriverUsecase(repository repositories.DriverRepository, generate func() string) DriverUsecase {
	return DriverUsecaseImpl{
		repository: repository,
		generate:   generate,
	}
}

func (usecase DriverUsecaseImpl) Register(name string, state interface{}) (string, error) {
	token := usecase.generate()
	err := usecase.repository.Create(name, token, state)
	return token, err
}

func (usecase DriverUsecaseImpl) Authorize(name string, token string) error {
	model, err := usecase.repository.Fetch(name)
	if err != nil {
		return err
	}
	if model.Token != token {
		return lib.ErrUnauthorized
	}
	return nil
}

func (usecase DriverUsecaseImpl) GetState(name string) (interface{}, error) {
	model, err := usecase.repository.Fetch(name)
	return model.State, err
}

func (usecase DriverUsecaseImpl) SetState(name string, state interface{}) error {
	model, err := usecase.repository.Fetch(name)
	if err != nil {
		return err
	}
	model.State = state
	return usecase.repository.Update(model)
}

func (usecase DriverUsecaseImpl) GetStatus(name string) (driver.Status, error) {
	model, err := usecase.repository.Fetch(name)
	if err != nil {
		return driver.Error, err
	}
	return model.Status, nil
}

func (usecase DriverUsecaseImpl) SetStatus(name string, status driver.Status) error {
	model, err := usecase.repository.Fetch(name)
	if err != nil {
		return err
	}
	model.Status = status
	model.Op = nil
	return usecase.repository.Update(model)
}

func (usecase DriverUsecaseImpl) GetOp(name string) (*driver.Op, error) {
	model, err := usecase.repository.Fetch(name)
	if err != nil {
		return nil, err
	}
	return model.Op, nil
}

func (usecase DriverUsecaseImpl) SetOp(name string, op driver.Op) error {
	model, err := usecase.repository.Fetch(name)
	if err != nil {
		return err
	}
	if model.Status != driver.Idle || model.Op != nil {
		return lib.ErrAlreadyExists
	}
	model.Status = driver.Busy
	model.Op = &op
	return usecase.repository.Update(model)
}
