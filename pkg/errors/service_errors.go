package errors

import (
	"net/http"

	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/grpc/codes"
)

type ServiceErrorCode code.Code

type ServiceError interface {
	Error() string
	ClientMessage() string
	Code() uint32
	HttpCode() uint32
	GrpcCode() codes.Code
}

type serviceError struct {
	code          ServiceErrorCode
	err           error
	clientMessage *string
}

func NewServiceError(code ServiceErrorCode, err error, clientMessage *string) ServiceError {
	return &serviceError{code, err, clientMessage}
}

func (e *serviceError) ClientMessage() string {
	if e.clientMessage == nil {
		return e.err.Error()
	}
	return *e.clientMessage
}

func (e *serviceError) Code() uint32 {
	return uint32(e.code)
}

func (e *serviceError) Error() string {
	return e.err.Error()
}

func (e *serviceError) HttpCode() uint32 {
	return e.code.toHttpCode()
}

func (e *serviceError) GrpcCode() codes.Code {
	return codes.Code(e.code)
}

func (e ServiceErrorCode) toHttpCode() uint32 {
	switch code.Code(e) {
	case code.Code_OK:
		return http.StatusOK
	case code.Code_CANCELLED:
		return http.StatusRequestTimeout
	case code.Code_UNKNOWN:
		return http.StatusInternalServerError
	case code.Code_INVALID_ARGUMENT:
		return http.StatusBadRequest
	case code.Code_DEADLINE_EXCEEDED:
		return http.StatusGatewayTimeout
	case code.Code_NOT_FOUND:
		return http.StatusNotFound
	case code.Code_ALREADY_EXISTS:
		return http.StatusConflict
	case code.Code_PERMISSION_DENIED:
		return http.StatusForbidden
	case code.Code_RESOURCE_EXHAUSTED:
		return http.StatusTooManyRequests
	case code.Code_FAILED_PRECONDITION:
		return http.StatusPreconditionFailed
	case code.Code_ABORTED:
		return http.StatusConflict
	case code.Code_OUT_OF_RANGE:
		return http.StatusRequestEntityTooLarge
	case code.Code_UNIMPLEMENTED:
		return http.StatusNotImplemented
	case code.Code_INTERNAL:
		return http.StatusInternalServerError
	case code.Code_UNAVAILABLE:
		return http.StatusServiceUnavailable
	case code.Code_DATA_LOSS:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
