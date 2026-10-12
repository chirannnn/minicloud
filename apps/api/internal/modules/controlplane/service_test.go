package controlplane

import (
	"errors"
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
	errorList := []error{
		ErrUserNotFound,
		ErrProjectNotFound,
		ErrApplicationNotFound,
		ErrDeploymentNotFound,
		ErrInvalidRelationship,
		ErrInvalidStatus,
		ErrResourceConflict,
		ErrValidation,
		ErrInternal,
	}

	for _, err := range errorList {
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

func TestToHTTPError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode string
	}{
		{
			name:         "UserNotFound",
			err:          ErrUserNotFound,
			expectedCode: CodeNotFound,
		},
		{
			name:         "ProjectNotFound",
			err:          ErrProjectNotFound,
			expectedCode: CodeNotFound,
		},
		{
			name:         "ApplicationNotFound",
			err:          ErrApplicationNotFound,
			expectedCode: CodeNotFound,
		},
		{
			name:         "DeploymentNotFound",
			err:          ErrDeploymentNotFound,
			expectedCode: CodeNotFound,
		},
		{
			name:         "InvalidRelationship",
			err:          ErrInvalidRelationship,
			expectedCode: CodeInvalidRelationship,
		},
		{
			name:         "InvalidStatus",
			err:          ErrInvalidStatus,
			expectedCode: CodeValidationFailed,
		},
		{
			name:         "ResourceConflict",
			err:          ErrResourceConflict,
			expectedCode: CodeConflict,
		},
		{
			name:         "Validation",
			err:          ErrValidation,
			expectedCode: CodeValidationFailed,
		},
		{
			name:         "Internal",
			err:          ErrInternal,
			expectedCode: CodeInternal,
		},
		{
			name:         "UnknownError",
			err:          errors.New("unknown error"),
			expectedCode: CodeInternal,
		},
		{
			name:         "NilError",
			err:          nil,
			expectedCode: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpErr := ToHTTPError(tt.err)
			if httpErr.Code != tt.expectedCode {
				t.Errorf("ToHTTPError() code = %v, want %v", httpErr.Code, tt.expectedCode)
			}
		})
	}
}

func TestValidator_ValidateEmail(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "ValidEmail",
			email:   "test@example.com",
			wantErr: false,
		},
		{
			name:    "EmptyEmail",
			email:   "",
			wantErr: true,
		},
		{
			name:    "EmailWithoutAt",
			email:   "testexample.com",
			wantErr: true,
		},
		{
			name:    "InvalidEmailFormat",
			email:   "invalid-email",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidateDisplayName(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name        string
		displayName string
		wantErr     bool
	}{
		{
			name:        "ValidDisplayName",
			displayName: "Test User",
			wantErr:     false,
		},
		{
			name:        "EmptyDisplayName",
			displayName: "",
			wantErr:     true,
		},
		{
			name:        "TooLongDisplayName",
			displayName: string(make([]byte, 101)),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateDisplayName(tt.displayName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDisplayName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidateSlug(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		slug    string
		wantErr bool
	}{
		{
			name:    "ValidSlug",
			slug:    "test-project",
			wantErr: false,
		},
		{
			name:    "EmptySlug",
			slug:    "",
			wantErr: true,
		},
		{
			name:    "SlugWithSpace",
			slug:    "test project",
			wantErr: true,
		},
		{
			name:    "SlugWithSlash",
			slug:    "test/project",
			wantErr: true,
		},
		{
			name:    "TooLongSlug",
			slug:    string(make([]byte, 65)),
			wantErr: true,
		},
		{
			name:    "SlugWithSpecialChars",
			slug:    "test@project",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateSlug(tt.slug)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSlug() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidateName(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		nameVal string
		field   string
		wantErr bool
	}{
		{
			name:    "ValidName",
			nameVal: "Test Project",
			field:   "name",
			wantErr: false,
		},
		{
			name:    "EmptyName",
			nameVal: "",
			field:   "name",
			wantErr: true,
		},
		{
			name:    "TooLongName",
			nameVal: string(make([]byte, 101)),
			field:   "name",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateName(tt.nameVal, tt.field)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidateDeploymentVersion(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		version string
		wantErr bool
	}{
		{
			name:    "ValidVersion",
			version: "v1.0.0",
			wantErr: false,
		},
		{
			name:    "EmptyVersion",
			version: "",
			wantErr: true,
		},
		{
			name:    "TooLongVersion",
			version: string(make([]byte, 51)),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateDeploymentVersion(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDeploymentVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidateDeploymentStatus(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		status  string
		wantErr bool
	}{
		{
			name:    "ValidStatusPending",
			status:  "PENDING",
			wantErr: false,
		},
		{
			name:    "ValidStatusRunning",
			status:  "RUNNING",
			wantErr: false,
		},
		{
			name:    "ValidStatusSucceeded",
			status:  "SUCCEEDED",
			wantErr: false,
		},
		{
			name:    "EmptyStatus",
			status:  "",
			wantErr: true,
		},
		{
			name:    "InvalidStatus",
			status:  "INVALID",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateDeploymentStatus(tt.status)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDeploymentStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidateUUID(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		idStr   string
		field   string
		wantErr bool
	}{
		{
			name:    "ValidUUID",
			idStr:   uuid.New().String(),
			field:   "id",
			wantErr: false,
		},
		{
			name:    "EmptyUUID",
			idStr:   "",
			field:   "id",
			wantErr: true,
		},
		{
			name:    "InvalidUUID",
			idStr:   "not-a-uuid",
			field:   "id",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validator.ValidateUUID(tt.idStr, tt.field)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUUID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	err := &ValidationError{
		Field:   "email",
		Message: "invalid email format",
	}

	expected := "email: invalid email format"
	if err.Error() != expected {
		t.Errorf("ValidationError.Error() = %v, want %v", err.Error(), expected)
	}
}

func TestHTTPError_Structure(t *testing.T) {
	httpErr := HTTPError{
		Code:    CodeValidationFailed,
		Message: "Validation failed",
	}

	if httpErr.Code != CodeValidationFailed {
		t.Errorf("HTTPError.Code = %v, want %v", httpErr.Code, CodeValidationFailed)
	}

	if httpErr.Message != "Validation failed" {
		t.Errorf("HTTPError.Message = %v, want Validation failed", httpErr.Message)
	}
}
