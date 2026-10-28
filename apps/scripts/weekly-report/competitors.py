import logging
from CompetitionModels.Record import Record
from CompetitionModels.CompetitorWithRecords import CompetitorWithRecords
from config import excludedCompetitorWcaIds, badges
from recordsManager import localNationalRecords
import requests
from concurrent.futures import ThreadPoolExecutor, as_completed
from CompetitionModels.Competition import Competition
from competitionCount import AddCompetitionToCompetitor
from delegatesCount import IsHungarianDelegate, AddCompetitionToDelegate
from solvesCount import AddSolvesToCompetitor, CountSuccessfulAttempts
import time
from utils import ReverseHungarianName

logger = logging.getLogger(__name__)

def UpdateRecords(record_type, value, competitor, event_id, localNationalRecords, badges):
    if value > 0:
        for index, nationalRecord in enumerate(localNationalRecords.values()):
            if competitor.WcaId in excludedCompetitorWcaIds and index == 2:
                continue
            if value <= nationalRecord[event_id][record_type]:
                nationalRecord[event_id][record_type] = value
                if event_id in competitor.Records:
                    competitor.Records[event_id].append(Record(record_type, value, badges[index]))
                else:
                    competitor.Records[event_id] = [Record(record_type, value, badges[index])]
                break

def ProcessPerson(person, events, isHungarianComp, resultsPosted):
    if person["countryIso2"] != "HU" and person["wcaId"] not in excludedCompetitorWcaIds:
        return None

    if not person.get("registration"):
        return None

    registered_event_ids = set(person["registration"]["eventIds"])
    display_name = person["name"]

    competitor = CompetitorWithRecords(display_name, person["wcaId"])

    has_any_result = False
    successful_solves = 0

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
            successful_solves += CountSuccessfulAttempts(result.get("attempts"))

    if not has_any_result:
        return None

    AddCompetitionToCompetitor(competitor)
    AddSolvesToCompetitor(competitor, successful_solves, bool(resultsPosted))

    if isHungarianComp and IsHungarianDelegate(competitor):
        AddCompetitionToDelegate(competitor)

    if person["countryIso2"] == "HU":
        display_name = ReverseHungarianName(display_name)
    competitor.Name = display_name

    return competitor


def IsHungarianCompetition(comp):
    return comp["country_iso2"] == "HU"

def EventsFromResults(flat_results, persons):
    wca_id_to_reg_id = {p["wcaId"]: p.get("registrantId") for p in persons if p.get("wcaId")}
    name_to_reg_id = {p["name"]: p.get("registrantId") for p in persons}

    events_dict = {}
    for r in flat_results:
        event_id = r.get("event_id")
        if event_id not in events_dict:
            events_dict[event_id] = {
                "id": event_id,
                "rounds": [{"id": "patched_round", "results": []}],
            }

        reg_id = wca_id_to_reg_id.get(r.get("wca_id"))
        if not reg_id:
            reg_id = name_to_reg_id.get(r.get("name"))

        if reg_id is not None:
            events_dict[event_id]["rounds"][0]["results"].append({
                "personId": reg_id,
                "best": r.get("best", -1),
                "average": r.get("average", -1),
                "attempts": r.get("attempts", []),
            })

    return list(events_dict.values())

def GetCompetitorsForCompetition(comp):
    comp_id = comp["id"]
    wcif_url = f"https://www.worldcubeassociation.org/api/v0/competitions/{comp_id}/wcif/public"
    results_url = f"https://www.worldcubeassociation.org/api/v0/competitions/{comp_id}/results"

    RETRY_STATUSES = {429, 500, 502, 503, 504}
    response = None
    for attempt in range(5):
        try:
            response = requests.get(wcif_url, timeout=10)
            if response.status_code == 200:
                break
            if response.status_code in RETRY_STATUSES:
                logger.warning("HTTP %s for %s WCIF, retrying", response.status_code, comp_id)
                time.sleep(4 ** attempt)
                continue

            logger.warning("HTTP %s for %s WCIF, skipping competition", response.status_code, comp_id)
            return []
        except requests.exceptions.Timeout:
            logger.warning("Timeout for %s WCIF, retrying", comp_id)
            time.sleep(4 ** attempt)

    if response is None or response.status_code != 200:
        logger.error("Could not download WCIF for %s, skipping competition", comp_id)
        return []

    competitionWCIF = response.json()
    persons = competitionWCIF.get("persons", [])
    events = competitionWCIF.get("events", [])

    try:
        res_response = requests.get(results_url, timeout=10)
        if res_response.status_code == 200:
            flat_results = res_response.json()
            if flat_results:
                events = EventsFromResults(flat_results, persons)
                logger.info("%s: using posted results (%d rows)", comp_id, len(flat_results))
    except Exception:
        logger.exception("Failed to process /results for %s, falling back to WCIF events", comp_id)

    with ThreadPoolExecutor() as executor:
        results = list(executor.map(
            lambda p: ProcessPerson(p, events, IsHungarianCompetition(comp), comp.get("results_posted_at")),
            persons,
        ))

    return [c for c in results if c]

def GetCompetitionsParallel(competitions):
    results = []

    def process(comp):
        competitors = GetCompetitorsForCompetition(comp)
        if competitors:
            logger.info("%s: %d Hungarian-interest competitors", comp["id"], len(competitors))
            return Competition(comp["name"], comp["country_iso2"], comp["start_date"], comp["end_date"], IsHungarianCompetition(comp), competitors)

    with ThreadPoolExecutor(max_workers=3) as executor:
        futures = [executor.submit(process, comp) for comp in competitions]
        for future in as_completed(futures):
            try:
                result = future.result()
                if result:
                    results.append(result)
            except Exception:
                logger.exception("Failed to process competition")

    return results
