<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="Traceway%20Logo%20White.png" />
    <source media="(prefers-color-scheme: light)" srcset="Traceway%20Logo.png" />
    <img src="Traceway Logo.png" alt="Traceway Logo" width="200" />
  </picture>
</p>

<h3 align="center">Open-source error tracking and performance monitoring for Go applications</h3>

<p align="center">
  <a href="https://tracewayapp.com">Website</a> · <a href="https://docs.tracewayapp.com">Docs</a> · <a href="https://github.com/tracewayapp/go-client">Go Client SDK</a>
</p>

---

Traceway is a self-hosted observability platform that ingests OpenTelemetry traces and metrics, groups exceptions automatically, and gives you endpoint performance, distributed tracing, and alerts — all in a single binary. No OTel Collector or separate time-series database required.

It can run as a standalone server with ClickHouse + PostgreSQL, or **embedded directly inside your Go application** with SQLite for zero-infrastructure local development.

<img width="2452" height="1966" alt="Traceway Dashboard" src="https://github.com/user-attachments/assets/30a4fa24-7d08-4b36-a8f3-42abc73692fd" />

## Features

- **OTel-Native** — Accepts OTLP/HTTP traces and metrics directly, no Collector needed
- **Embedded Mode** — Run Traceway inside your Go process with SQLite, zero external dependencies
- **Issue Tracking** — Automatic exception grouping with normalized stack trace hashing and contextual tags
- **Endpoint Performance** — P50, P95, P99 latency percentiles with Apdex scoring and impact scores
- **Distributed Tracing** — Full request traces with span breakdowns
- **Alerts & Notifications** — Configurable rules (error rate, latency thresholds, Apdex drops, metric thresholds) with email, Slack, GitHub, and webhook delivery
- **Session Replay** — Replay user sessions to see exactly what happened before an error
- **System Metrics** — CPU, memory, goroutines, and GC monitoring with custom metric support
- **Background Tasks** — Track and monitor async job performance
- **Multi-Tenant** — Organization-based access control with owner, admin, user, and read-only roles

## Quick Start

### Docker (standalone server)

```bash
docker compose up --build
```

Open `http://localhost` to access the dashboard.

See the [self-hosting docs](https://docs.tracewayapp.com/server/docker-compose) for configuration and deployment options.

### Embedded Mode (inside your Go app)

Run Traceway inside your Go process — no Docker, no external databases:

```bash
go get github.com/tracewayapp/traceway/backend
```

```go
import tracewaybackend "github.com/tracewayapp/traceway/backend"

func main() {
    go tracewaybackend.Run(
        tracewaybackend.WithPort(8082),
        tracewaybackend.WithDefaultUser("admin@localhost.com", "admin"),
        tracewaybackend.WithDefaultProject("My App", "go", "dev-token"),
    )

    // ... start your app, point OTel exporter to http://localhost:8082/api/otel/v1/traces
}
```

Open `http://localhost:8082`, log in, and hit your app to see traces appear.

See the [embedded mode guide](https://docs.tracewayapp.com/learn/embedded-mode) for the full step-by-step setup with OpenTelemetry, or check the [working example](./examples/embedded-backend-otel).

## Tech Stack

| Component | Technology |
|-----------|------------|
| Backend | Go 1.25, Gin |
| Frontend | SvelteKit 2, Svelte 5, Tailwind CSS v4 |
| Telemetry DB | ClickHouse (standalone) or SQLite (embedded) |
| Relational DB | PostgreSQL (standalone) or SQLite (embedded) |
| Client SDKs | [Go](https://github.com/tracewayapp/go-client), OpenTelemetry (any language) |

## Screenshots

| | |
|---|---|
| ![Issues](./printscreens/issues.png) | ![Endpoints](./printscreens/endpoints.png) |
| ![Spans](./printscreens/spans.png) | ![Metrics](./printscreens/metrics.png) |
| ![Session Replay](./printscreens/session-replay.png) | ![Attributes](./printscreens/attributes.png) |

## Project Structure

| Directory | Description |
|-----------|-------------|
| `backend/` | Go/Gin API server — telemetry ingestion, REST API, notifications, migrations |
| `frontend/` | SvelteKit 2 dashboard SPA |
| `docs/` | Documentation site (Nextra) |
| `examples/` | Working examples for embedded mode ([OTel](./examples/embedded-backend-otel), [Go client](./examples/embedded-backend-go-client)) |
| `website/` | Landing page |

## Build Tags

| Tag | Purpose |
|-----|---------|
| *(none)* | SQLite storage — embedded mode, zero dependencies. This is the default. |
| `pgch` | ClickHouse + PostgreSQL storage — standalone server mode. |
| `localdist` | Embeds frontend from `static/dist/` instead of `static/frontend/`. Used by traceway-cloud to inject billing UI. |

```bash
# Embedded mode (SQLite, default)
cd backend && go build ./cmd/traceway

# Standalone server (ClickHouse + PostgreSQL)
cd backend && go build -tags pgch ./cmd/traceway
```

## Running Tests

```bash
# SQLite tests (default, no tags needed)
cd backend && go test -v -count=1 ./app/repositories/

# ClickHouse + PostgreSQL tests (requires Docker)
./scripts/test-backend-pgch.sh
```

## Documentation

Full documentation at **[docs.tracewayapp.com](https://docs.tracewayapp.com)**:

- [**Client SDKs**](https://docs.tracewayapp.com/client) — Go, Node.js, and OpenTelemetry integration guides
- [**Self-Hosting**](https://docs.tracewayapp.com/server) — Docker Compose and deployment options
- [**Concepts**](https://docs.tracewayapp.com/learn) — How tracing, exception grouping, metrics, and alerts work
- [**Embedded Mode**](https://docs.tracewayapp.com/learn/embedded-mode) — Run Traceway inside your Go app

## Links

- [Website](https://tracewayapp.com)
- [Documentation](https://docs.tracewayapp.com)
- [Go Client SDK](https://github.com/tracewayapp/go-client)
