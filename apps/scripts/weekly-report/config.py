"""Environment configuration for the weekly-report job.

Single source of env config: the SAME root `.env.<ENV>` file the rest of the
stack uses (`.env.dev` / `.env.test` / `.env.prod`), not a per-app `.env`.
`ENV` selects it (default `dev`). Mirrors apps/api/internal/config/config.go:
no hardcoded secrets, everything comes from the environment.

How the file is found in each run mode:
  - container (compose `jobs` profile): the values already arrive as environment
    variables from `--env-file`, and the root file isn't copied into the image,
    so `load_dotenv` simply no-ops.
  - standalone (venv on the host): `load_dotenv` reads the repo-root `.env.<ENV>`.

DB host caveat: the root file targets the Docker network (`DB_HOST=db`,
`DATABASE_URL=...@db`). That only resolves inside a container. When this job runs
on the host (venv), the dev compose exposes the DB on `localhost:5432`, so we
detect "not in a container" and connect to localhost instead — see below.
"""

import os
from datetime import datetime

from dotenv import load_dotenv

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
# apps/scripts/weekly-report -> repo root
REPO_ROOT = os.path.abspath(os.path.join(SCRIPT_DIR, "..", "..", ".."))

ENV = os.getenv("ENV", "dev")  # dev | test | prod
load_dotenv(os.path.join(REPO_ROOT, f".env.{ENV}"))

# True inside a Docker container (the daemon creates /.dockerenv). Used to decide
# whether the DB is reachable at the compose service name or on the host.
IN_CONTAINER = os.path.exists("/.dockerenv")

# --- Report constants (not secrets) ---
badges = ["WR", "CR", "NR"]
excludedCompetitorWcaIds = ["2017TSVE02"]

# --- Email configuration ---
# EMAIL_TO can be pointed at a fixed test inbox in a test/staging context so
# real recipients are never emailed from there.
emailFrom = os.getenv("EMAIL_FROM", "")
emailTo = os.getenv("EMAIL_TO", "")
emailSubject = f"{datetime.now().strftime('%Y-%m-%d')} Weekly Hungarian competitors dump "
emailCode = os.getenv("EMAIL_APP_PASSWORD", "")  # Gmail app password - secret

# --- Database configuration ---
# The weekly report shares the SAME PostgreSQL database as the Go backend.
DB_PORT = os.getenv("DB_PORT", "5432")
DB_NAME = os.getenv("DB_NAME", "speedcubinghungary_dev")
DB_USER = os.getenv("DB_USER", "")
DB_PASSWORD = os.getenv("DB_PASSWORD", "")

if IN_CONTAINER:
    # Inside the compose network: use the root file's values as-is (DB_HOST=db,
    # DATABASE_URL=...@db). A full DATABASE_URL takes precedence when present.
    DATABASE_URL = os.getenv("DATABASE_URL")
    DB_HOST = os.getenv("DB_HOST", "db")
else:
    # On the host (venv): the container-internal `db` host isn't resolvable, so
    # ignore DATABASE_URL (it points at @db) and connect over the exposed port on
    # localhost using the DB_* credentials from the same root file.
    DATABASE_URL = None
    DB_HOST = "localhost"


def validate():
    """Fails fast if required config is missing, mirroring the Go backend's
    config.Validate()."""
    missing = [
        name
        for name, value in {
            "EMAIL_FROM": emailFrom,
            "EMAIL_TO": emailTo,
            "EMAIL_APP_PASSWORD": emailCode,
        }.items()
        if not value
    ]
    if not DATABASE_URL and not (DB_USER and DB_PASSWORD):
        missing.append("DATABASE_URL (or DB_USER + DB_PASSWORD)")

    if missing:
        raise RuntimeError(
            f"Missing required environment variables: {missing}\n"
            f"They come from the root .env.{ENV} file (see .env.{ENV}.example). "
            "Set ENV=dev|test|prod to pick the environment."
        )
