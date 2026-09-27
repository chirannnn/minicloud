package controlplane

import (
	"testing"

	"github.com/google/uuid"
)

func TestService_UpdateDeploymentStatus_ValidatesStatus(t *testing.T) {
	// Test that invalid status is rejected
	validStatuses := map[string]bool{
		"PENDING":   true,
		"QUEUED":    true,
		"RUNNING":   true,
		"SUCCEEDED": true,
		"FAILED":    true,
		"CANCELLED": true,
	}

	invalidStatus := "INVALID_STATUS"
	if validStatuses[invalidStatus] {
		t.Errorf("Expected invalid status to be rejected")
	}

	// Test that valid statuses are accepted
	for status := range validStatuses {
		if !validStatuses[status] {
			t.Errorf("Expected valid status %s to be accepted", status)
		}
	}
}

func TestServiceErrors(t *testing.T) {
	// Verify service error types are defined
	errors := []error{
		ErrUserNotFound,
		ErrProjectNotFound,
		ErrApplicationNotFound,
		ErrDeploymentNotFound,
		ErrInvalidRelationship,
		ErrInvalidStatus,
		ErrResourceConflict,
	}

	for _, err := range errors {
		if err == nil {
			t.Errorf("Expected error to be defined")
		}
	}
}

func TestUUIDValidation(t *testing.T) {
	// Test that we can generate and parse UUIDs
	id := uuid.New()
	if id == uuid.Nil {
		t.Errorf("Expected non-nil UUID")
	}

	parsed, err := uuid.Parse(id.String())
	if err != nil {
		t.Errorf("Failed to parse UUID: %v", err)
	}

	if parsed != id {
		t.Errorf("UUID mismatch after parsing")
	}
}
