package rest_easy

import (
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"
)

// Router represents a group of routes with shared prefix and middlewares
type Router struct {
	parentMux    *http.ServeMux
	Mux          *http.ServeMux
	prefix       string
	parentPrefix string
	middlewares  []func(http.Handler) http.Handler
}

func (r *Router) fullPath(path string) string {
	// If we have a parent prefix, use it, otherwise just use our prefix
	if r.parentPrefix != "" {
		return fmt.Sprintf("%s%s%s", r.parentPrefix, r.prefix, path)
	}
	return fmt.Sprintf("%s%s", r.prefix, path)
}

// HandleFunc registers a new route with the given pattern and handler
func (r *Router) HandleFunc(pattern string, handler http.HandlerFunc) {
	// Extract HTTP method and path from the pattern
	parts := strings.SplitN(pattern, " ", 2)
	if len(parts) != 2 {
		// TODO I'm not sure if using panic is how I want to handle this
		panic(fmt.Sprintf("invalid pattern: %s", pattern))
	}
	method, path := parts[0], parts[1]

	// Create the full path with group prefix
	fullPath := r.fullPath(path)

	// Register the handler with all middlewares applied
	finalHandler := http.Handler(handler)
	for _, middleware := range r.middlewares {
		finalHandler = middleware(finalHandler)
	}

	// Register with the router using ServeMux.HandleFunc pattern for Go 1.22+
	finalPattern := fmt.Sprintf("%s %s", method, fullPath)
	slog.Debug("Registering route", "pattern", finalPattern)
	r.Mux.Handle(finalPattern, finalHandler)
}

// HandleFunc registers a new route with the given pattern and handler
func (r *Router) Handle(pattern string, handler http.Handler) {
	// Extract HTTP method and path from the pattern
	parts := strings.SplitN(pattern, " ", 2)
	if len(parts) != 2 {
		// TODO I'm not sure if using panic is how I want to handle this
		panic(fmt.Sprintf("invalid pattern: %s", pattern))
	}
	method, path := parts[0], parts[1]

	// Create the full path with group prefix
	fullPath := r.fullPath(path)

	// Register the handler with all middlewares applied
	mdws := r.middlewares
	slices.Reverse(mdws)
	for _, middleware := range mdws {
		handler = middleware(handler)
	}

	// Register with the router using ServeMux.HandleFunc pattern for Go 1.22+
	finalPattern := fmt.Sprintf("%s %s", method, fullPath)
	slog.Debug("Registering route", "pattern", finalPattern)
	r.Mux.Handle(finalPattern, handler)
}

// Use adds a middleware to the route group
func (r *Router) Use(middleware func(http.Handler) http.Handler) {
	r.middlewares = append(r.middlewares, middleware)
}
