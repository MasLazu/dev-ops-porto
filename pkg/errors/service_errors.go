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
	HttpCode() int
	GrpcCode() codes.Code
}

type serviceError struct {
	code          ServiceErrorCode
	err           error
	clientMessage *string
}

func NewServiceErrorWithClientMessage(code code.Code, err error, clientMessage string) ServiceError {
	return &serviceError{ServiceErrorCode(code), err, &clientMessage}
}

func NewServiceError(code code.Code, err error) ServiceError {
	return &serviceError{ServiceErrorCode(code), err, nil}
}

func (e *serviceError) ClientMessage() string {
	if e.clientMessage == nil {
		return http.StatusText(int(e.code.toHttpCode()))
	}
	return *e.clientMessage
}

func (e *serviceError) Code() uint32 {
	return uint32(e.code)
}

func (e *serviceError) Error() string {
	if e.err != nil {
		return e.err.Error()
	}

	return http.StatusText(int(e.code.toHttpCode()))
}

func (e *serviceError) HttpCode() int {
	return int(e.code.toHttpCode())
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
