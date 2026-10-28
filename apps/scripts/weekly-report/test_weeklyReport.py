import json
from datetime import datetime
from types import SimpleNamespace

import pytest

import solvesCount
from emailSender import EmailCompetitions
from reportPost import RenderPostContent
from solvesCount import CountSuccessfulAttempts, GetImportantSolvers


def _competitor(name, records=None):
    return SimpleNamespace(CompetitorName=name, WcaId="2020TEST01", Records=records or {})


def _competition(name, is_hungarian, competitors):
    return SimpleNamespace(
        CompetitionName=name,
        CountryIso="HU" if is_hungarian else "AT",
        From=datetime(2026, 7, 18),
        To=datetime(2026, 7, 19),
        isHungarian=is_hungarian,
        CompetitorWithRecords=competitors,
    )


def _record():
    return SimpleNamespace(Type="single", Result=642, Badge="NR")


def test_email_keeps_only_record_holders_on_hungarian_competitions():
    hu = _competition("Miskolc Open 2026", True, [
        _competitor("Rekorder", {"333": [_record()]}),
        _competitor("Résztvevő"),
    ])

    [selected] = EmailCompetitions([hu])

    assert [c.CompetitorName for c in selected.CompetitorWithRecords] == ["Rekorder"]
    assert [c.CompetitorName for c in hu.CompetitorWithRecords] == ["Rekorder", "Résztvevő"]


def test_email_drops_hungarian_competition_without_records():
    hu = _competition("Miskolc Open 2026", True, [_competitor("Résztvevő")])

    assert EmailCompetitions([hu]) == []


def test_email_keeps_every_hungarian_on_foreign_competitions():
    foreign = _competition("Vienna Open 2026", False, [_competitor("Résztvevő")])

    assert EmailCompetitions([foreign]) == [foreign]


@pytest.mark.parametrize(
    "attempts,expected",
    [
        ([{"result": 1000}, {"result": -1}, {"result": 0}, {"result": -2}, {"result": 900}], 2),
        ([1000, -1, 900], 2),
        (None, 0),
    ],
)
def test_count_successful_attempts(attempts, expected):
    assert CountSuccessfulAttempts(attempts) == expected


@pytest.mark.parametrize(
    "count,milestone,achieved,remaining",
    [
        (5100, 5000, True, 0),
        (5250, 5000, True, 0),
        (4800, 5000, False, 200),
        (9900, 10000, False, 100),
        (100, 5000, False, 4900),
    ],
)
def test_solve_milestones(monkeypatch, count, milestone, achieved, remaining):
    monkeypatch.setattr(
        solvesCount,
        "get_all_important_solvers_from_db",
        lambda step, diff: [{"wca_id": "2020TEST01", "name": "Teszt Elek", "successful_solves_count": count}],
    )

    [solver] = GetImportantSolvers()

    assert solver["milestone"] == milestone
    assert solver["is_achieved"] is achieved
    assert solver["remaining"] == remaining


def test_post_payload_includes_solve_milestones():
    content = RenderPostContent({
        "competitions": [],
        "count_competitors": [],
        "delegate_milestones": [],
        "solve_milestones": [
            {
                "wca_id": "2020TEST01",
                "name": "Teszt Elek",
                "successful_solves_count": 4800,
                "milestone": 5000,
                "is_achieved": False,
                "remaining": 200,
            }
        ],
    })

    assert json.loads(content)["solveMilestones"] == [
        {"name": "Teszt Elek", "count": 4800, "milestone": 5000, "remaining": 200, "achieved": False}
    ]


def test_events_from_results_maps_people_to_registrant_ids():
    from competitors import EventsFromResults

    persons = [
        {"wcaId": "2020TEST01", "name": "Teszt Elek", "registrantId": 1},
        {"wcaId": None, "name": "Új Versenyző", "registrantId": 2},
    ]
    flat = [
        {"event_id": "333", "wca_id": "2020TEST01", "name": "Teszt Elek", "best": 900, "average": 1000, "attempts": [900, 1000]},
        {"event_id": "333", "wca_id": "2026UJVE01", "name": "Új Versenyző", "best": 1500, "average": -1, "attempts": [1500]},
        {"event_id": "222", "wca_id": "2019IDEG01", "name": "Idegen", "best": 300, "average": 400, "attempts": []},
    ]

    events = {e["id"]: e for e in EventsFromResults(flat, persons)}

    results = events["333"]["rounds"][0]["results"]
    assert [r["personId"] for r in results] == [1, 2]
    assert results[0]["attempts"] == [900, 1000]
    assert events["222"]["rounds"][0]["results"] == []
