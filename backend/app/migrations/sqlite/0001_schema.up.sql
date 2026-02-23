CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    token TEXT NOT NULL,
    framework TEXT NOT NULL DEFAULT 'custom',
    organization_id INTEGER REFERENCES organizations(id),
    source_map_token TEXT,
    created_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_projects_token ON projects(token);
CREATE INDEX IF NOT EXISTS idx_projects_organization_id ON projects(organization_id);

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    password TEXT NOT NULL,
    password_reset_token TEXT,
    password_reset_expires_at DATETIME,
    password_reset_requested_at DATETIME,
    created_at DATETIME DEFAULT (datetime('now'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email);

CREATE TABLE IF NOT EXISTS organizations (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    timezone TEXT NOT NULL DEFAULT 'UTC',
    created_at DATETIME DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS organization_users (
    id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    organization_id INTEGER NOT NULL REFERENCES organizations(id),
    role TEXT NOT NULL CHECK (role IN ('owner','admin','user','readonly')),
    created_at DATETIME DEFAULT (datetime('now')),
    UNIQUE(user_id, organization_id)
);

CREATE TABLE IF NOT EXISTS invitations (
    id INTEGER PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id),
    email TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('admin','user','readonly')),
    token TEXT NOT NULL UNIQUE,
    invited_by INTEGER NOT NULL REFERENCES users(id),
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','accepted','expired')),
    expires_at DATETIME NOT NULL,
    accepted_at DATETIME,
    created_at DATETIME DEFAULT (datetime('now'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_invitations_email_org_pending
    ON invitations(email, organization_id) WHERE status = 'pending';

CREATE TABLE IF NOT EXISTS source_maps (
    id INTEGER PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id),
    version TEXT NOT NULL,
    file_name TEXT NOT NULL,
    storage_key TEXT NOT NULL,
    file_size INTEGER NOT NULL,
    uploaded_at DATETIME NOT NULL DEFAULT (datetime('now')),
    UNIQUE(project_id, version, file_name)
);

CREATE INDEX IF NOT EXISTS idx_source_maps_project_version ON source_maps(project_id, version);

CREATE TABLE IF NOT EXISTS metric_registry (
    id INTEGER PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id),
    name TEXT NOT NULL,
    metric_type TEXT NOT NULL DEFAULT 'gauge',
    unit TEXT DEFAULT '',
    description TEXT DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    UNIQUE(project_id, name)
);

CREATE TABLE IF NOT EXISTS widget_groups (
    id INTEGER PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id),
    name TEXT NOT NULL,
    description TEXT DEFAULT '',
    is_default INTEGER NOT NULL DEFAULT 0,
    created_by INTEGER REFERENCES users(id),
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_widget_groups_project_id ON widget_groups(project_id);

CREATE TABLE IF NOT EXISTS widget_group_widgets (
    id INTEGER PRIMARY KEY,
    widget_group_id INTEGER NOT NULL REFERENCES widget_groups(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    widget_type TEXT NOT NULL,
    config TEXT NOT NULL DEFAULT '{}',
    position INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_widget_group_widgets_widget_group_id ON widget_group_widgets(widget_group_id);
