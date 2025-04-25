package rest_easy

import (
	"net/http"
)

// CustomResponseWriter wraps http.ResponseWriter to capture status code
type CustomResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// NewResponseWriter creates a new customResponseWriter
func NewResponseWriter(w http.ResponseWriter) *CustomResponseWriter {
	return &CustomResponseWriter{w, http.StatusOK}
}

// WriteHeader captures the status code
func (rw *CustomResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// StatusCode returns the captured status code
func (rw *CustomResponseWriter) StatusCode() int {
	return rw.statusCode
}
