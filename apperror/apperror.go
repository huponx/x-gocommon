package apperror

import (
	"errors"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Error struct {
	Code    codes.Code
	Message string
	err     error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	if e.err != nil {
		return e.err.Error()
	}
	return e.Code.String()
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.err
}

func (e *Error) GRPCStatus() *status.Status {
	if e == nil {
		return status.New(codes.OK, "")
	}
	return status.New(e.Code, e.Error())
}

func (e *Error) CodeName() string {
	if e == nil {
		return codes.OK.String()
	}
	return e.Code.String()
}

func (e *Error) HTTPStatus() int {
	if e == nil {
		return http.StatusOK
	}
	return httpStatus(e.Code)
}

func newError(code codes.Code, msg string) *Error {
	return &Error{Code: code, Message: msg}
}

func InvalidArgument(msg string) *Error    { return newError(codes.InvalidArgument, msg) }
func NotFound(msg string) *Error           { return newError(codes.NotFound, msg) }
func Unauthenticated(msg string) *Error    { return newError(codes.Unauthenticated, msg) }
func PermissionDenied(msg string) *Error   { return newError(codes.PermissionDenied, msg) }
func AlreadyExists(msg string) *Error      { return newError(codes.AlreadyExists, msg) }
func FailedPrecondition(msg string) *Error { return newError(codes.FailedPrecondition, msg) }
func Unavailable(msg string) *Error        { return newError(codes.Unavailable, msg) }
func Internal(msg string) *Error           { return newError(codes.Internal, msg) }

func Wrap(code codes.Code, msg string, err error) *Error {
	return &Error{Code: code, Message: msg, err: err}
}

func From(err error) *Error {
	if err == nil {
		return nil
	}
	var app *Error
	if errors.As(err, &app) {
		return app
	}
	st, ok := status.FromError(err)
	if !ok {
		return Wrap(codes.Internal, err.Error(), err)
	}
	return &Error{Code: st.Code(), Message: st.Message(), err: err}
}

func HTTPStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}
	return From(err).HTTPStatus()
}

func httpStatus(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.InvalidArgument, codes.OutOfRange:
		return http.StatusBadRequest
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists, codes.Aborted:
		return http.StatusConflict
	case codes.FailedPrecondition:
		return http.StatusPreconditionFailed
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Unavailable:
		return http.StatusBadGateway
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.Canceled:
		return 499
	default:
		return http.StatusInternalServerError
	}
}
