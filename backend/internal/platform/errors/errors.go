package apperrors

import (
	stderrors "errors"
	"fmt"
	"net/http"
	"runtime"
)

type AppError struct {
	Code       string
	Message    string
	HTTPStatus int
	Cause      error
	stack      []uintptr
}

func New(code, message string, httpStatus int, cause error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Cause:      cause,
		stack:      callers(),
	}
}

func NewAppError(code, message string, httpStatus int) *AppError {
	return New(code, message, httpStatus, nil)
}

func (e *AppError) WithCause(cause error) *AppError {
	if e == nil {
		return nil
	}
	e.Cause = cause
	return e
}

func (e *AppError) Error() string {
	if e == nil {
		return "internal error"
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func (e *AppError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprint(s, e.Message)
			if e.Cause != nil {
				fmt.Fprintf(s, ": %+v", e.Cause)
			}
			for _, pc := range e.stack {
				fn := runtime.FuncForPC(pc)
				if fn != nil {
					file, line := fn.FileLine(pc)
					fmt.Fprintf(s, "\n\t%s:%d", file, line)
				}
			}
			return
		}
		fallthrough
	case 's':
		fmt.Fprint(s, e.Message)
	case 'q':
		fmt.Fprintf(s, "%q", e.Message)
	}
}

func Status(err error) int {
	var appErr *AppError
	if stderrors.As(err, &appErr) {
		return appErr.HTTPStatus
	}
	return http.StatusInternalServerError
}

func Code(err error) string {
	var appErr *AppError
	if stderrors.As(err, &appErr) {
		return appErr.Code
	}
	return CodeInternal
}

func SafeMessage(err error) string {
	var appErr *AppError
	if stderrors.As(err, &appErr) {
		return appErr.Message
	}
	return "internal error"
}

func StackTrace(err error) []uintptr {
	var appErr *AppError
	if stderrors.As(err, &appErr) {
		return appErr.stack
	}
	return nil
}

func callers() []uintptr {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	return pcs[:n]
}
