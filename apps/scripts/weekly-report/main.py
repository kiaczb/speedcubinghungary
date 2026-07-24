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
    # Mindkét kapcsoló KÉZI/TESZT futtatáshoz való — a cron egyiket sem használja.
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
    # Fail fast if required env vars are missing (see config.py).
    config.validate()

    # Read-only preflight: confirm the shared backend DB is reachable and the
    # required tables exist. Raises (and aborts the job) if not.
    check_connection()

    start = time.time()

    # A riport a futás időpontjához képesti GÖRDÜLŐ 7 napot fedi le (nem naptári
    # hetet) — a WCA lekérdezés is ezt az ablakot használja, és a poszt címe is
    # ezt a két dátumot mutatja.
    period_end = datetime.now()
    period_start = period_end - timedelta(weeks=1)

    competitions_with_hungarians = GetCompetitionsWithHungarians(period_start)

    logger.info("Processed competitions in %.2f seconds", time.time() - start)

    # Egyszer állítjuk elő a riport adatait, és ugyanazt kapja az e-mail és a
    # weboldalra kerülő Post is (a mérföldkő-lekérdezés DB-t és WCA API-t is hív).
    report_data = BuildReportData(competitions_with_hungarians)

    # A Post létrehozása (idempotens: ugyanarra a hétre másodszor futtatva nem
    # jön létre újabb poszt) és a records mentése MINDIG fusson le, még akkor
    # is, ha az e-mail küldése alább meghiúsul — az e-mail egy külső SMTP
    # függőség, a weboldali poszt a feladat tényleges, tartós eredménye.
    PublishWeeklyReportPost(report_data, period_start, period_end, replace=force)
    SaveRecords()

    if skip_email:
        logger.info("E-mail küldés kihagyva (--skip-email).")
    else:
        # Ha ez dob (pl. hibás SMTP hitelesítés), a job hibával áll le — lásd
        # lent —, de a fenti poszt/records már elmentődött, nem vész el.
        WriteEmail(report_data)

    logger.info("Weekly report finished successfully.")


if __name__ == "__main__":
    args = parse_args()
    try:
        main(force=args.force, skip_email=args.skip_email)
    except Exception:
        logger.exception("Weekly report failed.")
        raise SystemExit(1)
