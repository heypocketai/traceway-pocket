CREATE TABLE IF NOT EXISTS log_records (
    id                   UUID CODEC(ZSTD(1)),
    project_id           UUID CODEC(ZSTD(1)),
    timestamp            DateTime64(9) CODEC(Delta, ZSTD(1)),
    timestamp_date       Date DEFAULT toDate(timestamp),
    trace_id             String CODEC(ZSTD(1)),
    span_id              String CODEC(ZSTD(1)),
    trace_flags          UInt8,
    severity_text        LowCardinality(String) CODEC(ZSTD(1)),
    severity_number      UInt8,
    service_name         LowCardinality(String) CODEC(ZSTD(1)),
    body                 String CODEC(ZSTD(1)),
    resource_schema_url  LowCardinality(String) CODEC(ZSTD(1)),
    resource_attributes  Map(LowCardinality(String), String) CODEC(ZSTD(1)),
    scope_schema_url     LowCardinality(String) CODEC(ZSTD(1)),
    scope_name           String CODEC(ZSTD(1)),
    scope_version        LowCardinality(String) CODEC(ZSTD(1)),
    scope_attributes     Map(LowCardinality(String), String) CODEC(ZSTD(1)),
    log_attributes       Map(LowCardinality(String), String) CODEC(ZSTD(1)),
    INDEX idx_trace_id       trace_id                     TYPE bloom_filter(0.001)     GRANULARITY 1,
    INDEX idx_res_attr_key   mapKeys(resource_attributes) TYPE bloom_filter(0.01)      GRANULARITY 1,
    INDEX idx_res_attr_value mapValues(resource_attributes) TYPE bloom_filter(0.01)    GRANULARITY 1,
    INDEX idx_log_attr_key   mapKeys(log_attributes)      TYPE bloom_filter(0.01)      GRANULARITY 1,
    INDEX idx_log_attr_value mapValues(log_attributes)    TYPE bloom_filter(0.01)      GRANULARITY 1,
    INDEX idx_body           body                         TYPE tokenbf_v1(32768, 3, 0) GRANULARITY 1
)
ENGINE = MergeTree
PARTITION BY timestamp_date
ORDER BY (project_id, service_name, severity_number, timestamp)
TTL timestamp_date + INTERVAL 30 DAY
SETTINGS index_granularity = 8192
