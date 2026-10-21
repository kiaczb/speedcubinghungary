# syntax=docker/dockerfile:1

FROM golang:1.25.5-alpine AS dev
WORKDIR /app
RUN go install github.com/air-verse/air@latest
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download
EXPOSE 8080
CMD ["air", "-c", ".air.toml"]

FROM golang:1.25.5-alpine AS build
WORKDIR /src
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download
COPY apps/api/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /server ./cmd/server

FROM alpine:3.20 AS prod
RUN apk add --no-cache ca-certificates su-exec && addgroup -S -g 1001 appuser && adduser -S -u 1001 -G appuser appuser
WORKDIR /app
COPY --from=build /server /app/server
COPY db/migrations /migrations
COPY infra/docker/api/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
EXPOSE 8080
ENTRYPOINT ["/entrypoint.sh"]
