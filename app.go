// Package spine is a thin HTTP framework for Go built on chi,
// in the style of Express / Fastify / Fiber.
//
// Handlers return error; responses are written through Context methods,
// not by writing to http.ResponseWriter directly.
package spine

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Handler is an HTTP handler that receives a Context and returns an error.
// Return nil after writing a response with Context methods, or return
// an *HTTPError for a structured JSON error response.
type Handler func(*Context) error

// Middleware wraps a Handler. The request flows through middleware
// left-to-right, then into the final handler.
type Middleware func(Handler) Handler

// Router registers routes, middleware, and nested groups.
// Both *App and nested groups implement Router.
type Router interface {
	Use(middlewares ...Middleware)
	GET(path string, handler Handler)
	POST(path string, handler Handler)
	PUT(path string, handler Handler)
	PATCH(path string, handler Handler)
	DELETE(path string, handler Handler)
	Group(prefix string, fn func(Router))
}

// Config configures a new App.
type Config struct {
	// ServiceName appears in the startup log when calling Run.
	ServiceName string
	// Logger is used for request and error logs. If nil, a text slog
	// logger writing to stdout is used.
	Logger *slog.Logger
}

// App is the root router and HTTP server entry point.
type App struct {
	group       *routeGroup
	logger      *slog.Logger
	serviceName string
}

type routeGroup struct {
	router      chi.Router
	logger      *slog.Logger
	middlewares []Middleware
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// New creates an App with the given config.
func New(cfg Config) *App {
	log := cfg.Logger
	if log == nil {
		log = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}

	root := &routeGroup{
		router: chi.NewRouter(),
		logger: log,
	}

	return &App{
		group:       root,
		logger:      log,
		serviceName: cfg.ServiceName,
	}
}

// Use appends middleware to the root router.
func (a *App) Use(middlewares ...Middleware) {
	a.group.Use(middlewares...)
}

// GET registers a GET route on the root router.
func (a *App) GET(path string, handler Handler) { a.group.GET(path, handler) }

// POST registers a POST route on the root router.
func (a *App) POST(path string, handler Handler) { a.group.POST(path, handler) }

// PUT registers a PUT route on the root router.
func (a *App) PUT(path string, handler Handler) { a.group.PUT(path, handler) }

// PATCH registers a PATCH route on the root router.
func (a *App) PATCH(path string, handler Handler) { a.group.PATCH(path, handler) }

// DELETE registers a DELETE route on the root router.
func (a *App) DELETE(path string, handler Handler) { a.group.DELETE(path, handler) }

// Group mounts a nested router under prefix.
func (a *App) Group(prefix string, fn func(Router)) {
	a.group.Group(prefix, fn)
}

// Handler returns the underlying http.Handler for use with
// http.ListenAndServe or a custom server.
func (a *App) Handler() http.Handler {
	return a.group.router
}

// Run starts an HTTP server on addr. It does not perform graceful shutdown.
func (a *App) Run(addr string) error {
	a.logger.Info("starting server", "service", a.serviceName, "addr", addr)
	return http.ListenAndServe(addr, a.Handler())
}

func (g *routeGroup) Use(middlewares ...Middleware) {
	g.middlewares = append(g.middlewares, middlewares...)
}

func (g *routeGroup) GET(path string, handler Handler) {
	g.register(http.MethodGet, path, handler)
}

func (g *routeGroup) POST(path string, handler Handler) {
	g.register(http.MethodPost, path, handler)
}

func (g *routeGroup) PUT(path string, handler Handler) {
	g.register(http.MethodPut, path, handler)
}

func (g *routeGroup) PATCH(path string, handler Handler) {
	g.register(http.MethodPatch, path, handler)
}

func (g *routeGroup) DELETE(path string, handler Handler) {
	g.register(http.MethodDelete, path, handler)
}

func (g *routeGroup) Group(prefix string, fn func(Router)) {
	g.router.Route(prefix, func(r chi.Router) {
		child := &routeGroup{
			router:      r,
			logger:      g.logger,
			middlewares: append([]Middleware{}, g.middlewares...),
		}

		fn(child)
	})
}

func (g *routeGroup) register(method, path string, handler Handler) {
	finalHandler := chain(handler, g.middlewares...)

	g.router.MethodFunc(method, path, func(w http.ResponseWriter, r *http.Request) {
		wrappedWriter := &statusWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		ctx := &Context{
			Writer:  wrappedWriter,
			Request: r,
			Logger:  g.logger,
		}

		if err := finalHandler(ctx); err != nil {
			g.logger.Error("request failed", "method", r.Method, "path", r.URL.Path, "error", err)
			if writeErr := writeError(ctx, err); writeErr != nil {
				g.logger.Error("failed to write error response", "error", writeErr)
			}
		}
	})
}

func chain(handler Handler, middlewares ...Middleware) Handler {
	chained := handler

	for i := len(middlewares) - 1; i >= 0; i-- {
		chained = middlewares[i](chained)
	}

	return chained
}

// RequestID is middleware that reads X-Request-Id or generates a UUID,
// stores it on Context.RequestID, and echoes it in the response header.
func RequestID() Middleware {
	return func(next Handler) Handler {
		return func(c *Context) error {
			requestID := c.Request.Header.Get("X-Request-Id")
			if requestID == "" {
				requestID = uuid.NewString()
			}

			c.RequestID = requestID
			c.Writer.Header().Set("X-Request-Id", requestID)

			return next(c)
		}
	}
}

// Logger is middleware that logs method, path, status, duration, and request ID
// after the handler completes.
func Logger() Middleware {
	return func(next Handler) Handler {
		return func(c *Context) error {
			startedAt := time.Now()
			err := next(c)

			status := http.StatusOK
			if wrappedWriter, ok := c.Writer.(*statusWriter); ok {
				status = wrappedWriter.status
			}

			c.Logger.Info(
				"request completed",
				"request_id", c.RequestID,
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"status", status,
				"duration", time.Since(startedAt).String(),
			)

			return err
		}
	}
}

// Recover is middleware that recovers from panics, logs the stack,
// and returns an Internal HTTPError.
func Recover() Middleware {
	return func(next Handler) Handler {
		return func(c *Context) (err error) {
			defer func() {
				if rec := recover(); rec != nil {
					c.Logger.Error(
						"panic recovered",
						"request_id", c.RequestID,
						"method", c.Request.Method,
						"path", c.Request.URL.Path,
						"panic", rec,
						"stack", string(debug.Stack()),
					)

					err = Internal(fmt.Sprintf("panic: %v", rec))
				}
			}()

			return next(c)
		}
	}
}
