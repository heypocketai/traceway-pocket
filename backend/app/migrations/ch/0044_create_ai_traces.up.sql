CREATE TABLE IF NOT EXISTS ai_traces
(
    `id` UUID,
    `project_id` UUID,
    `recorded_at` DateTime64(3),
    `duration` Int64,
    `status_code` UInt8,
    `model` LowCardinality(String) DEFAULT '',
    `response_model` LowCardinality(String) DEFAULT '',
    `provider` LowCardinality(String) DEFAULT '',
    `operation` LowCardinality(String) DEFAULT '',
    `input_tokens` Int64 DEFAULT 0,
    `output_tokens` Int64 DEFAULT 0,
    `total_tokens` Int64 DEFAULT 0,
    `cached_tokens` Int64 DEFAULT 0,
    `reasoning_tokens` Int64 DEFAULT 0,
    `input_cost` Float64 DEFAULT 0,
    `output_cost` Float64 DEFAULT 0,
    `total_cost` Float64 DEFAULT 0,
    `trace_name` LowCardinality(String) DEFAULT '',
    `user_id` String DEFAULT '',
    `finish_reason` LowCardinality(String) DEFAULT '',
    `server_name` LowCardinality(String) DEFAULT '',
    `app_version` LowCardinality(String) DEFAULT '',
    `storage_key` String DEFAULT '',
    `attributes` String DEFAULT '{}',
    INDEX idx_id id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_trace_name trace_name TYPE set(100) GRANULARITY 4,
    INDEX idx_model model TYPE set(100) GRANULARITY 4
)
ENGINE = MergeTree
PARTITION BY toYYYYMMDD(recorded_at)
ORDER BY (project_id, recorded_at, trace_name)
SETTINGS index_granularity = 8192
