#!/usr/bin/env sh
# A minimal plugin used in acceptance tests.
# If the first argument is an integer, it is used as the exit code.
# Otherwise, all arguments are echoed and the plugin exits 0.

if [ $# -gt 0 ] && echo "$1" | grep -qE '^[0-9]+$'; then
  code=$1
  shift
  echo "$@"
  exit "$code"
fi

echo "$@"
