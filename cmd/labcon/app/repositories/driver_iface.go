package repositories

import "github.com/ktnyt/labcon/cmd/labcon/app/models"

type DriverRepository interface {
	List() ([]string, error)
	Create(name, token string, state interface{}) error
	Fetch(name string) (models.DriverModel, error)
	Update(driver models.DriverModel) error
	Delete(name string) error
}
