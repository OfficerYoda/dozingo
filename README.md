# Dozingo

A bingo game for university lectures.

---

## Quick Start

No local toolchain needed — just Docker with Compose v2:

```bash
git clone https://github.com/OfficerYoda/dozingo.git
cd dozingo/deploy
cp .env.example .env
docker compose up -d
```

Open **http://localhost:4242** — the app starts with demo data already loaded.

Log in with `admin` / `password123`.

See [`deploy/README.md`](deploy/README.md) for more details.

---

## Features

### Gameplay

- **Bingo boards** in 4×4, 5×5, and 6×6 layouts with custom cell text
- **Server-side bingo detection** — rows, columns, and both diagonals checked on
  every move
- **Live timer** with game resume — elapsed time is persisted and restored on
  page reload
- **Cell persistence** — marked cells and completed lines survive a refresh
- **Victory celebration** — confetti, falling bingo balls, playing card
  animations, and a procedurally-generated techno beat
- **Share button** — native Web Share API to share any game directly

### Boards

- Browse, search, and sort the community board library (newest, most liked, most
  played, and more)
- Like/unlike boards with an optimistic UI
- Community activity stats (bingos, games, and boards created in the last week)

### Board Editor

- Spreadsheet-style entry table with keyboard navigation (Enter to add a row,
  Backspace to delete)
- Save as a reusable template or jump straight into a game

### Accounts & Security

- Registration, login, and password reset via email
- TOTP two-factor authentication — QR code setup, TOTP verification, one-time
  recovery codes, recovery code regeneration
- Session management — cookie-based sessions with sliding 30-day expiry, all
  sessions invalidated on password change
- Login notification emails
- Account deletion with password confirmation

### Profiles

- Auto-generated SVG avatars on registration (DiceBear)
- Custom avatar upload (PNG, JPEG, WEBP, max 20 MB)
- Inline username editing
- Game history and continue-where-you-left-off board slider
- Liked boards collection

### Accessibility & Personalisation

- **Dark mode** (persisted to localStorage)
- **Color correction modes** — red-green, blue-yellow, and grayscale (for color
  vision deficiencies)
- **English / German** localisation (vue-i18n)

---

## Tech Stack

| Layer          | Technology                                      |
| -------------- | ----------------------------------------------- |
| Backend        | Go 1.25, chi, huma v2 (OpenAPI)                 |
| Database       | PostgreSQL 16, pgx/v5, sqlc, golang-migrate     |
| Object Storage | Garage (self-hosted S3-compatible, for avatars) |
| Email          | Resend                                          |
| Frontend       | Vue 3, TypeScript, Vite                         |
| Serving        | Caddy (SPA + reverse proxy)                     |

---

## Repository Structure

```
dozingo/
├── backend/    ← Go API server (see backend/README.md)
├── frontend/   ← Vue 3 SPA (see frontend/README.md)
└── deploy/     ← Docker Compose setup for running the full stack
```

---

## Development Setup

Requires [mise](https://mise.jdx.dev/) and Docker.

```bash
# Install all tools (Go, just, sqlc, golang-migrate, golangci-lint)
mise install

# First-time setup: start infra, run migrations, codegen, install deps
just setup

# Start the backend
just backend run

# Start the frontend (separate terminal)
just frontend run
```

Run `just --list --list-submodules` for all available commands.

---

## AI Notice

All READMEs in this repository (`/README.md`, `backend/README.md`,
`frontend/README.md`, `deploy/README.md`) were generated with AI assistance.

The frontend test suite (`frontend/src/**/__tests__/`) and all backend tests
(`backend/internal/**/*_test.go`) were also written with AI assistance.
