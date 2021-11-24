package labcon

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/ktnyt/labcon/driver"
	"github.com/ktnyt/labcon/utils"
)

type Client struct {
	Addr string
}

func NewClient(addr string) *Client {
	return &Client{Addr: addr}
}

func (client *Client) List() ([]string, error) {
	url := fmt.Sprintf("%s/driver", client.Addr)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	buf := bytes.Buffer{}
	io.Copy(&buf, res.Body)

	if res.StatusCode != http.StatusOK {
		return nil, errors.New(buf.String())
	}

	var names []string
	err = json.Unmarshal(buf.Bytes(), &names)
	return names, err
}

func (client *Client) Register(name string, state interface{}) (string, error) {
	params := driver.RegisterParams{
		Name:  name,
		State: state,
	}

	body, err := utils.JsonMarshalToBuffer(params)
	if err != nil {
		return "", fmt.Errorf("failed to register driver %q: %v", name, err)
	}

	url := fmt.Sprintf("%s/driver", client.Addr)
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return "", fmt.Errorf("failed to register driver %q: %v", name, err)
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to register driver %q: %v", name, err)
	}

	buf := bytes.Buffer{}
	io.Copy(&buf, res.Body)

	if res.StatusCode != http.StatusOK {
		return "", errors.New(buf.String())
	}

	var token string
	err = json.Unmarshal(buf.Bytes(), &token)
	return token, err
}

func (client *Client) GetState(name string, state interface{}) error {
	url := fmt.Sprintf("%s/driver/%s/state", client.Addr, name)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to get state for driver %q: %v", name, err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get state for driver %q: %v", name, err)
	}

	buf := bytes.Buffer{}
	io.Copy(&buf, res.Body)

	if res.StatusCode != http.StatusOK {
		return errors.New(buf.String())
	}

	return json.Unmarshal(buf.Bytes(), state)
}

func (client *Client) SetState(name, token string, state interface{}) error {
	body, err := utils.JsonMarshalToBuffer(state)
	if err != nil {
		return fmt.Errorf("failed to register driver %q: %v", name, err)
	}

	url := fmt.Sprintf("%s/driver/%s/state", client.Addr, name)
	req, err := http.NewRequest(http.MethodPut, url, body)
	if err != nil {
		return fmt.Errorf("failed to set state for driver %q: %v", name, err)
	}
	req.Header.Add("X-Driver-Token", token)
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to set state for driver %q: %v", name, err)
	}

	if res.StatusCode != http.StatusOK {
		buf := bytes.Buffer{}
		io.Copy(&buf, res.Body)
		return errors.New(buf.String())
	}

	return nil
}

func (client *Client) GetStatus(name string) (driver.Status, error) {
	url := fmt.Sprintf("%s/driver/%s/status", client.Addr, name)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return driver.Error, fmt.Errorf("failed to get status for driver %q: %v", name, err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return driver.Error, fmt.Errorf("failed to get status for driver %q: %v", name, err)
	}

	buf := bytes.Buffer{}
	io.Copy(&buf, res.Body)

	if res.StatusCode != http.StatusOK {
		return driver.Error, errors.New(buf.String())
	}

	status := driver.Error
	err = json.Unmarshal(buf.Bytes(), &status)
	return status, err
}

func (client *Client) SetStatus(name, token string, status driver.Status) error {
	body, err := utils.JsonMarshalToBuffer(status)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/driver/%s/status", client.Addr, name)
	req, err := http.NewRequest(http.MethodPut, url, body)
	if err != nil {
		return err
	}
	req.Header.Add("X-Driver-Token", token)
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to set status for driver %q: %v", name, err)
	}

	if res.StatusCode != http.StatusOK {
		buf := bytes.Buffer{}
		io.Copy(&buf, res.Body)
		return errors.New(buf.String())
	}

	return nil
}

func (client *Client) Operation(name, token string) (*driver.Op, error) {
	url := fmt.Sprintf("%s/driver/%s/operation", client.Addr, name)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get operation for driver %q: %v", name, err)
	}
	req.Header.Add("X-Driver-Token", token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get operation for driver %q: %v", name, err)
	}

	buf := bytes.Buffer{}
	io.Copy(&buf, res.Body)

	if res.StatusCode != http.StatusOK {
		return nil, errors.New(buf.String())
	}

	op := new(driver.Op)
	err = json.Unmarshal(buf.Bytes(), &op)
	return op, err
}

func (client *Client) Dispatch(name string, op driver.Op) error {
	body, err := utils.JsonMarshalToBuffer(op)
	if err != nil {
		return fmt.Errorf("failed to dispatch to driver %q: %v", name, body)
	}

	url := fmt.Sprintf("%s/driver/%s/operation", client.Addr, name)
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return fmt.Errorf("failed to dispatch to driver %q: %v", name, body)
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to dispatch to driver %q: %v", name, body)
	}

	if res.StatusCode != http.StatusOK {
		buf := bytes.Buffer{}
		io.Copy(&buf, res.Body)
		return errors.New(buf.String())
	}

	return nil
}
