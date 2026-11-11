package controlplane

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minicloud/minicloud/apps/api/internal/database"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	// Check if test database is available
	testDBURL := os.Getenv("TEST_DATABASE_URL")
	if testDBURL == "" {
		// Default to main database for testing if test database not available
		// This is not ideal but allows running tests without separate test infrastructure
		testDBURL = "postgres://minicloud:minicloud_dev_password@localhost:5432/minicloud?sslmode=disable"
	}

	ctx := context.Background()
	var err error
	testPool, err = database.Open(ctx, testDBURL)
	if err != nil {
		// If database is not available, skip integration tests gracefully
		fmt.Printf("Integration tests skipped: test database not available (%v)\n", err)
		fmt.Println("To run integration tests:")
		fmt.Println("1. Set TEST_DATABASE_URL environment variable")
		fmt.Println("2. Or ensure PostgreSQL is running at localhost:5432 with credentials: minicloud/minicloud_dev_password")
		os.Exit(0)
	}
	defer testPool.Close()

	// Apply migrations
	migrationsDir := "../../migrations"
	if err := database.Migrate(ctx, testPool, migrationsDir); err != nil {
		fmt.Printf("Failed to apply migrations: %v\n", err)
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Clean up test database
	cleanupTestDatabase(ctx, testPool)

	os.Exit(code)
}

func cleanupTestDatabase(ctx context.Context, pool *pgxpool.Pool) {
	// Delete all test data
	tables := []string{"outbox_events", "audit_logs", "deployments", "applications", "projects", "users"}
	for _, table := range tables {
		pool.Exec(ctx, fmt.Sprintf("DELETE FROM %s WHERE true", table))
	}
}

func setupTestUser(ctx context.Context, t *testing.T) (uuid.UUID, error) {
	userID := uuid.New()
	email := fmt.Sprintf("test-%s@example.com", userID.String())
	displayName := "Test User"

	_, err := testPool.Exec(ctx, `
		INSERT INTO users (id, email, display_name) 
		VALUES ($1, $2, $3)
	`, userID, email, displayName)

	return userID, err
}

func TestIntegration_CreateUser(t *testing.T) {
	if testPool == nil {
		t.Skip("Test database not available")
	}

	ctx := context.Background()
	email := fmt.Sprintf("integration-test-%s@example.com", uuid.New().String())
	displayName := "Integration Test User"

	userID := uuid.New()
	_, err := testPool.Exec(ctx, `
		INSERT INTO users (id, email, display_name) 
		VALUES ($1, $2, $3)
	`, userID, email, displayName)

	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Verify user was created
	var createdEmail, createdDisplayName string
	err = testPool.QueryRow(ctx, `SELECT email, display_name FROM users WHERE id = $1`, userID).Scan(&createdEmail, &createdDisplayName)
	if err != nil {
		t.Fatalf("Failed to query user: %v", err)
	}

	if createdEmail != email {
		t.Errorf("Expected email %s, got %s", email, createdEmail)
	}
	if createdDisplayName != displayName {
		t.Errorf("Expected display name %s, got %s", displayName, createdDisplayName)
	}
}

func TestIntegration_CreateProject(t *testing.T) {
	if testPool == nil {
		t.Skip("Test database not available")
	}

	ctx := context.Background()
	ownerID, err := setupTestUser(ctx, t)
	if err != nil {
		t.Fatalf("Failed to setup test user: %v", err)
	}

	projectID := uuid.New()
	name := "Test Project"
	slug := "test-project"
	description := "A test project"

	// Create project in transaction with audit and outbox
	tx, err := testPool.Begin(ctx)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO projects (id, owner_id, name, slug, description, status) 
		VALUES ($1, $2, $3, $4, $5, 'active')
	`, projectID, ownerID, name, slug, description)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Insert audit log
	_, err = tx.Exec(ctx, `
		INSERT INTO audit_logs (project_id, actor_id, action, resource_type, resource_id, metadata) 
		VALUES ($1, $2, 'project.created', 'project', $3, '{}'::jsonb)
	`, projectID, ownerID, projectID)
	if err != nil {
		t.Fatalf("Failed to insert audit log: %v", err)
	}

	// Insert outbox event
	_, err = tx.Exec(ctx, `
		INSERT INTO outbox_events (event_type, aggregate_type, aggregate_id, payload) 
		VALUES ('minicloud.project.created', 'project', $1, '{}'::jsonb)
	`, projectID)
	if err != nil {
		t.Fatalf("Failed to insert outbox event: %v", err)
	}

	if err = tx.Commit(ctx); err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Verify all three records exist
	var projectName string
	err = testPool.QueryRow(ctx, `SELECT name FROM projects WHERE id = $1`, projectID).Scan(&projectName)
	if err != nil {
		t.Fatalf("Failed to query project: %v", err)
	}
	if projectName != name {
		t.Errorf("Expected project name %s, got %s", name, projectName)
	}

	var auditAction string
	err = testPool.QueryRow(ctx, `SELECT action FROM audit_logs WHERE resource_id = $1`, projectID).Scan(&auditAction)
	if err != nil {
		t.Fatalf("Failed to query audit log: %v", err)
	}
	if auditAction != "project.created" {
		t.Errorf("Expected audit action 'project.created', got %s", auditAction)
	}

	var eventType string
	err = testPool.QueryRow(ctx, `SELECT event_type FROM outbox_events WHERE aggregate_id = $1`, projectID).Scan(&eventType)
	if err != nil {
		t.Fatalf("Failed to query outbox event: %v", err)
	}
	if eventType != "minicloud.project.created" {
		t.Errorf("Expected event type 'minicloud.project.created', got %s", eventType)
	}
}

func TestIntegration_ProjectAuditOutboxAtomicity(t *testing.T) {
	if testPool == nil {
		t.Skip("Test database not available")
	}

	ctx := context.Background()
	ownerID, err := setupTestUser(ctx, t)
	if err != nil {
		t.Fatalf("Failed to setup test user: %v", err)
	}

	projectID := uuid.New()

	// Test: Successful transaction creates all records
	t.Run("Success", func(t *testing.T) {
		tx, err := testPool.Begin(ctx)
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO projects (id, owner_id, name, slug, description, status) 
			VALUES ($1, $2, 'Test', 'test', 'desc', 'active')
		`, projectID, ownerID)
		if err != nil {
			t.Fatalf("Failed to create project: %v", err)
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO audit_logs (project_id, actor_id, action, resource_type, resource_id, metadata) 
			VALUES ($1, $2, 'project.created', 'project', $1, '{}'::jsonb)
		`, projectID, ownerID)
		if err != nil {
			t.Fatalf("Failed to insert audit log: %v", err)
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO outbox_events (event_type, aggregate_type, aggregate_id, payload) 
			VALUES ('minicloud.project.created', 'project', $1, '{}'::jsonb)
		`, projectID)
		if err != nil {
			t.Fatalf("Failed to insert outbox event: %v", err)
		}

		if err = tx.Commit(ctx); err != nil {
			t.Fatalf("Failed to commit transaction: %v", err)
		}

		// Verify all records exist
		var count int
		testPool.QueryRow(ctx, `SELECT COUNT(*) FROM projects WHERE id = $1`, projectID).Scan(&count)
		if count != 1 {
			t.Errorf("Expected 1 project, got %d", count)
		}
		testPool.QueryRow(ctx, `SELECT COUNT(*) FROM audit_logs WHERE resource_id = $1`, projectID).Scan(&count)
		if count != 1 {
			t.Errorf("Expected 1 audit log, got %d", count)
		}
		testPool.QueryRow(ctx, `SELECT COUNT(*) FROM outbox_events WHERE aggregate_id = $1`, projectID).Scan(&count)
		if count != 1 {
			t.Errorf("Expected 1 outbox event, got %d", count)
		}

		// Cleanup
		testPool.Exec(ctx, `DELETE FROM projects WHERE id = $1`, projectID)
	})

	// Test: Rollback removes all records
	t.Run("Rollback", func(t *testing.T) {
		projectID2 := uuid.New()

		tx, err := testPool.Begin(ctx)
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO projects (id, owner_id, name, slug, description, status) 
			VALUES ($1, $2, 'Test2', 'test2', 'desc', 'active')
		`, projectID2, ownerID)
		if err != nil {
			t.Fatalf("Failed to create project: %v", err)
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO audit_logs (project_id, actor_id, action, resource_type, resource_id, metadata) 
			VALUES ($1, $2, 'project.created', 'project', $1, '{}'::jsonb)
		`, projectID2, ownerID)
		if err != nil {
			t.Fatalf("Failed to insert audit log: %v", err)
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO outbox_events (event_type, aggregate_type, aggregate_id, payload) 
			VALUES ('minicloud.project.created', 'project', $1, '{}'::jsonb)
		`, projectID2)
		if err != nil {
			t.Fatalf("Failed to insert outbox event: %v", err)
		}

		// Manually rollback to simulate failure
		_ = tx.Rollback(ctx)

		// Verify no records exist (transaction was rolled back)
		var count int
		testPool.QueryRow(ctx, `SELECT COUNT(*) FROM projects WHERE id = $1`, projectID2).Scan(&count)
		if count != 0 {
			t.Errorf("Expected 0 projects after rollback, got %d", count)
		}
		testPool.QueryRow(ctx, `SELECT COUNT(*) FROM audit_logs WHERE resource_id = $1`, projectID2).Scan(&count)
		if count != 0 {
			t.Errorf("Expected 0 audit logs after rollback, got %d", count)
		}
		testPool.QueryRow(ctx, `SELECT COUNT(*) FROM outbox_events WHERE aggregate_id = $1`, projectID2).Scan(&count)
		if count != 0 {
			t.Errorf("Expected 0 outbox events after rollback, got %d", count)
		}
	})
}

func TestIntegration_ApplicationLifecycle(t *testing.T) {
	if testPool == nil {
		t.Skip("Test database not available")
	}

	ctx := context.Background()
	ownerID, err := setupTestUser(ctx, t)
	if err != nil {
		t.Fatalf("Failed to setup test user: %v", err)
	}

	projectID := uuid.New()
	_, err = testPool.Exec(ctx, `
		INSERT INTO projects (id, owner_id, name, slug, description, status) 
		VALUES ($1, $2, 'Test Project', 'test-project', 'desc', 'active')
	`, projectID, ownerID)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Create application
	appID := uuid.New()
	_, err = testPool.Exec(ctx, `
		INSERT INTO applications (id, project_id, name, slug, description, status) 
		VALUES ($1, $2, 'Test App', 'test-app', 'desc', 'active')
	`, appID, projectID)
	if err != nil {
		t.Fatalf("Failed to create application: %v", err)
	}

	// Verify application exists
	var appName string
	err = testPool.QueryRow(ctx, `SELECT name FROM applications WHERE id = $1`, appID).Scan(&appName)
	if err != nil {
		t.Fatalf("Failed to query application: %v", err)
	}
	if appName != "Test App" {
		t.Errorf("Expected app name 'Test App', got %s", appName)
	}

	// Verify relationship
	var projectIDFromApp uuid.UUID
	err = testPool.QueryRow(ctx, `SELECT project_id FROM applications WHERE id = $1`, appID).Scan(&projectIDFromApp)
	if err != nil {
		t.Fatalf("Failed to query application project: %v", err)
	}
	if projectIDFromApp != projectID {
		t.Errorf("Expected project_id %s, got %s", projectID, projectIDFromApp)
	}

	// Delete application
	_, err = testPool.Exec(ctx, `DELETE FROM applications WHERE id = $1`, appID)
	if err != nil {
		t.Fatalf("Failed to delete application: %v", err)
	}

	// Verify deletion
	var count int
	testPool.QueryRow(ctx, `SELECT COUNT(*) FROM applications WHERE id = $1`, appID).Scan(&count)
	if count != 0 {
		t.Errorf("Expected 0 applications after deletion, got %d", count)
	}
}

func TestIntegration_DeploymentLifecycle(t *testing.T) {
	if testPool == nil {
		t.Skip("Test database not available")
	}

	ctx := context.Background()
	ownerID, err := setupTestUser(ctx, t)
	if err != nil {
		t.Fatalf("Failed to setup test user: %v", err)
	}

	projectID := uuid.New()
	_, err = testPool.Exec(ctx, `
		INSERT INTO projects (id, owner_id, name, slug, description, status) 
		VALUES ($1, $2, 'Test Project', 'test-project', 'desc', 'active')
	`, projectID, ownerID)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	appID := uuid.New()
	_, err = testPool.Exec(ctx, `
		INSERT INTO applications (id, project_id, name, slug, description, status) 
		VALUES ($1, $2, 'Test App', 'test-app', 'desc', 'active')
	`, appID, projectID)
	if err != nil {
		t.Fatalf("Failed to create application: %v", err)
	}

	// Create deployment
	deploymentID := uuid.New()
	version := "v1.0.0"
	config := map[string]interface{}{"image": "nginx:latest", "replicas": 3}
	configJSON, _ := json.Marshal(config)

	_, err = testPool.Exec(ctx, `
		INSERT INTO deployments (id, project_id, application_id, version, status, configuration) 
		VALUES ($1, $2, $3, $4, 'PENDING', $5)
	`, deploymentID, projectID, appID, version, configJSON)
	if err != nil {
		t.Fatalf("Failed to create deployment: %v", err)
	}

	// Verify deployment exists
	var deploymentVersion string
	err = testPool.QueryRow(ctx, `SELECT version FROM deployments WHERE id = $1`, deploymentID).Scan(&deploymentVersion)
	if err != nil {
		t.Fatalf("Failed to query deployment: %v", err)
	}
	if deploymentVersion != version {
		t.Errorf("Expected version %s, got %s", version, deploymentVersion)
	}

	// Update deployment status
	_, err = testPool.Exec(ctx, `
		UPDATE deployments SET status = 'RUNNING' WHERE id = $1
	`, deploymentID)
	if err != nil {
		t.Fatalf("Failed to update deployment status: %v", err)
	}

	// Verify status update
	var status string
	err = testPool.QueryRow(ctx, `SELECT status FROM deployments WHERE id = $1`, deploymentID).Scan(&status)
	if err != nil {
		t.Fatalf("Failed to query deployment status: %v", err)
	}
	if status != "RUNNING" {
		t.Errorf("Expected status 'RUNNING', got %s", status)
	}
}

func TestIntegration_Pagination(t *testing.T) {
	if testPool == nil {
		t.Skip("Test database not available")
	}

	ctx := context.Background()
	ownerID, err := setupTestUser(ctx, t)
	if err != nil {
		t.Fatalf("Failed to setup test user: %v", err)
	}

	// Create 25 projects
	projectIDs := make([]uuid.UUID, 25)
	for i := 0; i < 25; i++ {
		projectIDs[i] = uuid.New()
		_, err = testPool.Exec(ctx, `
			INSERT INTO projects (id, owner_id, name, slug, description, status) 
			VALUES ($1, $2, $3, $4, $5, 'active')
		`, projectIDs[i], ownerID, fmt.Sprintf("Project %d", i), fmt.Sprintf("project-%d", i), fmt.Sprintf("Description %d", i))
		if err != nil {
			t.Fatalf("Failed to create project %d: %v", i, err)
		}
	}

	// Test pagination with limit 10, offset 0
	rows, err := testPool.Query(ctx, `
		SELECT id, name FROM projects WHERE owner_id = $1 ORDER BY created_at LIMIT 10 OFFSET 0
	`, ownerID)
	if err != nil {
		t.Fatalf("Failed to query projects: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}
	if count != 10 {
		t.Errorf("Expected 10 projects with limit 10, got %d", count)
	}

	// Test pagination with limit 10, offset 10
	rows, err = testPool.Query(ctx, `
		SELECT id, name FROM projects WHERE owner_id = $1 ORDER BY created_at LIMIT 10 OFFSET 10
	`, ownerID)
	if err != nil {
		t.Fatalf("Failed to query projects with offset: %v", err)
	}
	defer rows.Close()

	count = 0
	for rows.Next() {
		count++
	}
	if count != 10 {
		t.Errorf("Expected 10 projects with limit 10 offset 10, got %d", count)
	}

	// Test pagination with limit 10, offset 20 (should return 5)
	rows, err = testPool.Query(ctx, `
		SELECT id, name FROM projects WHERE owner_id = $1 ORDER BY created_at LIMIT 10 OFFSET 20
	`, ownerID)
	if err != nil {
		t.Fatalf("Failed to query projects with offset 20: %v", err)
	}
	defer rows.Close()

	count = 0
	for rows.Next() {
		count++
	}
	if count != 5 {
		t.Errorf("Expected 5 projects with limit 10 offset 20, got %d", count)
	}

	// Cleanup
	for _, pid := range projectIDs {
		testPool.Exec(ctx, `DELETE FROM projects WHERE id = $1`, pid)
	}
}

func TestIntegration_NotFoundBehavior(t *testing.T) {
	if testPool == nil {
		t.Skip("Test database not available")
	}

	ctx := context.Background()
	nonExistentID := uuid.New()

	// Try to get non-existent user
	var email string
	err := testPool.QueryRow(ctx, `SELECT email FROM users WHERE id = $1`, nonExistentID).Scan(&email)
	if err == nil {
		t.Error("Expected error when querying non-existent user")
	}

	// Try to get non-existent project
	var name string
	err = testPool.QueryRow(ctx, `SELECT name FROM projects WHERE id = $1`, nonExistentID).Scan(&name)
	if err == nil {
		t.Error("Expected error when querying non-existent project")
	}

	// Try to delete non-existent application
	result, err := testPool.Exec(ctx, `DELETE FROM applications WHERE id = $1`, nonExistentID)
	if err != nil {
		t.Fatalf("Failed to execute delete: %v", err)
	}
	rowsAffected := result.RowsAffected()
	if rowsAffected != 0 {
		t.Errorf("Expected 0 rows affected, got %d", rowsAffected)
	}
}

func TestIntegration_ConstraintViolations(t *testing.T) {
	if testPool == nil {
		t.Skip("Test database not available")
	}

	ctx := context.Background()
	ownerID, err := setupTestUser(ctx, t)
	if err != nil {
		t.Fatalf("Failed to setup test user: %v", err)
	}

	projectID := uuid.New()
	_, err = testPool.Exec(ctx, `
		INSERT INTO projects (id, owner_id, name, slug, description, status) 
		VALUES ($1, $2, 'Test Project', 'test-project', 'desc', 'active')
	`, projectID, ownerID)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Try to create duplicate project with same slug
	duplicateProjectID := uuid.New()
	_, err = testPool.Exec(ctx, `
		INSERT INTO projects (id, owner_id, name, slug, description, status) 
		VALUES ($1, $2, 'Duplicate Project', 'test-project', 'desc', 'active')
	`, duplicateProjectID, ownerID)
	if err == nil {
		t.Error("Expected constraint violation for duplicate slug")
	}

	// Try to create application for non-existent project
	appID := uuid.New()
	nonExistentProjectID := uuid.New()
	_, err = testPool.Exec(ctx, `
		INSERT INTO applications (id, project_id, name, slug, description, status) 
		VALUES ($1, $2, 'Test App', 'test-app', 'desc', 'active')
	`, appID, nonExistentProjectID)
	if err == nil {
		t.Error("Expected constraint violation for non-existent project")
	}
}

func TestIntegration_AuditMetadataFields(t *testing.T) {
	if testPool == nil {
		t.Skip("Test database not available")
	}

	ctx := context.Background()
	ownerID, err := setupTestUser(ctx, t)
	if err != nil {
		t.Fatalf("Failed to setup test user: %v", err)
	}

	projectID := uuid.New()
	name := "Metadata Test Project"
	slug := "metadata-test"

	_, err = testPool.Exec(ctx, `
		INSERT INTO projects (id, owner_id, name, slug, description, status) 
		VALUES ($1, $2, $3, $4, 'desc', 'active')
	`, projectID, ownerID, name, slug)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Insert audit with metadata
	metadata := map[string]interface{}{
		"name":        name,
		"slug":        slug,
		"ownerId":     ownerID.String(),
		"description": "desc",
		"status":      "active",
	}
	metadataJSON, _ := json.Marshal(metadata)

	_, err = testPool.Exec(ctx, `
		INSERT INTO audit_logs (project_id, actor_id, action, resource_type, resource_id, metadata) 
		VALUES ($1, $2, 'project.created', 'project', $1, $3)
	`, projectID, ownerID, metadataJSON)
	if err != nil {
		t.Fatalf("Failed to insert audit log: %v", err)
	}

	// Verify metadata
	var retrievedMetadata string
	err = testPool.QueryRow(ctx, `SELECT metadata FROM audit_logs WHERE resource_id = $1`, projectID).Scan(&retrievedMetadata)
	if err != nil {
		t.Fatalf("Failed to query audit metadata: %v", err)
	}

	var parsedMetadata map[string]interface{}
	json.Unmarshal([]byte(retrievedMetadata), &parsedMetadata)

	if parsedMetadata["name"] != name {
		t.Errorf("Expected metadata name %s, got %v", name, parsedMetadata["name"])
	}
	if parsedMetadata["slug"] != slug {
		t.Errorf("Expected metadata slug %s, got %v", slug, parsedMetadata["slug"])
	}
}

func TestIntegration_OutboxEventPayload(t *testing.T) {
	if testPool == nil {
		t.Skip("Test database not available")
	}

	ctx := context.Background()
	ownerID, err := setupTestUser(ctx, t)
	if err != nil {
		t.Fatalf("Failed to setup test user: %v", err)
	}

	projectID := uuid.New()
	_, err = testPool.Exec(ctx, `
		INSERT INTO projects (id, owner_id, name, slug, description, status) 
		VALUES ($1, $2, 'Test Project', 'test-project', 'desc', 'active')
	`, projectID, ownerID)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Insert outbox event with payload
	payload := map[string]interface{}{
		"ownerId":     ownerID.String(),
		"name":        "Test Project",
		"slug":        "test-project",
		"description": "desc",
		"status":      "active",
	}
	payloadJSON, _ := json.Marshal(payload)

	_, err = testPool.Exec(ctx, `
		INSERT INTO outbox_events (event_type, aggregate_type, aggregate_id, payload) 
		VALUES ('minicloud.project.created', 'project', $1, $2)
	`, projectID, payloadJSON)
	if err != nil {
		t.Fatalf("Failed to insert outbox event: %v", err)
	}

	// Verify payload
	var retrievedPayload string
	err = testPool.QueryRow(ctx, `SELECT payload FROM outbox_events WHERE aggregate_id = $1`, projectID).Scan(&retrievedPayload)
	if err != nil {
		t.Fatalf("Failed to query outbox payload: %v", err)
	}

	var parsedPayload map[string]interface{}
	json.Unmarshal([]byte(retrievedPayload), &parsedPayload)

	if parsedPayload["name"] != "Test Project" {
		t.Errorf("Expected payload name 'Test Project', got %v", parsedPayload["name"])
	}
	if parsedPayload["slug"] != "test-project" {
		t.Errorf("Expected payload slug 'test-project', got %v", parsedPayload["slug"])
	}

	// Verify published_at is NULL
	var publishedAt *time.Time
	err = testPool.QueryRow(ctx, `SELECT published_at FROM outbox_events WHERE aggregate_id = $1`, projectID).Scan(&publishedAt)
	if err != nil {
		t.Fatalf("Failed to query published_at: %v", err)
	}
	if publishedAt != nil {
		t.Error("Expected published_at to be NULL")
	}
}
