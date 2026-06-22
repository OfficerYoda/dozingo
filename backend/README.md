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

````

## Setup

Make sure you have run the project-level setup first (see root README).

```bash
# From the project root:
just setup

# Or just the backend-specific parts:
just backend setup
````

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
[http://localhost:4242/api/docs](http://localhost:4242/api/docs)

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
