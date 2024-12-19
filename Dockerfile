FROM golang:1.23-alpine AS builder

RUN apk --no-cache add git ca-certificates
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o api-server .

FROM alpine:latest
RUN apk --no-cache add git ca-certificates

WORKDIR /app
COPY --from=builder /app/api-server .

RUN printf '#!/bin/sh\n\n' > /entrypoint.sh && \
    printf 'if [ -n "$GIT_USERNAME" ] && [ -n "$GITHUB_TOKEN" ]; then\n' >> /entrypoint.sh && \
    printf '    git config --global user.name "$GIT_USERNAME"\n' >> /entrypoint.sh && \
    printf '    git config --global user.email "${GIT_EMAIL:-no-reply@example.com}"\n' >> /entrypoint.sh && \
    printf '    git config --global credential.helper store\n' >> /entrypoint.sh && \
    printf '    echo "https://${GIT_USERNAME}:${GITHUB_TOKEN}@github.com" > /root/.git-credentials\n' >> /entrypoint.sh && \
    printf '    chmod 600 /root/.git-credentials\n' >> /entrypoint.sh && \
    printf 'fi\n\n' >> /entrypoint.sh && \
    # Add safe directories
    printf 'git config --global --add safe.directory /data/alaskartv/docker-ci\n' >> /entrypoint.sh && \
    printf 'git config --global --add safe.directory /data/alaskartv/androidtv-ci\n\n' >> /entrypoint.sh && \
    # Finally, run the main app
    printf 'exec /app/api-server\n' >> /entrypoint.sh && \
    chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]

