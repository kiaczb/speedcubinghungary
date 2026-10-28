import logging
import requests
from datetime import datetime
from competitors import GetCompetitionsParallel

logger = logging.getLogger(__name__)

def GetCompetitionsWithHungarians(date):
    competitions = GetCompetitions(date)
    logger.info("Found %d competitions since %s", len(competitions), date.strftime('%Y-%m-%d'))
    competition_results = GetCompetitionsParallel(competitions)

    if competition_results:
        return sorted(competition_results, key=lambda oput: oput.From)
    else:
        logger.info("Nincs elérhető verseny.")
        return []

def GetCompetitions(date):
    competitions = []
    pageCount = 1
    while True:
        competitionsUrl = f"https://www.worldcubeassociation.org/api/v0/competitions?sort=start_date&start={date.strftime('%Y-%m-%d')}&end={datetime.now().strftime('%Y-%m-%d')}&page={pageCount}"
        response = requests.get(competitionsUrl, timeout=10)
        if response.status_code != 200:
            logger.error("HTTP %s while listing competitions (page %d)", response.status_code, pageCount)
            break
        competitionData = response.json()
        if not competitionData:
            break
        competitions += competitionData
        pageCount += 1
    return competitions
