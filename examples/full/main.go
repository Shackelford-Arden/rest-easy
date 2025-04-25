package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	rest "github.com/Shackelford-Arden/rest-easy"
	"github.com/Shackelford-Arden/rest-easy/middlewares"
)

// setupRoutes configures all routes and middleware for the server
func BuildRoutes(s *rest.Server) {
	// v1 group
	v1Group := s.Group("/v1", nil)
	v1Group.Use(middlewares.AccessLogMiddleware)
	v1Group.Handle("GET /hello", rest.NewHandler(helloHandler))
	v1Group.Handle("GET /users", rest.NewHandler(getUsersHandler))

	// v2 group
	v2Group := s.Group("/v2", nil)
	v2Group.Use(middlewares.AccessLogMiddleware)
	v2Group.Handle("GET /hello", rest.NewHandler(helloV2Handler))
	v2Group.Handle("GET /users", rest.NewHandler(getUsersV2Handler))

	internalGroup := s.Group("/internal", nil)
	internalGroup.Use(middlewares.AccessLogMiddleware)
	internalGroup.HandleFunc("GET /metrics", metricsHandler)
	internalGroup.HandleFunc("GET /status", statusHandler)

	// showcase nested groups
	internalAdminGroup := s.Group("/admin", internalGroup)
	internalAdminGroup.Handle("GET /settings", rest.NewHandler(settingsHandler))
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
func settingsHandler(context.Context, rest.FuncReq) (AdminSettings, error) {
	return AdminSettings{Message: "Settings!"}, nil
}

// Request handlers
func helloHandler(context.Context, rest.FuncReq) (Hello, error) {
	return Hello{Message: "Hello from API v1"}, nil
}

func getUsersHandler(context.Context, rest.FuncReq) ([]User, error) {
	users := []User{
		{ID: 1, Name: "Alice", Version: "v1"},
		{ID: 2, Name: "Bob", Version: "v1"},
	}
	return users, nil
}

func helloV2Handler(context.Context, rest.FuncReq) (Hello, error) {
	return Hello{Message: "Hello from API v2"}, nil
}

func getUsersV2Handler(context.Context, rest.FuncReq) ([]UserV2, error) {
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

	server := rest.NewServer(
		rest.WithLogger(handler),
	)
	BuildRoutes(server)

	runErr := server.Run(context.Background())
	if runErr != nil {
		panic(runErr)
	}
}
