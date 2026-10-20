package controlplane

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minicloud/minicloud/apps/api/internal/database/generated"
)

type Service struct {
	db         *pgxpool.Pool
	repository *Repository
	Validator  *Validator
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{
		db:         db,
		repository: NewRepository(db),
		Validator:  NewValidator(),
	}
}

// Helper functions for audit and outbox event insertion

// insertAuditLog inserts an audit log record within a transaction
func (s *Service) insertAuditLog(ctx context.Context, tx pgx.Tx, projectID, actorID uuid.UUID, action, resourceType string, resourceID uuid.UUID, metadata map[string]any) error {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `INSERT INTO audit_logs(project_id,actor_id,action,resource_type,resource_id,metadata)VALUES($1,$2,$3,$4,$5,$6)`,
		projectID, actorID, action, resourceType, resourceID, metadataJSON)
	return err
}

// insertOutboxEvent inserts an outbox event record within a transaction
func (s *Service) insertOutboxEvent(ctx context.Context, tx pgx.Tx, eventType, aggregateType string, aggregateID uuid.UUID, payload map[string]any) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `INSERT INTO outbox_events(event_type,aggregate_type,aggregate_id,payload)VALUES($1,$2,$3,$4)`,
		eventType, aggregateType, aggregateID, payloadJSON)
	return err
}

// Users

func (s *Service) CreateUser(ctx context.Context, email, displayName string) (generated.User, error) {
	// Validate input
	if err := s.Validator.ValidateEmail(email); err != nil {
		return generated.User{}, ErrValidation
	}
	if err := s.Validator.ValidateDisplayName(displayName); err != nil {
		return generated.User{}, ErrValidation
	}

	user, err := s.repository.CreateUser(ctx, email, displayName)
	if err != nil {
		return generated.User{}, ErrResourceConflict
	}

	// Start transaction for audit and outbox
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return generated.User{}, err
	}
	defer tx.Rollback(ctx)

	// Insert audit log (no actor_id since authentication is not implemented)
	if err := s.insertAuditLog(ctx, tx, uuid.Nil, user.ID, "user.created", "user", user.ID, map[string]any{
		"email":       email,
		"displayName": displayName,
	}); err != nil {
		return generated.User{}, err
	}

	// Insert outbox event
	if err := s.insertOutboxEvent(ctx, tx, "minicloud.user.created", "user", user.ID, map[string]any{
		"email":       email,
		"displayName": displayName,
	}); err != nil {
		return generated.User{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return generated.User{}, err
	}

	return user, nil
}

func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (generated.User, error) {
	user, err := s.repository.GetUser(ctx, id)
	if err != nil {
		return generated.User{}, ErrUserNotFound
	}
	return user, nil
}

func (s *Service) ListUsers(ctx context.Context, limit, offset int32) ([]generated.User, error) {
	return s.repository.ListUsers(ctx, limit, offset)
}

// Projects

func (s *Service) CreateProject(ctx context.Context, ownerID uuid.UUID, name, slug, description string) (generated.Project, error) {
	// Validate input
	if err := s.Validator.ValidateName(name, "name"); err != nil {
		return generated.Project{}, ErrValidation
	}
	if err := s.Validator.ValidateSlug(slug); err != nil {
		return generated.Project{}, ErrValidation
	}
	if err := s.Validator.ValidateDescription(description, "description"); err != nil {
		return generated.Project{}, ErrValidation
	}

	// Verify owner exists
	_, err := s.repository.GetUser(ctx, ownerID)
	if err != nil {
		return generated.Project{}, ErrUserNotFound
	}

	// Create project
	project, err := s.repository.CreateProject(ctx, ownerID, name, slug, description)
	if err != nil {
		return generated.Project{}, ErrResourceConflict
	}

	// Start transaction for audit and outbox
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return generated.Project{}, err
	}
	defer tx.Rollback(ctx)

	// Insert audit log
	if err := s.insertAuditLog(ctx, tx, project.ID, ownerID, "project.created", "project", project.ID, map[string]any{
		"name":        name,
		"slug":        slug,
		"description": description,
	}); err != nil {
		return generated.Project{}, err
	}

	// Insert outbox event
	if err := s.insertOutboxEvent(ctx, tx, "minicloud.project.created", "project", project.ID, map[string]any{
		"ownerId":     ownerID.String(),
		"name":        name,
		"slug":        slug,
		"description": description,
	}); err != nil {
		return generated.Project{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return generated.Project{}, err
	}

	return project, nil
}

func (s *Service) GetProject(ctx context.Context, id uuid.UUID) (generated.Project, error) {
	project, err := s.repository.GetProject(ctx, id)
	if err != nil {
		return generated.Project{}, ErrProjectNotFound
	}
	return project, nil
}

func (s *Service) ListProjects(ctx context.Context, limit, offset int32) ([]generated.Project, error) {
	return s.repository.ListProjects(ctx, limit, offset)
}

func (s *Service) UpdateProject(ctx context.Context, id uuid.UUID, name, description, status string) (generated.Project, error) {
	// Validate input
	if err := s.Validator.ValidateName(name, "name"); err != nil {
		return generated.Project{}, ErrValidation
	}
	if err := s.Validator.ValidateDescription(description, "description"); err != nil {
		return generated.Project{}, ErrValidation
	}

	project, err := s.repository.UpdateProject(ctx, id, name, description, status)
	if err != nil {
		return generated.Project{}, ErrProjectNotFound
	}

	// Start transaction for audit and outbox
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return generated.Project{}, err
	}
	defer tx.Rollback(ctx)

	// Insert audit log (no actor_id since authentication is not implemented)
	if err := s.insertAuditLog(ctx, tx, project.ID, uuid.Nil, "project.updated", "project", project.ID, map[string]any{
		"name":        name,
		"description": description,
		"status":      status,
	}); err != nil {
		return generated.Project{}, err
	}

	// Insert outbox event
	if err := s.insertOutboxEvent(ctx, tx, "minicloud.project.updated", "project", project.ID, map[string]any{
		"name":        name,
		"description": description,
		"status":      status,
	}); err != nil {
		return generated.Project{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return generated.Project{}, err
	}

	return project, nil
}

func (s *Service) DeleteProject(ctx context.Context, id uuid.UUID) error {
	// Get project before deletion for audit/outbox
	project, err := s.repository.GetProject(ctx, id)
	if err != nil {
		return ErrProjectNotFound
	}

	err = s.repository.DeleteProject(ctx, id)
	if err != nil {
		return ErrProjectNotFound
	}

	// Start transaction for audit and outbox
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Insert audit log (no actor_id since authentication is not implemented)
	if err := s.insertAuditLog(ctx, tx, project.ID, uuid.Nil, "project.deleted", "project", project.ID, map[string]any{
		"name": project.Name,
		"slug": project.Slug,
	}); err != nil {
		return err
	}

	// Insert outbox event
	if err := s.insertOutboxEvent(ctx, tx, "minicloud.project.deleted", "project", project.ID, map[string]any{
		"name": project.Name,
		"slug": project.Slug,
	}); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

// Applications

func (s *Service) CreateApplication(ctx context.Context, projectID uuid.UUID, name, slug, description string) (generated.Application, error) {
	// Validate input
	if err := s.Validator.ValidateName(name, "name"); err != nil {
		return generated.Application{}, ErrValidation
	}
	if err := s.Validator.ValidateSlug(slug); err != nil {
		return generated.Application{}, ErrValidation
	}
	if err := s.Validator.ValidateDescription(description, "description"); err != nil {
		return generated.Application{}, ErrValidation
	}

	// Verify project exists
	_, err := s.repository.GetProject(ctx, projectID)
	if err != nil {
		return generated.Application{}, ErrProjectNotFound
	}

	// Create application
	app, err := s.repository.CreateApplication(ctx, projectID, name, slug, description)
	if err != nil {
		return generated.Application{}, ErrResourceConflict
	}

	// Start transaction for audit and outbox
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return generated.Application{}, err
	}
	defer tx.Rollback(ctx)

	// Insert audit log (no actor_id since authentication is not implemented)
	if err := s.insertAuditLog(ctx, tx, app.ProjectID, uuid.Nil, "application.created", "application", app.ID, map[string]any{
		"projectId":   projectID.String(),
		"name":        name,
		"slug":        slug,
		"description": description,
	}); err != nil {
		return generated.Application{}, err
	}

	// Insert outbox event
	if err := s.insertOutboxEvent(ctx, tx, "minicloud.application.created", "application", app.ID, map[string]any{
		"projectId":   projectID.String(),
		"name":        name,
		"slug":        slug,
		"description": description,
	}); err != nil {
		return generated.Application{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return generated.Application{}, err
	}

	return app, nil
}

func (s *Service) GetApplication(ctx context.Context, id, projectID uuid.UUID) (generated.Application, error) {
	app, err := s.repository.GetApplication(ctx, id, projectID)
	if err != nil {
		return generated.Application{}, ErrApplicationNotFound
	}
	return app, nil
}

func (s *Service) ListApplicationsByProject(ctx context.Context, projectID uuid.UUID, limit, offset int32) ([]generated.Application, error) {
	// Verify project exists
	_, err := s.repository.GetProject(ctx, projectID)
	if err != nil {
		return nil, ErrProjectNotFound
	}
	return s.repository.ListApplicationsByProject(ctx, projectID, limit, offset)
}

func (s *Service) UpdateApplication(ctx context.Context, id, projectID uuid.UUID, name, description, status string) (generated.Application, error) {
	// Validate input
	if err := s.Validator.ValidateName(name, "name"); err != nil {
		return generated.Application{}, ErrValidation
	}
	if err := s.Validator.ValidateDescription(description, "description"); err != nil {
		return generated.Application{}, ErrValidation
	}

	app, err := s.repository.UpdateApplication(ctx, id, projectID, name, description, status)
	if err != nil {
		return generated.Application{}, ErrApplicationNotFound
	}

	// Start transaction for audit and outbox
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return generated.Application{}, err
	}
	defer tx.Rollback(ctx)

	// Insert audit log (no actor_id since authentication is not implemented)
	if err := s.insertAuditLog(ctx, tx, app.ProjectID, uuid.Nil, "application.updated", "application", app.ID, map[string]any{
		"projectId":   projectID.String(),
		"name":        name,
		"description": description,
		"status":      status,
	}); err != nil {
		return generated.Application{}, err
	}

	// Insert outbox event
	if err := s.insertOutboxEvent(ctx, tx, "minicloud.application.updated", "application", app.ID, map[string]any{
		"projectId":   projectID.String(),
		"name":        name,
		"description": description,
		"status":      status,
	}); err != nil {
		return generated.Application{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return generated.Application{}, err
	}

	return app, nil
}

func (s *Service) DeleteApplication(ctx context.Context, id, projectID uuid.UUID) error {
	// Get application before deletion for audit/outbox
	app, err := s.repository.GetApplication(ctx, id, projectID)
	if err != nil {
		return ErrApplicationNotFound
	}

	err = s.repository.DeleteApplication(ctx, id, projectID)
	if err != nil {
		return ErrApplicationNotFound
	}

	// Start transaction for audit and outbox
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Insert audit log (no actor_id since authentication is not implemented)
	if err := s.insertAuditLog(ctx, tx, app.ProjectID, uuid.Nil, "application.deleted", "application", app.ID, map[string]any{
		"projectId": projectID.String(),
		"name":      app.Name,
		"slug":      app.Slug,
	}); err != nil {
		return err
	}

	// Insert outbox event
	if err := s.insertOutboxEvent(ctx, tx, "minicloud.application.deleted", "application", app.ID, map[string]any{
		"projectId": projectID.String(),
		"name":      app.Name,
		"slug":      app.Slug,
	}); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

// Deployments

func (s *Service) CreateDeployment(ctx context.Context, projectID, applicationID uuid.UUID, version string, configuration []byte) (generated.Deployment, error) {
	// Validate input
	if err := s.Validator.ValidateDeploymentVersion(version); err != nil {
		return generated.Deployment{}, ErrValidation
	}

	// Verify application exists and belongs to the project
	app, err := s.repository.GetApplication(ctx, applicationID, projectID)
	if err != nil {
		return generated.Deployment{}, ErrInvalidRelationship
	}

	// Create deployment
	deployment, err := s.repository.CreateDeployment(ctx, projectID, applicationID, version, configuration)
	if err != nil {
		return generated.Deployment{}, ErrResourceConflict
	}

	// Verify the deployment was created for the correct application
	if deployment.ApplicationID != app.ID || deployment.ProjectID != projectID {
		return generated.Deployment{}, ErrInvalidRelationship
	}

	// Start transaction for audit and outbox
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return generated.Deployment{}, err
	}
	defer tx.Rollback(ctx)

	// Insert audit log (no actor_id since authentication is not implemented)
	if err := s.insertAuditLog(ctx, tx, deployment.ProjectID, uuid.Nil, "deployment.created", "deployment", deployment.ID, map[string]any{
		"projectId":     projectID.String(),
		"applicationId": applicationID.String(),
		"version":       version,
		"status":        deployment.Status,
	}); err != nil {
		return generated.Deployment{}, err
	}

	// Insert outbox event (configuration is not included to avoid storing sensitive data)
	if err := s.insertOutboxEvent(ctx, tx, "minicloud.deployment.created", "deployment", deployment.ID, map[string]any{
		"projectId":     projectID.String(),
		"applicationId": applicationID.String(),
		"version":       version,
		"status":        deployment.Status,
	}); err != nil {
		return generated.Deployment{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return generated.Deployment{}, err
	}

	return deployment, nil
}

func (s *Service) GetDeployment(ctx context.Context, id uuid.UUID) (generated.Deployment, error) {
	deployment, err := s.repository.GetDeployment(ctx, id)
	if err != nil {
		return generated.Deployment{}, ErrDeploymentNotFound
	}
	return deployment, nil
}

func (s *Service) ListDeploymentsByApplication(ctx context.Context, projectID, applicationID uuid.UUID, limit, offset int32) ([]generated.Deployment, error) {
	// Verify application exists and belongs to the project
	_, err := s.repository.GetApplication(ctx, applicationID, projectID)
	if err != nil {
		return nil, ErrInvalidRelationship
	}
	return s.repository.ListDeploymentsByApplication(ctx, projectID, applicationID, limit, offset)
}

func (s *Service) UpdateDeploymentStatus(ctx context.Context, id uuid.UUID, status string) (generated.Deployment, error) {
	// Validate status
	if err := s.Validator.ValidateDeploymentStatus(status); err != nil {
		return generated.Deployment{}, ErrValidation
	}

	deployment, err := s.repository.UpdateDeploymentStatus(ctx, id, status)
	if err != nil {
		return generated.Deployment{}, ErrDeploymentNotFound
	}

	// Start transaction for audit and outbox
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return generated.Deployment{}, err
	}
	defer tx.Rollback(ctx)

	// Insert audit log (no actor_id since authentication is not implemented)
	if err := s.insertAuditLog(ctx, tx, deployment.ProjectID, uuid.Nil, "deployment.status_updated", "deployment", deployment.ID, map[string]any{
		"projectId":     deployment.ProjectID.String(),
		"applicationId": deployment.ApplicationID.String(),
		"version":       deployment.Version,
		"status":        status,
	}); err != nil {
		return generated.Deployment{}, err
	}

	// Insert outbox event
	if err := s.insertOutboxEvent(ctx, tx, "minicloud.deployment.status_updated", "deployment", deployment.ID, map[string]any{
		"projectId":     deployment.ProjectID.String(),
		"applicationId": deployment.ApplicationID.String(),
		"version":       deployment.Version,
		"status":        status,
	}); err != nil {
		return generated.Deployment{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return generated.Deployment{}, err
	}

	return deployment, nil
}
