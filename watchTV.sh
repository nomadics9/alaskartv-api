#!/bin/sh

if [ -f .botenv ]; then
    export $(grep -v '^#' .botenv | xargs)
fi

LAST_RELEASE_FILE=/data/alaskartv/androidtv-ci/release.txt

latest_release=$(curl -s "https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases/latest" | jq -r '.tag_name')
if [ -f "$LAST_RELEASE_FILE" ]; then
    read -r version < "$LAST_RELEASE_FILE"
else
    version=""
fi
if [ "$latest_release" != "$version" ]; then
  # echo "$latest_release" > "$LAST_RELEASE_FILE"
  curl -X POST "https://api.telegram.org/bot$BOT_TOKEN/sendMessage" \
    -d chat_id="$CHAT_ID" \
    -d text="🚀 New Release Detected: $latest_release on $REPO_OWNER/$REPO_NAME"
  curl -X POST "http://127.0.0.1:8080/alaskartv"
fi
