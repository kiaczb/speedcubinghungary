"""Environment configuration for the weekly-report job.

Loads `.env` (git-ignored; copy from `.env.example`) via python-dotenv, then
validates that required variables are present — mirroring
apps/api/internal/config/config.go so the Go backend and this job follow the
same rule: no hardcoded secrets, everything comes from the environment.
"""

import os
from datetime import datetime

from dotenv import load_dotenv

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
load_dotenv(os.path.join(SCRIPT_DIR, ".env"))

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
# A full DATABASE_URL takes precedence when present; otherwise the individual
# DB_* settings are used (mirrors apps/api/internal/db/postgres.go).
DATABASE_URL = os.getenv("DATABASE_URL")
DB_HOST = os.getenv("DB_HOST", "localhost")
DB_PORT = os.getenv("DB_PORT", "5432")
DB_NAME = os.getenv("DB_NAME", "speedcubinghungary_dev")
DB_USER = os.getenv("DB_USER", "")
DB_PASSWORD = os.getenv("DB_PASSWORD", "")


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
            "Copy .env.example to .env in apps/scripts/weekly-report and fill "
            "in the values."
        )
