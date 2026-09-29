import requests
from datetime import datetime
from competitors import GetCompetitionsParallel

def GetCompetitionsWithHungarians(date):
    competition_results = GetCompetitionsParallel(GetCompetitions(date))

    if competition_results:
        return sorted(competition_results, key=lambda oput: oput.From)
    else:
        print("Nincs elérhető verseny.")
        return []

def GetCompetitions(date):
    """A `date` és a mai nap között kezdődő versenyek (WCA API objektumok)."""
    competitions = []
    pageCount = 1
    end = False
    while not end:
        competitionsUrl = f"https://www.worldcubeassociation.org/api/v0/competitions?sort=start_date&start={date.strftime('%Y-%m-%d')}&end={datetime.now().strftime('%Y-%m-%d')}&page={pageCount}"
        response = requests.get(competitionsUrl)
        if response.status_code != 200:
            print("Error")
            end = True
        competitionData = response.json()
        # A lapozás végén a WCA API üres listát ad.
        if not competitionData:
            end = True
        competitions += competitionData
        pageCount += 1
    return competitions
