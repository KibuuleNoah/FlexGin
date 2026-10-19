package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"flexgin/internal/apperror"
)

// Envelope is the standard response envelope.
type Envelope[T any] struct {
	Data  T          `json:"data,omitempty"`
	Meta  *Meta      `json:"meta,omitempty"`
	Error *ErrorBody `json:"error,omitempty"`
}

// EnvelopD is for documentation purposes only by swagger.
type EnvelopD Envelope[any]

// ErrorBody is the standard error envelope.
type ErrorBody struct {
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Fields    map[string]string `json:"fields,omitempty"`
	RequestID string            `json:"request_id,omitempty"`
}

// Meta is the standard pagination envelope.
type Meta struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
	Total   int `json:"total"`
}

// OK renders the standard envelope.
func OK[T any](c *gin.Context, data T) {
	c.JSON(http.StatusOK, Envelope[T]{Data: data})
}

func Created[T any](c *gin.Context, data T) {
	c.JSON(http.StatusCreated, Envelope[T]{Data: data})
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func List[T any](c *gin.Context, data []T, meta Meta) {
	c.JSON(http.StatusOK, Envelope[[]T]{Data: data, Meta: &meta})
}

// WriteError renders any error as the standard envelope.
// Called by the error middleware; handlers just do c.Error(err) and return.
func WriteError(c *gin.Context, err error) {
	ae := apperror.From(err)
	c.AbortWithStatusJSON(ae.Kind.HTTPStatus(), Envelope[any]{
		Error: &ErrorBody{
			Code:      ae.Code,
			Message:   ae.Message,
			Fields:    ae.Fields,
			RequestID: c.GetString(RequestIDKey),
		},
	})
}

// RequestIDKey is the gin context key the request-ID middleware writes to.
const RequestIDKey = "request_id"
