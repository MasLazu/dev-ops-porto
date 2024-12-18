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
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func NewResponseWriter(tracer trace.Tracer) *ResponseWriter {
	return &ResponseWriter{tracer: tracer}
}

func (rw *ResponseWriter) WriteNotFoundResponse(ctx context.Context, w http.ResponseWriter) {
	code := http.StatusNotFound
	rw.WriteErrorResponse(ctx, w, code, http.StatusText(code))
}

func (rw *ResponseWriter) WriteUnauthorizedResponse(ctx context.Context, w http.ResponseWriter) {
	code := http.StatusUnauthorized
	rw.WriteErrorResponse(ctx, w, code, http.StatusText(code))
}

func (rw *ResponseWriter) WriteValidationErrorResponse(ctx context.Context, w http.ResponseWriter, err validationError) {
	rw.WriteJSONResponse(ctx, w, http.StatusUnprocessableEntity, "Validation Error", err.errors)
}

func (rw *ResponseWriter) WriteBadRequestResponse(ctx context.Context, w http.ResponseWriter) {
	code := http.StatusBadRequest
	rw.WriteErrorResponse(ctx, w, code, http.StatusText(code))
}

func (rw *ResponseWriter) WriteSuccessResponse(ctx context.Context, w http.ResponseWriter, data any) {
	code := http.StatusOK
	rw.WriteJSONResponse(ctx, w, code, http.StatusText(code), data)
}

func (rw *ResponseWriter) WriteErrorResponse(ctx context.Context, w http.ResponseWriter, statusCode int, message string) {
	rw.WriteJSONResponse(ctx, w, statusCode, message, nil)
}

func (rw *ResponseWriter) WriteJSONResponse(ctx context.Context, w http.ResponseWriter, statusCode int, message string, data any) {
	rw.writeJSONResponse(ctx, w, statusCode, message, data)
}

func (rw *ResponseWriter) WriteJSONResponseWithInternalError(ctx context.Context, w http.ResponseWriter, statusCode int, message string, data any, err error) {
	ctx, span := rw.tracer.Start(ctx, "writing response with internal error")
	defer span.End()

	span.SetAttributes(attribute.String("internalError.message", err.Error()))
	rw.writeJSONResponse(ctx, w, statusCode, message, data)
}

func (rw *ResponseWriter) writeJSONResponse(ctx context.Context, w http.ResponseWriter, statusCode int, message string, data any) {
	ctx, span := rw.tracer.Start(ctx, "writing response")
	defer span.End()

	response := response{
		Status:  statusCode,
		Success: statusCode >= 200 && statusCode < 300,
		Message: message,
		Data:    data,
	}

	jsonBytes, err := json.Marshal(response)
	if err != nil {
		rw.WriteInternalServerErrorResponse(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	span.SetAttributes(attribute.Int("status_code", statusCode))
	span.SetAttributes(attribute.String("data", string(jsonBytes)))

	w.WriteHeader(statusCode)
	w.Write(jsonBytes)
}

func (rw *ResponseWriter) WriteInternalServerErrorResponse(ctx context.Context, w http.ResponseWriter, err error) {
	_, span := rw.tracer.Start(ctx, "writing internal server error response")
	defer span.End()

	w.Header().Set("Content-Type", "application/json")

	response := `{"status": 500, "success": false, "message": "internal server error"}`
	span.SetAttributes(attribute.String("internalError.message", err.Error()))

	span.SetAttributes(attribute.Int("status_code", http.StatusInternalServerError))
	span.SetAttributes(attribute.String("data", response))
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(response))
}
