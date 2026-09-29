#!/usr/bin/env bash
# Heti riport futtatása cronból.
#
# A job egyetlen futásban:
#   1. legenerálja a riportot a WCA API + a közös DB alapján,
#   2. létrehozza a hozzá tartozó Post bejegyzést a weboldalon,
#   3. elküldi e-mailben (EMAIL_TO).
#
# A 2. lépés IDEMPOTENS: a posts.report_week UNIQUE, így ugyanarra a hétre
# másodszor futtatva nem jön létre újabb poszt (lásd reportPost.py). A cron
# ezért nyugodtan újrapróbálkozhat, és kézzel is indítható.
#
# Telepítés a szerveren (egyszer):
#   chmod +x infra/cron/weekly-report.sh
#   crontab -e
# majd minden hétfő 08:00-ra (teszt környezet):
#   0 8 * * 1 ENV=test /srv/speedcubinghungary-DEV/speedcubinghungary/infra/cron/weekly-report.sh >> /var/log/speedcubinghungary-weekly-report.log 2>&1
# éles környezetben ENV=prod és a prod deploy útvonala.

set -euo pipefail

# A repo gyökerét a script helyéből számoljuk — a cron nem ad értelmes cwd-t.
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

# --rm: a job egyszer fut le és eltűnik. A `jobs` profil miatt a normál
# `docker compose up` sosem indítja el ezt a szolgáltatást.
docker compose \
  -f infra/docker-compose.yml \
  -f "$OVERRIDE" \
  --env-file "$ENV_FILE" \
  --profile jobs run --rm weekly-report

echo "=== [$(date -Is)] weekly-report done (env=${ENV}) ==="
