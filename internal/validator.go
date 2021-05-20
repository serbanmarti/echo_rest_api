package internal

import (
	"errors"
	"strings"

	"github.com/go-playground/validator"
)

type (
	Validator struct {
		validator *validator.Validate
	}
)

var (
	roles = []string{"admin", "user"}
)

// CreateValidator creates a new input data validator instance
func CreateValidator() (*Validator, error) {
	v := validator.New()

	// Register role validation
	err := v.RegisterValidation("role", ValidateRole)
	if err != nil {
		return nil, errors.New("could not assign role validator")
	}

	// Register password validation
	err = v.RegisterValidation("password", ValidatePassword)
	if err != nil {
		return nil, errors.New("could not assign password validator")
	}

	return &Validator{validator: v}, nil
}

// Validate an input data structure
func (cv *Validator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

// ValidateRole will validate a role field
func ValidateRole(f validator.FieldLevel) bool {
	return InSlice(f.Field().String(), roles)
}

// ValidatePassword will validate a password field
func ValidatePassword(f validator.FieldLevel) bool {
	s := f.Field().String()
	return len(s) > 0 &&
		!strings.Contains(s, " ") &&
		strings.ContainsAny(s, "0123456789") &&
		strings.ContainsAny(s, "{}[]:;'\"\\/?.<>,=+-_()!@#$%^&*~`")
}
