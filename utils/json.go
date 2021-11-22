package utils

import (
	"bytes"
	"encoding/json"
)

func JsonMarshalToBuffer(p interface{}) (*bytes.Buffer, error) {
	b := &bytes.Buffer{}
	enc := json.NewEncoder(b)
	if err := enc.Encode(p); err != nil {
		return nil, err
	}
	return b, nil
}
