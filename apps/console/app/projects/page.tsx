"use client";

import { useEffect, useState } from "react";
import { api } from "@/app/api";

interface Project {
  id: string;
  ownerId: string;
  name: string;
  slug: string;
  description: string;
  status: string;
}

export default function ProjectsPage() {
  const [projects, setProjects] = useState<Project[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [formData, setFormData] = useState({ ownerId: "", name: "", slug: "", description: "" });

  const loadProjects = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await api.listProjects(100, 0);
      setProjects(data.items);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Failed to load projects");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadProjects();
  }, []);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await api.createProject(formData);
      setShowCreate(false);
      setFormData({ ownerId: "", name: "", slug: "", description: "" });
      loadProjects();
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Failed to create project");
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm("Are you sure you want to delete this project?")) return;
    try {
      await api.deleteProject(id);
      loadProjects();
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Failed to delete project");
    }
  };

  return (
    <main>
      <p className="eyebrow">DEVELOPMENT CONSOLE</p>
      <h1>Projects</h1>
      <p className="intro">Manage your MiniCloud projects</p>

      <section>
        <button onClick={() => setShowCreate(!showCreate)}>{showCreate ? "Cancel" : "Create Project"}</button>
      </section>

      {showCreate && (
        <section>
          <h2>Create Project</h2>
          <form onSubmit={handleCreate}>
            <div>
              <label>Owner ID:</label>
              <input
                value={formData.ownerId}
                onChange={(e) => setFormData({ ...formData, ownerId: e.target.value })}
                required
              />
            </div>
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
      ) : projects && projects.length > 0 ? (
        <section>
          <h2>Projects ({projects.length})</h2>
          <ul>
            {projects.map((project) => (
              <li key={project.id}>
                <strong>{project.name}</strong> ({project.slug})
                <br />
                <small>{project.description}</small>
                <br />
                <a href={`/projects/${project.id}/applications`}>View Applications</a>
                <button onClick={() => handleDelete(project.id)} style={{ marginLeft: "10px" }}>Delete</button>
              </li>
            ))}
          </ul>
        </section>
      ) : (
        <section><p>No projects found</p></section>
      )}

      <section>
        <a href="/dashboard">Back to Dashboard</a>
      </section>
    </main>
  );
}
