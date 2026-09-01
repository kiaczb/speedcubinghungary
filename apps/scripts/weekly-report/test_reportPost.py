"""A heti riport → Post publikálás tesztjei.

Futtatás:
    cd apps/scripts/weekly-report
    pip install -r requirements-dev.txt
    python -m pytest

Az itteni logika DB nélkül tesztelhető: az idempotencia-kulcs számítása és a
sablonrenderelés tiszta függvények (a tényleges INSERT-et a Postgres UNIQUE
megszorítása védi, lásd db/migrations/000003).
"""

import json
from datetime import date, datetime, timedelta
from types import SimpleNamespace

import pytest

from reportData import BuildReportData
from reportPost import BuildTitle, RenderPostContent, WeekStart


# --- Idempotencia-kulcs (a hét hétfője) ---------------------------------------

@pytest.mark.parametrize(
    "today,expected_monday",
    [
        (date(2026, 7, 20), date(2026, 7, 20)),  # hétfő
        (date(2026, 7, 24), date(2026, 7, 20)),  # péntek
        (date(2026, 7, 26), date(2026, 7, 20)),  # vasárnap
        (date(2026, 7, 27), date(2026, 7, 27)),  # a következő hét hétfője
    ],
)
def test_week_start_returns_monday(today, expected_monday):
    assert WeekStart(today) == expected_monday


def test_same_week_yields_same_key():
    """Ez a tényleges idempotencia-garancia: a héten belül bármikor futva
    ugyanaz a kulcs keletkezik, így a második INSERT ütközik és nem hoz létre
    újabb posztot."""
    days = [date(2026, 7, 20), date(2026, 7, 22), date(2026, 7, 26)]

    assert {WeekStart(d) for d in days} == {date(2026, 7, 20)}


def test_consecutive_weeks_yield_different_keys():
    assert WeekStart(date(2026, 7, 26)) != WeekStart(date(2026, 7, 27))


def test_build_title_shows_the_actual_queried_period():
    """A cím a TÉNYLEGES lekérdezett ablakot mutassa (gördülő 7 nap), ne a
    naptári hetet — különben olyan napokat állítana, amikre nincs is adat."""
    period_end = datetime(2026, 7, 24)  # péntek
    period_start = period_end - timedelta(weeks=1)  # előző péntek

    title = BuildTitle(period_start, period_end)

    assert "2026. 07. 17." in title
    assert "2026. 07. 24." in title


def test_build_title_does_not_claim_the_calendar_week():
    """Regresszió: a cím korábban a naptári hetet (hétfő-vasárnap) írta ki, ami
    egy pénteki futásnál a jövőbe is belógott."""
    period_end = datetime(2026, 7, 24)  # péntek; a naptári hét vasárnapja 07. 26.
    period_start = period_end - timedelta(weeks=1)

    title = BuildTitle(period_start, period_end)

    assert "2026. 07. 26." not in title  # a jövőbeli vasárnap nem jelenhet meg
    assert "2026. 07. 20." not in title  # a naptári hét hétfője sem


# --- JSON payload renderelés --------------------------------------------------

def _competition(name="Miskolc Open 2026", iso="HU", is_hungarian=True, competitors=None):
    return SimpleNamespace(
        CompetitionName=name,
        CountryIso=iso,
        From=datetime(2026, 7, 18),
        To=datetime(2026, 7, 19),
        isHungarian=is_hungarian,
        CompetitorWithRecords=competitors if competitors is not None else [],
    )


def _competitor(name="Teszt Elek", records=None):
    return SimpleNamespace(
        CompetitorName=name,
        WcaId="2020TEST01",
        Records=records if records is not None else {},
    )


def test_render_content_is_valid_json():
    """A riport content JSON (a frontend WeeklyReportView komponense rendereli)."""
    content = RenderPostContent(
        {"competitions": [], "count_competitors": [], "delegate_milestones": []}
    )

    payload = json.loads(content)  # nem dobhat
    assert payload == {
        "competitions": [],
        "competitorMilestones": [],
        "delegateMilestones": [],
    }


def test_render_content_includes_results_and_milestones():
    content = RenderPostContent({
        "competitions": [
            _competition(competitors=[
                _competitor(records={"333": [SimpleNamespace(Type="single", Result=642, Badge="NR")]})
            ])
        ],
        "count_competitors": [
            {
                "name": "Nagy Anna",
                "competition_count": 99,
                "estimated_competition": "Budapest Open 2026",
                "estimated_competition_date": "2026-08-01",
                "is_achieved": False,
            }
        ],
        "delegate_milestones": [
            {"name": "Kiss Péter", "delegated_competitions_count": 49}
        ],
    })

    payload = json.loads(content)

    comp = payload["competitions"][0]
    assert comp["name"] == "Miskolc Open 2026"
    assert comp["from"] == "2026-07-18"
    result = comp["competitors"][0]["results"][0]
    assert result == {"event": "333", "type": "single", "result": "6.42", "badge": "NR"}

    assert payload["competitorMilestones"][0]["name"] == "Nagy Anna"
    assert payload["competitorMilestones"][0]["estimated"] == "Budapest Open 2026"
    assert payload["delegateMilestones"][0] == {"name": "Kiss Péter", "count": 49}


def test_every_competitor_included_even_without_records():
    """A user kérése: MINDEN versenyző jelenjen meg, ne csak a rekordosok. A
    rekord nélküli versenyző neve is bekerül, üres results-szal."""
    content = RenderPostContent({
        "competitions": [
            _competition(competitors=[
                _competitor(name="Rekorder", records={"333": [SimpleNamespace(Type="single", Result=642, Badge="NR")]}),
                _competitor(name="Nincs rekordja", records={}),
            ])
        ],
        "count_competitors": [],
        "delegate_milestones": [],
    })

    competitors = json.loads(content)["competitions"][0]["competitors"]
    names = [c["name"] for c in competitors]

    assert names == ["Rekorder", "Nincs rekordja"]
    assert competitors[0]["results"]  # a rekordosnak van eredménye
    assert competitors[1]["results"] == []  # a rekord nélkülinek üres, de ott a neve


# --- BuildReportData: nincs szűrés, minden versenyző marad --------------------

def test_build_report_data_keeps_all_competitors(monkeypatch):
    """A BuildReportData ne szűrjön rekordra — a rekord nélküli versenyzők is
    maradjanak (a mérföldkő-lekérdezéseket kimockoljuk, hogy ne hívjanak DB-t)."""
    import reportData

    monkeypatch.setattr(reportData, "GetImportantCompetitors", lambda: [])
    monkeypatch.setattr(reportData, "GetImportantDelegates", lambda: [])

    comp = _competition(is_hungarian=True, competitors=[
        _competitor(name="Rekorder", records={"333": [SimpleNamespace(Type="single", Result=642, Badge="NR")]}),
        _competitor(name="Nincs rekordja", records={}),
    ])

    data = BuildReportData([comp])

    assert data["competitions"] == [comp]
    # A versenyzőlista érintetlen — semmit nem szűrtünk ki.
    assert [c.CompetitorName for c in data["competitions"][0].CompetitorWithRecords] == [
        "Rekorder",
        "Nincs rekordja",
    ]
