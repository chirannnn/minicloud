CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(), email TEXT NOT NULL UNIQUE,
  display_name TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE projects (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(), owner_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  name TEXT NOT NULL, slug TEXT NOT NULL, description TEXT NOT NULL DEFAULT '', status TEXT NOT NULL DEFAULT 'ACTIVE',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now(), UNIQUE(owner_id, slug)
);
CREATE TABLE applications (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(), project_id UUID NOT NULL REFERENCES projects(id) ON DELETE RESTRICT,
  name TEXT NOT NULL, slug TEXT NOT NULL, description TEXT NOT NULL DEFAULT '', status TEXT NOT NULL DEFAULT 'ACTIVE',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now(), UNIQUE(project_id, slug)
);
CREATE TABLE deployments (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(), project_id UUID NOT NULL REFERENCES projects(id) ON DELETE RESTRICT,
  application_id UUID NOT NULL REFERENCES applications(id) ON DELETE RESTRICT, version TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING','QUEUED','RUNNING','SUCCEEDED','FAILED','CANCELLED')),
  configuration JSONB NOT NULL DEFAULT '{}', created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE audit_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(), project_id UUID REFERENCES projects(id) ON DELETE RESTRICT, actor_id UUID REFERENCES users(id) ON DELETE RESTRICT,
  action TEXT NOT NULL, resource_type TEXT NOT NULL, resource_id UUID NOT NULL, metadata JSONB NOT NULL DEFAULT '{}', created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE outbox_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(), event_type TEXT NOT NULL, aggregate_type TEXT NOT NULL, aggregate_id UUID NOT NULL,
  payload JSONB NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT now(), published_at TIMESTAMPTZ
);
CREATE INDEX projects_owner_id_idx ON projects(owner_id);
CREATE INDEX applications_project_id_idx ON applications(project_id);
CREATE INDEX deployments_project_id_idx ON deployments(project_id);
CREATE INDEX deployments_application_id_idx ON deployments(application_id);
CREATE INDEX deployments_status_idx ON deployments(status);
CREATE INDEX audit_logs_project_id_idx ON audit_logs(project_id);
CREATE INDEX audit_logs_resource_id_idx ON audit_logs(resource_id);
CREATE INDEX outbox_events_unpublished_idx ON outbox_events(published_at) WHERE published_at IS NULL;
