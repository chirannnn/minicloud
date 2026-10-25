"use client";

import { useCallback, useEffect, useState } from "react";
import { api } from "@/app/api";

interface Deployment {
  id: string;
  projectId: string;
  applicationId: string;
  version: string;
  status: string;
}

export default function DeploymentsPage({ params }: { params: Promise<{ projectId: string; applicationId: string }> }) {
  const [resolvedParams, setResolvedParams] = useState<{ projectId: string; applicationId: string } | null>(null);
  const [deployments, setDeployments] = useState<Deployment[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [formData, setFormData] = useState({ version: "", configuration: "{}" });
  const [showUpdateStatus, setShowUpdateStatus] = useState(false);
  const [statusForm, setStatusForm] = useState({ status: "" });

  useEffect(() => {
    params.then(setResolvedParams);
  }, [params]);

  const loadDeployments = useCallback(async () => {
    if (!resolvedParams) return;
    setLoading(true);
    setError(null);
    try {
      const data = await api.listDeployments(resolvedParams.projectId, resolvedParams.applicationId, 100, 0);
      setDeployments(data.items);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Failed to load deployments");
    } finally {
      setLoading(false);
    }
  }, [resolvedParams]);

  useEffect(() => {
    loadDeployments();
  }, [loadDeployments]);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!resolvedParams) return;
    try {
      const config = JSON.parse(formData.configuration);
      await api.createDeployment(resolvedParams.projectId, resolvedParams.applicationId, { version: formData.version, configuration: config });
      setShowCreate(false);
      setFormData({ version: "", configuration: "{}" });
      loadDeployments();
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Failed to create deployment");
    }
  };

  const handleUpdateStatus = async (deploymentId: string) => {
    try {
      await api.updateDeploymentStatus(deploymentId, statusForm);
      setShowUpdateStatus(false);
      setStatusForm({ status: "" });
      loadDeployments();
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Failed to update deployment status");
    }
  };

  if (!resolvedParams) {
    return <main><p>Loading...</p></main>;
  }

  return (
    <main>
      <p className="eyebrow">DEVELOPMENT CONSOLE</p>
      <h1>Deployments</h1>
      <p className="intro">Manage deployments for application {resolvedParams.applicationId}</p>

      <section>
        <button onClick={() => setShowCreate(!showCreate)}>{showCreate ? "Cancel" : "Create Deployment"}</button>
      </section>

      {showCreate && (
        <section>
          <h2>Create Deployment</h2>
          <form onSubmit={handleCreate}>
            <div>
              <label>Version:</label>
              <input
                value={formData.version}
                onChange={(e) => setFormData({ ...formData, version: e.target.value })}
                required
              />
            </div>
            <div>
              <label>Configuration (JSON):</label>
              <textarea
                value={formData.configuration}
                onChange={(e) => setFormData({ ...formData, configuration: e.target.value })}
                rows={5}
                required
              />
            </div>
            <button type="submit">Create</button>
          </form>
        </section>
      )}

      {showUpdateStatus && (
        <section>
          <h2>Update Deployment Status</h2>
          <form onSubmit={(e) => { e.preventDefault(); handleUpdateStatus(""); }}>
            <div>
              <label>Status:</label>
              <select
                value={statusForm.status}
                onChange={(e) => setStatusForm({ status: e.target.value })}
                required
              >
                <option value="">Select status</option>
                <option value="PENDING">PENDING</option>
                <option value="QUEUED">QUEUED</option>
                <option value="RUNNING">RUNNING</option>
                <option value="SUCCEEDED">SUCCEEDED</option>
                <option value="FAILED">FAILED</option>
                <option value="CANCELLED">CANCELLED</option>
              </select>
            </div>
            <button type="submit">Update</button>
            <button type="button" onClick={() => setShowUpdateStatus(false)}>Cancel</button>
          </form>
        </section>
      )}

      {error && <section><p style={{ color: "#fca5a5" }}>{error}</p></section>}

      {loading ? (
        <section><p>Loading...</p></section>
      ) : deployments && deployments.length > 0 ? (
        <section>
          <h2>Deployments ({deployments.length})</h2>
          <ul>
            {deployments.map((deployment) => (
              <li key={deployment.id}>
                <strong>Version {deployment.version}</strong>
                <br />
                <small>Status: {deployment.status}</small>
                <br />
                <button onClick={() => { setShowUpdateStatus(true); setStatusForm({ status: "" }); }}>
                  Update Status
                </button>
              </li>
            ))}
          </ul>
        </section>
      ) : (
        <section><p>No deployments found</p></section>
      )}

      <section>
        <a href={`/projects/${resolvedParams.projectId}/applications`}>Back to Applications</a>
      </section>
    </main>
  );
}
