#!/bin/sh

log() {
    echo "[CRON][$(date '+%Y-%m-%d %H:%M:%S')] $*"
}

# Load .botenv if present
if [ -f .botenv ]; then
    log "Loading environment from .botenv"
    export $(grep -v '^#' .botenv | xargs)
fi

LAST_RELEASE_FILE=/data/alaskartv-forge/alaskartv-app/release.txt

# Get latest release from GitHub
latest_release=$(curl -s "https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases/latest" | jq -r '.tag_name')
log "Latest Release: '$latest_release'"

# Read current version
if [ -f "$LAST_RELEASE_FILE" ]; then
    read -r version < "$LAST_RELEASE_FILE"
else
    version=""
fi
log "Current Version: '$version'"

# Compare and act
if [ "$latest_release" != "$version" ]; then
    log "New version detected, sending notification and triggering update..."
    
    curl -s -X POST "https://api.telegram.org/bot$BOT_TOKEN/sendMessage" \
        -d chat_id="$CHAT_ID" \
        -d text="🚀 New Release Detected: $latest_release on $REPO_OWNER/$REPO_NAME" | jq

    curl -s -X POST "http://api-server:9090/api/alaskartv" | jq

    # Optional: update version file (uncomment if needed)
    # echo "$latest_release" > "$LAST_RELEASE_FILE"
else
    log "No update needed."
fi

