import json
import logging
from datetime import date, timedelta
from uuid import uuid4

from database.db import insert_weekly_report_post
from reportPayload import BuildReportPayload

logger = logging.getLogger(__name__)


def WeekStart(today=None):
    today = today or date.today()
    return today - timedelta(days=today.weekday())


def BuildTitle(period_start, period_end):
    return (
        f"Heti riport – {period_start.strftime('%Y. %m. %d.')} – "
        f"{period_end.strftime('%Y. %m. %d.')}"
    )


def RenderPostContent(report_data):
    return json.dumps(BuildReportPayload(report_data), ensure_ascii=False)


def PublishWeeklyReportPost(report_data, period_start, period_end, replace=False):
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
