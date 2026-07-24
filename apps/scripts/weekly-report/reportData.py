"""A heti riport adathalmaza.

Egyszer állítjuk elő, és UGYANAZT használja az e-mail és a weboldalra kerülő
Post is. Így nem duplikálódik sem a szűrés, sem a mérföldkő-lekérdezés — utóbbi
az adatbázist és a WCA API-t is terheli, ezért kifejezetten drága kétszer futtatni.
"""

import logging

from competitionCount import GetImportantCompetitors
from delegatesCount import GetImportantDelegates

logger = logging.getLogger(__name__)


def BuildReportData(competitionsWithHungarians):
    """A sablonoknak átadható adathalmaz (a kulcsok a sablonváltozók nevei)."""
    return {
        "competitions": FilterCompetitions(competitionsWithHungarians),
        "count_competitors": GetImportantCompetitors(),
        "delegate_milestones": GetImportantDelegates(),
    }


def FilterCompetitions(competitionsWithHungarians):
    """Csak a megjelenítendő versenyeket tartja meg.

    A nem magyar versenyek mind bekerülnek; a magyar versenyek viszont csak
    akkor, ha van rekordot elért versenyzőjük (és ilyenkor a versenyzőlistát is
    a rekordosokra szűkítjük). Korábban ez a logika a HTML renderelőben ült.
    """
    filtered = []

    for comp in competitionsWithHungarians:
        if not comp.isHungarian:
            filtered.append(comp)
            continue

        recorders = [
            person for person in comp.CompetitorWithRecords if person.Records
        ]
        comp.CompetitorWithRecords = recorders

        if recorders:
            filtered.append(comp)

    return filtered
