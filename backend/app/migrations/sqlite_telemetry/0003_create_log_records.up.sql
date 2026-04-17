CREATE TABLE IF NOT EXISTS log_records (
    id                  TEXT NOT NULL,
    project_id          TEXT NOT NULL,
    timestamp           DATETIME NOT NULL,
    trace_id            TEXT NOT NULL DEFAULT '',
    span_id             TEXT NOT NULL DEFAULT '',
    trace_flags         INTEGER NOT NULL DEFAULT 0,
    severity_text       TEXT NOT NULL DEFAULT '',
    severity_number     INTEGER NOT NULL DEFAULT 0,
    service_name        TEXT NOT NULL DEFAULT '',
    body                TEXT NOT NULL DEFAULT '',
    resource_schema_url TEXT NOT NULL DEFAULT '',
    resource_attributes TEXT NOT NULL DEFAULT '{}',
    scope_schema_url    TEXT NOT NULL DEFAULT '',
    scope_name          TEXT NOT NULL DEFAULT '',
    scope_version       TEXT NOT NULL DEFAULT '',
    scope_attributes    TEXT NOT NULL DEFAULT '{}',
    log_attributes      TEXT NOT NULL DEFAULT '{}'
);
CREATE INDEX IF NOT EXISTS idx_log_records_project_timestamp ON log_records(project_id, timestamp);
CREATE INDEX IF NOT EXISTS idx_log_records_project_trace     ON log_records(project_id, trace_id);
CREATE INDEX IF NOT EXISTS idx_log_records_project_service   ON log_records(project_id, service_name, timestamp);
