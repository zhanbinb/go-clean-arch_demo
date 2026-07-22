// Package server provides HTTP and gRPC server lifecycle helpers with
// graceful shutdown semantics.
package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// HTTPServer wraps http.Server with graceful shutdown.
type HTTPServer struct {
	srv *http.Server
	log *zap.Logger
}

// NewHTTPServer builds an HTTP server with sane timeouts.
func NewHTTPServer(addr string, handler http.Handler, log *zap.Logger) *HTTPServer {
	return &HTTPServer{
		srv: &http.Server{
			Addr:              addr,
			Handler:           handler,
			ReadHeaderTimeout: 10 * time.Second,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       120 * time.Second,
		},
		log: log,
	}
}

// Run blocks until the server stops. ctx cancellation triggers a graceful
// shutdown with a 30-second drain timeout.
func (h *HTTPServer) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		h.log.Info("http server starting", zap.String("addr", h.srv.Addr))
		if err := h.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err, ok := <-errCh:
		if !ok {
			return nil
		}
		return err
	case <-ctx.Done():
		h.log.Info("http server graceful shutdown initiated")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := h.srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("http shutdown: %w", err)
		}
		h.log.Info("http server stopped")
		return nil
	}
}
