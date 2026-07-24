"""A heti riport → Post publikálás tesztjei.

Futtatás:
    cd apps/scripts/weekly-report
    pip install -r requirements-dev.txt
    python -m pytest

Az itteni logika DB nélkül tesztelhető: az idempotencia-kulcs számítása és a
sablonrenderelés tiszta függvények (a tényleges INSERT-et a Postgres UNIQUE
megszorítása védi, lásd db/migrations/000003).
"""

from datetime import date, datetime, timedelta
from types import SimpleNamespace

import pytest

from reportData import FilterCompetitions
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


# --- Sablonrenderelés ---------------------------------------------------------

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


def test_render_content_without_results():
    content = RenderPostContent(
        {"competitions": [], "count_competitors": [], "delegate_milestones": []}
    )

    assert "Ezen a héten nem volt magyar érdekeltségű versenyeredmény." in content


def test_render_content_is_plain_text_not_html():
    """A frontend a post.content-et SZÖVEGKÉNT rendereli, ezért markup nem kerülhet bele."""
    content = RenderPostContent({
        "competitions": [
            _competition(competitors=[
                _competitor(records={"333": [SimpleNamespace(Type="single", Result=642, Badge="NR")]})
            ])
        ],
        "count_competitors": [],
        "delegate_milestones": [],
    })

    assert "<" not in content and ">" not in content


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

    assert "Miskolc Open 2026" in content
    assert "Teszt Elek" in content
    assert "6.42" in content  # ConvertResult: 642 -> "6.42"
    assert "NR" in content
    assert "Nagy Anna" in content
    assert "Budapest Open 2026" in content
    assert "Kiss Péter" in content


# --- Versenyszűrés (az e-mail és a poszt is ezt használja) --------------------

def test_filter_keeps_non_hungarian_competitions_as_is():
    without_records = _competitor(records={})
    comp = _competition(is_hungarian=False, competitors=[without_records])

    assert FilterCompetitions([comp]) == [comp]


def test_filter_drops_hungarian_competition_without_record_holders():
    comp = _competition(is_hungarian=True, competitors=[_competitor(records={})])

    assert FilterCompetitions([comp]) == []


def test_filter_narrows_hungarian_competition_to_record_holders():
    with_records = _competitor(
        name="Rekorder", records={"333": [SimpleNamespace(Type="single", Result=642, Badge="NR")]}
    )
    without_records = _competitor(name="Nincs rekordja", records={})
    comp = _competition(is_hungarian=True, competitors=[with_records, without_records])

    result = FilterCompetitions([comp])

    assert result == [comp]
    assert comp.CompetitorWithRecords == [with_records]
