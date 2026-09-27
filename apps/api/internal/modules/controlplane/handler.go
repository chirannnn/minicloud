package controlplane

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
	"strings"
)

type Handler struct {
	db      *pgxpool.Pool
	service *Service
}

func New(db *pgxpool.Pool) *Handler {
	return &Handler{
		db:      db,
		service: NewService(db),
	}
}
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
	user, err := h.service.CreateUser(r.Context(), v.Email, v.DisplayName)
	if err != nil {
		if err == ErrResourceConflict {
			out(w, 409, map[string]string{"code": "ALREADY_EXISTS", "message": "User already exists"})
		} else {
			out(w, 500, map[string]string{"code": "INTERNAL", "message": "Internal server error"})
		}
		return
	}
	out(w, 201, map[string]string{"id": user.ID.String(), "email": user.Email, "displayName": user.DisplayName})
}
func (h *Handler) listUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.ListUsers(r.Context(), 100, 0)
	if err != nil {
		out(w, 500, map[string]string{"code": "INTERNAL", "message": "Internal server error"})
		return
	}
	items := []any{}
	for _, user := range users {
		items = append(items, map[string]string{"id": user.ID.String(), "email": user.Email, "displayName": user.DisplayName})
	}
	out(w, 200, map[string]any{"items": items, "limit": 100, "offset": 0})
}
func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) {
	x, ok := id(r, "id")
	if !ok {
		bad(w, "invalid user id")
		return
	}
	user, err := h.service.GetUser(r.Context(), x)
	if err != nil {
		if err == ErrUserNotFound {
			missing(w)
		} else {
			out(w, 500, map[string]string{"code": "INTERNAL", "message": "Internal server error"})
		}
		return
	}
	out(w, 200, map[string]string{"id": user.ID.String(), "email": user.Email, "displayName": user.DisplayName})
}
func (h *Handler) createProject(w http.ResponseWriter, r *http.Request) {
	var v struct{ OwnerID, Name, Slug, Description string }
	if decode(r, &v) != nil || v.OwnerID == "" || v.Name == "" || !valid(v.Slug) {
		bad(w, "ownerId, name and valid slug are required")
		return
	}
	owner, _ := uuid.Parse(v.OwnerID)
	project, err := h.service.CreateProject(r.Context(), owner, v.Name, v.Slug, v.Description)
	if err != nil {
		if err == ErrUserNotFound {
			out(w, 404, map[string]string{"code": "NOT_FOUND", "message": "Owner not found"})
		} else if err == ErrResourceConflict {
			out(w, 409, map[string]string{"code": "CONFLICT", "message": "Project already exists"})
		} else {
			out(w, 500, map[string]string{"code": "INTERNAL", "message": "Internal server error"})
		}
		return
	}
	out(w, 201, map[string]string{"id": project.ID.String()})
}
func valid(s string) bool { return s != "" && len(s) < 64 && !strings.ContainsAny(s, " /\\") }
func (h *Handler) listProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := h.service.ListProjects(r.Context(), 100, 0)
	if err != nil {
		out(w, 500, map[string]string{"code": "INTERNAL", "message": "Internal server error"})
		return
	}
	items := []any{}
	for _, project := range projects {
		items = append(items, map[string]string{"id": project.ID.String(), "ownerId": project.OwnerID.String(), "name": project.Name, "slug": project.Slug, "description": project.Description, "status": project.Status})
	}
	out(w, 200, map[string]any{"items": items, "limit": 100, "offset": 0})
}
func (h *Handler) getProject(w http.ResponseWriter, r *http.Request) {
	x, ok := id(r, "id")
	if !ok {
		bad(w, "invalid project id")
		return
	}
	project, err := h.service.GetProject(r.Context(), x)
	if err != nil {
		if err == ErrProjectNotFound {
			missing(w)
		} else {
			out(w, 500, map[string]string{"code": "INTERNAL", "message": "Internal server error"})
		}
		return
	}
	out(w, 200, map[string]string{"id": project.ID.String(), "ownerId": project.OwnerID.String(), "name": project.Name, "slug": project.Slug, "description": project.Description, "status": project.Status})
}
func (h *Handler) listApps(w http.ResponseWriter, r *http.Request) {
	p, ok := id(r, "projectId")
	if !ok {
		bad(w, "invalid project id")
		return
	}
	apps, err := h.service.ListApplicationsByProject(r.Context(), p, 100, 0)
	if err != nil {
		if err == ErrProjectNotFound {
			missing(w)
		} else {
			out(w, 500, map[string]string{"code": "INTERNAL", "message": "Internal server error"})
		}
		return
	}
	items := []any{}
	for _, app := range apps {
		items = append(items, map[string]string{"id": app.ID.String(), "projectId": app.ProjectID.String(), "name": app.Name, "slug": app.Slug, "description": app.Description, "status": app.Status})
	}
	out(w, 200, map[string]any{"items": items})
}
func (h *Handler) createApp(w http.ResponseWriter, r *http.Request) {
	p, ok := id(r, "projectId")
	var v struct{ Name, Slug, Description string }
	if !ok || decode(r, &v) != nil || v.Name == "" || !valid(v.Slug) {
		bad(w, "name and valid slug required")
		return
	}
	app, err := h.service.CreateApplication(r.Context(), p, v.Name, v.Slug, v.Description)
	if err != nil {
		if err == ErrProjectNotFound {
			out(w, 404, map[string]string{"code": "NOT_FOUND", "message": "Project not found"})
		} else if err == ErrResourceConflict {
			out(w, 409, map[string]string{"code": "CONFLICT", "message": "Application conflict"})
		} else {
			out(w, 500, map[string]string{"code": "INTERNAL", "message": "Internal server error"})
		}
		return
	}
	out(w, 201, map[string]string{"id": app.ID.String()})
}
func (h *Handler) getApp(w http.ResponseWriter, r *http.Request) {
	p, _ := id(r, "projectId")
	a, _ := id(r, "applicationId")
	app, err := h.service.GetApplication(r.Context(), a, p)
	if err != nil {
		if err == ErrApplicationNotFound {
			missing(w)
		} else {
			out(w, 500, map[string]string{"code": "INTERNAL", "message": "Internal server error"})
		}
		return
	}
	out(w, 200, map[string]string{"id": app.ID.String(), "projectId": app.ProjectID.String(), "name": app.Name, "slug": app.Slug, "description": app.Description, "status": app.Status})
}
func (h *Handler) listDeployments(w http.ResponseWriter, r *http.Request) {
	p, _ := id(r, "projectId")
	a, _ := id(r, "applicationId")
	deployments, err := h.service.ListDeploymentsByApplication(r.Context(), p, a, 100, 0)
	if err != nil {
		if err == ErrInvalidRelationship {
			out(w, 400, map[string]string{"code": "INVALID_RELATIONSHIP", "message": "Invalid application/project relationship"})
		} else {
			out(w, 500, map[string]string{"code": "INTERNAL", "message": "Internal server error"})
		}
		return
	}
	items := []any{}
	for _, deployment := range deployments {
		items = append(items, map[string]string{"id": deployment.ID.String(), "version": deployment.Version, "status": deployment.Status})
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
	deployment, err := h.service.CreateDeployment(r.Context(), p, a, v.Version, v.Configuration)
	if err != nil {
		if err == ErrInvalidRelationship {
			out(w, 400, map[string]string{"code": "INVALID_RELATIONSHIP", "message": "Invalid application/project relationship"})
		} else if err == ErrResourceConflict {
			out(w, 409, map[string]string{"code": "CONFLICT", "message": "Deployment conflict"})
		} else {
			missing(w)
		}
		return
	}
	out(w, 201, map[string]string{"id": deployment.ID.String()})
}
func (h *Handler) getDeployment(w http.ResponseWriter, r *http.Request) {
	d, _ := id(r, "id")
	deployment, err := h.service.GetDeployment(r.Context(), d)
	if err != nil {
		if err == ErrDeploymentNotFound {
			missing(w)
		} else {
			out(w, 500, map[string]string{"code": "INTERNAL", "message": "Internal server error"})
		}
		return
	}
	out(w, 200, map[string]string{"id": deployment.ID.String(), "projectId": deployment.ProjectID.String(), "applicationId": deployment.ApplicationID.String(), "version": deployment.Version, "status": deployment.Status})
}
func (h *Handler) updateProject(w http.ResponseWriter, r *http.Request) {
	p, ok := id(r, "id")
	var v struct{ Name, Description, Status string }
	if !ok || decode(r, &v) != nil || v.Name == "" {
		bad(w, "name required")
		return
	}
	_, err := h.service.UpdateProject(r.Context(), p, v.Name, v.Description, v.Status)
	if err != nil {
		if err == ErrProjectNotFound {
			missing(w)
		} else {
			out(w, 500, map[string]string{"code": "INTERNAL", "message": "Internal server error"})
		}
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
	err := h.service.DeleteProject(r.Context(), p)
	if err != nil {
		if err == ErrProjectNotFound {
			missing(w)
		} else {
			out(w, 409, map[string]string{"code": "CONFLICT", "message": "Project has child resources"})
		}
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
	_, err := h.service.UpdateApplication(r.Context(), a, p, v.Name, v.Description, v.Status)
	if err != nil {
		if err == ErrApplicationNotFound {
			missing(w)
		} else {
			out(w, 500, map[string]string{"code": "INTERNAL", "message": "Internal server error"})
		}
		return
	}
	out(w, 200, map[string]string{"id": a.String()})
}
func (h *Handler) deleteApp(w http.ResponseWriter, r *http.Request) {
	p, _ := id(r, "projectId")
	a, _ := id(r, "applicationId")
	err := h.service.DeleteApplication(r.Context(), a, p)
	if err != nil {
		if err == ErrApplicationNotFound {
			missing(w)
		} else {
			out(w, 409, map[string]string{"code": "CONFLICT", "message": "Application has deployments"})
		}
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
	_, err := h.service.UpdateDeploymentStatus(r.Context(), d, v.Status)
	if err != nil {
		if err == ErrDeploymentNotFound {
			missing(w)
		} else if err == ErrInvalidStatus {
			bad(w, "invalid deployment status")
		} else {
			out(w, 500, map[string]string{"code": "INTERNAL", "message": "Internal server error"})
		}
		return
	}
	out(w, 200, map[string]string{"id": d.String(), "status": v.Status})
}
