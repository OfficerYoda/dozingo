# Dozingo Backend

Go API server for the Dozingo bingo game.

---

## Tech Stack

| Layer           | Technology                                   |
| --------------- | -------------------------------------------- |
| Language        | Go 1.25                                      |
| HTTP Router     | chi/v5                                       |
| API Framework   | huma/v2 (automatic validation, OpenAPI docs) |
| Database        | PostgreSQL 16                                |
| DB Driver       | pgx/v5 (connection pooling via pgxpool)      |
| Code Generation | sqlc (type-safe Go from SQL queries)         |
| Migrations      | golang-migrate                               |
| Object Storage  | Garage (S3-compatible, for avatars)          |
| Email           | Resend                                       |
| Config          | caarlos0/env + godotenv                      |
| Linting         | golangci-lint                                |

## Architecture

```
HTTP Request
     │
     ▼
  chi router + middleware (auth, sessions, rate limiting, logging)
     │
     ▼
  huma handler (automatic request validation, OpenAPI docs)
     │
     ▼
  service layer (bingo detection, avatar generation, email dispatch)
     │
     ▼
  repository (sqlc-generated type-safe queries)
     │
     ▼
  pgx → PostgreSQL
```

## Setup

Make sure you have run the project-level setup first (see root README).

```bash
# From the project root:
just setup

# Or just the backend-specific parts:
just backend setup
```

This will:

1. Copy `.env.example` to `.env` (if not present)
2. Run database migrations
3. Generate Go code from SQL queries
4. Install Go dependencies

## Running

```bash
just backend run
```

The server runs at [http://localhost:4242](http://localhost:4242)

API docs (OpenAPI/Swagger UI) are at
[http://localhost:4242/docs](http://localhost:4242/docs)

## Environment Variables

Copy `.env.example` to `.env` and adjust as needed:

| Variable              | Description                                                     | Default                                                                            |
| --------------------- | --------------------------------------------------------------- | ---------------------------------------------------------------------------------- |
| `DATABASE_URL`        | PostgreSQL connection string                                    | `postgres://dozingo_user:dozingo_pass@localhost:5432/dozingo?sslmode=disable`      |
| `TEST_DATABASE_URL`   | Test database connection string                                 | `postgres://dozingo_user:dozingo_pass@localhost:5432/dozingo_test?sslmode=disable` |
| `PORT`                | Server port                                                     | `4242`                                                                             |
| `SECURE_COOKIE`       | Set `Secure` flag on session cookie (disable for local HTTP)    | `true`                                                                             |
| `TOTP_ENCRYPTION_KEY` | Base64-encoded 32-byte AES-256-GCM key for TOTP secrets at rest | —                                                                                  |
| `RESEND_API_KEY`      | Resend API key for transactional email                          | —                                                                                  |
| `MAIL_SENDER_ADDRESS` | From address for outgoing emails                                | —                                                                                  |
| `GARAGE_ENDPOINT`     | Internal S3 API URL for Garage                                  | —                                                                                  |
| `GARAGE_PUBLIC_URL`   | Public base URL for serving avatar images                       | —                                                                                  |
| `GARAGE_ACCESS_KEY`   | Garage S3 access key                                            | —                                                                                  |
| `GARAGE_SECRET_KEY`   | Garage S3 secret key                                            | —                                                                                  |
| `GARAGE_BUCKET_NAME`  | S3 bucket name for avatars                                      | `profile-pictures`                                                                 |

## Daily Workflow

```bash
# Start infrastructure (if not already running)
just infra-up

# Work on the server
just backend run
```

### Making a database change

```bash
# 1. Create migration files
just backend migrate-create add_lecturers_table

# 2. Write your SQL in the new .up.sql and .down.sql files
#    (in internal/db/migrations/)

# 3. Apply the migration
just backend migrate-up

# 4. Write queries in internal/db/queries/

# 5. Regenerate Go code
just backend generate

# 6. Use the generated code in your handlers
```

### Adding a new API endpoint

1. Write SQL queries in `internal/db/queries/*.sql`
2. Run `just backend generate`
3. Create or edit a handler in `internal/handler/`
4. Register the route in `cmd/api/main.go`
5. API docs update automatically at `/docs`

## Background Workers

Five periodic goroutines run in the background:

| Worker                       | Interval | What it does                                                 |
| ---------------------------- | -------- | ------------------------------------------------------------ |
| `session_cleanup`            | 1 hour   | Deletes expired session rows                                 |
| `verification_token_cleanup` | 1 hour   | Deletes expired email verification and password reset tokens |
| `avatar_orphan_cleanup`      | 1 hour   | Removes S3 objects no longer referenced by any user          |
| `game_abandon`               | 1 hour   | Marks games as abandoned after 6 hours of inactivity         |
| `game_session_cleanup`       | 1 minute | Closes stale game sessions (no heartbeat for >1 minute)      |
