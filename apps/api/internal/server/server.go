package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/minicloud/minicloud/apps/api/internal/config"
)

func Run(ctx context.Context, cfg config.Config, handler http.Handler, log *slog.Logger) error {
	srv := &http.Server{Addr: cfg.Address(), Handler: handler, ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 60 * time.Second}
	errCh := make(chan error, 1)
	go func() { log.Info("API server started", "address", cfg.Address()); errCh <- srv.ListenAndServe() }()
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
