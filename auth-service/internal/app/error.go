package app

type ServiceError interface {
	Error() string
	Code() int
}

type InternalServiceError struct {
	err error
}

func (e InternalServiceError) Code() int {
	return 500
}

func (e InternalServiceError) Error() string {
	return e.err.Error()
}

func NewInternalServiceError(err error) ServiceError {
	return &InternalServiceError{err}
}

type ClientError struct {
	code int
	err  error
}

func (e ClientError) Code() int {
	return e.code
}

func (e ClientError) Error() string {
	return e.err.Error()
}

func NewClientError(code int, err error) ServiceError {
	return &ClientError{code, err}
}
