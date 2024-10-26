package util

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"go.opentelemetry.io/otel/trace"
)

type validationError struct {
	errors map[string]string
}

type Validator struct {
	validate *validator.Validate
	tracer   trace.Tracer
}

func NewValidator(tracer trace.Tracer) *Validator {
	return &Validator{validate: validator.New(), tracer: tracer}
}

func (v *Validator) Validate(ctx context.Context, data interface{}) *validationError {
	_, span := v.tracer.Start(ctx, "validating request")
	defer span.End()

	if err := v.validate.Struct(data); err != nil {
		validationErrorsString := make(map[string]string)
		validationErrors := err.(validator.ValidationErrors)

		for _, e := range validationErrors {
			validationErrorsString[e.Field()] = getErrorMessage(e)
		}

		return &validationError{errors: validationErrorsString}
	}
	return nil
}

func getErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("The %s field is required", err.Field())
	case "email":
		return fmt.Sprintf("The %s field must be a valid email", err.Field())
	case "gte":
		return fmt.Sprintf("The %s field must be greater than or equal to %s", err.Field(), err.Param())
	default:
		return fmt.Sprintf("The %s field is invalid", err.Field())
	}
}
