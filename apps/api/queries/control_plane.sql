-- name: CreateUser :one
INSERT INTO users (email, display_name) VALUES ($1, $2) RETURNING *;
-- name: GetUser :one
SELECT * FROM users WHERE id = $1;
-- name: ListUsers :many
SELECT * FROM users ORDER BY created_at LIMIT $1 OFFSET $2;

-- name: CreateProject :one
INSERT INTO projects (owner_id, name, slug, description) VALUES ($1, $2, $3, $4) RETURNING *;
-- name: GetProject :one
SELECT * FROM projects WHERE id = $1;
-- name: ListProjects :many
SELECT * FROM projects ORDER BY created_at LIMIT $1 OFFSET $2;
-- name: UpdateProject :one
UPDATE projects SET name = $2, description = $3, status = COALESCE(NULLIF($4, ''), status), updated_at = now() WHERE id = $1 RETURNING *;
-- name: DeleteProject :exec
DELETE FROM projects WHERE id = $1;

-- name: CreateApplication :one
INSERT INTO applications (project_id, name, slug, description) VALUES ($1, $2, $3, $4) RETURNING *;
-- name: GetApplication :one
SELECT * FROM applications WHERE id = $1 AND project_id = $2;
-- name: ListApplicationsByProject :many
SELECT * FROM applications WHERE project_id = $1 ORDER BY created_at LIMIT $2 OFFSET $3;
-- name: UpdateApplication :one
UPDATE applications SET name = $3, description = $4, status = COALESCE(NULLIF($5, ''), status), updated_at = now() WHERE id = $1 AND project_id = $2 RETURNING *;
-- name: DeleteApplication :exec
DELETE FROM applications WHERE id = $1 AND project_id = $2;

-- name: CreateDeployment :one
INSERT INTO deployments (project_id, application_id, version, configuration) SELECT $1, $2, $3, $4 WHERE EXISTS(SELECT 1 FROM applications WHERE id = $2 AND project_id = $1) RETURNING *;
-- name: GetDeployment :one
SELECT * FROM deployments WHERE id = $1;
-- name: ListDeploymentsByApplication :many
SELECT * FROM deployments WHERE project_id = $1 AND application_id = $2 ORDER BY created_at LIMIT $3 OFFSET $4;
-- name: UpdateDeploymentStatus :one
UPDATE deployments SET status = $2, updated_at = now() WHERE id = $1 RETURNING *;
