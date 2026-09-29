from CompetitionModels.Record import Record
from CompetitionModels.CompetitorWithRecords import CompetitorWithRecords
from config import excludedCompetitorWcaIds, badges
from recordsManager import localNationalRecords
import requests
from concurrent.futures import ThreadPoolExecutor, as_completed
from CompetitionModels.Competition import Competition
from competitionCount import AddCompetitionToCompetitor
from delegatesCount import IsHungarianDelegate, AddCompetitionToDelegate
import time
from utils import ReverseHungarianName
def UpdateRecords(record_type, value, competitor, event_id, localNationalRecords, badges):
    #Frissíti a versenyző rekordjait és kezeli a nemzeti és speciális rekordokat.
    if value > 0:
        for index, nationalRecord in enumerate(localNationalRecords.values()):
            if competitor.WcaId in excludedCompetitorWcaIds and index == 2: #If the person is not representing Hungary but a member of the Hungarian community
                continue
            if value <= nationalRecord[event_id][record_type]:
                nationalRecord[event_id][record_type] = value
                if event_id in competitor.Records:
                    competitor.Records[event_id].append(Record(record_type, value, badges[index]))
                else:
                    competitor.Records[event_id] = [Record(record_type, value, badges[index])]
                break

def ProcessPerson(person, events ,isHungarianComp):
    if person["countryIso2"] != "HU" and person["wcaId"] not in excludedCompetitorWcaIds:
        return None

    if not person.get("registration"):
        return None

    registered_event_ids = set(person["registration"]["eventIds"])
    display_name = person["name"]

    competitor = CompetitorWithRecords(display_name, person["wcaId"])

    has_any_result = False

    def IsPersonResult(result):
        return result.get("personId") == person["registrantId"]

    for event in events:
        if event["id"] not in registered_event_ids:
            continue

        round_result_pairs = [
            (_round, result)
            for _round in event.get("rounds", [])
            for result in _round.get("results", [])
            if IsPersonResult(result)
        ]

        if not round_result_pairs:
            continue

        has_any_result = True

        for _round, result in round_result_pairs:
            UpdateRecords("average", result["average"], competitor, event["id"], localNationalRecords, badges)
            UpdateRecords("single", result["best"], competitor, event["id"], localNationalRecords, badges)

    # Csak az esik ki, akinek egyetlen versenyszámban sincs eredménye.
    if not has_any_result:
        return None

    AddCompetitionToCompetitor(competitor)

    if isHungarianComp and IsHungarianDelegate(competitor):
        AddCompetitionToDelegate(competitor)
    
    if person["countryIso2"] == "HU":
        display_name = ReverseHungarianName(display_name)
    competitor.Name = display_name 

    return competitor


def IsHungarianCompetition(comp):
    return comp["country_iso2"] == "HU"

def GetCompetitorsForCompetition(comp):
    url = f"https://www.worldcubeassociation.org/api/v0/competitions/{comp['id']}/wcif/public"

    RETRY_STATUSES = {429, 500, 502, 503, 504}
    for attempt in range(5):
        try:
            response = requests.get(url)
            if response.status_code == 200:
                break
            if response.status_code in RETRY_STATUSES:
                wait = 4 ** attempt
                time.sleep(wait)
                continue

            print(f"HTTP {response.status_code} for {comp['id']}")
            return []
        
        except requests.exceptions.Timeout:
            wait = 4 ** attempt
            time.sleep(wait)

    competitionWCIF = response.json()
    persons = competitionWCIF["persons"]
    events = competitionWCIF["events"]

    with ThreadPoolExecutor() as executor:
        results = list(executor.map(lambda p: ProcessPerson(p, events, IsHungarianCompetition(comp)), persons))

    return [c for c in results if c]

def GetCompetitionsParallel(competitions):
    results = []

    def process(comp):
        competitors = GetCompetitorsForCompetition(comp)
        if competitors:
            return Competition(comp["name"], comp["country_iso2"], comp["start_date"], comp["end_date"], competitors)

    with ThreadPoolExecutor(max_workers=3) as executor:
        futures = [executor.submit(process, comp) for comp in competitions]
        for future in as_completed(futures):
            try:
                result = future.result()
                if result:
                    results.append(result)
            except Exception as e:
                print(f"Hiba a verseny feldolgozásakor: {e}")

    return results