#!/bin/sh
set -eu

ADDRESS="${HEALTHCHECK_ADDR:-127.0.0.1:56000}"
TIMEOUT="${HEALTHCHECK_TIMEOUT:-5s}"

set -- -healthcheck "$ADDRESS" -healthcheck-timeout "$TIMEOUT"

if [ "${WRAP_MODE:-false}" = "true" ]; then
  : "${WRAP_KEY:?WRAP_KEY is required when WRAP_MODE=true}"
  set -- "$@" -wrap -wrap-key "$WRAP_KEY"
fi

exec ./vk-turn-proxy "$@"
