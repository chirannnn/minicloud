package router

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minicloud/minicloud/apps/api/internal/config"
	"github.com/minicloud/minicloud/apps/api/internal/modules/controlplane"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func New(cfg config.Config, log *slog.Logger, databases ...*pgxpool.Pool) http.Handler {
	var db *pgxpool.Pool
	if len(databases) > 0 {
		db = databases[0]
	}
	mux := http.NewServeMux()
	health := func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": cfg.ServiceName})
	}
	mux.HandleFunc("GET /health", health)
	mux.HandleFunc("GET /api/v1/health", health)
	mux.HandleFunc("GET /ready", readiness(cfg))
	if db != nil {
		controlplane.New(db).Register(mux)
	}
	return otelhttp.NewHandler(requestID(cors(cfg.ConsoleOrigin, log, mux)), "minicloud-api")
}

func readiness(cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dependencies := make(map[string]string, len(cfg.Dependencies))
		ready := true
		for name, address := range cfg.Dependencies {
			connection, err := (&net.Dialer{}).DialContext(r.Context(), "tcp", address)
			if err != nil {
				dependencies[name] = "unavailable"
				ready = false
				continue
			}
			_ = connection.Close()
			dependencies[name] = "ok"
		}
		if !ready {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"code": "dependency_unavailable", "message": "one or more dependencies are unavailable", "requestId": requestIDFrom(r), "dependencies": dependencies})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "dependencies": dependencies})
	}
}

type requestIDKey struct{}

func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = time.Now().UTC().Format("20060102T150405.000000000Z")
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey{}, id)))
	})
}
func requestIDFrom(r *http.Request) string {
	id, _ := r.Context().Value(requestIDKey{}).(string)
	return id
}
func cors(origin string, log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Request-ID")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		started := time.Now()
		next.ServeHTTP(w, r)
		log.Debug("http request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(started).String())
	})
}
func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
