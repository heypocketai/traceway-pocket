export async function register() {
  if (process.env.NEXT_RUNTIME === "nodejs") {
    const { NodeSDK } = await import("@opentelemetry/sdk-node");
    const { OTLPTraceExporter } = await import(
      "@opentelemetry/exporter-trace-otlp-http"
    );
    const { OTLPMetricExporter } = await import(
      "@opentelemetry/exporter-metrics-otlp-http"
    );
    const { PeriodicExportingMetricReader } = await import(
      "@opentelemetry/sdk-metrics"
    );
    const { getNodeAutoInstrumentations } = await import(
      "@opentelemetry/auto-instrumentations-node"
    );

    const tracewayUrl = process.env.TRACEWAY_URL || "http://localhost:8082";
    const tracewayToken = process.env.TRACEWAY_TOKEN || "backend-dev-token";

    const { Resource } = await import("@opentelemetry/resources");

    const sdk = new NodeSDK({
      resource: new Resource({
        "service.name": "nextjs-otel-example",
        "service.version": "1.0.0",
      }),

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

      instrumentations: [getNodeAutoInstrumentations()],
    });

    sdk.start();
  }
}
