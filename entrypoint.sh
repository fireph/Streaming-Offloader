#!/usr/bin/env bash
set -e

# Default to UID/GID 1000 if not provided
RUN_USER=${RUN_USER:-1000}
RUN_GROUP=${RUN_GROUP:-1000}

# Create group if it doesn’t exist
if ! getent group $RUN_GROUP >/dev/null; then
  groupadd -g $RUN_GROUP streamer
fi
# Create user if it doesn’t exist
if ! id -u streamer >/dev/null 2>&1; then
  useradd -u $RUN_USER -g $RUN_GROUP -M streamer
fi

# Ensure config directory exists
mkdir -p /config

# If no config.yaml exists, populate default and set ownership
if [ ! -f /config/config.yaml ]; then
  cp /app/default-config.yaml /config/config.yaml
  chown ${RUN_USER}:${RUN_GROUP} /config/config.yaml || true
fi

# Drop privileges and launch
exec gosu streamer "$@"
