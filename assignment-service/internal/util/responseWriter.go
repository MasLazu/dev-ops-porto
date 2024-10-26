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
	rw.WriteErrorResponse(ctx, w, http.StatusInternalServerError, "internal server error")
}

func (rw *ResponseWriter) WriteUnauthorizedResponse(ctx context.Context, w http.ResponseWriter) {
	rw.WriteErrorResponse(ctx, w, http.StatusUnauthorized, "unauthorized")
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
		response := `{"status": 500, "success": false, "message": "internal server error"}`
		w.WriteHeader(http.StatusInternalServerError)
		span.SetAttributes(attribute.Int("status_code", http.StatusInternalServerError))
		span.SetAttributes(attribute.String("error.message", err.Error()))
		span.SetAttributes(attribute.String("data", response))
		w.Write([]byte(response))
		return
	}

	span.SetAttributes(attribute.Int("status_code", statusCode))
	span.SetAttributes(attribute.String("data", string(jsonBytes)))

	w.WriteHeader(statusCode)
	w.Write(jsonBytes)
}
