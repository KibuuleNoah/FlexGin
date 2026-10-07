package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"flex/internal/apperror" // replace "yourmodule" with the module path from go.mod
)

type Envelope struct {
	Data  any        `json:"data,omitempty"`
	Meta  *Meta      `json:"meta,omitempty"`
	Error *ErrorBody `json:"error,omitempty"`
}

type ErrorBody struct {
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Fields    map[string]string `json:"fields,omitempty"`
	RequestID string            `json:"request_id,omitempty"`
}

type Meta struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
	Total   int `json:"total"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Envelope{Data: data})
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, Envelope{Data: data})
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func List(c *gin.Context, data any, meta Meta) {
	c.JSON(http.StatusOK, Envelope{Data: data, Meta: &meta})
}

// WriteError renders any error as the standard envelope.
// Called by the error middleware; handlers just do c.Error(err) and return.
func WriteError(c *gin.Context, err error) {
	ae := apperror.From(err)
	c.AbortWithStatusJSON(ae.Kind.HTTPStatus(), Envelope{
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
