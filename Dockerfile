# syntax=docker/dockerfile:1.6

# ---- Build stage ----
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /src

# Cache deps separately for faster rebuilds
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build both the API server and the migration CLI as static binaries
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/api ./cmd/api \
 && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/migrate ./cmd/migrate

# ---- Runtime stage ----
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata curl \
 && addgroup -S app && adduser -S app -G app \
 && mkdir -p /tmp/tusd && chown app:app /tmp/tusd

WORKDIR /app

# Binaries
COPY --from=builder /out/api     /app/api
COPY --from=builder /out/migrate /app/migrate

# Static assets needed at runtime (SQL migrations + swagger docs)
COPY --from=builder /src/migrations /app/migrations
COPY --from=builder /src/docs       /app/docs

# Default port (overridable via env)
ENV PORT=8080
EXPOSE 8080

USER app

# Healthcheck against the public /api/health route
HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
    CMD curl -fsS "http://localhost:${PORT}/api/health" || exit 1

ENTRYPOINT ["/app/api"]
