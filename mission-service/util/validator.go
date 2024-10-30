package util

import (
	"context"
	"fmt"
	"time"

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
	validator := validator.New()
	validator.RegisterValidation("future", futureDate)
	return &Validator{validate: validator, tracer: tracer}
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

func futureDate(fl validator.FieldLevel) bool {
	if date, ok := fl.Field().Interface().(time.Time); ok {
		return date.After(time.Now())
	}
	return false
}

func getErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("The %s field is required", err.Field())
	case "email":
		return fmt.Sprintf("The %s field must be a valid email", err.Field())
	case "gte":
		return fmt.Sprintf("The %s field must be greater than or equal to %s", err.Field(), err.Param())
	case "future":
		return fmt.Sprintf("The %s field must be a future date", err.Field())
	default:
		return fmt.Sprintf("The %s field is invalid", err.Field())
	}
}
