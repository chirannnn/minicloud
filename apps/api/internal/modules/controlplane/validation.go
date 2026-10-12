package controlplane

import (
	"fmt"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/google/uuid"
)

// Pagination constants
const (
	DefaultLimit = 20
	MaxLimit     = 100
)

// Validator provides validation methods for control-plane resources
type Validator struct{}

// NewValidator creates a new Validator instance
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateEmail validates email format
func (v *Validator) ValidateEmail(email string) *ValidationError {
	if email == "" {
		return &ValidationError{Field: "email", Message: "email is required"}
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		return &ValidationError{Field: "email", Message: "invalid email format"}
	}

	if !strings.Contains(email, "@") {
		return &ValidationError{Field: "email", Message: "email must contain @"}
	}

	return nil
}

// ValidateDisplayName validates display name
func (v *Validator) ValidateDisplayName(displayName string) *ValidationError {
	if displayName == "" {
		return &ValidationError{Field: "displayName", Message: "displayName is required"}
	}

	if len(displayName) > 100 {
		return &ValidationError{Field: "displayName", Message: "displayName must be less than 100 characters"}
	}

	return nil
}

// ValidateSlug validates slug format
func (v *Validator) ValidateSlug(slug string) *ValidationError {
	if slug == "" {
		return &ValidationError{Field: "slug", Message: "slug is required"}
	}

	if len(slug) > 64 {
		return &ValidationError{Field: "slug", Message: "slug must be less than 64 characters"}
	}

	if strings.ContainsAny(slug, " /\\") {
		return &ValidationError{Field: "slug", Message: "slug cannot contain spaces or slashes"}
	}

	// Only allow alphanumeric, hyphens, and underscores
	validSlug := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validSlug.MatchString(slug) {
		return &ValidationError{Field: "slug", Message: "slug can only contain letters, numbers, hyphens, and underscores"}
	}

	return nil
}

// ValidateName validates resource name
func (v *Validator) ValidateName(name string, field string) *ValidationError {
	if name == "" {
		return &ValidationError{Field: field, Message: fmt.Sprintf("%s is required", field)}
	}

	if len(name) > 100 {
		return &ValidationError{Field: field, Message: fmt.Sprintf("%s must be less than 100 characters", field)}
	}

	return nil
}

// ValidateDescription validates description (optional field)
func (v *Validator) ValidateDescription(description string, field string) *ValidationError {
	if len(description) > 500 {
		return &ValidationError{Field: field, Message: fmt.Sprintf("%s must be less than 500 characters", field)}
	}

	return nil
}

// ValidateUUID validates UUID format
func (v *Validator) ValidateUUID(idStr string, field string) (uuid.UUID, *ValidationError) {
	if idStr == "" {
		return uuid.Nil, &ValidationError{Field: field, Message: fmt.Sprintf("%s is required", field)}
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, &ValidationError{Field: field, Message: fmt.Sprintf("invalid %s format", field)}
	}

	return id, nil
}

// ValidateDeploymentVersion validates deployment version
func (v *Validator) ValidateDeploymentVersion(version string) *ValidationError {
	if version == "" {
		return &ValidationError{Field: "version", Message: "version is required"}
	}

	if len(version) > 50 {
		return &ValidationError{Field: "version", Message: "version must be less than 50 characters"}
	}

	return nil
}

// ValidateDeploymentStatus validates deployment status
func (v *Validator) ValidateDeploymentStatus(status string) *ValidationError {
	if status == "" {
		return &ValidationError{Field: "status", Message: "status is required"}
	}

	// Only allow the specific status values that are defined in the database schema
	validStatuses := map[string]bool{
		"PENDING":   true,
		"QUEUED":    true,
		"RUNNING":   true,
		"SUCCEEDED": true,
		"FAILED":    true,
		"CANCELLED": true,
	}

	if !validStatuses[status] {
		return &ValidationError{Field: "status", Message: "invalid status, must be one of: PENDING, QUEUED, RUNNING, SUCCEEDED, FAILED, CANCELLED"}
	}

	return nil
}

// ValidateStringForAlphaNumeric validates that a string contains only alphanumeric characters
func (v *Validator) ValidateStringForAlphaNumeric(s string, field string) *ValidationError {
	if s == "" {
		return &ValidationError{Field: field, Message: fmt.Sprintf("%s is required", field)}
	}

	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return &ValidationError{Field: field, Message: fmt.Sprintf("%s must contain only letters and numbers", field)}
		}
	}

	return nil
}

// ValidateStringLength validates string length constraints
func (v *Validator) ValidateStringLength(s string, field string, min, max int) *ValidationError {
	if s == "" {
		return &ValidationError{Field: field, Message: fmt.Sprintf("%s is required", field)}
	}

	if len(s) < min {
		return &ValidationError{Field: field, Message: fmt.Sprintf("%s must be at least %d characters", field, min)}
	}

	if len(s) > max {
		return &ValidationError{Field: field, Message: fmt.Sprintf("%s must be less than %d characters", field, max)}
	}

	return nil
}

// ValidatePagination validates and returns limit and offset with defaults
func (v *Validator) ValidatePagination(limitStr, offsetStr string) (int32, int32, *ValidationError) {
	// Parse limit
	limit := int32(DefaultLimit)
	if limitStr != "" {
		parsedLimit, err := strconv.ParseInt(limitStr, 10, 32)
		if err != nil {
			return 0, 0, &ValidationError{Field: "limit", Message: "limit must be a valid integer"}
		}
		if parsedLimit < 0 {
			return 0, 0, &ValidationError{Field: "limit", Message: "limit must be non-negative"}
		}
		if parsedLimit > int64(MaxLimit) {
			return 0, 0, &ValidationError{Field: "limit", Message: fmt.Sprintf("limit must not exceed %d", MaxLimit)}
		}
		limit = int32(parsedLimit)
	}

	// Parse offset
	offset := int32(0)
	if offsetStr != "" {
		parsedOffset, err := strconv.ParseInt(offsetStr, 10, 32)
		if err != nil {
			return 0, 0, &ValidationError{Field: "offset", Message: "offset must be a valid integer"}
		}
		if parsedOffset < 0 {
			return 0, 0, &ValidationError{Field: "offset", Message: "offset must be non-negative"}
		}
		offset = int32(parsedOffset)
	}

	return limit, offset, nil
}
