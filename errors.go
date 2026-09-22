package spine

import (
	"errors"
	"net/http"
)

// HTTPError is a handler error that Spine turns into a JSON response
// with the corresponding HTTP status.
type HTTPError struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *HTTPError) Error() string {
	return e.Message
}

// BadRequest returns a 400 bad_request error.
func BadRequest(message string) *HTTPError {
	return &HTTPError{
		Status:  http.StatusBadRequest,
		Code:    "bad_request",
		Message: message,
	}
}

// Unauthorized returns a 401 unauthorized error.
func Unauthorized(message string) *HTTPError {
	return &HTTPError{
		Status:  http.StatusUnauthorized,
		Code:    "unauthorized",
		Message: message,
	}
}

// Forbidden returns a 403 forbidden error.
func Forbidden(message string) *HTTPError {
	return &HTTPError{
		Status:  http.StatusForbidden,
		Code:    "forbidden",
		Message: message,
	}
}

// NotFound returns a 404 not_found error.
func NotFound(message string) *HTTPError {
	return &HTTPError{
		Status:  http.StatusNotFound,
		Code:    "not_found",
		Message: message,
	}
}

// ValidationError returns a 422 validation_error error.
func ValidationError(message string) *HTTPError {
	return &HTTPError{
		Status:  http.StatusUnprocessableEntity,
		Code:    "validation_error",
		Message: message,
	}
}

// Internal returns a 500 internal_error error.
func Internal(message string) *HTTPError {
	return &HTTPError{
		Status:  http.StatusInternalServerError,
		Code:    "internal_error",
		Message: message,
	}
}

func writeError(c *Context, err error) error {
	if err == nil {
		return nil
	}

	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return c.JSON(httpErr.Status, httpErr)
	}

	return c.JSON(http.StatusInternalServerError, map[string]string{
		"code":    "internal_error",
		"message": err.Error(),
	})
}
