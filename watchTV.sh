#!/bin/sh
if [ -f .botenv ]; then
    export $(grep -v '^#' .botenv | xargs)
fi

LAST_RELEASE_FILE=/data/alaskartv/androidtv-ci/release.txt
latest_release=$(curl -s "https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases/latest" | jq -r '.tag_name')

echo "Latest Release: '$latest_release'"

if [ -f "$LAST_RELEASE_FILE" ]; then
    read -r version < "$LAST_RELEASE_FILE"
else
    version=""
fi

echo "Current Version: '$version'"

echo "Comparison result: $latest_release != $version"
if [ "$latest_release" != "$version" ]; then
    echo "Versions are different, updating..."
    curl -X POST "https://api.telegram.org/bot$BOT_TOKEN/sendMessage" \
        -d chat_id="$CHAT_ID" \
        -d text="🚀 New Release Detected: $latest_release on $REPO_OWNER/$REPO_NAME"
    curl -X POST "http://api-server:9090/api/alaskartv"
else
    echo "Versions are the same, no update needed"
fi
