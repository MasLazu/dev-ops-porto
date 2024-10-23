package util

import (
	"context"

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
			validationErrorsString[e.Field()] = e.Tag()
		}

		return &validationError{errors: validationErrorsString}
	}
	return nil
}
