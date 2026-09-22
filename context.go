package spine

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Context wraps the request and response. Handlers should not write to
// Writer directly — use the methods on Context or return an *HTTPError.
type Context struct {
	Writer    http.ResponseWriter
	Request   *http.Request
	Logger    *slog.Logger
	RequestID string
}

// Context returns the request's context.Context for use with services and DB calls.
func (c *Context) Context() context.Context {
	return c.Request.Context()
}

// Param returns a path parameter by name (chi syntax, e.g. {id}).
func (c *Context) Param(key string) string {
	return chi.URLParam(c.Request, key)
}

// Query returns a query string value by key.
func (c *Context) Query(key string) string {
	return c.Request.URL.Query().Get(key)
}

// BindJSON decodes the request body into dst.
// Unknown JSON fields result in a 400 BadRequest *HTTPError.
func (c *Context) BindJSON(dst any) error {
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return BadRequest(err.Error())
	}

	return nil
}

// JSON writes a JSON response with the given status code.
func (c *Context) JSON(status int, data any) error {
	c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	c.Writer.WriteHeader(status)

	if err := json.NewEncoder(c.Writer).Encode(data); err != nil {
		return err
	}

	return nil
}

// Text writes a plain-text response with the given status code.
func (c *Context) Text(status int, value string) error {
	c.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.Writer.WriteHeader(status)

	if _, err := c.Writer.Write([]byte(value)); err != nil {
		return err
	}

	return nil
}

// OK writes a 200 JSON response.
func (c *Context) OK(data any) error {
	return c.JSON(http.StatusOK, data)
}

// Created writes a 201 JSON response.
func (c *Context) Created(data any) error {
	return c.JSON(http.StatusCreated, data)
}

// NoContent writes a response with the given status and no body
// (typically 204).
func (c *Context) NoContent(status int) error {
	c.Writer.WriteHeader(status)
	return nil
}
