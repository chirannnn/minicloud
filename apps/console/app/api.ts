const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export async function apiRequest<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const response = await fetch(`${API_URL}${endpoint}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...options.headers,
    },
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ code: "INTERNAL", message: "Request failed" }));
    throw error;
  }

  return response.json();
}

export const api = {
  // Projects
  listProjects: (limit = 20, offset = 0) =>
    apiRequest<{ items: Array<{ id: string; ownerId: string; name: string; slug: string; description: string; status: string }>; limit: number; offset: number }>(
      `/api/v1/projects?limit=${limit}&offset=${offset}`
    ),
  getProject: (id: string) =>
    apiRequest<{ id: string; ownerId: string; name: string; slug: string; description: string; status: string }>(
      `/api/v1/projects/${id}`
    ),
  createProject: (data: { ownerId: string; name: string; slug: string; description: string }) =>
    apiRequest<{ id: string }>(`/api/v1/projects`, {
      method: "POST",
      body: JSON.stringify(data),
    }),
  updateProject: (id: string, data: { name: string; description: string; status: string }) =>
    apiRequest<{ id: string }>(`/api/v1/projects/${id}`, {
      method: "PATCH",
      body: JSON.stringify(data),
    }),
  deleteProject: (id: string) =>
    fetch(`${API_URL}/api/v1/projects/${id}`, { method: "DELETE" }),

  // Applications
  listApplications: (projectId: string, limit = 20, offset = 0) =>
    apiRequest<{ items: Array<{ id: string; projectId: string; name: string; slug: string; description: string; status: string }>; limit: number; offset: number }>(
      `/api/v1/projects/${projectId}/applications?limit=${limit}&offset=${offset}`
    ),
  getApplication: (projectId: string, applicationId: string) =>
    apiRequest<{ id: string; projectId: string; name: string; slug: string; description: string; status: string }>(
      `/api/v1/projects/${projectId}/applications/${applicationId}`
    ),
  createApplication: (projectId: string, data: { name: string; slug: string; description: string }) =>
    apiRequest<{ id: string }>(`/api/v1/projects/${projectId}/applications`, {
      method: "POST",
      body: JSON.stringify(data),
    }),
  updateApplication: (projectId: string, applicationId: string, data: { name: string; description: string; status: string }) =>
    apiRequest<{ id: string }>(`/api/v1/projects/${projectId}/applications/${applicationId}`, {
      method: "PATCH",
      body: JSON.stringify(data),
    }),
  deleteApplication: (projectId: string, applicationId: string) =>
    fetch(`${API_URL}/api/v1/projects/${projectId}/applications/${applicationId}`, { method: "DELETE" }),

  // Deployments
  listDeployments: (projectId: string, applicationId: string, limit = 20, offset = 0) =>
    apiRequest<{ items: Array<{ id: string; projectId: string; applicationId: string; version: string; status: string }>; limit: number; offset: number }>(
      `/api/v1/projects/${projectId}/applications/${applicationId}/deployments?limit=${limit}&offset=${offset}`
    ),
  getDeployment: (id: string) =>
    apiRequest<{ id: string; projectId: string; applicationId: string; version: string; status: string }>(
      `/api/v1/deployments/${id}`
    ),
  createDeployment: (projectId: string, applicationId: string, data: { version: string; configuration: object }) =>
    apiRequest<{ id: string }>(`/api/v1/projects/${projectId}/applications/${applicationId}/deployments`, {
      method: "POST",
      body: JSON.stringify(data),
    }),
  updateDeploymentStatus: (id: string, data: { status: string }) =>
    apiRequest<{ id: string; status: string }>(`/api/v1/deployments/${id}/status`, {
      method: "PATCH",
      body: JSON.stringify(data),
    }),
};
