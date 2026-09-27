package controlplane

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minicloud/minicloud/apps/api/internal/database/generated"
)

// Service layer errors
var (
	ErrUserNotFound        = errors.New("user not found")
	ErrProjectNotFound     = errors.New("project not found")
	ErrApplicationNotFound = errors.New("application not found")
	ErrDeploymentNotFound  = errors.New("deployment not found")
	ErrInvalidRelationship = errors.New("invalid resource relationship")
	ErrInvalidStatus       = errors.New("invalid status transition")
	ErrResourceConflict    = errors.New("resource already exists")
)

type Service struct {
	db         *pgxpool.Pool
	repository *Repository
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{
		db:         db,
		repository: NewRepository(db),
	}
}

// Users

func (s *Service) CreateUser(ctx context.Context, email, displayName string) (generated.User, error) {
	user, err := s.repository.CreateUser(ctx, email, displayName)
	if err != nil {
		return generated.User{}, ErrResourceConflict
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
	_, err = tx.Exec(ctx, `INSERT INTO audit_logs(project_id,actor_id,action,resource_type,resource_id)VALUES($1,$2,'project.created','project',$1)`, project.ID, ownerID)
	if err != nil {
		return generated.Project{}, err
	}

	// Insert outbox event
	_, err = tx.Exec(ctx, `INSERT INTO outbox_events(event_type,aggregate_type,aggregate_id,payload)VALUES('minicloud.project.created','project',$1,$2)`, project.ID, []byte(`{}`))
	if err != nil {
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
	project, err := s.repository.UpdateProject(ctx, id, name, description, status)
	if err != nil {
		return generated.Project{}, ErrProjectNotFound
	}
	return project, nil
}

func (s *Service) DeleteProject(ctx context.Context, id uuid.UUID) error {
	err := s.repository.DeleteProject(ctx, id)
	if err != nil {
		return ErrProjectNotFound
	}
	return nil
}

// Applications

func (s *Service) CreateApplication(ctx context.Context, projectID uuid.UUID, name, slug, description string) (generated.Application, error) {
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

	// Insert audit log
	_, err = tx.Exec(ctx, `INSERT INTO audit_logs(project_id,action,resource_type,resource_id)VALUES($1,'application.created','application',$2)`, projectID, app.ID)
	if err != nil {
		return generated.Application{}, err
	}

	// Insert outbox event
	_, err = tx.Exec(ctx, `INSERT INTO outbox_events(event_type,aggregate_type,aggregate_id,payload)VALUES('minicloud.application.created','application',$1,'{}')`, app.ID)
	if err != nil {
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
	app, err := s.repository.UpdateApplication(ctx, id, projectID, name, description, status)
	if err != nil {
		return generated.Application{}, ErrApplicationNotFound
	}
	return app, nil
}

func (s *Service) DeleteApplication(ctx context.Context, id, projectID uuid.UUID) error {
	err := s.repository.DeleteApplication(ctx, id, projectID)
	if err != nil {
		return ErrApplicationNotFound
	}
	return nil
}

// Deployments

func (s *Service) CreateDeployment(ctx context.Context, projectID, applicationID uuid.UUID, version string, configuration []byte) (generated.Deployment, error) {
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
	validStatuses := map[string]bool{
		"PENDING":   true,
		"QUEUED":    true,
		"RUNNING":   true,
		"SUCCEEDED": true,
		"FAILED":    true,
		"CANCELLED": true,
	}
	if !validStatuses[status] {
		return generated.Deployment{}, ErrInvalidStatus
	}

	deployment, err := s.repository.UpdateDeploymentStatus(ctx, id, status)
	if err != nil {
		return generated.Deployment{}, ErrDeploymentNotFound
	}
	return deployment, nil
}
