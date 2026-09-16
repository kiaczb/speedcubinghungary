# syntax=docker/dockerfile:1
# Multi-stage build for the Go API.
# Build context is the REPO ROOT (see docker-compose.yml) so the prod image can
# bake in db/migrations. Stage selection is driven by the compose `target:`:
#   - dev  -> air live-reload, source bind-mounted at runtime (docker-compose.dev.yml)
#   - prod -> minimal static-binary image (default; used by test/prod overrides)

# ---------- dev: hot reload via air (source bind-mounted at runtime) ----------
FROM golang:1.25.5-alpine AS dev
WORKDIR /app
RUN go install github.com/air-verse/air@latest
# Pre-warm the module cache as its own layer. At runtime the source tree is
# bind-mounted over /app and modules are cached in the go_mod_cache volume.
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download
EXPOSE 8080
CMD ["air", "-c", ".air.toml"]

# ---------- build: compile a fully static binary ----------
FROM golang:1.25.5-alpine AS build
WORKDIR /src
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download
COPY apps/api/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /server ./cmd/server

# ---------- prod: minimal runtime image ----------
FROM alpine:3.20 AS prod
# ca-certificates: required for outbound HTTPS (WCA API).
# su-exec: lets the entrypoint start as root (to chown a fresh uploads volume,
# which Docker creates root:root) and then drop to appuser before exec'ing.
RUN apk add --no-cache ca-certificates su-exec && addgroup -S -g 1001 appuser && adduser -S -u 1001 -G appuser appuser
WORKDIR /app
COPY --from=build /server /app/server
# golang-migrate reads file:///migrations on startup (see cmd/server/main.go).
COPY db/migrations /migrations
COPY infra/docker/api/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
# Stays root here on purpose — entrypoint.sh drops to appuser itself.
EXPOSE 8080
ENTRYPOINT ["/entrypoint.sh"]
