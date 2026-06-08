# syntax=docker/dockerfile:1.7

# ---- Build stage ----
FROM golang:1.25-alpine AS build

# Keep this in lockstep with:
#   - .mise.toml              -> gomigrate = "4.18.3"
#   - .github/workflows/backend.yml  -> MIGRATE_VERSION: v4.18.3
ARG MIGRATE_VERSION=v4.18.3

WORKDIR /src

# Cache module downloads as a separate layer
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
  go mod download

# Copy the rest of the source
COPY . .

# Build api and seed as static, stripped binaries.
# Cache mounts speed up incremental rebuilds.
RUN --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
  CGO_ENABLED=0 GOOS=linux \
  go build -trimpath -ldflags="-s -w" -o /out/dozingo       ./cmd/api && \
  CGO_ENABLED=0 GOOS=linux \
  go build -trimpath -ldflags="-s -w" -o /out/seed          ./cmd/seed && \
  CGO_ENABLED=0 GOOS=linux \
  go build -trimpath -ldflags="-s -w" -o /out/sweep-orphans ./cmd/sweep-orphans

# Build the golang-migrate CLI with only the postgres driver to keep
# the binary small (~5 MB vs ~15 MB for all drivers).
RUN --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
  CGO_ENABLED=0 GOOS=linux \
  go install -tags 'postgres' -trimpath -ldflags="-s -w" \
  github.com/golang-migrate/migrate/v4/cmd/migrate@${MIGRATE_VERSION}

# ---- Runtime stage ----
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata \
  && addgroup -S app \
  && adduser  -S -G app -H -s /sbin/nologin app

COPY --from=build /out/dozingo                 /usr/local/bin/dozingo
COPY --from=build /out/seed                    /usr/local/bin/seed
COPY --from=build /out/sweep-orphans           /usr/local/bin/sweep-orphans
COPY --from=build /go/bin/migrate              /usr/local/bin/migrate
COPY --from=build /src/internal/db/migrations  /app/migrations

USER app
WORKDIR /home/app

# Hits the /api/health endpoint on the configured PORT (defaults to 4242).
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- "http://127.0.0.1:${PORT:-4242}/api/health" >/dev/null || exit 1

ENTRYPOINT ["/usr/local/bin/dozingo"]
