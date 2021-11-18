package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
)

func RunCase(t *testing.T, label interface{}, f func(t *testing.T)) {
	t.Helper()
	if i, ok := label.(int); ok {
		label = i + 1
	}
	t.Run(fmt.Sprintf("case %v", label), f)
}

func JsonMarshalToBuffer(t *testing.T, p interface{}) *bytes.Buffer {
	t.Helper()
	b := &bytes.Buffer{}
	enc := json.NewEncoder(b)
	if err := enc.Encode(p); err != nil {
		t.Fatalf("failed to encode JSON: %v", err)
	}
	return b
}
