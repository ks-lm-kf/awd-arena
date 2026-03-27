package validator

import "github.com/go-playground/validator/v10"

// New creates a new validator instance.
func New() *validator.Validate {
	return validator.New()
}

// Validate validates a struct.
func Validate(v *validator.Validate, s interface{}) error {
	return v.Struct(s)
}
