package labcon

import "github.com/ktnyt/labcon/driver"

type Driver struct {
	client *Client
	name   string
	token  string
}

func NewDriver(client *Client, name string, state interface{}) (Driver, error) {
	token, err := client.Register(name, state)
	return Driver{
		client: client,
		name:   name,
		token:  token,
	}, err
}

func (driver Driver) GetState(state interface{}) error {
	return driver.client.GetState(driver.name, state)
}

func (driver Driver) SetState(state interface{}) error {
	return driver.client.SetState(driver.name, driver.token, state)
}

func (driver Driver) GetStatus() (driver.Status, error) {
	return driver.client.GetStatus(driver.name)
}

func (driver Driver) SetStatus(status driver.Status) error {
	return driver.client.SetStatus(driver.name, driver.token, status)
}

func (driver Driver) Operation() (*driver.Op, error) {
	return driver.client.Operation(driver.name, driver.token)
}

func (driver Driver) Dispatch(op driver.Op) error {
	return driver.client.Dispatch(driver.name, op)
}
