import { serve } from "@hono/node-server";
import { Hono } from "hono";
import { otel } from "@hono/otel";
import { trace, SpanStatusCode } from "@opentelemetry/api";
import Database from "better-sqlite3";

const tracer = trace.getTracer("hono-otel-example");

// SQLite setup
const db = new Database(":memory:");
db.exec(`
  CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT NOT NULL
  )
`);
db.exec(`
  INSERT INTO users (name, email) VALUES
    ('Alice', 'alice@example.com'),
    ('Bob', 'bob@example.com'),
    ('Charlie', 'charlie@example.com')
`);

// Helper: wrap SQLite calls in a manual span (no auto-instrumentation for SQLite)
function dbSpan(name, query, fn) {
  return tracer.startActiveSpan(name, (span) => {
    span.setAttribute("db.system", "sqlite");
    span.setAttribute("db.statement", query);
    try {
      const result = fn();
      span.end();
      return result;
    } catch (error) {
      span.recordException(error);
      span.setStatus({ code: SpanStatusCode.ERROR, message: error.message });
      span.end();
      throw error;
    }
  });
}

const app = new Hono();

app.use(otel());

// GET /hono/api/users — list all users (manual DB span)
app.get("/hono/api/users", (c) => {
  const users = dbSpan("db.query", "SELECT * FROM users", () =>
    db.prepare("SELECT * FROM users").all()
  );
  return c.json(users);
});

// GET /hono/api/users/:id — get user by ID (manual DB span)
app.get("/hono/api/users/:id", (c) => {
  const id = c.req.param("id");
  const user = dbSpan("db.query", "SELECT * FROM users WHERE id = ?", () =>
    db.prepare("SELECT * FROM users WHERE id = ?").get(id)
  );
  if (!user) {
    return c.json({ error: "User not found" }, 404);
  }
  return c.json(user);
});

// POST /hono/api/users — create user (manual DB span)
app.post("/hono/api/users", async (c) => {
  const body = await c.req.json();
  const result = dbSpan(
    "db.query",
    "INSERT INTO users (name, email) VALUES (?, ?)",
    () =>
      db
        .prepare("INSERT INTO users (name, email) VALUES (?, ?)")
        .run(body.name, body.email)
  );
  const user = { id: Number(result.lastInsertRowid), ...body };
  return c.json(user, 201);
});

// GET /hono/api/slow — simulates slow work (manual span)
app.get("/hono/api/slow", async (c) => {
  await tracer.startActiveSpan("slow-operation", async (span) => {
    await new Promise((resolve) => setTimeout(resolve, 300));
    span.end();
  });
  return c.json({ message: "Slow response" });
});

// GET /hono/api/external — outgoing fetch (auto-instrumented by instrumentation-undici)
app.get("/hono/api/external", async (c) => {
  const res = await fetch("https://jsonplaceholder.typicode.com/posts/1");
  const post = await res.json();
  return c.json(post);
});

// GET /hono/api/test-error — throws error (auto-captured by @hono/otel)
app.get("/hono/api/test-error", () => {
  throw new Error("Test error from Hono");
});

serve({ fetch: app.fetch, port: 3002 }, () => {
  console.log("Hono listening on http://localhost:3002");
});
