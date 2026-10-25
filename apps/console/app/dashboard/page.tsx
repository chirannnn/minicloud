"use client";

import { useEffect, useState } from "react";
import { api } from "@/app/api";

type Status = "checking" | "connected" | "unavailable";

interface Project {
  id: string;
  ownerId: string;
  name: string;
  slug: string;
  description: string;
  status: string;
}

interface Application {
  id: string;
  projectId: string;
  name: string;
  slug: string;
  description: string;
  status: string;
}

interface Deployment {
  id: string;
  projectId: string;
  applicationId: string;
  version: string;
  status: string;
}

export default function Dashboard() {
  const [status, setStatus] = useState<Status>("checking");
  const [projects, setProjects] = useState<Project[] | null>(null);
  const [applications, setApplications] = useState<Application[] | null>(null);
  const [deployments, setDeployments] = useState<Deployment[] | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadData = async () => {
      try {
        const health = await fetch(`${process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080"}/health`);
        setStatus(health.ok ? "connected" : "unavailable");

        const projectsData = await api.listProjects(100, 0);
        setProjects(projectsData.items);

        const allApps: Application[] = [];
        for (const project of projectsData.items) {
          const appsData = await api.listApplications(project.id, 100, 0);
          allApps.push(...appsData.items);
        }
        setApplications(allApps);

        const allDeployments: Deployment[] = [];
        for (const app of allApps) {
          const deploymentsData = await api.listDeployments(app.projectId, app.id, 100, 0);
          allDeployments.push(...deploymentsData.items);
        }
        setDeployments(allDeployments);
      } catch (err: unknown) {
        setError(err instanceof Error ? err.message : "Failed to load data");
        setStatus("unavailable");
      }
    };

    loadData();
  }, []);

  return (
    <main>
      <p className="eyebrow">DEVELOPMENT CONSOLE</p>
      <h1>MiniCloud</h1>
      <p className="intro">Phase 1 Core Control Plane - Real API Integration</p>

      <section>
        <h2>API connectivity</h2>
        <p data-status={status}>{status}</p>
      </section>

      {error && (
        <section>
          <h2>Error</h2>
          <p style={{ color: "#fca5a5" }}>{error}</p>
        </section>
      )}

      {status === "connected" && (
        <>
          <section>
            <h2>Resources</h2>
            <p>Projects: {projects?.length ?? 0}</p>
            <p>Applications: {applications?.length ?? 0}</p>
            <p>Deployments: {deployments?.length ?? 0}</p>
          </section>

          <section>
            <h2>Quick Links</h2>
            <ul>
              <li><a href="/projects">View Projects</a></li>
            </ul>
          </section>
        </>
      )}
    </main>
  );
}
