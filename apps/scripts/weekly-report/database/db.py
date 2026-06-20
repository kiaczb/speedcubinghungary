import logging

import psycopg
from psycopg.rows import dict_row

from env import (
    DATABASE_URL,
    DB_HOST,
    DB_NAME,
    DB_PASSWORD,
    DB_PORT,
    DB_USER,
)

logger = logging.getLogger(__name__)

# Connection settings for the SHARED backend database. These mirror the env
# vars used by the Go backend (apps/api/internal/db/postgres.go).
DB_CONFIG = {
    "dbname": DB_NAME,
    "user": DB_USER,
    "password": DB_PASSWORD,
    "host": DB_HOST,
    "port": DB_PORT,
}

# Tables owned by the Go backend that this job reads/writes. The script must
# NEVER create or alter them; they are provisioned by the backend migrations.
#  - `users`   : competitor roster (wca_id, name, competition_count,
#                delegated_count). The backend's equivalent of the old
#                stand-alone `persons` table.
#  - `records` : world/continental/national records (region, event_id,
#                single, average).
REQUIRED_TABLES = ("users", "records")


def _connect(row_factory=None):
    """Opens a connection to the shared backend database.

    Uses DATABASE_URL when provided, otherwise the individual DB_* settings.
    """
    kwargs = {}
    if row_factory is not None:
        kwargs["row_factory"] = row_factory

    if DATABASE_URL:
        return psycopg.connect(DATABASE_URL, **kwargs)
    return psycopg.connect(**DB_CONFIG, **kwargs)


def get_connection():
    """Returns a connection to the shared backend database."""
    return _connect()


def check_connection():
    """Verifies the shared database is reachable and the required tables exist.

    This is a read-only preflight: it never creates databases, tables or
    migrations (the Go backend owns the schema). It raises on failure so the
    weekly job fails loudly instead of silently running against a broken DB.
    """
    try:
        with _connect() as conn:
            with conn.cursor() as cur:
                for table in REQUIRED_TABLES:
                    cur.execute("SELECT to_regclass(%s)", (f"public.{table}",))
                    (exists,) = cur.fetchone()
                    if exists is None:
                        raise RuntimeError(
                            f"Required table '{table}' is missing from the shared "
                            f"database '{DB_NAME}'. Run the Go backend migrations first."
                        )
        logger.info(
            "Connected to shared database '%s'; required tables present.", DB_NAME
        )
    except Exception:
        logger.exception("Database preflight check failed.")
        raise


def get_person_data(wca_id):
    """Returns a competitor row from the shared `users` table as a dict, or None."""
    try:
        with _connect(row_factory=dict_row) as conn:
            with conn.cursor() as cur:
                cur.execute(
                    """
                    SELECT wca_id, name, competition_count, delegated_count
                    FROM users
                    WHERE wca_id = %s
                    """,
                    (wca_id,),
                )
                return cur.fetchone()
    except Exception:
        logger.exception("Failed to read competitor %s", wca_id)
        return None


def upsert_person(wca_id, name, comp_count=None, del_count=None):
    """Inserts or updates a competitor in the shared `users` table (upsert).

    Only the competitor-roster columns are written. Auth columns (email, roles)
    are deliberately left untouched so this never clobbers backend login data:
    on insert they take their schema defaults (email NULL, roles '[]').
    """
    try:
        with get_connection() as conn:
            with conn.cursor() as cur:
                cur.execute(
                    """
                    INSERT INTO users (wca_id, name, competition_count, delegated_count)
                    VALUES (%s, %s, COALESCE(%s, 0), COALESCE(%s, 0))
                    ON CONFLICT (wca_id) DO UPDATE SET
                        name = EXCLUDED.name,
                        -- Keep the existing value when the new one is NULL (None).
                        competition_count = COALESCE(%s, users.competition_count),
                        delegated_count = COALESCE(%s, users.delegated_count),
                        updated_at = NOW();
                    """,
                    (wca_id, name, comp_count, del_count, comp_count, del_count),
                )
            conn.commit()
    except Exception:
        logger.exception("Upsert failed for competitor %s", wca_id)


def upsert_record(region, event_id, single, average):
    """Inserts or updates a record in the shared `records` table (upsert)."""
    try:
        with get_connection() as conn:
            with conn.cursor() as cur:
                cur.execute(
                    """
                    INSERT INTO records (region, event_id, single, average)
                    VALUES (%s, %s, %s, %s)
                    ON CONFLICT (region, event_id) DO UPDATE SET
                        single = EXCLUDED.single,
                        average = EXCLUDED.average;
                    """,
                    (region, event_id, single, average),
                )
            conn.commit()
    except Exception:
        logger.exception("Record upsert failed (%s - %s)", region, event_id)


def get_all_records():
    """Returns all records as the JSON-shaped dict the report templates expect."""
    records_dict = {
        "world_records": {},
        "european_records": {},
        "hungarian_records": {},
    }
    mapping = {
        "world": "world_records",
        "europe": "european_records",
        "hungary": "hungarian_records",
    }

    try:
        with _connect(row_factory=dict_row) as conn:
            with conn.cursor() as cur:
                cur.execute("SELECT region, event_id, single, average FROM records")
                rows = cur.fetchall()
                for row in rows:
                    mapped_region = mapping.get(row["region"])
                    if mapped_region:
                        records_dict[mapped_region][row["event_id"]] = {
                            "single": row["single"],
                            "average": row["average"],
                        }
    except Exception:
        logger.exception("Failed to read records")

    return records_dict


def get_all_important_competitors_from_db(milestones, difference):
    """Returns competitors whose competition_count is near a milestone."""
    conditions = []
    params = []
    for ms in milestones:
        conditions.append("competition_count BETWEEN %s AND %s")
        params.extend([ms - difference, ms])

    where_clause = " OR ".join(conditions)

    try:
        with _connect(row_factory=dict_row) as conn:
            with conn.cursor() as cur:
                cur.execute(
                    f"""
                    SELECT wca_id, name, competition_count, delegated_count
                    FROM users
                    WHERE {where_clause}
                    """,
                    params,
                )
                return cur.fetchall()
    except Exception:
        logger.exception("Failed to read milestone competitors")
        return []


def get_all_important_delegates_from_db(milestones, difference):
    """Returns delegates whose delegated_count is near a milestone."""
    conditions = []
    params = []
    for ms in milestones:
        conditions.append("delegated_count BETWEEN %s AND %s")
        params.extend([ms - difference, ms])

    where_clause = " OR ".join(conditions)

    try:
        with _connect(row_factory=dict_row) as conn:
            with conn.cursor() as cur:
                cur.execute(
                    f"""
                    SELECT wca_id, name, delegated_count
                    FROM users
                    WHERE {where_clause}
                    """,
                    params,
                )
                rows = cur.fetchall()
                # Reshape to the structure the email sender expects.
                return [
                    {
                        "wca_id": r["wca_id"],
                        "name": r["name"],
                        "delegated_competitions_count": r["delegated_count"],
                    }
                    for r in rows
                ]
    except Exception:
        logger.exception("Failed to read milestone delegates")
        return []
