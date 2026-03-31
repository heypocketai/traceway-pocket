package otelcontrollers

import (
	"encoding/base64"
	"encoding/hex"
	"strconv"
	"time"

	"github.com/google/uuid"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
)

// otelTraceIDToUUID converts a 16-byte OTEL trace ID to a UUID.
// OTLP JSON uses hex encoding for trace_id, but protojson base64-decodes it,
// producing 24 bytes instead of 16. We roundtrip through base64 to recover the hex string.
func otelTraceIDToUUID(traceID []byte) uuid.UUID {
	if len(traceID) == 16 {
		var u uuid.UUID
		copy(u[:], traceID)
		return u
	}
	if len(traceID) == 24 {
		hexStr := base64.StdEncoding.EncodeToString(traceID)
		if decoded, err := hex.DecodeString(hexStr); err == nil && len(decoded) == 16 {
			var u uuid.UUID
			copy(u[:], decoded)
			return u
		}
	}
	return uuid.Nil
}

// otelSpanIDToUUID converts an 8-byte OTEL span ID to a UUID by zero-padding the first 8 bytes.
// Same base64/hex roundtrip as otelTraceIDToUUID for 12-byte inputs.
func otelSpanIDToUUID(spanID []byte) uuid.UUID {
	if len(spanID) == 8 {
		var u uuid.UUID
		copy(u[8:], spanID)
		return u
	}
	if len(spanID) == 12 {
		hexStr := base64.StdEncoding.EncodeToString(spanID)
		if decoded, err := hex.DecodeString(hexStr); err == nil && len(decoded) == 8 {
			var u uuid.UUID
			copy(u[8:], decoded)
			return u
		}
	}
	return uuid.Nil
}

func nanoToTime(nanos uint64) time.Time {
	return time.Unix(0, int64(nanos))
}

func extractAttributes(attrs []*commonpb.KeyValue) map[string]string {
	if len(attrs) == 0 {
		return nil
	}
	m := make(map[string]string, len(attrs))
	for _, kv := range attrs {
		if kv.Value == nil {
			continue
		}
		switch v := kv.Value.Value.(type) {
		case *commonpb.AnyValue_StringValue:
			m[kv.Key] = v.StringValue
		case *commonpb.AnyValue_IntValue:
			m[kv.Key] = strconv.FormatInt(v.IntValue, 10)
		case *commonpb.AnyValue_DoubleValue:
			m[kv.Key] = strconv.FormatFloat(v.DoubleValue, 'g', -1, 64)
		case *commonpb.AnyValue_BoolValue:
			m[kv.Key] = strconv.FormatBool(v.BoolValue)
		}
	}
	return m
}

func getStringAttribute(attrs []*commonpb.KeyValue, key string) string {
	for _, kv := range attrs {
		if kv.Key == key && kv.Value != nil {
			if sv, ok := kv.Value.Value.(*commonpb.AnyValue_StringValue); ok {
				return sv.StringValue
			}
		}
	}
	return ""
}

func getIntAttribute(attrs []*commonpb.KeyValue, key string) (int64, bool) {
	for _, kv := range attrs {
		if kv.Key == key && kv.Value != nil {
			if iv, ok := kv.Value.Value.(*commonpb.AnyValue_IntValue); ok {
				return iv.IntValue, true
			}
		}
	}
	return 0, false
}

func getFloatAttribute(attrs []*commonpb.KeyValue, key string) float64 {
	for _, kv := range attrs {
		if kv.Key == key && kv.Value != nil {
			switch v := kv.Value.Value.(type) {
			case *commonpb.AnyValue_DoubleValue:
				return v.DoubleValue
			case *commonpb.AnyValue_IntValue:
				return float64(v.IntValue)
			}
		}
	}
	return 0
}
