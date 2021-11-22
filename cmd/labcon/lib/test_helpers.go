package lib

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/ktnyt/labcon/utils"
)

func RunCase(t *testing.T, label interface{}, f func(t *testing.T)) {
	t.Helper()
	if i, ok := label.(int); ok {
		label = i + 1
	}
	t.Run(fmt.Sprintf("case %v", label), f)
}

func MustJsonMarshalToBuffer(t *testing.T, p interface{}) *bytes.Buffer {
	t.Helper()
	b, err := utils.JsonMarshalToBuffer(p)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
