# Embedded Backend Example (Go Client)

A single-file Go example that runs the Traceway backend and a Gin app instrumented with the Traceway Go client SDK in one process. No external databases required — everything runs on SQLite.

## What it does

1. Starts the Traceway backend on `:8082` with a pre-seeded user and project
2. Configures the Traceway Gin middleware to send traces and exceptions to the backend
3. Runs a Gin server on `:8080` with a single `/hello/:name` endpoint that optionally records an error

## Running

```bash
go build .
./embedded-backend-go-client
```

Then try:

- `http://localhost:8080/hello/world` — successful request
- `http://localhost:8080/hello/error` — request that captures an exception

Open the dashboard at `http://localhost:8082` and log in with `admin@localhost.com` / `admin` to see the traces.
