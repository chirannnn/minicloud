package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/minicloud/minicloud/apps/api/internal/config"
	"github.com/minicloud/minicloud/apps/api/internal/logger"
)

func TestHealth(t *testing.T) {
	h := New(config.Config{ServiceName: "minicloud-api", ConsoleOrigin: "http://localhost:3000"}, logger.New())
	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected request ID")
	}
}

func TestReadinessUnavailableWithoutDependencies(t *testing.T) {
	h := New(config.Config{ServiceName: "minicloud-api", ConsoleOrigin: "http://localhost:3000", Dependencies: map[string]string{"postgres": "127.0.0.1:1"}}, logger.New())
	r := httptest.NewRequest(http.MethodGet, "/ready", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}
