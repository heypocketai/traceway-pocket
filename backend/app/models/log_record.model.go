package models

import (
	"time"

	"github.com/google/uuid"
)

type LogRecord struct {
	Id                 uuid.UUID         `json:"id" ch:"id"`
	ProjectId          uuid.UUID         `json:"projectId" ch:"project_id"`
	Timestamp          time.Time         `json:"timestamp" ch:"timestamp"`
	TraceId            string            `json:"traceId" ch:"trace_id"`
	SpanId             string            `json:"spanId" ch:"span_id"`
	TraceFlags         uint8             `json:"traceFlags" ch:"trace_flags"`
	SeverityText       string            `json:"severityText" ch:"severity_text"`
	SeverityNumber     uint8             `json:"severityNumber" ch:"severity_number"`
	ServiceName        string            `json:"serviceName" ch:"service_name"`
	Body               string            `json:"body" ch:"body"`
	ResourceSchemaUrl  string            `json:"resourceSchemaUrl" ch:"resource_schema_url"`
	ResourceAttributes map[string]string `json:"resourceAttributes" ch:"resource_attributes"`
	ScopeSchemaUrl     string            `json:"scopeSchemaUrl" ch:"scope_schema_url"`
	ScopeName          string            `json:"scopeName" ch:"scope_name"`
	ScopeVersion       string            `json:"scopeVersion" ch:"scope_version"`
	ScopeAttributes    map[string]string `json:"scopeAttributes" ch:"scope_attributes"`
	LogAttributes      map[string]string `json:"logAttributes" ch:"log_attributes"`
}
