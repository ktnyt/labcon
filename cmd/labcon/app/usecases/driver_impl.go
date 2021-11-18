package usecases

import (
	"github.com/ktnyt/labcon/cmd/labcon/app/models"
	"github.com/ktnyt/labcon/cmd/labcon/app/repositories"
	"github.com/ktnyt/labcon/cmd/labcon/lib"
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

func (usecase DriverUsecaseImpl) GetStatus(name string) (models.DriverStatus, error) {
	model, err := usecase.repository.Fetch(name)
	if err != nil {
		return models.DriverError, err
	}
	return model.Status, nil
}

func (usecase DriverUsecaseImpl) SetStatus(name string, status models.DriverStatus) error {
	model, err := usecase.repository.Fetch(name)
	if err != nil {
		return err
	}
	model.Status = status
	if status == models.DriverIdle {
		model.Op = nil
	}
	return usecase.repository.Update(model)
}

func (usecase DriverUsecaseImpl) GetOp(name string) (*models.DriverOp, error) {
	model, err := usecase.repository.Fetch(name)
	if err != nil {
		return nil, err
	}
	return model.Op, nil
}

func (usecase DriverUsecaseImpl) SetOp(name string, op models.DriverOp) error {
	model, err := usecase.repository.Fetch(name)
	if err != nil {
		return err
	}
	model.Status = models.DriverBusy
	model.Op = &op
	return usecase.repository.Update(model)
}
