DROP INDEX IF EXISTS idx_metric_points_project_name;
CREATE INDEX IF NOT EXISTS idx_metric_points_project_name_value ON metric_points(project_id, name, recorded_at, value);
