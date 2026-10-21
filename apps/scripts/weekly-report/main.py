import argparse
import logging
import time
from datetime import datetime, timedelta

import config
from database.db import check_connection
from competition import GetCompetitionsWithHungarians
from emailSender import WriteEmail
from reportData import BuildReportData
from reportPost import PublishWeeklyReportPost
from recordsManager import SaveRecords

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s %(levelname)s %(name)s: %(message)s",
)
logger = logging.getLogger("weekly-report")


def parse_args():
    parser = argparse.ArgumentParser(
        description="Heti riport: WCA-lekérés, weboldali poszt és e-mail küldés."
    )
    parser.add_argument(
        "--force",
        action="store_true",
        help="A hét meglévő riport-posztját felülírja az új tartalommal "
        "(alapból a hét első riportja marad meg).",
    )
    parser.add_argument(
        "--skip-email",
        action="store_true",
        help="Kihagyja az e-mail küldést (teszteléshez, SMTP nélkül).",
    )
    return parser.parse_args()


def main(force=False, skip_email=False):
    config.validate()

    check_connection()

    start = time.time()

    period_end = datetime.now()
    period_start = period_end - timedelta(weeks=1)

    competitions_with_hungarians = GetCompetitionsWithHungarians(period_start)

    logger.info("Processed competitions in %.2f seconds", time.time() - start)

    report_data = BuildReportData(competitions_with_hungarians)

    PublishWeeklyReportPost(report_data, period_start, period_end, replace=force)
    SaveRecords()

    if skip_email:
        logger.info("E-mail küldés kihagyva (--skip-email).")
    else:
        WriteEmail(report_data)

    logger.info("Weekly report finished successfully.")


if __name__ == "__main__":
    args = parse_args()
    try:
        main(force=args.force, skip_email=args.skip_email)
    except Exception:
        logger.exception("Weekly report failed.")
        raise SystemExit(1)
