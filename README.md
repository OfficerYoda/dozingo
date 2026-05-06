# Dozingo

A bingo game for university lectures.

---

## Repository Structure

This is a monorepo containing both the backend and frontend:

```
dozingo/
├── backend/    ← Go API server (see backend/README.md)
├── frontend/   ← Web frontend (see frontend/README.md)
├── justfile    ← Project-wide commands + submodule imports
└── docker-compose.yml ← Shared infrastructure (PostgreSQL)
```

## Prerequisites

Install these two things manually. Everything else is handled by
[mise](https://mise.jdx.dev/) (a tool version manager), Docker, and package managers.

| Tool                      | What it's for                                                  | Install                                                                      |
| ------------------------- | -------------------------------------------------------------- | ---------------------------------------------------------------------------- |
| **mise**                  | Manages all dev tools (Go, just, sqlc, golang-migrate, linter) | `brew install mise` or [see docs](https://mise.jdx.dev/getting-started.html) |
| **Docker** (with Compose) | Runs PostgreSQL locally                                        | `brew install docker` or [see docs](https://docs.docker.com/)                |

> **How it works:** Tool versions are pinned in `.mise.toml` at the project
> root. Running `mise install` (or `just setup`) installs exactly the right
> versions of Go, just, sqlc, golang-migrate, and golangci-lint. No more "works
> on my machine" issues.

### Activating mise (one-time)

After installing mise, you need to activate it in your shell so that tools are
available automatically. Run **one** of these depending on your shell, then
**restart your terminal**:

```bash
# bash
echo 'eval "$(mise activate bash)"' >> ~/.bashrc

# zsh
echo 'eval "$(mise activate zsh)"' >> ~/.zshrc

# fish
echo 'mise activate fish | source' >> ~/.config/fish/config.fish
```

Verify everything works:

```bash
mise --version       # any recent version
docker --version     # any recent version
```

## Getting Started

```bash
# 1. Clone the project
git clone https://github.com/officeryoda/dozingo.git
cd dozingo

# 2. Install all tools (Go, just, sqlc, etc.)
mise install

# 3. Run first-time setup (infra, migrations, codegen, deps)
just setup

# 4. Start the backend server
just backend run
```

The API server runs at [http://localhost:4242](http://localhost:4242)

API docs are at [http://localhost:4242/docs](http://localhost:4242/docs)

## Available Commands

Run `just` (with no arguments) to see top-level commands.
Run `just --list --list-submodules` to see all commands including submodules.

### Root commands (infrastructure)

```bash
just setup          # First-time setup (tools, infra, backend setup)
just tools          # Install/update project tools via mise
just infra-up       # Start postgres
just infra-down     # Stop postgres
just infra-reset    # Wipe database and restart from scratch
just db-shell       # Open a psql shell to the local database
```

### Backend commands (`just backend <command>`)

```bash
just backend run             # Start the Go server
just backend test            # Run all tests
just backend lint            # Run linter
just backend migrate-up      # Apply pending migrations
just backend migrate-down    # Roll back the last migration
just backend migrate-create  # Create a new migration file
just backend generate        # Regenerate Go code from SQL
just backend seed            # Seed database with sample data
just backend setup           # Backend-specific setup (env, migrations, codegen)
```

### Frontend commands (`just frontend <command>`)

```bash
just frontend run    # Serve the frontend locally
just frontend build  # Build for production
just frontend test   # Run tests
```
