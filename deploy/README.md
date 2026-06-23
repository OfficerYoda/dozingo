# Dozingo — Quick Start

Run the full application with a single command. Requires only **Docker** (with
Compose v2).

## Setup

```bash
cp .env.example .env
docker compose up -d
```

Open **http://localhost:4242**

## What happens on first run

| Step         | What it does                                                         |
| ------------ | -------------------------------------------------------------------- |
| **postgres** | Starts the database                                                  |
| **garage**   | Starts S3-compatible object storage (for avatars)                    |
| **migrate**  | Applies all database migrations                                      |
| **seed**     | Populates demo users, boards, and games                              |
| **backend**  | Starts the Go API server                                             |
| **frontend** | Starts Caddy, serves the SPA on :4242, proxies `/api` to the backend |

## Demo credentials

| Username | Password      |
| -------- | ------------- |
| `admin`  | `password123` |

## Stopping

```bash
docker compose down        # stop, keep data
docker compose down -v     # stop and wipe all data (fresh start)
```

## Skipping the seed

To start with an empty database (no demo data):

```bash
docker compose up -d --scale seed=0
```

## Notes

- Avatar images load from **http://localhost:3902** (Garage web endpoint) — this
  port must be free on your machine
- Email features (password reset, 2FA setup) won't deliver mail with the dummy
  `RESEND_API_KEY`; the app still runs normally. For a production workflow with
  working emails, see [dozingo.de](https://dozingo.de)
- `TOTP_ENCRYPTION_KEY` in `.env.example` is a pre-generated local-demo key
