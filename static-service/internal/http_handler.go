package internal

import (
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/MasLazu/dev-ops-porto/pkg/util"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/trace"
)

type HttpHandler struct {
	tracer         trace.Tracer
	responseWriter *util.ResponseWriter
	handlerTracer  *util.HandlerTracer
	s3Client       *s3.Client
	bucketNames    bucketNames
}

func NewHttpHandler(
	tracer trace.Tracer,
	responseWriter *util.ResponseWriter,
	handlerTracer *util.HandlerTracer,
	s3Client *s3.Client,
	bucketNames bucketNames,
) *HttpHandler {
	return &HttpHandler{
		tracer:         tracer,
		responseWriter: responseWriter,
		handlerTracer:  handlerTracer,
		s3Client:       s3Client,
		bucketNames:    bucketNames,
	}
}

func (h *HttpHandler) setupRoutes(c *chi.Mux) http.Handler {
	c.Get("/health", h.HealthCheck)
	c.Get("/*", h.ServeStaticFile)

	return c
}

func (h *HttpHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.handlerTracer.TraceHttpHandler(r, "HttpHandler.HealthCheck")
	defer span.End()

	h.responseWriter.WriteSuccessResponse(ctx, w, "OK")
}

func (h *HttpHandler) ServeStaticFile(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.handlerTracer.TraceHttpHandler(r, "HttpHandler.ServeStaticFile")
	defer span.End()

	key := chi.URLParam(r, "*")
	key = path.Clean(key)
	key = strings.TrimPrefix(key, "/")

	result, err := h.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &h.bucketNames.profilePictures,
		Key:    &key,
	})
	if err != nil {
		http.Error(w, "File Not Found", http.StatusNotFound)
		return
	}
	defer result.Body.Close()

	if result.ContentType != nil {
		w.Header().Set("Content-Type", *result.ContentType)
	} else {
		w.Header().Set("Content-Type", "image/jpeg")
	}

	w.Header().Set("Cache-Control", "max-age=86400")
	w.Header().Set("Content-Disposition", "inline")

	if _, err := io.Copy(w, result.Body); err != nil {
		http.Error(w, "Error serving file", http.StatusInternalServerError)
		return
	}
}
