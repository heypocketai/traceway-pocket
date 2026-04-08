# Add Traceway to a Project

Add OpenTelemetry tracing to an existing project so it reports to a Traceway instance.

## What Traceway Needs

For the integration to work correctly, the instrumentation MUST capture:

1. **Endpoints grouped by route pattern** — `GET /api/users/1` and `GET /api/users/2` must appear as a single `GET /api/users/:id` endpoint, NOT as separate entries. This requires `http.route` to be set on the root span. Without it, the Traceway dashboard explodes with thousands of unique URL entries.

2. **Status codes** — `http.response.status_code` must be set on spans so Traceway can track error rates, 4xx/5xx breakdowns, and Apdex scores.

3. **Exceptions with stack traces** — Thrown errors must be recorded as span events with `exception.type`, `exception.message`, and `exception.stacktrace` attributes. These appear as **Issues** in Traceway.

4. **Scheduled/long-running tasks** — Background jobs (cron, queues, consumers) must create root spans with `SpanKind.CONSUMER`. This is how Traceway distinguishes Tasks from Endpoints. Without the correct span kind, background work either gets misclassified as an Endpoint or dropped entirely.

### How Traceway classifies spans

| OTel Span | Condition | Traceway Concept |
|---|---|---|
| Root span | `SpanKind = SERVER` or `INTERNAL` with HTTP attributes | **Endpoint** |
| Root span | `SpanKind = CONSUMER` | **Task** |
| Non-root span | Has a parent span ID | **Span** |
| Exception event | Event named `"exception"` on any span | **Issue** |

## Step 1: Identify the Framework

Detect the framework by reading `package.json` (Node.js), `go.mod` (Go), `composer.json` (PHP), or asking the user.

## Step 2: Follow the Framework-Specific Guide

### Hono (Node.js)
Follow `skills/add-traceway-to-hono-project.md`. Uses `@hono/otel` middleware — do NOT use `@opentelemetry/instrumentation-http` (it doesn't work with Hono's ESM imports on Node 22+).
- Endpoints: `@hono/otel` sets `http.route` automatically
- Status codes: `@hono/otel` sets them automatically
- Exceptions: `@hono/otel` records thrown errors automatically
- Tasks: No built-in scheduler — use `SpanKind.CONSUMER` manually for background work

### NestJS (Node.js)
Follow `skills/add-traceway-to-nestjs-project.md`. Simplest integration — Express auto-instrumentation handles everything.
- Endpoints: `instrumentation-express` sets `http.route` automatically
- Status codes: `instrumentation-http` sets them automatically
- Exceptions: Express error handling records them automatically
- Tasks: Wrap `@nestjs/schedule` cron jobs and `@nestjs/bull` queue consumers with `SpanKind.CONSUMER` spans

### Next.js (Node.js)
Follow `skills/add-traceway-to-nextjs-project.md`. Requires `withRoute()` wrapper for API routes and `@prisma/instrumentation` for database tracing.
- Endpoints: `withRoute()` helper must be added manually to every API route handler
- Status codes: Set by the HTTP instrumentation
- Exceptions: `withRoute()` catches and records thrown errors
- Tasks: No built-in scheduler — use `SpanKind.CONSUMER` manually for background work

### Express (Node.js)
- Install: `@opentelemetry/sdk-node @opentelemetry/auto-instrumentations-node @opentelemetry/exporter-trace-otlp-http @opentelemetry/exporter-metrics-otlp-http @opentelemetry/api`
- Create `instrumentation.js` at project root with `NodeSDK` + `getNodeAutoInstrumentations()`
- No app code changes needed — auto-instrumentation captures routes, status codes, errors
- Start with `node --import ./instrumentation.js server.js`
- Tasks: Use `SpanKind.CONSUMER` manually for background work
- Full docs: `docs/pages/client/node-sdk/index.mdx`

### Gin / Chi / Fiber / FastHTTP / stdlib (Go)
- Install the framework-specific middleware: `go get go.tracewayapp.com/tracewaygin` (or `tracewaychi`, `tracewayfiber`, `tracewayfasthttp`, `tracewayhttp`)
- Add middleware: `r.Use(tracewaygin.New("token@http://traceway:8082/api/report"))`
- Reports via Traceway's native protocol (`/api/report`), not OTel
- Endpoints, status codes, exceptions, and tasks are all handled by the Go SDK automatically
- Full docs: `docs/pages/client/gin-middleware/index.mdx` (or the corresponding framework directory)

### Symfony (PHP)
- Install: `composer require traceway/opentelemetry-symfony open-telemetry/exporter-otlp php-http/guzzle7-adapter`
- Configure via `.env` with `OTEL_*` variables
- Add `\OpenTelemetry\SDK\SdkAutoloader::autoload()` to `public/index.php`
- Endpoints and status codes: handled by Symfony OTel auto-instrumentation
- Tasks: Symfony Messenger consumers are auto-instrumented as Tasks
- Full docs: `docs/pages/client/symfony/index.mdx`

### React / Vue / Svelte / jQuery (Frontend)
- Install the framework-specific Traceway SDK: `npm install @tracewayapp/react` (or `@tracewayapp/vue`, `@tracewayapp/svelte`, `@tracewayapp/jquery`)
- These are client-side SDKs that report to `/api/report`, not OTel
- Full docs: `docs/pages/client/react/index.mdx` (or the corresponding framework directory)

### Cloudflare Workers
- Uses Cloudflare's built-in OTLP export, not the Node SDK
- Scheduled handlers (`scheduled` event) create root spans automatically
- Full docs: `docs/pages/client/cloudflare/index.mdx`

### Any Other Language (Generic OTel)
- Use any OpenTelemetry SDK for the language
- Export via OTLP/HTTP to `https://<traceway-instance>/api/otel/v1/traces` and `/v1/metrics`
- Set `Authorization: Bearer <project-token>` header
- Ensure `http.route` is set on root SERVER spans (not just `url.path`)
- Use `SpanKind.CONSUMER` for background/scheduled work
- Full docs: `docs/pages/client/otel/index.mdx`

## Instrumenting Background Tasks (All Frameworks)

For any framework, background work (cron jobs, queue consumers, scheduled tasks) must create a root span with `SpanKind.CONSUMER` to appear as a **Task** in Traceway:

```typescript
import { trace, SpanKind, SpanStatusCode } from "@opentelemetry/api";

const tracer = trace.getTracer("my-app");

async function runScheduledJob() {
  await tracer.startActiveSpan(
    "cleanup-expired-sessions",
    { kind: SpanKind.CONSUMER },
    async (span) => {
      try {
        await doWork();
        span.setStatus({ code: SpanStatusCode.OK });
      } catch (error) {
        span.recordException(error);
        span.setStatus({ code: SpanStatusCode.ERROR, message: error.message });
        throw error;
      } finally {
        span.end();
      }
    }
  );
}
```

Without `SpanKind.CONSUMER`, the span would either be classified as an Endpoint (wrong) or dropped.

## Common Across All Node.js Frameworks

- **Traceway URL**: `https://<instance>/api/otel/v1/traces` and `/v1/metrics`
- **Auth header**: `Authorization: Bearer <project-token>`
- **Environment variables**: `TRACEWAY_URL` and `TRACEWAY_TOKEN` (or standard `OTEL_*` vars)
- **Auto-instrumented child spans** (CJS packages only): `pg`, `mysql2`, `mongodb`, `ioredis`, `redis`, Prisma (with `@prisma/instrumentation`), outgoing `fetch()` via `instrumentation-undici`
- **Not auto-instrumented**: SQLite (`better-sqlite3`), custom business logic — use `tracer.startActiveSpan()` manually
