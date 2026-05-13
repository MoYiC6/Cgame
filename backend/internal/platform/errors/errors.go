package apperrors

import (
	stderrors "errors"
	"net/http"
)

type AppError struct {
	Code       string
	Message    string
	HTTPStatus int
	Cause      error
}

func NewAppError(code, message string, httpStatus int) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: httpStatus}
}

func Wrap(base *AppError, cause error) *AppError {
	if base == nil {
		return &AppError{Code: "INTERNAL_ERROR", Message: "internal error", HTTPStatus: http.StatusInternalServerError, Cause: cause}
	}
	clone := *base
	clone.Cause = cause
	return &clone
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
	return "INTERNAL_ERROR"
}

func SafeMessage(err error) string {
	var appErr *AppError
	if stderrors.As(err, &appErr) {
		return appErr.Message
	}
	return "internal error"
}
