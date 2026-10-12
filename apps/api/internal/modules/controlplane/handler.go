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
	out(w, 400, HTTPError{Code: CodeValidationFailed, Message: m})
}

func missing(w http.ResponseWriter) {
	out(w, 404, HTTPError{Code: CodeNotFound, Message: "Resource was not found"})
}
func (h *Handler) createUser(w http.ResponseWriter, r *http.Request) {
	var v struct{ Email, DisplayName string }
	if decode(r, &v) != nil {
		bad(w, "invalid request body")
		return
	}

	user, err := h.service.CreateUser(r.Context(), v.Email, v.DisplayName)
	if err != nil {
		httpErr := ToHTTPError(err)
		status := 500
		if httpErr.Code == CodeValidationFailed {
			status = 400
		} else if httpErr.Code == CodeConflict {
			status = 409
		}
		out(w, status, httpErr)
		return
	}
	out(w, 201, map[string]string{"id": user.ID.String(), "email": user.Email, "displayName": user.DisplayName})
}
func (h *Handler) listUsers(w http.ResponseWriter, r *http.Request) {
	// Parse and validate pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit, offset, validationErr := h.service.Validator.ValidatePagination(limitStr, offsetStr)
	if validationErr != nil {
		bad(w, validationErr.Message)
		return
	}

	users, err := h.service.ListUsers(r.Context(), limit, offset)
	if err != nil {
		httpErr := ToHTTPError(err)
		out(w, 500, httpErr)
		return
	}
	items := []any{}
	for _, user := range users {
		items = append(items, map[string]string{"id": user.ID.String(), "email": user.Email, "displayName": user.DisplayName})
	}
	out(w, 200, map[string]any{"items": items, "limit": limit, "offset": offset})
}
func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) {
	x, ok := id(r, "id")
	if !ok {
		bad(w, "invalid user id")
		return
	}
	user, err := h.service.GetUser(r.Context(), x)
	if err != nil {
		httpErr := ToHTTPError(err)
		status := 500
		if httpErr.Code == CodeNotFound {
			status = 404
		}
		out(w, status, httpErr)
		return
	}
	out(w, 200, map[string]string{"id": user.ID.String(), "email": user.Email, "displayName": user.DisplayName})
}
func (h *Handler) createProject(w http.ResponseWriter, r *http.Request) {
	var v struct{ OwnerID, Name, Slug, Description string }
	if decode(r, &v) != nil {
		bad(w, "invalid request body")
		return
	}

	owner, err := uuid.Parse(v.OwnerID)
	if err != nil {
		bad(w, "invalid owner id")
		return
	}

	project, err := h.service.CreateProject(r.Context(), owner, v.Name, v.Slug, v.Description)
	if err != nil {
		httpErr := ToHTTPError(err)
		status := 500
		if httpErr.Code == CodeValidationFailed {
			status = 400
		} else if httpErr.Code == CodeNotFound {
			status = 404
		} else if httpErr.Code == CodeConflict {
			status = 409
		}
		out(w, status, httpErr)
		return
	}
	out(w, 201, map[string]string{"id": project.ID.String()})
}
func valid(s string) bool { return s != "" && len(s) < 64 && !strings.ContainsAny(s, " /\\") }
func (h *Handler) listProjects(w http.ResponseWriter, r *http.Request) {
	// Parse and validate pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit, offset, validationErr := h.service.Validator.ValidatePagination(limitStr, offsetStr)
	if validationErr != nil {
		bad(w, validationErr.Message)
		return
	}

	projects, err := h.service.ListProjects(r.Context(), limit, offset)
	if err != nil {
		httpErr := ToHTTPError(err)
		out(w, 500, httpErr)
		return
	}
	items := []any{}
	for _, project := range projects {
		items = append(items, map[string]string{"id": project.ID.String(), "ownerId": project.OwnerID.String(), "name": project.Name, "slug": project.Slug, "description": project.Description, "status": project.Status})
	}
	out(w, 200, map[string]any{"items": items, "limit": limit, "offset": offset})
}
func (h *Handler) getProject(w http.ResponseWriter, r *http.Request) {
	x, ok := id(r, "id")
	if !ok {
		bad(w, "invalid project id")
		return
	}
	project, err := h.service.GetProject(r.Context(), x)
	if err != nil {
		httpErr := ToHTTPError(err)
		status := 500
		if httpErr.Code == CodeNotFound {
			status = 404
		}
		out(w, status, httpErr)
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

	// Parse and validate pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit, offset, validationErr := h.service.Validator.ValidatePagination(limitStr, offsetStr)
	if validationErr != nil {
		bad(w, validationErr.Message)
		return
	}

	apps, err := h.service.ListApplicationsByProject(r.Context(), p, limit, offset)
	if err != nil {
		httpErr := ToHTTPError(err)
		status := 500
		if httpErr.Code == CodeNotFound {
			status = 404
		}
		out(w, status, httpErr)
		return
	}
	items := []any{}
	for _, app := range apps {
		items = append(items, map[string]string{"id": app.ID.String(), "projectId": app.ProjectID.String(), "name": app.Name, "slug": app.Slug, "description": app.Description, "status": app.Status})
	}
	out(w, 200, map[string]any{"items": items, "limit": limit, "offset": offset})
}
func (h *Handler) createApp(w http.ResponseWriter, r *http.Request) {
	p, ok := id(r, "projectId")
	var v struct{ Name, Slug, Description string }
	if !ok || decode(r, &v) != nil {
		bad(w, "invalid request or project id")
		return
	}

	app, err := h.service.CreateApplication(r.Context(), p, v.Name, v.Slug, v.Description)
	if err != nil {
		httpErr := ToHTTPError(err)
		status := 500
		if httpErr.Code == CodeValidationFailed {
			status = 400
		} else if httpErr.Code == CodeNotFound {
			status = 404
		} else if httpErr.Code == CodeConflict {
			status = 409
		}
		out(w, status, httpErr)
		return
	}
	out(w, 201, map[string]string{"id": app.ID.String()})
}
func (h *Handler) getApp(w http.ResponseWriter, r *http.Request) {
	p, _ := id(r, "projectId")
	a, _ := id(r, "applicationId")
	app, err := h.service.GetApplication(r.Context(), a, p)
	if err != nil {
		httpErr := ToHTTPError(err)
		status := 500
		if httpErr.Code == CodeNotFound {
			status = 404
		}
		out(w, status, httpErr)
		return
	}
	out(w, 200, map[string]string{"id": app.ID.String(), "projectId": app.ProjectID.String(), "name": app.Name, "slug": app.Slug, "description": app.Description, "status": app.Status})
}
func (h *Handler) listDeployments(w http.ResponseWriter, r *http.Request) {
	p, _ := id(r, "projectId")
	a, _ := id(r, "applicationId")

	// Parse and validate pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit, offset, validationErr := h.service.Validator.ValidatePagination(limitStr, offsetStr)
	if validationErr != nil {
		bad(w, validationErr.Message)
		return
	}

	deployments, err := h.service.ListDeploymentsByApplication(r.Context(), p, a, limit, offset)
	if err != nil {
		httpErr := ToHTTPError(err)
		status := 500
		if httpErr.Code == CodeInvalidRelationship {
			status = 400
		}
		out(w, status, httpErr)
		return
	}
	items := []any{}
	for _, deployment := range deployments {
		items = append(items, map[string]string{"id": deployment.ID.String(), "version": deployment.Version, "status": deployment.Status})
	}
	out(w, 200, map[string]any{"items": items, "limit": limit, "offset": offset})
}
func (h *Handler) createDeployment(w http.ResponseWriter, r *http.Request) {
	p, po := id(r, "projectId")
	a, ao := id(r, "applicationId")
	var v struct {
		Version       string
		Configuration json.RawMessage
	}
	if !po || !ao || decode(r, &v) != nil {
		bad(w, "invalid request or missing ids")
		return
	}

	deployment, err := h.service.CreateDeployment(r.Context(), p, a, v.Version, v.Configuration)
	if err != nil {
		httpErr := ToHTTPError(err)
		status := 500
		if httpErr.Code == CodeValidationFailed {
			status = 400
		} else if httpErr.Code == CodeInvalidRelationship {
			status = 400
		} else if httpErr.Code == CodeConflict {
			status = 409
		} else if httpErr.Code == CodeNotFound {
			status = 404
		}
		out(w, status, httpErr)
		return
	}
	out(w, 201, map[string]string{"id": deployment.ID.String()})
}
func (h *Handler) getDeployment(w http.ResponseWriter, r *http.Request) {
	d, _ := id(r, "id")
	deployment, err := h.service.GetDeployment(r.Context(), d)
	if err != nil {
		httpErr := ToHTTPError(err)
		status := 500
		if httpErr.Code == CodeNotFound {
			status = 404
		}
		out(w, status, httpErr)
		return
	}
	out(w, 200, map[string]string{"id": deployment.ID.String(), "projectId": deployment.ProjectID.String(), "applicationId": deployment.ApplicationID.String(), "version": deployment.Version, "status": deployment.Status})
}
func (h *Handler) updateProject(w http.ResponseWriter, r *http.Request) {
	p, ok := id(r, "id")
	var v struct{ Name, Description, Status string }
	if !ok || decode(r, &v) != nil {
		bad(w, "invalid request or project id")
		return
	}
	_, err := h.service.UpdateProject(r.Context(), p, v.Name, v.Description, v.Status)
	if err != nil {
		httpErr := ToHTTPError(err)
		status := 500
		if httpErr.Code == CodeNotFound {
			status = 404
		}
		out(w, status, httpErr)
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
		httpErr := ToHTTPError(err)
		status := 500
		if httpErr.Code == CodeNotFound {
			status = 404
		} else if httpErr.Code == CodeConflict {
			status = 409
		}
		out(w, status, httpErr)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (h *Handler) updateApp(w http.ResponseWriter, r *http.Request) {
	p, _ := id(r, "projectId")
	a, _ := id(r, "applicationId")
	var v struct{ Name, Description, Status string }
	if decode(r, &v) != nil {
		bad(w, "invalid request body")
		return
	}
	_, err := h.service.UpdateApplication(r.Context(), a, p, v.Name, v.Description, v.Status)
	if err != nil {
		httpErr := ToHTTPError(err)
		status := 500
		if httpErr.Code == CodeNotFound {
			status = 404
		}
		out(w, status, httpErr)
		return
	}
	out(w, 200, map[string]string{"id": a.String()})
}
func (h *Handler) deleteApp(w http.ResponseWriter, r *http.Request) {
	p, _ := id(r, "projectId")
	a, _ := id(r, "applicationId")
	err := h.service.DeleteApplication(r.Context(), a, p)
	if err != nil {
		httpErr := ToHTTPError(err)
		status := 500
		if httpErr.Code == CodeNotFound {
			status = 404
		} else if httpErr.Code == CodeConflict {
			status = 409
		}
		out(w, status, httpErr)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (h *Handler) updateDeploymentStatus(w http.ResponseWriter, r *http.Request) {
	d, ok := id(r, "id")
	var v struct{ Status string }
	if !ok || decode(r, &v) != nil {
		bad(w, "invalid request or deployment id")
		return
	}
	_, err := h.service.UpdateDeploymentStatus(r.Context(), d, v.Status)
	if err != nil {
		httpErr := ToHTTPError(err)
		status := 500
		if httpErr.Code == CodeValidationFailed {
			status = 400
		} else if httpErr.Code == CodeNotFound {
			status = 404
		}
		out(w, status, httpErr)
		return
	}
	out(w, 200, map[string]string{"id": d.String(), "status": v.Status})
}
