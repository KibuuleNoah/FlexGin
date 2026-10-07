package apperror

import (
	"errors"
	"net/http"
)

type Kind int

const (
	KindInternal Kind = iota
	KindNotFound
	KindConflict
	KindValidation
	KindUnauthorized
	KindForbidden
)

func (k Kind) HTTPStatus() int {
	switch k {
	case KindNotFound:
		return http.StatusNotFound
	case KindConflict:
		return http.StatusConflict
	case KindValidation:
		return http.StatusUnprocessableEntity
	case KindUnauthorized:
		return http.StatusUnauthorized
	case KindForbidden:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

// Error is the only error type handlers and middleware need to understand.
// Err is the wrapped cause: logged, never sent to the client.
type Error struct {
	Kind    Kind
	Code    string            // stable machine-readable code, e.g. "article_not_found"
	Message string            // safe to show to clients
	Fields  map[string]string // field -> problem, for validation errors
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return e.Code + ": " + e.Message + ": " + e.Err.Error()
	}
	return e.Code + ": " + e.Message
}

func (e *Error) Unwrap() error { return e.Err }

func NotFound(code, msg string) *Error {
	return &Error{Kind: KindNotFound, Code: code, Message: msg}
}

func Conflict(code, msg string) *Error {
	return &Error{Kind: KindConflict, Code: code, Message: msg}
}

func Validation(msg string, fields map[string]string) *Error {
	return &Error{Kind: KindValidation, Code: "validation_failed", Message: msg, Fields: fields}
}

func Unauthorized(msg string) *Error {
	return &Error{Kind: KindUnauthorized, Code: "unauthorized", Message: msg}
}

func Forbidden(msg string) *Error {
	return &Error{Kind: KindForbidden, Code: "forbidden", Message: msg}
}

// Internal wraps an unexpected error. The cause is logged but never exposed.
func Internal(err error) *Error {
	return &Error{Kind: KindInternal, Code: "internal_error", Message: "internal server error", Err: err}
}

// From converts any error into *Error. Unknown errors become Internal.
func From(err error) *Error {
	var ae *Error
	if errors.As(err, &ae) {
		return ae
	}
	return Internal(err)
}
