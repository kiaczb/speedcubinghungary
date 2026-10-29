import logging

import requests

from database.db import upsert_person, get_all_important_solvers_from_db

logger = logging.getLogger(__name__)

milestone_step = 5000
difference = 250

def CountSuccessfulAttempts(attempts):
    count = 0
    for attempt in attempts or []:
        value = attempt.get("result") if isinstance(attempt, dict) else attempt
        if isinstance(value, (int, float)) and value > 0:
            count += 1
    return count

def FetchTotalSolves(wca_id):
    if not wca_id:
        return None

    url = f"https://www.worldcubeassociation.org/api/v0/persons/{wca_id}"
    try:
        response = requests.get(url, timeout=5)
        if response.status_code == 200:
            return response.json().get("total_solves")
        logger.warning("HTTP %s while fetching total solves for %s", response.status_code, wca_id)
    except Exception:
        logger.exception("Failed to fetch total solves for %s", wca_id)
    return None

def AddSolvesToCompetitor(competitor, solves_this_competition, results_posted):
    if not competitor.WcaId:
        return

    total_solves = FetchTotalSolves(competitor.WcaId)
    if total_solves is None:
        return

    if not results_posted:
        total_solves += solves_this_competition

    upsert_person(competitor.WcaId, competitor.CompetitorName, solve_count=total_solves)

def GetImportantSolvers():
    persons = get_all_important_solvers_from_db(milestone_step, difference)

    results = []
    for person in persons:
        count = person.get("successful_solves_count", 0)
        remainder = count % milestone_step
        last_crossed_milestone = count - remainder

        if remainder <= difference and last_crossed_milestone > 0:
            milestone = last_crossed_milestone
            is_achieved = True
            remaining = 0
        else:
            milestone = last_crossed_milestone + milestone_step
            is_achieved = False
            remaining = milestone - count

        results.append({
            "wca_id": person["wca_id"],
            "name": person["name"],
            "successful_solves_count": count,
            "milestone": milestone,
            "is_achieved": is_achieved,
            "remaining": remaining,
        })

    return results
