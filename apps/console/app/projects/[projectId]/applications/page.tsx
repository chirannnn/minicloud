"use client";

import { useCallback, useEffect, useState } from "react";
import { api } from "@/app/api";

interface Application {
  id: string;
  projectId: string;
  name: string;
  slug: string;
  description: string;
  status: string;
}

export default function ApplicationsPage({ params }: { params: Promise<{ projectId: string }> }) {
  const [resolvedParams, setResolvedParams] = useState<{ projectId: string } | null>(null);
  const [applications, setApplications] = useState<Application[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [formData, setFormData] = useState({ name: "", slug: "", description: "" });

  useEffect(() => {
    params.then(setResolvedParams);
  }, [params]);

  const loadApplications = useCallback(async () => {
    if (!resolvedParams) return;
    setLoading(true);
    setError(null);
    try {
      const data = await api.listApplications(resolvedParams.projectId, 100, 0);
      setApplications(data.items);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Failed to load applications");
    } finally {
      setLoading(false);
    }
  }, [resolvedParams]);

  useEffect(() => {
    loadApplications();
  }, [loadApplications]);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!resolvedParams) return;
    try {
      await api.createApplication(resolvedParams.projectId, formData);
      setShowCreate(false);
      setFormData({ name: "", slug: "", description: "" });
      loadApplications();
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Failed to create application");
    }
  };

  const handleDelete = async (applicationId: string) => {
    if (!resolvedParams) return;
    if (!confirm("Are you sure you want to delete this application?")) return;
    try {
      await api.deleteApplication(resolvedParams.projectId, applicationId);
      loadApplications();
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Failed to delete application");
    }
  };

  if (!resolvedParams) {
    return <main><p>Loading...</p></main>;
  }

  return (
    <main>
      <p className="eyebrow">DEVELOPMENT CONSOLE</p>
      <h1>Applications</h1>
      <p className="intro">Manage applications for project {resolvedParams.projectId}</p>

      <section>
        <button onClick={() => setShowCreate(!showCreate)}>{showCreate ? "Cancel" : "Create Application"}</button>
      </section>

      {showCreate && (
        <section>
          <h2>Create Application</h2>
          <form onSubmit={handleCreate}>
            <div>
              <label>Name:</label>
              <input
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                required
              />
            </div>
            <div>
              <label>Slug:</label>
              <input
                value={formData.slug}
                onChange={(e) => setFormData({ ...formData, slug: e.target.value })}
                required
              />
            </div>
            <div>
              <label>Description:</label>
              <input
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              />
            </div>
            <button type="submit">Create</button>
          </form>
        </section>
      )}

      {error && <section><p style={{ color: "#fca5a5" }}>{error}</p></section>}

      {loading ? (
        <section><p>Loading...</p></section>
      ) : applications && applications.length > 0 ? (
        <section>
          <h2>Applications ({applications.length})</h2>
          <ul>
            {applications.map((app) => (
              <li key={app.id}>
                <strong>{app.name}</strong> ({app.slug})
                <br />
                <small>{app.description}</small>
                <br />
                <a href={`/projects/${resolvedParams.projectId}/applications/${app.id}/deployments`}>View Deployments</a>
                <button onClick={() => handleDelete(app.id)} style={{ marginLeft: "10px" }}>Delete</button>
              </li>
            ))}
          </ul>
        </section>
      ) : (
        <section><p>No applications found</p></section>
      )}

      <section>
        <a href="/projects">Back to Projects</a>
      </section>
    </main>
  );
}
