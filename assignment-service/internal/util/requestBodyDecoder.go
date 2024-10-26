package util

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type RequestBodyDecoder struct {
	tracer trace.Tracer
}

func NewRequestBodyDecoder(tracer trace.Tracer) *RequestBodyDecoder {
	return &RequestBodyDecoder{tracer: tracer}
}

func (rbd *RequestBodyDecoder) Decode(ctx context.Context, r *http.Request, v interface{}) error {
	_, span := rbd.tracer.Start(ctx, "decoding request body")
	defer span.End()

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	span.SetAttributes(attribute.String("data", string(bodyBytes)))

	err = json.Unmarshal(bodyBytes, &v)
	if err != nil {
		return err
	}

	return nil
}
