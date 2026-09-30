from utils import ConvertResult


def BuildReportPayload(report_data):
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
