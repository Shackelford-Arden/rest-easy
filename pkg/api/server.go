package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	logger          *slog.Logger
	port            int
	shutdownTimeout time.Duration
	RootMux         *http.ServeMux
	server          *http.Server
	routers         []Router
}

// NewServer creates a new parentMux instance
func NewServer(options ...ServerOption) *Server {
	server := &Server{
		RootMux:         http.NewServeMux(),
		port:            8080,
		shutdownTimeout: 15 * time.Second,
		server: &http.Server{
			Addr: fmt.Sprintf(":%d", 8080),
		},
	}

	for _, opt := range options {
		opt(server)
	}

	// Set our RootMux as the server's handler
	server.server.Handler = server.RootMux

	return server
}

// Group creates a new route group with the given prefix
func (s *Server) Group(prefix string, parent *Router) *Router {
	// Create a new mux for this group
	groupMux := http.NewServeMux()

	// If no parent Mux is passed in, assume we're
	// mounting it to the root Mux
	mountMux := s.RootMux
	if parent != nil {
		mountMux = parent.Mux
	}

	// Calculate the full prefix path
	rPrefix := prefix
	parentPrefix := ""
	if parent != nil {
		parentPrefix = parent.prefix
		rPrefix = fmt.Sprintf("%s/%s", parent.prefix, rPrefix)
	}

	rtr := Router{
		parentMux:    mountMux,
		Mux:          groupMux,
		prefix:       prefix,
		parentPrefix: parentPrefix,
	}

	s.routers = append(s.routers, rtr)

	return &rtr
}

type ServerOption func(*Server)

// WithLogger sets up the root slog handler with the given handler
func WithLogger(handler slog.Handler) ServerOption {
	return func(r *Server) {
		// Set it as the default handler
		slog.SetDefault(slog.New(handler))

		// Store the logger for server-specific use if needed
		r.logger = slog.Default()
	}
}

// WithPort sets the port the parentMux listens on
func WithPort(port int) ServerOption {
	return func(r *Server) {
		r.server.Addr = fmt.Sprintf(":%d", port)
	}
}

func WithShutdownTimeout(timeout time.Duration) ServerOption {
	return func(r *Server) {
		r.shutdownTimeout = timeout
	}
}

func (s *Server) mount() error {
	// Go through the groups/routers and mount
	// them to the root mux
	for _, r := range s.routers {
		// Calculate the full prefix path
		fullPrefix := r.prefix
		if r.parentPrefix != "" {
			fullPrefix = fmt.Sprintf("%s%s", r.parentPrefix, r.prefix)
		}
		fullPrefix = fullPrefix + "/"
		s.RootMux.Handle(fullPrefix, r.Mux)
	}

	return nil
}

// Run hosts the parentMux and handles a graceful shutdown.
func (s *Server) Run(ctx context.Context) error {

	// Initialize Logging w/ defined logger

	mountErr := s.mount()
	if mountErr != nil {
		return mountErr
	}

	// watch for shutdown signals
	go func() {
		slog.InfoContext(ctx, "Starting server", "addr", s.server.Addr)
		if shutDownErr := s.server.ListenAndServe(); shutDownErr != nil && !errors.Is(shutDownErr, http.ErrServerClosed) {
			slog.ErrorContext(ctx, "starting http parentMux ran into an error", "error", shutDownErr)
		}

		slog.InfoContext(ctx, "shutting down, no longer accepting new connections")
	}()

	servChan := make(chan os.Signal, 1)
	signal.Notify(servChan, syscall.SIGINT, syscall.SIGTERM)
	<-servChan

	// set default shutdownTimeout of 10 seconds
	shutdownTimeout := s.shutdownTimeout
	if s.shutdownTimeout == 0 {
		shutdownTimeout = time.Second * 10
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("failed to shutdown parentMux: %w", err)
	}

	slog.InfoContext(ctx, "shutdown complete")

	return nil

}
