package controlplane

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
	"strings"
)

type Handler struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Handler { return &Handler{db} }
func (h *Handler) Register(m *http.ServeMux) {
	m.HandleFunc("GET /api/v1/users", h.listUsers)
	m.HandleFunc("POST /api/v1/users", h.createUser)
	m.HandleFunc("GET /api/v1/users/{id}", h.getUser)
	m.HandleFunc("GET /api/v1/projects", h.listProjects)
	m.HandleFunc("POST /api/v1/projects", h.createProject)
	m.HandleFunc("GET /api/v1/projects/{id}", h.getProject)
	m.HandleFunc("PATCH /api/v1/projects/{id}", h.updateProject)
	m.HandleFunc("DELETE /api/v1/projects/{id}", h.deleteProject)
	m.HandleFunc("GET /api/v1/projects/{projectId}/applications", h.listApps)
	m.HandleFunc("POST /api/v1/projects/{projectId}/applications", h.createApp)
	m.HandleFunc("GET /api/v1/projects/{projectId}/applications/{applicationId}", h.getApp)
	m.HandleFunc("PATCH /api/v1/projects/{projectId}/applications/{applicationId}", h.updateApp)
	m.HandleFunc("DELETE /api/v1/projects/{projectId}/applications/{applicationId}", h.deleteApp)
	m.HandleFunc("GET /api/v1/projects/{projectId}/applications/{applicationId}/deployments", h.listDeployments)
	m.HandleFunc("POST /api/v1/projects/{projectId}/applications/{applicationId}/deployments", h.createDeployment)
	m.HandleFunc("GET /api/v1/deployments/{id}", h.getDeployment)
	m.HandleFunc("PATCH /api/v1/deployments/{id}/status", h.updateDeploymentStatus)
}
func decode(r *http.Request, v any) error { return json.NewDecoder(r.Body).Decode(v) }
func out(w http.ResponseWriter, s int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(s)
	_ = json.NewEncoder(w).Encode(v)
}
func id(r *http.Request, k string) (uuid.UUID, bool) {
	x, e := uuid.Parse(r.PathValue(k))
	return x, e == nil
}
func bad(w http.ResponseWriter, m string) {
	out(w, 400, map[string]string{"code": "VALIDATION_FAILED", "message": m})
}
func missing(w http.ResponseWriter) {
	out(w, 404, map[string]string{"code": "NOT_FOUND", "message": "Resource was not found"})
}
func (h *Handler) createUser(w http.ResponseWriter, r *http.Request) {
	var v struct{ Email, DisplayName string }
	if decode(r, &v) != nil || !strings.Contains(v.Email, "@") || v.DisplayName == "" {
		bad(w, "email and displayName are required")
		return
	}
	var x struct{ ID, Email, DisplayName string }
	e := h.db.QueryRow(r.Context(), `INSERT INTO users(email,display_name)VALUES($1,$2) RETURNING id,email,display_name`, v.Email, v.DisplayName).Scan(&x.ID, &x.Email, &x.DisplayName)
	if e != nil {
		out(w, 409, map[string]string{"code": "ALREADY_EXISTS", "message": "User already exists"})
		return
	}
	out(w, 201, x)
}
func (h *Handler) listUsers(w http.ResponseWriter, r *http.Request) {
	rows, e := h.db.Query(r.Context(), `SELECT id,email,display_name FROM users ORDER BY created_at LIMIT 100`)
	if e != nil {
		out(w, 500, map[string]string{"code": "INTERNAL", "message": "Internal server error"})
		return
	}
	defer rows.Close()
	a := []any{}
	for rows.Next() {
		var x struct{ ID, Email, DisplayName string }
		_ = rows.Scan(&x.ID, &x.Email, &x.DisplayName)
		a = append(a, x)
	}
	out(w, 200, map[string]any{"items": a, "limit": 100, "offset": 0})
}
func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) {
	x, ok := id(r, "id")
	if !ok {
		bad(w, "invalid user id")
		return
	}
	var v struct{ ID, Email, DisplayName string }
	if h.db.QueryRow(r.Context(), `SELECT id,email,display_name FROM users WHERE id=$1`, x).Scan(&v.ID, &v.Email, &v.DisplayName) != nil {
		missing(w)
		return
	}
	out(w, 200, v)
}
func (h *Handler) createProject(w http.ResponseWriter, r *http.Request) {
	var v struct{ OwnerID, Name, Slug, Description string }
	if decode(r, &v) != nil || v.OwnerID == "" || v.Name == "" || !valid(v.Slug) {
		bad(w, "ownerId, name and valid slug are required")
		return
	}
	owner, _ := uuid.Parse(v.OwnerID)
	tx, e := h.db.Begin(r.Context())
	if e != nil {
		out(w, 500, nil)
		return
	}
	defer tx.Rollback(r.Context())
	var x struct{ ID string }
	e = tx.QueryRow(r.Context(), `INSERT INTO projects(owner_id,name,slug,description)VALUES($1,$2,$3,$4)RETURNING id`, owner, v.Name, v.Slug, v.Description).Scan(&x.ID)
	if e != nil {
		out(w, 409, map[string]string{"code": "CONFLICT", "message": "Project already exists or owner is invalid"})
		return
	}
	_, e = tx.Exec(r.Context(), `INSERT INTO audit_logs(project_id,actor_id,action,resource_type,resource_id)VALUES($1,$2,'project.created','project',$1)`, x.ID, owner)
	if e == nil {
		_, e = tx.Exec(r.Context(), `INSERT INTO outbox_events(event_type,aggregate_type,aggregate_id,payload)VALUES('minicloud.project.created','project',$1,$2)`, x.ID, []byte(`{}`))
	}
	if e != nil {
		out(w, 500, map[string]string{"code": "INTERNAL", "message": "Internal server error"})
		return
	}
	_ = tx.Commit(r.Context())
	out(w, 201, x)
}
func valid(s string) bool { return s != "" && len(s) < 64 && !strings.ContainsAny(s, " /\\") }
func (h *Handler) listProjects(w http.ResponseWriter, r *http.Request) {
	rows, _ := h.db.Query(r.Context(), `SELECT id,owner_id,name,slug,description,status FROM projects ORDER BY created_at LIMIT 100`)
	defer rows.Close()
	a := []any{}
	for rows.Next() {
		var x struct{ ID, OwnerID, Name, Slug, Description, Status string }
		_ = rows.Scan(&x.ID, &x.OwnerID, &x.Name, &x.Slug, &x.Description, &x.Status)
		a = append(a, x)
	}
	out(w, 200, map[string]any{"items": a, "limit": 100, "offset": 0})
}
func (h *Handler) getProject(w http.ResponseWriter, r *http.Request) {
	x, ok := id(r, "id")
	if !ok {
		bad(w, "invalid project id")
		return
	}
	var v struct{ ID, OwnerID, Name, Slug, Description, Status string }
	if h.db.QueryRow(r.Context(), `SELECT id,owner_id,name,slug,description,status FROM projects WHERE id=$1`, x).Scan(&v.ID, &v.OwnerID, &v.Name, &v.Slug, &v.Description, &v.Status) != nil {
		missing(w)
		return
	}
	out(w, 200, v)
}
func (h *Handler) listApps(w http.ResponseWriter, r *http.Request) {
	p, ok := id(r, "projectId")
	if !ok {
		bad(w, "invalid project id")
		return
	}
	rows, e := h.db.Query(r.Context(), `SELECT id,project_id,name,slug,description,status FROM applications WHERE project_id=$1`, p)
	if e != nil {
		missing(w)
		return
	}
	defer rows.Close()
	a := []any{}
	for rows.Next() {
		var x struct{ ID, ProjectID, Name, Slug, Description, Status string }
		_ = rows.Scan(&x.ID, &x.ProjectID, &x.Name, &x.Slug, &x.Description, &x.Status)
		a = append(a, x)
	}
	out(w, 200, map[string]any{"items": a})
}
func (h *Handler) createApp(w http.ResponseWriter, r *http.Request) {
	p, ok := id(r, "projectId")
	var v struct{ Name, Slug, Description string }
	if !ok || decode(r, &v) != nil || v.Name == "" || !valid(v.Slug) {
		bad(w, "name and valid slug required")
		return
	}
	tx, _ := h.db.Begin(r.Context())
	defer tx.Rollback(r.Context())
	var aid string
	e := tx.QueryRow(r.Context(), `INSERT INTO applications(project_id,name,slug,description)VALUES($1,$2,$3,$4)RETURNING id`, p, v.Name, v.Slug, v.Description).Scan(&aid)
	if e != nil {
		out(w, 409, map[string]string{"code": "CONFLICT", "message": "Application conflict or project missing"})
		return
	}
	_, e = tx.Exec(r.Context(), `INSERT INTO audit_logs(project_id,action,resource_type,resource_id)VALUES($1,'application.created','application',$2)`, p, aid)
	if e == nil {
		_, e = tx.Exec(r.Context(), `INSERT INTO outbox_events(event_type,aggregate_type,aggregate_id,payload)VALUES('minicloud.application.created','application',$1,'{}')`, aid)
	}
	if e != nil {
		out(w, 500, nil)
		return
	}
	_ = tx.Commit(r.Context())
	out(w, 201, map[string]string{"id": aid})
}
func (h *Handler) getApp(w http.ResponseWriter, r *http.Request) {
	p, _ := id(r, "projectId")
	a, _ := id(r, "applicationId")
	var v struct{ ID, ProjectID, Name, Slug, Description, Status string }
	if h.db.QueryRow(r.Context(), `SELECT id,project_id,name,slug,description,status FROM applications WHERE id=$1 AND project_id=$2`, a, p).Scan(&v.ID, &v.ProjectID, &v.Name, &v.Slug, &v.Description, &v.Status) != nil {
		missing(w)
		return
	}
	out(w, 200, v)
}
func (h *Handler) listDeployments(w http.ResponseWriter, r *http.Request) {
	p, _ := id(r, "projectId")
	a, _ := id(r, "applicationId")
	rows, _ := h.db.Query(r.Context(), `SELECT id,version,status FROM deployments WHERE project_id=$1 AND application_id=$2`, p, a)
	defer rows.Close()
	items := []any{}
	for rows.Next() {
		var v struct{ ID, Version, Status string }
		_ = rows.Scan(&v.ID, &v.Version, &v.Status)
		items = append(items, v)
	}
	out(w, 200, map[string]any{"items": items})
}
func (h *Handler) createDeployment(w http.ResponseWriter, r *http.Request) {
	p, po := id(r, "projectId")
	a, ao := id(r, "applicationId")
	var v struct {
		Version       string
		Configuration json.RawMessage
	}
	if !po || !ao || decode(r, &v) != nil || v.Version == "" {
		bad(w, "version required")
		return
	}
	var did string
	e := h.db.QueryRow(r.Context(), `INSERT INTO deployments(project_id,application_id,version,configuration) SELECT $1,$2,$3,$4 WHERE EXISTS(SELECT 1 FROM applications WHERE id=$2 AND project_id=$1) RETURNING id`, p, a, v.Version, v.Configuration).Scan(&did)
	if e != nil {
		missing(w)
		return
	}
	out(w, 201, map[string]string{"id": did})
}
func (h *Handler) getDeployment(w http.ResponseWriter, r *http.Request) {
	d, _ := id(r, "id")
	var v struct{ ID, ProjectID, ApplicationID, Version, Status string }
	if h.db.QueryRow(r.Context(), `SELECT id,project_id,application_id,version,status FROM deployments WHERE id=$1`, d).Scan(&v.ID, &v.ProjectID, &v.ApplicationID, &v.Version, &v.Status) != nil {
		missing(w)
		return
	}
	out(w, 200, v)
}
func (h *Handler) updateProject(w http.ResponseWriter, r *http.Request) {
	p, ok := id(r, "id")
	var v struct{ Name, Description, Status string }
	if !ok || decode(r, &v) != nil || v.Name == "" {
		bad(w, "name required")
		return
	}
	tag, e := h.db.Exec(r.Context(), `UPDATE projects SET name=$2,description=$3,status=COALESCE(NULLIF($4,''),status),updated_at=now() WHERE id=$1`, p, v.Name, v.Description, v.Status)
	if e != nil {
		out(w, 500, map[string]string{"code": "INTERNAL", "message": "Internal server error"})
		return
	}
	if tag.RowsAffected() == 0 {
		missing(w)
		return
	}
	out(w, 200, map[string]string{"id": p.String()})
}
func (h *Handler) deleteProject(w http.ResponseWriter, r *http.Request) {
	p, ok := id(r, "id")
	if !ok {
		bad(w, "invalid project id")
		return
	}
	tag, e := h.db.Exec(r.Context(), `DELETE FROM projects WHERE id=$1`, p)
	if e != nil {
		out(w, 409, map[string]string{"code": "CONFLICT", "message": "Project has child resources"})
		return
	}
	if tag.RowsAffected() == 0 {
		missing(w)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (h *Handler) updateApp(w http.ResponseWriter, r *http.Request) {
	p, _ := id(r, "projectId")
	a, _ := id(r, "applicationId")
	var v struct{ Name, Description, Status string }
	if decode(r, &v) != nil || v.Name == "" {
		bad(w, "name required")
		return
	}
	tag, _ := h.db.Exec(r.Context(), `UPDATE applications SET name=$3,description=$4,status=COALESCE(NULLIF($5,''),status),updated_at=now() WHERE id=$1 AND project_id=$2`, a, p, v.Name, v.Description, v.Status)
	if tag.RowsAffected() == 0 {
		missing(w)
		return
	}
	out(w, 200, map[string]string{"id": a.String()})
}
func (h *Handler) deleteApp(w http.ResponseWriter, r *http.Request) {
	p, _ := id(r, "projectId")
	a, _ := id(r, "applicationId")
	tag, e := h.db.Exec(r.Context(), `DELETE FROM applications WHERE id=$1 AND project_id=$2`, a, p)
	if e != nil {
		out(w, 409, map[string]string{"code": "CONFLICT", "message": "Application has deployments"})
		return
	}
	if tag.RowsAffected() == 0 {
		missing(w)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (h *Handler) updateDeploymentStatus(w http.ResponseWriter, r *http.Request) {
	d, ok := id(r, "id")
	var v struct{ Status string }
	if !ok || decode(r, &v) != nil || !map[string]bool{"PENDING": true, "QUEUED": true, "RUNNING": true, "SUCCEEDED": true, "FAILED": true, "CANCELLED": true}[v.Status] {
		bad(w, "invalid deployment status")
		return
	}
	tag, _ := h.db.Exec(r.Context(), `UPDATE deployments SET status=$2,updated_at=now() WHERE id=$1`, d, v.Status)
	if tag.RowsAffected() == 0 {
		missing(w)
		return
	}
	out(w, 200, map[string]string{"id": d.String(), "status": v.Status})
}
