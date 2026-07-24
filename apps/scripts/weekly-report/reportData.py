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
    """A riport adathalmaza az e-mail és a weboldali Post számára.

    MINDEN magyar érdekeltségű versenyző bekerül — nem csak a rekordot elértek.
    A `competitors.py` (`GetCompetitionsParallel`) már csak azokat a versenyeket
    adja vissza, amiknek van eredménnyel rendelkező magyar versenyzőjük, és minden
    ilyen versenyzőt tartalmaz (a `Records` üres, ha az illetőnek nincs rekordja).
    Ezért itt már nincs külön szűrés.
    """
    return {
        "competitions": competitionsWithHungarians,
        "count_competitors": GetImportantCompetitors(),
        "delegate_milestones": GetImportantDelegates(),
    }
