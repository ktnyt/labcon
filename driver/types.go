package driver

type Status string

const (
	Idle  Status = "idle"
	Busy  Status = "busy"
	Lost  Status = "lost"
	Error Status = "error"
)

type Op struct {
	Name string      `json:"name" validate:"required"`
	Arg  interface{} `json:"arg,omitempty"`
}

type RegisterParams struct {
	Name  string      `json:"name" validate:"required"`
	State interface{} `json:"state" validate:"required"`
}
