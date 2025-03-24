package middlewares

import (
	"log/slog"
	"net/http"

	"github.com/Shackelford-Arden/rest-easy/pkg/api"
)

func AccessLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Wrap the response writer to capture the status code
		wrapped := api.NewResponseWriter(w)

		// Process the request
		next.ServeHTTP(wrapped, r)

		// Log request details
		// TODO Consider using OTEL semantic attribute naming, though maybe in another middleware?
		slog.InfoContext(r.Context(), "request completed",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", wrapped.StatusCode()),
			slog.String("ip", r.RemoteAddr),
			slog.String("user-agent", r.UserAgent()),
		)
	})
}
