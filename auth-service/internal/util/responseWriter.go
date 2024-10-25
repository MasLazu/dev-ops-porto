package util

import (
	"context"
	"encoding/json"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type ResponseWriter struct {
	tracer trace.Tracer
}

type response struct {
	Status  int         `json:"status"`
	Suceess bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func NewResponseWriter(tracer trace.Tracer) *ResponseWriter {
	return &ResponseWriter{tracer: tracer}
}

func (rw *ResponseWriter) WriteInternalServerErrorResponse(ctx context.Context, w http.ResponseWriter, error error) {
	_, span := rw.tracer.Start(ctx, "writing response")
	span.SetAttributes(attribute.Int("status_code", http.StatusInternalServerError))
	defer span.End()

	w.Header().Set("Content-Type", "application/json")

	response := response{
		Status:  500,
		Suceess: false,
		Message: "Internal Server Error",
	}

	jsonBytes, _ := json.Marshal(response)

	span.SetAttributes(attribute.String("data", string(jsonBytes)))
	span.SetAttributes(attribute.String("error.message", error.Error()))

	w.WriteHeader(500)
	w.Write(jsonBytes)
}

func (rw *ResponseWriter) WriteValidationErrorResponse(ctx context.Context, w http.ResponseWriter, err validationError) {
	rw.WriteJSONResponse(ctx, w, http.StatusBadRequest, "validation error", err.errors)
}

func (rw *ResponseWriter) WriteBadRequestResponse(ctx context.Context, w http.ResponseWriter) {
	rw.WriteErrorResponse(ctx, w, http.StatusBadRequest, "bad request")
}

func (rw *ResponseWriter) WriteSuccessResponse(ctx context.Context, w http.ResponseWriter, data any) {
	rw.WriteJSONResponse(ctx, w, http.StatusOK, "success", data)
}

func (rw *ResponseWriter) WriteErrorResponse(ctx context.Context, w http.ResponseWriter, statusCode int, message string) {
	rw.WriteJSONResponse(ctx, w, statusCode, message, nil)
}

func (rw *ResponseWriter) WriteJSONResponse(ctx context.Context, w http.ResponseWriter, statusCode int, message string, data any) {
	_, span := rw.tracer.Start(ctx, "writing response")
	span.SetAttributes(attribute.Int("status_code", statusCode))
	defer span.End()

	w.Header().Set("Content-Type", "application/json")

	response := response{
		Status:  statusCode,
		Suceess: statusCode >= 200 && statusCode < 300,
		Message: message,
		Data:    data,
	}

	jsonBytes, err := json.Marshal(response)
	if err != nil {
		rw.WriteInternalServerErrorResponse(ctx, w, err)
		return
	}

	span.SetAttributes(attribute.String("data", string(jsonBytes)))

	w.WriteHeader(statusCode)
	w.Write(jsonBytes)
}
