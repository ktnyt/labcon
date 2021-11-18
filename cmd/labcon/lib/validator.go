package lib

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/schema"
)

var decoder = schema.NewDecoder()

var validate *validator.Validate

func init() {
	validate = validator.New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		fieldName := fld.Tag.Get("json")
		if fieldName == "-" {
			return ""
		}
		return fieldName
	})
}

type CustomValidator interface {
	Validate() error
}

var (
	ErrSchemaDecode = errors.New("failed to decode values")
)

type ValidationError []validator.FieldError

func (err ValidationError) Error() string {
	msgs := make([]string, len(err))
	for i, field := range err {
		msgs[i] = fmt.Sprintf(
			"validation failed on field %q for constraint %q",
			field.Field(),
			field.Tag(),
		)
	}
	return strings.Join(msgs, "\n")
}

// Validate validates an object using the global validator.
func Validate(p interface{}) error {
	if err := validate.Struct(p); err != nil {
		if fields, ok := err.(validator.ValidationErrors); ok {
			err = ValidationError(fields)
		}
		return err
	}

	if val, ok := p.(CustomValidator); ok {
		if err := val.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// ValidateQuery decodes and validates an object from query values using the
// global decoder and validator.
func ValidateQuery(p interface{}, v url.Values) error {
	if err := decoder.Decode(p, v); err != nil {
		return ErrSchemaDecode
	}
	return Validate(p)
}
