FROM golang:1.23-alpine AS builder

RUN apk --no-cache add git ca-certificates
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o api-server .

FROM alpine:latest
RUN apk --no-cache add git ca-certificates jq curl openssh

WORKDIR /app
COPY --from=builder --chown=1000:1000 /app/api-server .

RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser


RUN printf '#!/bin/sh\n\n' > /entrypoint.sh && \
    # printf 'if [ -n "$GIT_USERNAME" ] && [ -n "$FORGEJO_TOKEN" ]; then\n' >> /entrypoint.sh && \
    printf '    git config --global user.name "AlaskarTV-Bot"\n' >> /entrypoint.sh && \
    printf '    git config --global user.email "no-reply@askar.tv}"\n' >> /entrypoint.sh && \
    printf '    git config --global credential.helper store\n' >> /entrypoint.sh && \
    # printf '    echo "https://${GIT_USERNAME}:${FORGEJO_TOKEN}@git.askar.tv" > /root/.git-credentials\n' >> /entrypoint.sh && \
    # printf '    chmod 600 /root/.git-credentials\n' >> /entrypoint.sh && \
    # printf 'fi\n\n' >> /entrypoint.sh && \
    # # Add safe directories
     printf 'git config --global --add safe.directory /data/alaskartv-forge/alaskartv-docker\n' >> /entrypoint.sh && \
     printf 'git config --global --add safe.directory /data/alaskartv-forge/alaskartv-app\n\n' >> /entrypoint.sh && \
    # Finally, run the main app
    printf 'crond -b\n' >> /entrypoint.sh && \
    printf 'exec /app/api-server\n' >> /entrypoint.sh && \
    chmod +x /entrypoint.sh

COPY watchTV.sh /app/watchTV.sh
COPY frontend.html /app/index.html
COPY .botenv /app/.botenv
RUN chmod +x /app/watchTV.sh && \
    echo "0 0 * * * /app/watchTV.sh" > /etc/crontabs/root

USER appuser
ENTRYPOINT ["/entrypoint.sh"]

