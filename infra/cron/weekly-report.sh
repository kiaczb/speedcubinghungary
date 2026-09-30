#!/usr/bin/env bash

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$REPO_ROOT"

ENV="${ENV:-test}"
ENV_FILE=".env.${ENV}"
OVERRIDE="infra/docker-compose.${ENV}.yml"

if [[ ! -f "$ENV_FILE" ]]; then
  echo "ERROR: $ENV_FILE not found in $REPO_ROOT" >&2
  exit 1
fi
if [[ ! -f "$OVERRIDE" ]]; then
  echo "ERROR: $OVERRIDE not found (valid ENV values: dev, test, prod)" >&2
  exit 1
fi

echo "=== [$(date -Is)] weekly-report start (env=${ENV}) ==="

docker compose \
  -f infra/docker-compose.yml \
  -f "$OVERRIDE" \
  --env-file "$ENV_FILE" \
  --profile jobs run --rm weekly-report

echo "=== [$(date -Is)] weekly-report done (env=${ENV}) ==="
