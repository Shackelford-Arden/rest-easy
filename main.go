package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/Shackelford-Arden/rest-easy/pkg/api"
	"github.com/Shackelford-Arden/rest-easy/pkg/middlewares"
)

// setupRoutes configures all routes and middleware for the server
func BuildRoutes(s *api.Server) {
	// v1 group
	v1Group := s.Group("/v1", nil)
	v1Group.Use(middlewares.AccessLogMiddleware)
	v1Group.Handle("GET /hello", NewHandler(helloHandler))
	v1Group.Handle("GET /users", NewHandler(getUsersHandler))

	// v2 group
	v2Group := s.Group("/v2", nil)
	v2Group.Use(middlewares.AccessLogMiddleware)
	v2Group.Handle("GET /hello", NewHandler(helloV2Handler))
	v2Group.Handle("GET /users", NewHandler(getUsersV2Handler))

	// internal group with auth middleware
	internalGroup := s.Group("/internal", nil)
	internalGroup.Use(middlewares.AccessLogMiddleware)
	internalGroup.HandleFunc("GET /metrics", metricsHandler)
	internalGroup.HandleFunc("GET /status", statusHandler)
	// showcase nested groups
	internalAdminGroup := s.Group("/admin", internalGroup)
	internalAdminGroup.Handle("GET /settings", NewHandler(settingsHandler))
}

type FuncReq struct {
	Body    any
	Request *http.Request
}

type CustomHandlerFunc[In any, Out any] func(ctx context.Context, in FuncReq) (Out, error)

// Man I really don't like the way Generics look
func NewHandler[In FuncReq, Out any](f CustomHandlerFunc[In, Out]) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var funcIn FuncReq

		funcIn.Request = r

		// TODO: Consider parsing headers, query params, etc and adding them to `FuncReq`
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			err := json.NewDecoder(r.Body).Decode(&funcIn.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}

		out, err := f(r.Context(), funcIn)
		// TODO: I want to consider options for returning different types of errors
		// instead of always assuming http.StatusInternalServerError
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// TODO What if there is no body? Like on a DELETE, the status code may all that's needed
		// Guess this is one of the flexibility challenges of this pattern of using
		// a single "handler" with generics
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(out)
	})
}

type Hello struct {
	Message string `json:"message"`
}

type User struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

type UserV2 struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Version string `json:"version"`
}

type AdminSettings struct {
	Message string `json:"message"`
}

// AdminSettings handler
func settingsHandler(context.Context, FuncReq) (AdminSettings, error) {
	return AdminSettings{Message: "Settings!"}, nil
}

// Request handlers
func helloHandler(context.Context, FuncReq) (Hello, error) {
	return Hello{Message: "Hello from API v1"}, nil
}

func getUsersHandler(context.Context, FuncReq) ([]User, error) {
	users := []User{
		{ID: 1, Name: "Alice", Version: "v1"},
		{ID: 2, Name: "Bob", Version: "v1"},
	}
	return users, nil
}

func helloV2Handler(context.Context, FuncReq) (Hello, error) {
	return Hello{Message: "Hello from API v2"}, nil
}

func getUsersV2Handler(context.Context, FuncReq) ([]UserV2, error) {
	users := []UserV2{
		{ID: 1, Name: "Alice", Email: "alice@example.com", Version: "v2"},
		{ID: 2, Name: "Bob", Email: "bob@example.com", Version: "v2"},
	}

	return users, nil

}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by auth middleware)
	user := r.Context().Value("user").(string)

	metrics := map[string]interface{}{
		"uptime":     "24h",
		"requests":   12345,
		"errors":     123,
		"accessedBy": user,
	}
	json.NewEncoder(w).Encode(metrics)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	status := map[string]string{
		"status":      "healthy",
		"version":     "1.0.3",
		"environment": "production",
	}
	json.NewEncoder(w).Encode(status)
}

func main() {
	// Configure logging with text handler (human-readable)
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	server := api.NewServer(
		api.WithLogger(handler),
	)
	BuildRoutes(server)

	runErr := server.Run(context.Background())
	if runErr != nil {
		panic(runErr)
	}
}
