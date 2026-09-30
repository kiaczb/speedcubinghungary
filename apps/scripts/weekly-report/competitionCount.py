from database.db import upsert_person, get_person_data, get_all_important_competitors_from_db
from utils import ConvertDate
import requests
from datetime import datetime

difference = 3
milestones = [100, 200]

def GetImportantCompetitors():
    competitors = get_all_important_competitors_from_db(milestones, difference)

    for comp in competitors:
        comp_name, comp_date, achieved = EstimateMilestoneCompetition(comp["wca_id"], comp.get("competition_count", 0))

        comp["estimated_competition"] = comp_name
        comp["estimated_competition_date"] = comp_date
        comp["is_achieved"] = achieved

    return competitors

def AddCompetitionToCompetitor(competitor):
    if not competitor.WcaId:
        return
    current = get_person_data(competitor.WcaId)
    new_count = (current['competition_count'] + 1) if current else 1

    upsert_person(competitor.WcaId, competitor.CompetitorName, comp_count=new_count)

def EstimateMilestoneCompetition(wca_id, current_count):
    if not wca_id:
        return "Ismeretlen", "", False

    next_milestone = None
    for ms in milestones:
        if current_count <= ms:
            next_milestone = ms
            break

    if not next_milestone:
        return "Már elérte a maximumot", "", False

    needed = next_milestone - current_count

    url = f"https://www.worldcubeassociation.org/api/v0/users/{wca_id}?upcoming_competitions=true"
    try:
        response = requests.get(url, timeout=5)
        if response.status_code == 200:
            data = response.json()
            upcoming = data.get("upcoming_competitions", [])
            upcoming_sorted = sorted(upcoming, key=lambda c: c.get("start_date", ""))

            if needed == 0:
                if upcoming_sorted:
                    target_comp = upcoming_sorted[0]
                    comp_name = target_comp.get("name", "Ismeretlen verseny")
                    start_str = target_comp.get("start_date")
                    end_str = target_comp.get("end_date")

                    formatted_date = ""
                    if start_str and end_str:
                        start_dt = datetime.strptime(start_str, "%Y-%m-%d")
                        end_dt = datetime.strptime(end_str, "%Y-%m-%d")
                        formatted_date = ConvertDate(start_dt, end_dt)

                    return comp_name, formatted_date, True
                else:
                    return "Elérte a mérföldkövet! 🎉", "", True

            if len(upcoming_sorted) >= needed:
                target_comp = upcoming_sorted[needed - 1]
                comp_name = target_comp.get("name", "Ismeretlen verseny")

                start_str = target_comp.get("start_date")
                end_str = target_comp.get("end_date")

                formatted_date = ""
                if start_str and end_str:
                    start_dt = datetime.strptime(start_str, "%Y-%m-%d")
                    end_dt = datetime.strptime(end_str, "%Y-%m-%d")
                    formatted_date = ConvertDate(start_dt, end_dt)

                return comp_name, formatted_date, False
            else:
                return f"Nincs elég regisztrált verseny (még {needed} kellene)", "", False
        else:
            return "WCA API nem elérhető", "", False
    except Exception as e:
        print(f"Hiba a mérföldkő becslésénél ({wca_id}): {e}")
        return "Hiba a lekérdezés során", "", False
