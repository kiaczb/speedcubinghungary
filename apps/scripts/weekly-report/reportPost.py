"""A heti riport publikálása a MEGLÉVŐ Post rendszerbe.

A riport nem külön tartalomkezelőbe kerül, hanem a posts táblába — így örökli a
listázást, a részletező oldalt (/posts/<id>) és a megjelenítést, külön frontend
implementáció nélkül. A poszt tagje 'report' (lásd db/migrations/000003).

Idempotencia: a posts.report_week oszlop UNIQUE, és az INSERT ON CONFLICT DO
NOTHING — ugyanarra a hétre másodszor futtatva NEM jön létre újabb poszt. Ez
azért is fontos, mert a cron újrapróbálkozhat vagy kézzel is indítható.
"""

import json
import logging
from datetime import date, timedelta
from uuid import uuid4

from database.db import insert_weekly_report_post
from reportPayload import BuildReportPayload

logger = logging.getLogger(__name__)


def WeekStart(today=None):
    """A riportolt hét hétfője (ISO hét kezdete) — ez az idempotencia kulcsa.

    Bármikor is fut a job a héten belül, ugyanazt a hétfőt adja, így a hét
    második futása már ütközik a UNIQUE megszorításba.
    """
    today = today or date.today()
    return today - timedelta(days=today.weekday())


def BuildTitle(period_start, period_end):
    """A poszt címe a TÉNYLEGESEN lekérdezett intervallummal.

    Fontos: ez NEM a naptári hét. A job gördülő 7 napos ablakot kérdez le a WCA
    API-tól (futás időpontja - 1 hét ... futás időpontja, lásd main.py és
    competition.py), ezért a cím is ezt a két dátumot mutatja — különben a cím
    olyan napokat állítana, amikre nézve nincs is adat.

    Az idempotencia-kulcs (WeekStart) ettől függetlenül a naptári hét hétfője:
    az a "hetente egy riport" garancia, és nem jelenik meg a címben.
    """
    return (
        f"Heti riport – {period_start.strftime('%Y. %m. %d.')} – "
        f"{period_end.strftime('%Y. %m. %d.')}"
    )


def RenderPostContent(report_data):
    """A riport tartalma JSON stringként — a frontend WeeklyReportView komponense
    rendereli (lásd reportPayload.py). ensure_ascii=False: a magyar ékezetek
    olvashatóan (UTF-8) kerüljenek a DB-be, ne \\u-escape-elve."""
    return json.dumps(BuildReportPayload(report_data), ensure_ascii=False)


def PublishWeeklyReportPost(report_data, period_start, period_end, replace=False):
    """Létrehozza a heti riport posztot, ha arra a hétre még nincs.

    period_start / period_end: a ténylegesen lekérdezett intervallum (a címhez).
    replace=True: a hét meglévő riportját FELÜLÍRJA az új tartalommal. Ezt csak
    kézi/teszt futtatáshoz használjuk (main.py --force) — a cron sosem.

    Visszatérés: True, ha új poszt jött létre vagy felülírtuk; False, ha már
    létezett és érintetlenül hagytuk.
    """
    week_start = WeekStart(period_end.date())
    content = RenderPostContent(report_data)

    written = insert_weekly_report_post(
        post_id=str(uuid4()),
        title=BuildTitle(period_start, period_end),
        content=content,
        report_week=week_start,
        replace=replace,
    )

    if written and replace:
        logger.info("Weekly report post written/replaced (week starting %s).", week_start)
    elif written:
        logger.info("Weekly report post created (week starting %s).", week_start)
    else:
        logger.info(
            "Weekly report post already exists for week starting %s — skipped "
            "(use --force to overwrite).",
            week_start,
        )

    return written
