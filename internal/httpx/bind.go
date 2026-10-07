package httpx

import (
	"errors"
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"flexgin/internal/apperror"
)

// UseJSONFieldNames makes validation errors report json tag names instead of Go field names.
// Call once at startup, before serving requests.
func UseJSONFieldNames() {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}
	v.RegisterTagNameFunc(func(f reflect.StructField) string {
		name := strings.SplitN(f.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return f.Name
		}
		return name
	})
}

// BindJSON decodes and validates the request body into dst.
func BindJSON(c *gin.Context, dst any) error {
	return toAppError(c.ShouldBindJSON(dst), "invalid request body")
}

// BindQuery decodes and validates query params into dst.
func BindQuery(c *gin.Context, dst any) error {
	return toAppError(c.ShouldBindQuery(dst), "invalid query parameters")
}

// ParamInt64 parses a positive int64 path param.
func ParamInt64(c *gin.Context, name string) (int64, error) {
	n, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || n < 1 {
		return 0, apperror.Validation("invalid path parameter", map[string]string{name: "must be a positive integer"})
	}
	return n, nil
}

func toAppError(err error, msg string) error {
	if err == nil {
		return nil
	}
	var verrs validator.ValidationErrors
	if errors.As(err, &verrs) {
		fields := make(map[string]string, len(verrs))
		for _, fe := range verrs {
			fields[fe.Field()] = describe(fe)
		}
		return apperror.Validation("validation failed", fields)
	}
	return apperror.Validation(msg, nil)
}

func describe(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "min":
		return "must be at least " + fe.Param()
	case "max":
		return "must be at most " + fe.Param()
	case "email":
		return "must be a valid email"
	case "oneof":
		return "must be one of: " + fe.Param()
	default:
		return "is invalid (" + fe.Tag() + ")"
	}
}
