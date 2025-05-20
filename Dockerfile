# ---------- Stage 1: Build ----------
FROM golang:alpine AS builder

RUN apk --no-cache add git ca-certificates
WORKDIR /app

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o api-server .

# ---------- Stage 2: Runtime ----------
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add git ca-certificates jq curl openssh

# Create app user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Copy built binary and resources
COPY --from=builder /app/api-server .
COPY watchTV.sh index.html .botenv ./

# Permissions
RUN chmod +x /app/watchTV.sh && \
    chown -R appuser:appuser /app

# Setup cron (no output redirection to file → output goes to stdout)
RUN echo "0 0 * * * /app/watchTV.sh" > /etc/crontabs/appuser

# Git config setup + entrypoint
RUN printf '#!/bin/sh\n\n' > /entrypoint.sh && \
    printf 'git config --global user.name "AlaskarTV-Bot"\n' >> /entrypoint.sh && \
    printf 'git config --global user.email "no-reply@askar.tv"\n' >> /entrypoint.sh && \
    printf 'git config --global credential.helper store\n' >> /entrypoint.sh && \
    printf 'git config --global --add safe.directory /data/alaskartv-forge/alaskartv-docker\n' >> /entrypoint.sh && \
    printf 'git config --global --add safe.directory /data/alaskartv-forge/alaskartv-app\n\n' >> /entrypoint.sh && \
    # Start cron in background
    printf 'crond -b -l 8\n' >> /entrypoint.sh && \
    # Start main Go app in foreground
    printf 'exec /app/api-server\n' >> /entrypoint.sh && \
    chmod +x /entrypoint.sh

USER appuser
ENTRYPOINT ["/entrypoint.sh"]

