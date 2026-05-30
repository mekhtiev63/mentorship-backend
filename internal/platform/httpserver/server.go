package httpserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-mentorship-platform/backend/internal/platform/config"
)

// Server wraps http.Server with graceful shutdown.
type Server struct {
	cfg    config.HTTP
	log    *slog.Logger
	server *http.Server
}

// New creates an HTTP server with the given handler.
func New(cfg config.HTTP, log *slog.Logger, handler http.Handler) *Server {
	return &Server{
		cfg: cfg,
		log: log,
		server: &http.Server{
			Addr:         cfg.Addr(),
			Handler:      handler,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
	}
}

// Start listens until the context is cancelled or the server fails.
func (s *Server) Start(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		s.log.Info("http server listening", "addr", s.cfg.Addr())
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		return s.Shutdown()
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("http server: %w", err)
		}
		return nil
	}
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
	defer cancel()

	s.log.Info("http server shutting down")
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("http shutdown: %w", err)
	}
	return nil
}

// Addr returns the configured listen address.
func (s *Server) Addr() string {
	return s.cfg.Addr()
}

// SetShutdownTimeout overrides shutdown timeout (used in tests).
func (s *Server) SetShutdownTimeout(d time.Duration) {
	s.cfg.ShutdownTimeout = d
}
