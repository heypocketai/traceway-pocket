import { NodeSDK } from "@opentelemetry/sdk-node";
import { OTLPTraceExporter } from "@opentelemetry/exporter-trace-otlp-http";
import { OTLPMetricExporter } from "@opentelemetry/exporter-metrics-otlp-http";
import { PeriodicExportingMetricReader } from "@opentelemetry/sdk-metrics";
import { getNodeAutoInstrumentations } from "@opentelemetry/auto-instrumentations-node";

const tracewayUrl = process.env.TRACEWAY_URL || "http://localhost:8082";
const tracewayToken = process.env.TRACEWAY_TOKEN || "backend-dev-token";

const sdk = new NodeSDK({
  traceExporter: new OTLPTraceExporter({
    url: `${tracewayUrl}/api/otel/v1/traces`,
    headers: { Authorization: `Bearer ${tracewayToken}` },
  }),

  metricReader: new PeriodicExportingMetricReader({
    exporter: new OTLPMetricExporter({
      url: `${tracewayUrl}/api/otel/v1/metrics`,
      headers: { Authorization: `Bearer ${tracewayToken}` },
    }),
    exportIntervalMillis: 10_000,
  }),

  // Disable instrumentation-http — @hono/otel handles HTTP spans at the middleware level.
  // Keep other auto-instrumentations (database, redis, fetch, etc.)
  instrumentations: [
    getNodeAutoInstrumentations({
      "@opentelemetry/instrumentation-http": { enabled: false },
    }),
  ],
});

sdk.start();
