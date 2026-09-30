import logging

from competitionCount import GetImportantCompetitors
from delegatesCount import GetImportantDelegates

logger = logging.getLogger(__name__)


def BuildReportData(competitionsWithHungarians):
    return {
        "competitions": competitionsWithHungarians,
        "count_competitors": GetImportantCompetitors(),
        "delegate_milestones": GetImportantDelegates(),
    }
