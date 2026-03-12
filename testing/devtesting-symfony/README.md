# devtesting-symfony

Symfony test application for Traceway via OpenTelemetry auto-instrumentation. Mirrors the Go `devtesting` reference app to validate that OTLP trace and metric ingestion works correctly.

## Setup (macOS)

### 1. Install PHP 8.2+

```bash
brew install php
```

### 2. Install the OpenTelemetry PHP extension

```bash
pecl install opentelemetry
```

Verify it's enabled:

```bash
php -m | grep opentelemetry
```

If not listed, add to your `php.ini` (`php --ini` to find it):

```ini
extension=opentelemetry
```

### 3. Install Composer (if not installed)

```bash
brew install composer
```

### 4. Install dependencies

```bash
cd testing/devtesting-symfony
composer install
```

### 5. Configure

Edit `.env` and set your project token:

```
OTEL_EXPORTER_OTLP_HEADERS=Authorization=Bearer YOUR_PROJECT_TOKEN
```

### 6. Run

```bash
php -S localhost:8080 public/index.php
```

## Endpoints

### Test Controller

| Method | Path | Description |
|--------|------|-------------|
| GET | `/test-ok` | Returns 200 `{ status: "ok" }` |
| GET | `/test-not-found` | Returns 404 |
| GET | `/test-param/{param}` | Returns the route parameter |
| GET | `/test-exception` | Throws `RuntimeException` (auto-captured) |
| GET | `/test-spans` | Manual child spans (db.query, cache.set, http.external_api) |
| GET | `/test-metrics` | Manual metrics (gauge + histogram) |
| GET | `/test-message` | Span events (10 messages + 1 exception) |
| GET | `/test-json` | Span with JSON attribute |
| GET | `/test-self-report-attributes` | Exception with custom attributes |
| GET | `/test-cerror-simple` | Simple exception |
| GET | `/test-cerror-wrapped` | Nested exceptions (previous chain) |
| GET | `/test-cerror-custom` | Custom exception class |
| GET | `/test-cerror-nested` | Error from nested function calls |
| POST | `/test-recording/{param}` | JSON body parsing, throws if name != "good" |

### Users Controller (CRUD)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/users` | List all users |
| GET | `/users/{id}` | Get user by ID |
| POST | `/users` | Create user (`{ first_name, last_name, email }`) |
| PUT | `/users/{id}` | Update user |
| DELETE | `/users/{id}` | Delete user |

Uses SQLite file-based DB (`var/devtesting.db`), auto-created on first request. The `opentelemetry-auto-pdo` package auto-instruments all DB queries as spans.

## Testing

```bash
# Basic endpoints
curl http://localhost:8080/test-ok
curl http://localhost:8080/test-exception
curl http://localhost:8080/test-spans
curl http://localhost:8080/test-metrics

# User CRUD
curl -X POST http://localhost:8080/users \
  -H 'Content-Type: application/json' \
  -d '{"first_name":"John","last_name":"Doe","email":"john@example.com"}'

curl http://localhost:8080/users
```

Check the Traceway dashboard for endpoints, exceptions, spans, and metrics.
