import os
from datetime import datetime

from dotenv import load_dotenv

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
REPO_ROOT = os.path.abspath(os.path.join(SCRIPT_DIR, "..", "..", ".."))

ENV = os.getenv("ENV", "dev")
load_dotenv(os.path.join(REPO_ROOT, f".env.{ENV}"))

IN_CONTAINER = os.path.exists("/.dockerenv")

badges = ["WR", "CR", "NR"]
excludedCompetitorWcaIds = ["2017TSVE02"]

emailFrom = os.getenv("EMAIL_FROM", "")
emailTo = os.getenv("EMAIL_TO", "")
emailSubject = f"{datetime.now().strftime('%Y-%m-%d')} Weekly Hungarian competitors dump "
emailCode = os.getenv("EMAIL_APP_PASSWORD", "")

DB_PORT = os.getenv("DB_PORT", "5432")
DB_NAME = os.getenv("DB_NAME", "speedcubinghungary_dev")
DB_USER = os.getenv("DB_USER", "")
DB_PASSWORD = os.getenv("DB_PASSWORD", "")

if IN_CONTAINER:
    DATABASE_URL = os.getenv("DATABASE_URL")
    DB_HOST = os.getenv("DB_HOST", "db")
else:
    DATABASE_URL = None
    DB_HOST = "localhost"


def validate():
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
