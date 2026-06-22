# syntax=docker/dockerfile:1.7
# Build context: repo root (so COPY frontend/ works)

# ---- Build stage ----
FROM node:22-alpine AS build

WORKDIR /app
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# ---- Runtime stage ----
FROM caddy:2-alpine

COPY frontend/Caddyfile /etc/caddy/Caddyfile
COPY --from=build /app/dist /srv
