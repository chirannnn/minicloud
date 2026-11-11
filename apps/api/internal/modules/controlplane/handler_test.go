package controlplane

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Handler tests for HTTP behavior
func TestHandlerMalformedJSON(t *testing.T) {
	// Create a mock handler with nil DB (will fail on service calls, but that's okay for JSON parsing test)
	handler := New(nil)
	mux := http.NewServeMux()
	handler.Register(mux)

	// Test malformed JSON
	body := bytes.NewBufferString("{invalid json}")
	req := httptest.NewRequest("POST", "/api/v1/users", body)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for malformed JSON, got %d", w.Code)
	}

	var resp HTTPError
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	if resp.Code != CodeValidationFailed {
		t.Errorf("Expected VALIDATION_FAILED code, got %s", resp.Code)
	}
}

func TestHandlerInvalidRequestBody(t *testing.T) {
	handler := New(nil)
	mux := http.NewServeMux()
	handler.Register(mux)

	// Test valid JSON but missing required fields
	body := bytes.NewBufferString(`{"displayName": "Test"}`)
	req := httptest.NewRequest("POST", "/api/v1/users", body)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	// Should return validation error (email is required)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for missing required field, got %d", w.Code)
	}
}

func TestHandlerInvalidUUID(t *testing.T) {
	handler := New(nil)
	mux := http.NewServeMux()
	handler.Register(mux)

	// Test invalid UUID in path
	req := httptest.NewRequest("GET", "/api/v1/users/not-a-uuid", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	// Should return 404 because UUID parsing fails (path value not found)
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for invalid UUID, got %d", w.Code)
	}
}

func TestHandlerPaginationErrors(t *testing.T) {
	handler := New(nil)
	mux := http.NewServeMux()
	handler.Register(mux)

	// Test invalid pagination parameter
	req := httptest.NewRequest("GET", "/api/v1/users?limit=abc", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	// Should return validation error
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for invalid pagination, got %d", w.Code)
	}

	var resp HTTPError
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	if resp.Code != CodeValidationFailed {
		t.Errorf("Expected VALIDATION_FAILED code, got %s", resp.Code)
	}
}

func TestHandlerConsistentErrorResponse(t *testing.T) {
	handler := New(nil)
	mux := http.NewServeMux()
	handler.Register(mux)

	// Test error response structure
	body := bytes.NewBufferString(`{"displayName": "Test"}`)
	req := httptest.NewRequest("POST", "/api/v1/users", body)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	var resp HTTPError
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	// Verify error response has required fields
	if resp.Code == "" {
		t.Error("Expected error code to be set")
	}
	if resp.Message == "" {
		t.Error("Expected error message to be set")
	}

	// Verify Content-Type is JSON
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
	}
}

func TestHandlerWithMockDB(t *testing.T) {
	// Create a handler with a mock DB connection
	// This test verifies handler doesn't panic with nil DB
	handler := New(nil)
	mux := http.NewServeMux()
	handler.Register(mux)

	// Test that handler doesn't panic on valid request structure
	body := bytes.NewBufferString(`{"email": "test@example.com", "displayName": "Test User"}`)
	req := httptest.NewRequest("POST", "/api/v1/users", body)
	w := httptest.NewRecorder()

	// This will fail on DB call but shouldn't panic
	mux.ServeHTTP(w, req)

	// Should return internal error since DB is nil
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500 with nil DB, got %d", w.Code)
	}

	var resp HTTPError
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	if resp.Code != CodeInternal {
		t.Errorf("Expected INTERNAL code, got %s", resp.Code)
	}
}
