import logging
import time
from datetime import datetime, timedelta

import config
from database.db import check_connection
from competition import GetCompetitionsWithHungarians
from emailSender import WriteEmail
from recordsManager import SaveRecords

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s %(levelname)s %(name)s: %(message)s",
)
logger = logging.getLogger("weekly-report")


def main():
    # Fail fast if required env vars are missing (see config.py).
    config.validate()

    # Read-only preflight: confirm the shared backend DB is reachable and the
    # required tables exist. Raises (and aborts the job) if not.
    check_connection()

    start = time.time()

    date = datetime.now() - timedelta(weeks=1)
    competitions_with_hungarians = GetCompetitionsWithHungarians(date)

    logger.info("Processed competitions in %.2f seconds", time.time() - start)

    WriteEmail(competitions_with_hungarians)
    SaveRecords()

    logger.info("Weekly report finished successfully.")


if __name__ == "__main__":
    try:
        main()
    except Exception:
        logger.exception("Weekly report failed.")
        raise SystemExit(1)
