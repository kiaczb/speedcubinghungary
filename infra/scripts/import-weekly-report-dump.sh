#!/usr/bin/env bash

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

DUMP="${1:-}"
if [[ -z "$DUMP" || ! -f "$DUMP" ]]; then
  echo "usage: ENV=dev|test|prod $0 <dump.sql>" >&2
  exit 1
fi
DUMP="$(cd "$(dirname "$DUMP")" && pwd)/$(basename "$DUMP")"

cd "$REPO_ROOT"

ENV="${ENV:-dev}"
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

compose() {
  docker compose -f infra/docker-compose.yml -f "$OVERRIDE" --env-file "$ENV_FILE" "$@"
}

if [[ -z "$(compose ps --status running -q db)" ]]; then
  echo "ERROR: the db container is not running (env=${ENV})" >&2
  exit 1
fi

PERSONS="$(grep -c '^INSERT INTO public\.persons VALUES ' "$DUMP" || true)"
RECORDS="$(grep -c '^INSERT INTO public\.records VALUES ' "$DUMP" || true)"
if [[ "$PERSONS" -eq 0 ]]; then
  echo "ERROR: no 'INSERT INTO public.persons' rows in $DUMP" >&2
  exit 1
fi
echo "=== importing ${PERSONS} persons and ${RECORDS} records into env=${ENV} ==="

{
  cat <<'SQL'
\set ON_ERROR_STOP on
BEGIN;

CREATE TEMP TABLE import_persons (
    wca_id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    competition_count INTEGER,
    delegated_count INTEGER,
    successful_solves_count INTEGER
) ON COMMIT DROP;

CREATE TEMP TABLE import_records (
    region TEXT NOT NULL,
    event_id TEXT NOT NULL,
    single INTEGER,
    average INTEGER,
    PRIMARY KEY (region, event_id)
) ON COMMIT DROP;
SQL

  grep -E '^INSERT INTO public\.(persons|records) VALUES ' "$DUMP" \
    | sed -e 's/^INSERT INTO public\.persons VALUES /INSERT INTO import_persons VALUES /' \
          -e 's/^INSERT INTO public\.records VALUES /INSERT INTO import_records VALUES /'

  cat <<'SQL'

SELECT
    count(*) FILTER (WHERE u.wca_id IS NULL) AS new_users,
    count(*) FILTER (WHERE u.wca_id IS NOT NULL) AS updated_users
FROM import_persons p
LEFT JOIN users u ON u.wca_id = p.wca_id;

INSERT INTO users (wca_id, name, competition_count, delegated_count, successful_solves_count)
SELECT wca_id, name, COALESCE(competition_count, 0), COALESCE(delegated_count, 0), COALESCE(successful_solves_count, 0)
FROM import_persons
ON CONFLICT (wca_id) DO UPDATE SET
    name = COALESCE(NULLIF(users.name, ''), EXCLUDED.name),
    competition_count = EXCLUDED.competition_count,
    delegated_count = EXCLUDED.delegated_count,
    successful_solves_count = CASE
        WHEN EXCLUDED.successful_solves_count > 0 THEN EXCLUDED.successful_solves_count
        ELSE users.successful_solves_count
    END,
    updated_at = NOW();

INSERT INTO records (region, event_id, single, average)
SELECT region, event_id, single, average
FROM import_records
ON CONFLICT (region, event_id) DO NOTHING;

SELECT
    (SELECT count(*) FROM users) AS users_total,
    (SELECT count(*) FROM users WHERE competition_count > 0) AS with_competitions,
    (SELECT count(*) FROM users WHERE delegated_count > 0) AS delegates,
    (SELECT count(*) FROM users WHERE successful_solves_count > 0) AS with_solves,
    (SELECT count(*) FROM records) AS records_total;

COMMIT;
SQL
} | compose exec -T db sh -c 'psql -X -q -U "$POSTGRES_USER" -d "$POSTGRES_DB"'

echo "=== import done (env=${ENV}) ==="
