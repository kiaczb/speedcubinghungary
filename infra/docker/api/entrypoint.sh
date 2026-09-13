#!/bin/sh
# Runs as root so it can fix ownership of the uploads volume (a fresh named
# volume is created root:root by Docker, but the server runs as appuser), then
# drops privileges before exec'ing the server. See CLAUDE.md §17 upload notes.
set -e

UPLOAD_DIR="${UPLOAD_DIR:-/data/uploads}"
mkdir -p "$UPLOAD_DIR"
chown -R appuser:appuser "$UPLOAD_DIR"

exec su-exec appuser /app/server
