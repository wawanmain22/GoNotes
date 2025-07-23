package utils

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// Validator wraps the validator instance for dependency injection
type Validator struct {
	validator *validator.Validate
}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{
		validator: validate,
	}
}

// ValidateStruct validates a struct using validation tags
func (v *Validator) ValidateStruct(s interface{}) error {
	return v.validator.Struct(s)
}

// ValidateStruct validates a struct using validation tags
func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}

// FormatValidationError formats validation errors into a readable string
func FormatValidationError(err error) string {
	if err == nil {
		return ""
	}

	var errors []string

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, validationError := range validationErrors {
			field := strings.ToLower(validationError.Field())

			switch validationError.Tag() {
			case "required":
				errors = append(errors, fmt.Sprintf("%s is required", field))
			case "email":
				errors = append(errors, fmt.Sprintf("%s must be a valid email address", field))
			case "min":
				errors = append(errors, fmt.Sprintf("%s must be at least %s characters long", field, validationError.Param()))
			case "max":
				errors = append(errors, fmt.Sprintf("%s must be at most %s characters long", field, validationError.Param()))
			default:
				errors = append(errors, fmt.Sprintf("%s is invalid", field))
			}
		}
	} else {
		errors = append(errors, err.Error())
	}

	return strings.Join(errors, ", ")
}
