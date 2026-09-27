package controlplane

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minicloud/minicloud/apps/api/internal/database/generated"
)

type Repository struct {
	db *pgxpool.Pool
	q  *generated.Queries
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
		q:  generated.New(db),
	}
}

// Users
func (r *Repository) CreateUser(ctx context.Context, email, displayName string) (generated.User, error) {
	return r.q.CreateUser(ctx, generated.CreateUserParams{
		Email:       email,
		DisplayName: displayName,
	})
}

func (r *Repository) GetUser(ctx context.Context, id uuid.UUID) (generated.User, error) {
	return r.q.GetUser(ctx, id)
}

func (r *Repository) ListUsers(ctx context.Context, limit, offset int32) ([]generated.User, error) {
	return r.q.ListUsers(ctx, generated.ListUsersParams{
		Limit:  limit,
		Offset: offset,
	})
}

// Projects
func (r *Repository) CreateProject(ctx context.Context, ownerID uuid.UUID, name, slug, description string) (generated.Project, error) {
	return r.q.CreateProject(ctx, generated.CreateProjectParams{
		OwnerID:     ownerID,
		Name:        name,
		Slug:        slug,
		Description: description,
	})
}

func (r *Repository) GetProject(ctx context.Context, id uuid.UUID) (generated.Project, error) {
	return r.q.GetProject(ctx, id)
}

func (r *Repository) ListProjects(ctx context.Context, limit, offset int32) ([]generated.Project, error) {
	return r.q.ListProjects(ctx, generated.ListProjectsParams{
		Limit:  limit,
		Offset: offset,
	})
}

func (r *Repository) UpdateProject(ctx context.Context, id uuid.UUID, name, description, status string) (generated.Project, error) {
	return r.q.UpdateProject(ctx, generated.UpdateProjectParams{
		ID:          id,
		Name:        name,
		Description: description,
		Status:      status,
	})
}

func (r *Repository) DeleteProject(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteProject(ctx, id)
}

// Applications
func (r *Repository) CreateApplication(ctx context.Context, projectID uuid.UUID, name, slug, description string) (generated.Application, error) {
	return r.q.CreateApplication(ctx, generated.CreateApplicationParams{
		ProjectID:   projectID,
		Name:        name,
		Slug:        slug,
		Description: description,
	})
}

func (r *Repository) GetApplication(ctx context.Context, id, projectID uuid.UUID) (generated.Application, error) {
	return r.q.GetApplication(ctx, generated.GetApplicationParams{
		ID:        id,
		ProjectID: projectID,
	})
}

func (r *Repository) ListApplicationsByProject(ctx context.Context, projectID uuid.UUID, limit, offset int32) ([]generated.Application, error) {
	return r.q.ListApplicationsByProject(ctx, generated.ListApplicationsByProjectParams{
		ProjectID: projectID,
		Limit:     limit,
		Offset:    offset,
	})
}

func (r *Repository) UpdateApplication(ctx context.Context, id, projectID uuid.UUID, name, description, status string) (generated.Application, error) {
	return r.q.UpdateApplication(ctx, generated.UpdateApplicationParams{
		ID:          id,
		ProjectID:   projectID,
		Name:        name,
		Description: description,
		Status:      status,
	})
}

func (r *Repository) DeleteApplication(ctx context.Context, id, projectID uuid.UUID) error {
	return r.q.DeleteApplication(ctx, generated.DeleteApplicationParams{
		ID:        id,
		ProjectID: projectID,
	})
}

// Deployments
func (r *Repository) CreateDeployment(ctx context.Context, projectID, applicationID uuid.UUID, version string, configuration []byte) (generated.Deployment, error) {
	return r.q.CreateDeployment(ctx, generated.CreateDeploymentParams{
		ProjectID:     projectID,
		ApplicationID: applicationID,
		Version:       version,
		Configuration: configuration,
	})
}

func (r *Repository) GetDeployment(ctx context.Context, id uuid.UUID) (generated.Deployment, error) {
	return r.q.GetDeployment(ctx, id)
}

func (r *Repository) ListDeploymentsByApplication(ctx context.Context, projectID, applicationID uuid.UUID, limit, offset int32) ([]generated.Deployment, error) {
	return r.q.ListDeploymentsByApplication(ctx, generated.ListDeploymentsByApplicationParams{
		ProjectID:     projectID,
		ApplicationID: applicationID,
		Limit:         limit,
		Offset:        offset,
	})
}

func (r *Repository) UpdateDeploymentStatus(ctx context.Context, id uuid.UUID, status string) (generated.Deployment, error) {
	return r.q.UpdateDeploymentStatus(ctx, generated.UpdateDeploymentStatusParams{
		ID:     id,
		Status: status,
	})
}

// Audit and Outbox (kept as raw SQL for now)
func (r *Repository) InsertAuditLog(ctx context.Context, projectID, actorID uuid.UUID, action, resourceType string, resourceID uuid.UUID, metadata []byte) error {
	_, err := r.db.Exec(ctx, `INSERT INTO audit_logs(project_id,actor_id,action,resource_type,resource_id,metadata)VALUES($1,$2,$3,$4,$5,$6)`, projectID, actorID, action, resourceType, resourceID, metadata)
	return err
}

func (r *Repository) InsertOutboxEvent(ctx context.Context, eventType, aggregateType string, aggregateID uuid.UUID, payload []byte) error {
	_, err := r.db.Exec(ctx, `INSERT INTO outbox_events(event_type,aggregate_type,aggregate_id,payload)VALUES($1,$2,$3,$4)`, eventType, aggregateType, aggregateID, payload)
	return err
}
