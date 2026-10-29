import logging

from competitionCount import GetImportantCompetitors
from delegatesCount import GetImportantDelegates
from solvesCount import GetImportantSolvers

logger = logging.getLogger(__name__)


def BuildReportData(competitionsWithHungarians):
    return {
        "competitions": competitionsWithHungarians,
        "count_competitors": GetImportantCompetitors(),
        "delegate_milestones": GetImportantDelegates(),
        "solve_milestones": GetImportantSolvers(),
    }
