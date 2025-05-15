#!/usr/bin/env bash
set -e
# default to nobody if not set
RUN_USER=${RUN_USER:-65534}
RUN_GROUP=${RUN_GROUP:-65534}
# create group if missing
if ! getent group $RUN_GROUP >/dev/null; then
  groupadd -g $RUN_GROUP streamer
fi
# create user if missing
if ! id streamer >/dev/null 2>&1; then
  useradd -u $RUN_USER -g streamer -M streamer
fi
# drop privileges
exec gosu streamer "$@"
