"""A heti riport strukturált adata a weboldali Post számára.

A poszt `content`-je a riportnál NEM markdown, hanem JSON: a frontend a
`tag === 'report'` posztokat egy dedikált komponenssel (WeeklyReportView)
rendereli, ami ebből a struktúrából épít stílusozott megjelenítést. Így az
adat (Python) és a megjelenítés (React) tisztán szétválik.

A séma stabil kell legyen — ha bővíted, tartsd visszafelé kompatibilisen, mert
régebbi riport-posztok is ezt a struktúrát tárolják. A frontend oldali típus:
apps/web/app/posts/types.ts (WeeklyReport*).
"""

from utils import ConvertResult


def BuildReportPayload(report_data):
    """A report_data (Competition/CompetitorWithRecords objektumok + mérföldkő
    dict-ek) átalakítása JSON-serializálható, a frontend által várt struktúrává.

    Az eredményeket itt formázzuk stringgé (ConvertResult), mert a formázási
    logika (single/average, fmc, mbld) a Pythonban él — a frontend ne duplikálja.
    """
    return {
        "competitions": [
            {
                "name": comp.CompetitionName,
                "countryIso": comp.CountryIso,
                "from": comp.From.strftime("%Y-%m-%d"),
                "to": comp.To.strftime("%Y-%m-%d"),
                "competitors": [
                    {
                        "name": competitor.CompetitorName,
                        # Üres, ha a versenyzőnek nem volt rekordja — a nevét
                        # ilyenkor is megjelenítjük (a rekord csak a rekordosoknál).
                        "results": [
                            {
                                "event": event,
                                "type": record.Type,
                                "result": ConvertResult(record.Result, event),
                                "badge": record.Badge,
                            }
                            for event, records in competitor.Records.items()
                            for record in records
                        ],
                    }
                    for competitor in comp.CompetitorWithRecords
                ],
            }
            for comp in report_data["competitions"]
        ],
        "competitorMilestones": [
            {
                "name": milestone["name"],
                "count": milestone["competition_count"],
                "estimated": milestone.get("estimated_competition"),
                "estimatedDate": milestone.get("estimated_competition_date"),
                "achieved": milestone.get("is_achieved", False),
            }
            for milestone in report_data["count_competitors"]
        ],
        "delegateMilestones": [
            {
                "name": delegate["name"],
                "count": delegate["delegated_competitions_count"],
            }
            for delegate in report_data["delegate_milestones"]
        ],
    }
